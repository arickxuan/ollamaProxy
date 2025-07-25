package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
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

	if XConfig != nil && XConfig.IsTls {
		go TlsServer()
	}

	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		log.Println("收到根路径请求")
		c.String(http.StatusOK, "Ollama is running ok")
	})

	// LM Studio
	router.GET("/api/v0/models", getLMModels)
	router.POST("/api/v0/chat/completions", chatHandlerSteam)

	// ollama api
	router.GET("/api/tags", getModels)
	router.POST("/api/chat", chatHandlerSteam)
	router.POST("/v1/chat/completions", chatHandlerSteam)

	//open ai
	router.POST("/openai/v1/chat/completions", OpenaiHandler)
	router.GET("/openai/v1/models", GetGptModels)
	router.GET("/openai/models", GetGptModels)
	router.POST("/openai/chat/completions", OpenaiHandler)

	router.POST("/proxy/openai/v1/chat/completions", ProxyChatHandle)
	router.GET("/proxy/openai/v1/models", GetGptModels)

	// imgreduce openai
	router.GET("/imgreduce/openai", func(c *gin.Context) {
		log.Println("收到openai根路径请求")
		c.String(http.StatusOK, "Ollama is running ok")
	})
	router.POST("/imgreduce/openai/v1/chat/completions", OpenaiHandler)
	router.GET("/imgreduce/openai/v1/models", GetGptModels)

	// imgreduce ollama
	router.GET("/imgreduce/ollama", func(c *gin.Context) {
		log.Println("收到ollama根路径请求")
		c.String(http.StatusOK, "Ollama is running ok")
	})
	router.GET("/imgreduce/ollama/api/tags", getModels)
	router.POST("/imgreduce/ollama/api/chat", chatHandlerSteam)

	//imgreduce lm studio
	router.GET("/imgreduce/lmstudio", func(c *gin.Context) {
		log.Println("收到lmstudio根路径请求")
		c.String(http.StatusOK, "Ollama is running ok")
	})
	router.GET("/imgreduce/lmstudio/api/v0/models", getLMModels)
	router.POST("/imgreduce/lmstudio/api/v0/chat/completions", chatHandlerSteam)

	router.GET("/imgreduce/lmstudio/v1/models", GetGptModels)
	router.POST("/imgreduce/lmstudio/v1/chat/completions", OpenaiHandler)

	// claude
	router.POST("/claude/v1/messages", ClaudeHandlerSteam)
	router.GET("/claude/v1/models", getModels)

	router.POST("/upload/oss", Upload)
	router.GET("/upload/oss/list", OssList)

	log.Println("Claude proxy server running at :" + strconv.Itoa(XConfig.Port))
	router.Run(":" + strconv.Itoa(XConfig.Port))
}

func getModels(c *gin.Context) {
	log.Println("收到 /api/tags 请求 ChatType:", XConfig.ChatType)
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
	} else if XConfig != nil && XConfig.ChatType == "dify" {
		for key := range XConfig.DifyAppMap {
			family := strings.Split(key, "-")[0]
			model := map[string]interface{}{
				"name":        key,
				"model":       key,
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
		for key := range XConfig.DifyAppMapProd {
			family := strings.Split(key, "-")[0]
			model := map[string]interface{}{
				"name":        key,
				"model":       key,
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
	} else {
		var models []map[string]interface{}
		list, err := getModelsByUrl()
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"models": models})
		}
		for _, v := range list.Data {
			family := ""
			arr := strings.Split(v.ID, "-")
			if len(arr) > 0 {
				family = arr[0]
			} else {
				arr := strings.Split(v.ID, " ")
				if len(arr) > 0 {
					family = arr[0]
				}
			}

			model := map[string]interface{}{
				"name":        v.ID,
				"model":       v.ID,
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
	}
	c.JSON(http.StatusOK, gin.H{"models": models})
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

	models := make([]string, 0)
	hasModels := make([]string, 0)
	for model := range XConfig.DifyAppMapProd {
		models = append(models, model)
		if model == input.Model {
			XConfig.IsProd = true
			break
		} else {
			hasModels = append(models, model)
		}
	}
	if len(hasModels) == len(models) {
		XConfig.IsProd = false
	}

	url := XConfig.APIURL
	if XConfig.IsProd {
		url = XConfig.APIURLProd
	}

	// 构造 API 请求 Ollama to Dify or Anthropic 依赖 XConfig.IsProd
	payload, err := GenRequest(&input)
	if err != nil {
		log.Println("Encode error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode request"})
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		log.Println("Request error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}
	if XConfig.ChatType == "dify" {
		//log.Println("current DifyToken:", XConfig.DifyToken)
		req.Header.Set("Authorization", "Bearer "+XConfig.DifyTokenMap[input.Model])
	} else {
		req.Header.Set("Authorization", "Bearer "+XConfig.APIKey)
	}
	req.Header.Set("x-api-key", XConfig.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", "2023-06-01")

	// 发起请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Request error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "API request failed"})
		return
	}
	defer resp.Body.Close()
	log.Println("Claude API response status:", resp.Status, XConfig.APIURL, string(payload))

	// 设置为流式响应
	// c.Header("content-Type", "text/event-stream")
	c.Header("content-Type", "application/x-ndjson")
	c.Header("cache-control", "no-cache")
	c.Header("Connection", "keep-alive")

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(string(line)) < 6 {
			continue
		}
		if string(line)[:5] != "data:" && string(line)[:6] != "[DONE]" {
			continue
		}
		//fmt.Println("收到一行:", string(line))
		jsonStr, err := GenResponse(line, &input)
		//log.Println("返回一行:", string(jsonStr))
		if err != nil {
			log.Println("Encode error:", err)
			continue
		}
		if string(jsonStr) == "null" {
			continue
		}
		_, writeErr := c.Writer.Write(jsonStr)
		if writeErr != nil {
			log.Println("Write error:", writeErr)
			continue

		}
		_, writeErr = c.Writer.Write([]byte("\r\n"))
		if writeErr != nil {
			log.Println("Write error:", writeErr)
			continue

		}
		c.Writer.Flush()

	}

}

func GenRequest(input *OllamaChatRequest) ([]byte, error) {
	if XConfig == nil {
		return nil, nil
	}
	switch XConfig.ChatType {
	case "dify":
		if XConfig.DifyTokenMap[input.Model] == "" {
			err := getDifyToken(input.Model)
			if err != nil {
				return nil, err
			}
		}
		re := ToDityRequest(input)
		return json.Marshal(re)
	case "claude":
		re := toClaudeRequest(input.Messages)
		return json.Marshal(re)
	default:
		return nil, nil
	}

}

func GenResponse(input []byte, req *OllamaChatRequest) ([]byte, error) {
	if XConfig == nil {
		return nil, nil
	}
	switch XConfig.ChatType {
	case "dify":
		re, err := DifyToOllamaResponse(input, req)
		if err != nil {
			return nil, nil
		}
		return json.Marshal(re)
	case "claude":
		re, err := ClaudeBlockToOllamaResponse(input, req)
		if err != nil {
			return nil, nil
		}
		return json.Marshal(re)
	default:
		return nil, nil
	}

}
