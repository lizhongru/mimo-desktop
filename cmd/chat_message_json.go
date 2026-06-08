package cmd

import (
	"encoding/json"
	"time"
)

// chatMessageJSON 用于 JSON 序列化（Go 未导出字段无法被 encoding/json 序列化）
type chatMessageJSON struct {
	Role      string        `json:"role"`
	Content   string        `json:"content"`
	Tool      string        `json:"tool"`
	Args      string        `json:"args"`
	Result    string        `json:"result"`
	Tokens    int           `json:"tokens"`
	ToolCalls int           `json:"tool_calls"`
	Duration  time.Duration `json:"duration"`
	Thinking  string        `json:"thinking"`
	ToolLines []string      `json:"tool_lines"`
}

func (m chatMessage) MarshalJSON() ([]byte, error) {
	return json.Marshal(chatMessageJSON{
		Role: m.role, Content: m.content, Tool: m.tool, Args: m.args,
		Result: m.result, Tokens: m.tokens, ToolCalls: m.toolCalls, Duration: m.duration,
		Thinking: m.thinking, ToolLines: m.toolLines,
	})
}

func (m *chatMessage) UnmarshalJSON(data []byte) error {
	var j chatMessageJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	m.role, m.content, m.tool, m.args, m.result = j.Role, j.Content, j.Tool, j.Args, j.Result
	m.tokens, m.toolCalls, m.duration = j.Tokens, j.ToolCalls, j.Duration
	m.thinking, m.toolLines = j.Thinking, j.ToolLines
	return nil
}
