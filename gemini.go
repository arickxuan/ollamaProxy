package main

type GeminiResponse struct {
	ID                string         `json:"id"`
	Object            string         `json:"object"`
	Created           int64          `json:"created"` // 时间戳使用 int64 表示
	Model             string         `json:"model"`
	SystemFingerprint *string        `json:"system_fingerprint"` // 使用指针以支持 null 值
	Choices           []GeminiChoice `json:"choices"`
	Usage             *Usage         `json:"usage"` // 使用指针以支持 null 值
}

type GeminiChoice struct {
	Delta        GeminiDelta `json:"delta"`
	Logprobs     *Logprobs   `json:"logprobs"`      // 使用指针以支持 null 值
	FinishReason *string     `json:"finish_reason"` // 使用指针以支持 null 值
	Index        int         `json:"index"`
}

type GeminiDelta struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}
