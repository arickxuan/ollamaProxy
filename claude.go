package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type ClaudeDataItem struct {
	CreatedAt   string `json:"created_at"`
	DisplayName string `json:"display_name"`
	ID          string `json:"id"`
	Type        string `json:"type"`
}

type ClaudeModelResponse struct {
	Data    []ClaudeDataItem `json:"data"`
	FirstID string           `json:"first_id"`
	HasMore bool             `json:"has_more"`
	LastID  string           `json:"last_id"`
}

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

func ClaudeHandlerSteam(c *gin.Context) {
	var input ClaudeRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if XConfig != nil && XConfig.Mock {
		//c.Header("content-Type", "application/x-ndjson")
		c.Header("content-Type", "application/json")
		//c.Header("content-Type", "text/event-stream")
		c.Header("cache-control", "no-cache")
		c.Header("Connection", "keep-alive")
		//c.Writer.WriteHeader(http.StatusOK)
		//jsonStr, err := json.Marshal(MockGPTResponse())
		//if err != nil {
		//	log.Println("Encode error:", err)
		//}
		//_, _ = c.Writer.Write(jsonStr)
		c.JSON(200, MockGPTResponse())
		return
	}

	req, err := http.NewRequest("POST", XConfig.APIURL, bytes.NewBuffer([]byte{}))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("x-api-key", "no-cache")

	// 发起请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Request error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "API request failed"})
		return
	}
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, InitialScannerBufferSize), MaxScannerBufferSize)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		data := scanner.Text()
		if XConfig.Debug {
			println(data)
		}
		if len(data) < 6 {
			continue
		}
		if data[:5] != "data:" && data[:6] != "[DONE]" {
			continue
		}
		data = data[5:]
		data = strings.TrimLeft(data, " ")
		data = strings.TrimSuffix(data, "\r")

		if !strings.HasPrefix(data, "[DONE]") {
			success := dataHandler(data)
			if !success {
				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		if err != io.EOF {
			log.Println("scanner error: " + err.Error())
		}
	}

}
