package script

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io"
	"log"
	"scripts-llm/internal"
	"scripts-llm/internal/chat"
	"scripts-llm/internal/role"
	"strings"
)

func Upload(ctx *gin.Context) {
	// 1. 定义结构体来解析 title 和 content-type
	type Request struct {
		Title       string `form:"title" binding:"required"`
		ContentType string `form:"content-type"`
	}
	var req Request

	// 2. 绑定请求参数
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(400, gin.H{
			"code":    1,
			"status":  "error",
			"message": "参数绑定失败",
		})
		return
	}

	// 3. 获取上传的文件
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(400, gin.H{
			"code":    1,
			"status":  "error",
			"message": "文件上传失败",
		})
		return
	}

	// 4. 验证 title 是否存在
	if req.Title == "" {
		ctx.JSON(400, gin.H{
			"code":    1,
			"status":  "error",
			"message": "标题不能为空",
		})
		return
	}

	// 5. 处理上传的文件
	// 例如：读取文件内容
	fileBytes, err := file.Open()
	if err != nil {
		ctx.JSON(500, gin.H{
			"code":    1,
			"status":  "error",
			"message": "文件打开失败",
		})
		return
	}
	defer fileBytes.Close()

	content, err := io.ReadAll(fileBytes)
	if err != nil {
		ctx.JSON(500, gin.H{
			"code":    1,
			"status":  "error",
			"message": "文件读取失败",
		})
		return
	}

	// 6. 检查 content-type（如果需要）
	if req.ContentType != "" && req.ContentType != "text/plain" {
		ctx.JSON(400, gin.H{
			"code":    1,
			"status":  "error",
			"message": "不支持的文件类型",
		})
		return
	}

	// 7. 保存 sqlite
	script := &internal.Script{
		Sid:     uuid.New().String(),
		Title:   req.Title,
		Content: string(content),
	}
	internal.DB().Create(script)

	// 7. 返回成功响应
	ctx.JSON(200, gin.H{
		"code":    0,
		"status":  "success",
		"message": "上传成功",
		"data": gin.H{
			"script_id":      script.Sid,
			"script_title":   script.Title,
			"script_content": script.Content,
		},
	})
}

func Analyze(ctx *gin.Context) {
	// 定义请求结构体
	type Request struct {
		ScriptId string `json:"script_id" binding:"required"`
	}
	var req Request

	// 绑定请求参数
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, gin.H{
			"code":    1,
			"status":  "error",
			"message": "参数错误",
		})
		return
	}

	// 验证 script_id 的有效性
	var script internal.Script
	result := internal.DB().Where("sid = ?", req.ScriptId).First(&script)
	if result.Error != nil {
		ctx.JSON(400, gin.H{
			"code":    1,
			"status":  "error",
			"message": "剧本ID不存在",
		})
		return
	}

	// 获取剧本内容
	scriptContent := script.Content

	// 分析剧本（这里假设有一个 AnalyzeScript 函数来实现具体的分析逻辑）
	analysisResult, err := AnalyzeScript(scriptContent)
	if err != nil || len(analysisResult) == 0 {
		ctx.JSON(400, gin.H{
			"code":    1,
			"status":  "error",
			"message": "没有找到角色",
		})
		return
	}

	for i := range analysisResult {
		analysisResult[i].Rid = uuid.New().String()
		analysisResult[i].Sid = req.ScriptId
	}

	// 保存角色到数据库
	for _, role := range analysisResult {
		if err := internal.DB().Create(&role).Error; err != nil {
			ctx.JSON(500, gin.H{
				"code":    1,
				"status":  "error",
				"message": "角色保存失败",
			})
			return
		}
	}

	resp := role.ToVo(analysisResult)

	// 返回分析结果
	ctx.JSON(200, gin.H{
		"code":    0,
		"status":  "success",
		"message": "分析成功",
		"data": gin.H{
			"roles": resp,
		},
	})
}

// AnalyzeScript 是一个示例函数，用于分析剧本内容并返回角色信息
func AnalyzeScript(content string) ([]internal.Role, error) {
	reader, err := getPayload(content)
	if err != nil {
		return nil, err
	}
	res := chat.Chat(reader)
	res = removeFirstAndLastLines(res)

	var resp Response
	if err := json.Unmarshal([]byte(res), &resp); err != nil {
		log.Println(err, "script.go, analyzeScript")
	}
	roles := make([]internal.Role, len(resp.Roles))
	for i := range resp.Roles {
		roles[i] = internal.Role{
			Name:          resp.Roles[i].Name,
			Character:     resp.Roles[i].Character,
			LanguageHabit: resp.Roles[i].LanguageHabit,
		}
	}
	return roles, nil
}

func removeFirstAndLastLines(input string) string {
	lines := strings.Split(input, "\n")
	if len(lines) <= 2 { // 如果只有2行或更少，返回空字符串
		return ""
	}
	// 删除第一行和最后一行
	return strings.Join(lines[1:len(lines)-1], "\n")
}

func getPayload(content string) (io.Reader, error) {
	type sentence struct {
		Content string `json:"content"`
		Role    string `json:"role"`
	}
	payload := chat.NewPayload()
	payload.Messages = append(payload.Messages, sentence{
		Content: chat.AnalyzeSystemPrompt,
		Role:    "system",
	})
	payload.Messages = append(payload.Messages, sentence{
		Content: content,
		Role:    "user",
	})
	bytes, err := json.Marshal(payload)
	if err != nil {
		log.Println(err, " script.go, getPayload")
		return nil, err
	}
	str := string(bytes)
	return strings.NewReader(str), nil
}

type Response struct {
	Roles []struct {
		Name          string `json:"name"`
		Character     string `json:"character"`
		LanguageHabit string `json:"language_habit"`
	} `json:"roles"`
}
