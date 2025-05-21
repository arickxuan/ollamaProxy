package main

import (
	"encoding/json"
	"log"
	"time"
)

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

func ToClaudeRequest(input *OllamaChatRequest) *ClaudeRequest {
	claudeReq := ClaudeRequest{
		Model:     input.Model, //XConfig.Model,
		Messages:  toClaudeRequest(input.Messages),
		Stream:    true,
		MaxTokens: 1024,
	}
	return &claudeReq
}

func ClaudeBlockToOllamaResponse(input []byte, req *OllamaChatRequest) (*OllamaResponse, error) {
	msg := OllamaResponse{}
	claudeResponse := ClaudeBlockResponse{}
	if len(input) < 6 {
		return nil, nil
	}
	log.Println("Received request:", string(input[6:]))
	err := json.Unmarshal(input[6:], &claudeResponse)
	if err != nil {
		return nil, err
	}
	if claudeResponse.Type == "content_block_stop" || claudeResponse.Type == "message_stop" {
		msg.Done = true
		msg.Model = req.Model
		msg.CreatedAt = time.Now().UTC().Format(time.RFC3339Nano)
		msg.Message = OllamaMessage{
			Role:    "assistant",
			Content: "",
		}
		msg.DoneReason = "stop"
		msg.TotalDuration = 13937866250
		msg.LoadDuration = 5978299625
		msg.PromptEvalCount = 9
		msg.PromptEvalDuration = 3912791542
		msg.EvalCount = 12
		msg.EvalDuration = 10937866250
		return &msg, err
	}
	if claudeResponse.Type == "message_start" || claudeResponse.Type == "content_block_start" {
		msg.Done = false
		return nil, err
	}
	if claudeResponse.Delta == nil {
		return nil, err
	}
	msg.Model = req.Model
	msg.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	msg.Message = OllamaMessage{
		Role:    "assistant",
		Content: claudeResponse.Delta.Text,
	}
	msg.Done = false
	return &msg, nil
}
