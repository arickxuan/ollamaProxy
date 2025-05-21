package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type ModelList struct {
	Data   []Model `json:"data"`
	Object string  `json:"object"`
}

type Model struct {
	Created    int64        `json:"created"`
	ID         string       `json:"id"`
	Object     string       `json:"object"`
	OwnedBy    string       `json:"owned_by"`
	Parent     interface{}  `json:"parent"` // Using interface{} since it can be null
	Permission []Permission `json:"permission"`
	Root       string       `json:"root"`
}

type Permission struct {
	AllowCreateEngine  bool        `json:"allow_create_engine"`
	AllowFineTuning    bool        `json:"allow_fine_tuning"`
	AllowLogprobs      bool        `json:"allow_logprobs"`
	AllowSampling      bool        `json:"allow_sampling"`
	AllowSearchIndices bool        `json:"allow_search_indices"`
	AllowView          bool        `json:"allow_view"`
	Created            int64       `json:"created"`
	Group              interface{} `json:"group"` // Using interface{} since it can be null
	ID                 string      `json:"id"`
	IsBlocking         bool        `json:"is_blocking"`
	Object             string      `json:"object"`
	Organization       string      `json:"organization"`
}

func RandString(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

func MockOllamaResponse() *OllamaResponse {
	msg := OllamaResponse{}
	msg.Done = true
	msg.Model = "claude-3-7-sonnet-latest"
	msg.CreatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	msg.Message = OllamaMessage{
		Role:    "assistant",
		Content: "ok ,so easy!!",
	}
	msg.DoneReason = "stop"
	msg.TotalDuration = 13937866250
	msg.LoadDuration = 5978299625
	msg.PromptEvalCount = 9
	msg.PromptEvalDuration = 3912791542
	msg.EvalCount = 12
	msg.EvalDuration = 10937866250

	return &msg
}

func getModelsByUrl() (*ModelList, error) {
	if XConfig == nil {
		return nil, fmt.Errorf("XConfig is nil")
	}
	client := &http.Client{}
	req, err := http.NewRequest("GET", XConfig.ModelsURL, nil)
	if err != nil {
		log.Printf("创建请求失败: %v", err)
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+XConfig.APIKey)
	req.Header.Set("x-api-key", XConfig.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("发送请求失败: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("读取响应失败: %v", err)
		return nil, err
	}
	re := ModelList{}
	err = json.Unmarshal(body, &re)

	return &re, nil
}
