package main

type OllamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OllamaChatRequest struct {
	Model      string          `json:"model"`
	Messages   []OllamaMessage `json:"messages"`
	KeepAlives bool            `json:"keep_alives"`
	Stream     bool            `json:"stream"`
	Options    struct {
		Context []string `json:"context"`
		NumCtx  int      `json:"num_ctx"`
		NumGpu  int      `json:"num_gpu"`
		NumGqa  int      `json:"num_gqa"`
		NumMp   int      `json:"num_mp"`
	} `json:"options"`
}

// Ollama响应结构
type OllamaResponse struct {
	Model              string        `json:"model"`
	CreatedAt          string        `json:"created_at"`
	Message            OllamaMessage `json:"message"`
	Done               bool          `json:"done"`
	DoneReason         string        `json:"done_reason,omitempty"`          // 停止原因
	TotalDuration      int64         `json:"total_duration,omitempty"`       // 使用 int64 表示可能较大的时间值（单位可能为纳秒）
	LoadDuration       int64         `json:"load_duration,omitempty"`        // 同样为 int64 类型
	PromptEvalCount    int           `json:"prompt_eval_count,omitempty"`    // prompt 评估的计数
	PromptEvalDuration int64         `json:"prompt_eval_duration,omitempty"` // prompt 评估的时间
	EvalCount          int           `json:"eval_count,omitempty"`           // 评估的次数
	EvalDuration       int64         `json:"eval_duration,omitempty"`        // 评估的耗时（单位可能为纳秒）
	Error              string        `json:"error,omitempty"`
}
