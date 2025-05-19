package main

type ChatGPTResponse struct {
	ID                string      `json:"id"`
	Object            string      `json:"object"`
	Created           int64       `json:"created"`
	Model             string      `json:"model"`
	SystemFingerprint string      `json:"system_fingerprint"`
	Choices           []GPTChoice `json:"choices"`
	Usage             Usage       `json:"usage"`
}

type GPTChoice struct {
	Index        int        `json:"index"`
	Message      GptMessage `json:"message"`
	Logprobs     *Logprobs  `json:"logprobs"` // Assuming it could be nil.
	FinishReason string     `json:"finish_reason"`
}

type GptMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Logprobs struct {
	// Define fields here if the structure is known; otherwise, use interface{} or leave undefined.
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
