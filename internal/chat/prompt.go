package chat

import (
	"encoding/json"
	"log"
)

type Payload struct {
	Messages []struct {
		Content string `json:"content"`
		Role    string `json:"role"`
	} `json:"messages"`
	Model            string `json:"model"`
	FrequencyPenalty int    `json:"frequency_penalty"`
	MaxTokens        int    `json:"max_tokens"`
	PresencePenalty  int    `json:"presence_penalty"`
	ResponseFormat   struct {
		Type string `json:"type"`
	} `json:"response_format"`
	Stream      bool `json:"stream"`
	Temperature int  `json:"temperature"`
	TopP        int  `json:"top_p"`
	Logprobs    bool `json:"logprobs"`
}

func NewPayload() Payload {
	var p Payload
	if err := json.Unmarshal([]byte(`{
  "model": "deepseek-chat",
  "frequency_penalty": 0,
  "max_tokens": 2048,
  "presence_penalty": 0,
  "response_format": {
    "type": "text"
  },
  "stop": null,
  "stream": false,
  "stream_options": null,
  "temperature": 1,
  "top_p": 1,
  "tools": null,
  "tool_choice": "none",
  "logprobs": false,
  "top_logprobs": null
}`), &p); err != nil {
		log.Println(err, "prompt.go, NewPayload")
	}
	return p
}

const AnalyzeSystemPrompt = "你是一个专业的剧本角色解析器，专门从剧本文本中提取角色信息。请严格遵循以下规则：\n\n1. **输出格式**：仅输出标准的 JSON 对象，结构为：\n{\n  \"roles\": [\n    {\n      \"name\": \"角色全名\",\n      \"character\": \"性格描述\",\n      \"language_habit\": \"语言习惯描述\"\n    }\n  ]\n}\n\n2. **数据处理原则**：\n- ✅ 从剧本对话/动作描述中提取出现的所有角色\n- ✅ 性格(character)：用 2-5 个形容词概括核心特质（如：谨慎多疑、冲动易怒）\n- ✅ 语言习惯(language_habit)：根据对话特征总结（如：爱用文言文、习惯简短句式）\n- ❌ 不存在的字段用空字符串 \"\" 填充\n- ❌ 不添加注释/额外说明\n\n3. **特殊情况处理**：\n- 同一角色多个称谓 → 合并为一条记录\n- 模糊描述 → 根据上下文合理推断（如\"暴躁老人\" → character:\"暴躁\"）\n- 未直接出现 → 忽略该角色\n\n4. **质量控制**：\n- 确保JSON可直接被 `json.Marshal()` 解析\n- 数组按角色重要性降序排列\n- 字符串值禁用换行符和特殊符号\n\n现在开始解析以下剧本内容："
