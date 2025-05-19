package main

type ClaudeBlockResponse struct {
	Type  string       `json:"type"`
	Index int          `json:"index"`
	Delta *ClaudeDelta `json:"delta"`
}
type ClaudeStartResponse struct {
	Type    string `json:"type"`
	Message struct {
		Id      string        `json:"id"`
		Type    string        `json:"type"`
		Role    string        `json:"role"`
		Content []interface{} `json:"content"`
		Model   string        `json:"model"`
	} `json:"message"`
}

type ClaudeDelta struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Claude 请求体
type ClaudeRequest struct {
	Model       string              `json:"model"`
	Messages    []ClaudeMessageItem `json:"messages"`
	Stream      bool                `json:"stream"`
	MaxTokens   int                 `json:"max_tokens"`
	Temperature float32             `json:"temperature"`
}

type ClaudeMessageItem struct {
	Role    string                 `json:"role"`
	Content []ClaudeMessageContent `json:"content"`
}

type ClaudeMessageContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
