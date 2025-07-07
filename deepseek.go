package main

type DeepSeekResponse struct {
	ID                string              `json:"id"`
	Object            string              `json:"object"`
	Created           int64               `json:"created"`
	Model             string              `json:"model"`
	Choices           []DeepSeekChoice    `json:"choices"`
	SystemFingerprint string              `json:"system_fingerprint"`
	Usage             DeepSeekChoiceUsage `json:"usage"`
}

type DeepSeekChoice struct {
	Index        int           `json:"index"`
	Delta        DeepSeekDelta `json:"delta"`
	FinishReason string        `json:"finish_reason,omitempty"`
}

type DeepSeekDelta struct {
	Content          *string `json:"content"`
	ReasoningContent string  `json:"reasoning_content"`
	Role             string  `json:"role"`
}

type DeepSeekChoiceUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
