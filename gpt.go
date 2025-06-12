package main

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	InitialScannerBufferSize = 1 << 20  // 1MB (1*1024*1024)
	MaxScannerBufferSize     = 10 << 20 // 10MB (10*1024*1024)
	DefaultPingInterval      = 10 * time.Second
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type ChatGPTRequest struct {
	Messages []Message `json:"messages"`
	Model    string    `json:"model"`
	Steam    bool      `json:"steam"`
}

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

func MockGPTResponse() *ChatGPTResponse {

	c := GPTChoice{
		Index: 0,
		Message: GptMessage{
			Role:    "assistant",
			Content: "我可以帮助你完成许多任务，比如回答问题、提供建议、解决问题、生成内容（如文章、代码、总结等）、翻译语言、分析数据等等。如果你有具体的需求，比如需要查询信息、计算、写作辅助或工具使用，都可以告诉我！",
		},
		FinishReason: "stop",
	}

	return &ChatGPTResponse{
		ID:                "dfgdfg",
		Object:            "dfgdfg",
		Created:           123123123,
		Model:             "gpt-4.1",
		SystemFingerprint: "dfgdfg",
		Choices:           []GPTChoice{c},
		Usage:             Usage{},
	}
}

func GPTServer() {
	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		log.Println("收到根路径请求")
		c.String(http.StatusOK, "Ollama is running ok")
	})
	router.GET("/api/tags", getModels)
	router.GET("/api/models", getModels)
	router.POST("/api/chat", chatHandlerSteam)
	router.POST("/v1/chat/completions", OpenaiHandlerSteam)

	log.Println("openai proxy server running at :" + strconv.Itoa(XConfig.OpenaiPort))
	router.Run(":" + strconv.Itoa(XConfig.OpenaiPort))
}

func OpenaiHandlerSteam(c *gin.Context) {
	var input ChatGPTRequest
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

func dataHandler(data string) bool {
	return true

}
