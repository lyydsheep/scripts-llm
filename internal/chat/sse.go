package chat

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"os"
	"scripts-llm/internal"
	"strings"
)

const (
	Visitor = "visitor"
)

type Req struct {
	RoleIdUser      string `json:"role_id_user"`
	RoleIdAssistant string `json:"role_id_assistant"`
	Content         string `json:"content"`
}

// 游客模式/角色扮演
func SSE(ctx *gin.Context) {
	var request Req
	if err := ctx.Bind(&request); err != nil {
		log.Println("解析请求参数失败: ", "err", err)
		ctx.JSON(400, gin.H{"error": "参数绑定失败", "status": "error", "message": "参数错误"})
		return
	}
	if request.RoleIdUser == "" {
		request.RoleIdUser = Visitor
	}

	url := "https://api.deepseek.com/chat/completions"
	method := "POST"

	client := &http.Client{}
	payload, err := getPayload(ctx, request)
	if err != nil {
		log.Println(err)
		ctx.JSON(400, gin.H{
			"code":    1,
			"status":  "error",
			"message": err.Error(),
		})
		return
	}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+os.Getenv("KEY"))

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	scanner := bufio.NewScanner(res.Body)
	// 设置 Header
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("Access-Control-Allow-Origin", "*")

	ch := make(chan string)
	ch2 := make(chan string)
	// 异步传输消息
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			// 仅处理包含有效数据的行
			if strings.HasPrefix(line, "data:") {
				eventData := strings.TrimPrefix(line, "data: ")
				if eventData == "[DONE]" { // 流结束标记
					close(ch)
					close(ch2)
					break
				}
				// 解析 JSON 并提取内容
				var event Event
				if err := json.Unmarshal([]byte(eventData), &event); err == nil {
					if len(event.Choices) > 0 {
						//fmt.Println(event.Choices[0].Delta.Content) // 实时输出内容
						fmt.Println(eventData)
						ch <- eventData
						ch2 <- event.Choices[0].Delta.Content
					}
				}
			}
		}
		if err := scanner.Err(); err != nil {
			log.Fatal(nil, "读取流失败: ", "err", err)
		}
	}()
	go func() {
		str := ""
		for msg := range ch2 {
			str += msg
		}
		internal.DB().Model(&internal.Sentence{}).Create(&internal.Sentence{
			RoleIdUser:      request.RoleIdUser,
			RoleIdAssistant: request.RoleIdAssistant,
			Content:         request.Content,
			Role:            "user",
		})
		internal.DB().Model(&internal.Sentence{}).Create(&internal.Sentence{
			RoleIdUser:      request.RoleIdUser,
			RoleIdAssistant: request.RoleIdAssistant,
			Content:         str,
			Role:            "assistant",
		})
	}()
	// 保持连接并发送消息
	ctx.Stream(func(w io.Writer) bool {
		if msg, ok := <-ch; ok {
			// SSE 数据格式：data: <content>\n\n
			ctx.SSEvent("message", msg) // Gin 内置的 SSE 辅助方法
			return true
		}
		return false // 关闭连接
	})
}

// History 获取历史对话
func History(ctx *gin.Context) {
	roleIdAssistant, roleIdUser := ctx.Query("role_id_assistant"), ctx.Query("role_id_user")
	if roleIdUser == "" {
		roleIdUser = Visitor
	}
	sentences := getHistory(roleIdAssistant, roleIdUser)
	ctx.JSON(200, gin.H{
		"code":    0,
		"status":  "success",
		"message": "查询成功",
		"data": gin.H{
			"history":           sentences,
			"role_id_user":      roleIdUser,
			"role_id_assistant": roleIdAssistant,
		},
	})
}

// 设置 systemPrompt 和历史对话信息
func getPayload(ctx *gin.Context, req Req) (io.Reader, error) {
	payload := NewPayload()
	// sse
	payload.Stream = true

	// TODO 设置 prompt
	var (
		assistant internal.Role
		user      internal.Role
	)
	// 查询 llm 扮演的角色
	if err := internal.DB().Model(&internal.Role{}).Where("rid = ?", req.RoleIdAssistant).First(&assistant).Error; err != nil {
		log.Println("查询人物角色失败: ", "err", err)
		return nil, err
	}

	if req.RoleIdUser != Visitor {
		if err := internal.DB().Model(&internal.Role{}).Where("rid = ?", req.RoleIdUser).First(&user).Error; err != nil {
			log.Println("查询人物角色失败: ", "err", err)
			return nil, err
		}
	}

	systemPrompt := setSystemPrompt(assistant, user)

	// 添加历史对话
	sentences := getHistory(req.RoleIdAssistant, req.RoleIdUser)
	payload.Messages = make([]struct {
		Content string `json:"content"`
		Role    string `json:"role"`
	}, len(sentences)+1)
	payload.Messages[0].Role, payload.Messages[0].Content = "system", systemPrompt
	for i := range sentences {
		payload.Messages[i+1].Role, payload.Messages[i+1].Content = sentences[i].Role, sentences[i].Content
	}
	// 添加当前对话
	payload.Messages = append(payload.Messages, sentence{
		Content: req.Content,
		Role:    "user",
	})

	fmt.Println(payload.Messages)

	bytes, err := json.Marshal(payload)
	if err != nil {
		log.Println("序列化请求参数失败: ", "err", err)
		return nil, err
	}
	return strings.NewReader(string(bytes)), nil
}

func setSystemPrompt(assistant, user internal.Role) string {
	var script internal.Script
	internal.DB().Model(&internal.Script{}).Where("sid = ?", assistant.Sid).First(&script)
	return fmt.Sprintf("%s\n\n%s\n\n%s\n\n剧本为：\n%s", ChatSystemPrompt, getRoleStr("你", assistant), getRoleStr("用户", user), script.Content)
}

func getRoleStr(identity string, role internal.Role) string {
	if role.Rid == "" {
		return fmt.Sprintf("%s的身份是游客", identity)
	}
	return fmt.Sprintf("%s扮演的人物是%s，性格特点为%s，说话习惯是 %s", identity, role.Name, role.Character, role.LanguageHabit)
}

type sentence struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}

// 获取历史对话
func getHistory(assistantId, userId string) []sentence {
	db := internal.DB()
	var entities []internal.Sentence
	result := db.Where("role_id_assistant = ? AND role_id_user = ?", assistantId, userId).Order("id ASC").Find(&entities)
	if result.Error != nil {
		log.Println("查询对话记录失败: ", "err", result.Error)
		// 只记录，不返回错误
	}
	sentences := make([]sentence, len(entities))
	for i := range entities {
		sentences[i] = sentence{
			Content: entities[i].Content,
			Role:    entities[i].Role,
		}
	}
	return sentences
}

type Event struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}
