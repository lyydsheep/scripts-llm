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

func SSE(ctx *gin.Context) {
	url := "https://api.deepseek.com/chat/completions"
	method := "POST"

	client := &http.Client{}
	payload, err := getPayload(ctx)
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
	// 异步传输消息
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			// 仅处理包含有效数据的行
			if strings.HasPrefix(line, "data:") {
				eventData := strings.TrimPrefix(line, "data: ")
				if eventData == "[DONE]" { // 流结束标记
					close(ch)
					break
				}
				// 解析 JSON 并提取内容
				var event Event
				if err := json.Unmarshal([]byte(eventData), &event); err == nil {
					if len(event.Choices) > 0 {
						//fmt.Println(event.Choices[0].Delta.Content) // 实时输出内容
						fmt.Println(eventData)
						ch <- eventData
					}
				}
			}
		}
		if err := scanner.Err(); err != nil {
			log.Fatal(nil, "读取流失败: ", "err", err)
		}
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

func getPayload(ctx *gin.Context) (io.Reader, error) {
	type Req struct {
		RoleIdUser      string `json:"role_id_user"`
		RoleIdAssistant string `json:"role_id_assistant"`
		Content         string `json:"content"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		log.Println("解析请求参数失败: ", "err", err)
		return nil, err
	}

	payload := NewPayload()
	// sse
	payload.Stream = true

	// 添加历史对话
	sentences := getHistory(req.RoleIdAssistant, req.RoleIdUser)
	payload.Messages = make([]struct {
		Content string `json:"content"`
		Role    string `json:"role"`
	}, len(sentences))
	for i := range sentences {
		payload.Messages[i].Role, payload.Messages[i].Content = sentences[i].Role, sentences[i].Content
	}
	payload.Messages = append(payload.Messages, sentence{
		Content: req.Content,
		Role:    "user",
	})

	// TODO 设置 prompt
	// 1. 查询人物性格特点
	// 2. 设置 system prompt

	bytes, err := json.Marshal(payload)
	if err != nil {
		log.Println("序列化请求参数失败: ", "err", err)
		return nil, err
	}
	return strings.NewReader(string(bytes)), nil
}

type sentence struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}

func getHistory(assistantId, userId string) []sentence {
	db := internal.DB()
	var entities []internal.Sentence
	result := db.Where("assistant_id = ? OR user_id = ?", assistantId, userId).Order("id ASC").Find(&entities)
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
