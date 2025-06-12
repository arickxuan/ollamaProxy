package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type DifyToken struct {
	AccessToken string `json:"access_token"`
}

type DifyChatRequest struct {
	ResponseMode   string                 `json:"response_mode"`
	ConversationID string                 `json:"conversation_id"`
	Query          string                 `json:"query"`
	Inputs         map[string]interface{} `json:"inputs"`
}

type DifyAgentThoughtEvent struct {
	Event          string                 `json:"event"`
	ConversationID string                 `json:"conversation_id"`
	MessageID      string                 `json:"message_id"`
	CreatedAt      int64                  `json:"created_at"`
	TaskID         string                 `json:"task_id"`
	ID             string                 `json:"id"`
	Position       int                    `json:"position,omitempty"`
	Thought        string                 `json:"thought,omitempty"`
	Observation    string                 `json:"observation,omitempty"`
	Tool           string                 `json:"tool,omitempty"`
	Answer         string                 `json:"answer,omitempty"`
	ToolLabels     map[string]interface{} `json:"tool_labels,omitempty"`
	ToolInput      string                 `json:"tool_input,omitempty"`
	MessageFiles   []interface{}          `json:"message_files,omitempty"`
	Metadata       MessageMetadata        `json:"metadata,omitempty"`
}
type MessageMetadata struct {
	Usage UsageInfo `json:"usage"`
}

type UsageInfo struct {
	PromptTokens        int     `json:"prompt_tokens"`
	PromptUnitPrice     string  `json:"prompt_unit_price"`
	PromptPriceUnit     string  `json:"prompt_price_unit"`
	PromptPrice         string  `json:"prompt_price"`
	CompletionTokens    int     `json:"completion_tokens"`
	CompletionUnitPrice string  `json:"completion_unit_price"`
	CompletionPriceUnit string  `json:"completion_price_unit"`
	CompletionPrice     string  `json:"completion_price"`
	TotalTokens         int     `json:"total_tokens"`
	TotalPrice          string  `json:"total_price"`
	Currency            string  `json:"currency"`
	Latency             float64 `json:"latency"`
}

func ToDityRequest(input *OllamaChatRequest) *DifyChatRequest {
	index := len(input.Messages) - 1
	req := DifyChatRequest{
		ResponseMode:   "streaming",
		ConversationID: "",
		Query:          input.Messages[index].Content,
		Inputs:         map[string]interface{}{},
	}
	return &req
}

func DifyToOllamaResponse(input []byte, req *OllamaChatRequest) (*OllamaResponse, error) {
	msg := OllamaResponse{}
	response := DifyAgentThoughtEvent{}
	if len(input) < 6 {
		return nil, nil
	}
	log.Println("Received request:", string(input[6:])) //去除 data:
	err := json.Unmarshal(input[6:], &response)
	if err != nil {
		log.Println("Unmarshal error:", err)
		return nil, nil
	}
	if response.Event == "message_end" {
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
	if response.Event == "agent_thought" || response.Event == "content_block_start" {
		if response.Thought == "" {
			msg.Done = false
			return nil, err
		}
		msg.Done = false
		msg.Message = OllamaMessage{
			Role:    "assistant",
			Content: response.Thought,
		}
		msg.Model = req.Model
		msg.CreatedAt = time.Now().UTC().Format(time.RFC3339Nano)
		return &msg, err
	}
	if response.Answer == "" {
		return nil, err
	}
	msg.Model = req.Model
	msg.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	msg.Message = OllamaMessage{
		Role:    "assistant",
		Content: response.Answer,
	}
	msg.Done = false
	return &msg, nil
}

func getDifyToken(model string) error {
	if XConfig == nil {
		return fmt.Errorf("XConfig is nil")
	}
	client := &http.Client{}
	url := XConfig.DifyTokenUrl
	if XConfig.IsProd {
		url = XConfig.DifyTokenUrlProd
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("创建请求失败: %v", err)
		return err
	}
	if app, ok := XConfig.DifyAppMapProd[model]; ok {
		req.Header.Add("X-App-Code", app)
	}
	if app, ok := XConfig.DifyAppMap[model]; ok {
		req.Header.Add("X-App-Code", app)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("发送请求失败: %v", err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("读取响应失败: %v", err)
		return err
	}
	token := DifyToken{}
	err = json.Unmarshal(body, &token)
	if XConfig.DifyTokenMap == nil {
		XConfig.DifyTokenMap = make(map[string]string)
	}
	XConfig.DifyTokenMap[model] = token.AccessToken
	//log.Println("获取到的token:", token.AccessToken)

	return nil
}
