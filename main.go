package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// API constants
const (
	ClaudeAPIURL = "https://test.wisdgod.com/v1/messages"
	ClaudeModel  = "claude-3-7-sonnet-latest"
)

// 环境变量：CLAUDE_API_KEY
// var claudeAPIKey = os.Getenv("CLAUDE_API_KEY")
var claudeAPIKey = "k-ant-this-is-a-test-for-cursor"

var enabledModels = []string{
	"gpt-4",
	"Claude-37-Sonnet",
	"gemini-2.0-flash",
	"grok-3-beta",
	"DeepSeek-V3",
}

var XConfig *Config

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.json", "path to config file")
	flag.Parse()
	if configPath != "" {
		log.Println("使用配置文件:", configPath)
		loadConfig(configPath)
	}
	if claudeAPIKey == "" {
		log.Fatal("Missing CLAUDE_API_KEY environment variable")
	}

	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		log.Println("收到根路径请求")
		c.String(http.StatusOK, "Ollama is running ok")
	})
	router.GET("/api/tags", getModels)
	router.POST("/api/chat", chatHandlerSteam)
	router.POST("/v1/chat/completions", chatHandlerSteam)

	log.Println("Claude proxy server running at :" + strconv.Itoa(XConfig.Port))
	router.Run(":" + strconv.Itoa(XConfig.Port))
}

func toClaudeRequest(input []OllamaMessage) []ClaudeMessageItem {
	msg := make([]ClaudeMessageItem, 0, len(input))
	for _, m := range input {
		if m.Role == "system" {
			m.Role = "assistant"
		}
		msg = append(msg, ClaudeMessageItem{
			Role:    m.Role,
			Content: []ClaudeMessageContent{{Type: "text", Text: m.Content}},
		})
	}
	return msg
}

func ClaudeBlockToOllamaResponse(input []byte, req OllamaChatRequest) (*OllamaResponse, error) {
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

func getModels(c *gin.Context) {
	log.Println("收到 /api/tags 请求")
	var models []map[string]interface{}
	if XConfig != nil && XConfig.Mock {

		for _, name := range enabledModels {
			family := strings.Split(name, "-")[0]
			model := map[string]interface{}{
				"name":        name,
				"model":       name,
				"modified_at": time.Now().UTC().Format(time.RFC3339),
				"size":        rand.Int63n(1e10),
				"digest":      RandString(12),
				"details": map[string]interface{}{
					"format":             "unknown",
					"family":             family,
					"families":           []string{family},
					"parameter_size":     "unknown",
					"quantization_level": "unknown",
				},
			}
			models = append(models, model)
		}
		c.JSON(http.StatusOK, gin.H{"models": models})
		return
	}

}

func chatHandlerSteam(c *gin.Context) {

	var input OllamaChatRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	if XConfig != nil && XConfig.Mock {
		c.Header("content-Type", "application/x-ndjson")
		// c.Header("content-Type", "text/event-stream")
		c.Header("cache-control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Writer.WriteHeader(http.StatusOK)
		jsonStr, err := json.Marshal(MockOllamaResponse())
		if err != nil {
			log.Println("Encode error:", err)
		}
		_, _ = c.Writer.Write(jsonStr)
		return
	}
	//log.Println("Received request:", input)

	// 构造 Claude API 请求
	claudeReq := ClaudeRequest{
		Model:     XConfig.Model,
		Messages:  toClaudeRequest(input.Messages),
		Stream:    true,
		MaxTokens: 1024,
	}
	payload, err := json.Marshal(claudeReq)
	if err != nil {
		log.Println("Encode error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode request"})
		return
	}

	req, err := http.NewRequest("POST", XConfig.APIURL, bytes.NewBuffer(payload))
	if err != nil {
		log.Println("Request error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}
	req.Header.Set("Authorization", "Bearer "+XConfig.APIKey)
	req.Header.Set("x-api-key", XConfig.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", "2023-06-01")

	// 发起请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Request error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Claude API request failed"})
		return
	}
	defer resp.Body.Close()
	log.Println("Claude API response status:", resp.Status, ClaudeAPIURL, string(payload))

	// 设置为流式响应
	// c.Header("content-Type", "text/event-stream")
	c.Header("content-Type", "application/x-ndjson")
	c.Header("cache-control", "no-cache")
	c.Header("Connection", "keep-alive")

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Bytes()
		fmt.Println("收到一行:", string(line))

		re, err := ClaudeBlockToOllamaResponse(line, input)
		if err != nil {
			//log.Println("Decode error:", err)
			continue
		}
		if re == nil {
			continue
		}
		jsonStr, err := json.Marshal(re)
		if err != nil {
			log.Println("Encode error:", err)
			continue
		}
		_, writeErr := c.Writer.Write(jsonStr)
		if writeErr != nil {
			log.Println("Write error:", writeErr)
			continue

		}
		_, writeErr = c.Writer.Write([]byte("\n"))
		if writeErr != nil {
			log.Println("Write error:", writeErr)
			continue

		}
		c.Writer.Flush()

	}

}
