package main

import (
	"math/rand"
	"time"
)

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
