package main

import (
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type LMStudioModelResp struct {
	Object string              `json:"object"`
	Data   []LMStudioModelData `json:"data"`
}

type LMStudioModelData struct {
	ID                string `json:"id"`                 // 模型的唯一标识符
	Object            string `json:"object"`             // 类型 (如 "model")
	Type              string `json:"type"`               // 模型类型 (如 "vlm")
	Publisher         string `json:"publisher"`          // 发布者
	Arch              string `json:"arch"`               // 架构 (如 "qwen2_vl")
	CompatibilityType string `json:"compatibility_type"` // 兼容性类型
	Quantization      string `json:"quantization"`       // 量化 (如 "4bit")
	State             string `json:"state"`              // 当前加载状态
	MaxContextLength  int    `json:"max_context_length"` // 最大上下文长度
}
type LMStudioChatReq struct {
	Model    string `json:"model"`
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
	Temperature int  `json:"temperature"`
	MaxTokens   int  `json:"max_tokens"`
	Stream      bool `json:"stream"`
}

type LMStudioChatResp struct {
	Choices []struct {
		FinishReason string `json:"finish_reason"`
		Index        int    `json:"index"`
		Logprobs     any    `json:"logprobs"`
		Text         string `json:"text"`
	} `json:"choices"`
	Created int    `json:"created"`
	ID      string `json:"id"`
	Model   string `json:"model"`
	Object  string `json:"object"`
	Usage   struct {
		CompletionTokens int `json:"completion_tokens"`
		PromptTokens     int `json:"prompt_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func getLMModels(c *gin.Context) {
	log.Println("收到 /api/tags 请求 ChatType:", XConfig.ChatType)
	var models []map[string]interface{}
	if XConfig != nil && XConfig.Mock {
		for _, name := range enabledModels {
			family := strings.Split(name, "-")[0]
			model := map[string]interface{}{
				"id":                 name,
				"model":              "model",
				"type":               "llm",
				"publisher":          family,
				"arch":               "llama",
				"compatibility_type": "gguf",
				"quantization":       "Q4_K_M",
				"state":              "not-loaded",
				"max_context_length": 131072,
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
