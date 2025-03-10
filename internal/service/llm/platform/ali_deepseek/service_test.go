package ali_deepseek

import (
	"context"
	"gitee.com/flycash/ai-gateway-demo/internal/domain"
	"github.com/stretchr/testify/require"
	"log"
	"testing"
)

func TestService(t *testing.T) {
	apiKey := "sk-1ff9e16afa654f50a0a9c759bd59274d"
	service := NewService(apiKey)
	msgChan, err := service.Stream(context.Background(), domain.Request{
		Messages: []domain.Msg{
			{
				Type:    domain.ChatMessageTypeUser,
				Content: "请简述一下苏州天气",
			},
		},
		Model: "deepseek-r1",
	})
	require.NoError(t, err)
	for {
		select {
		case event, ok := <-msgChan:
			if !ok {
				// 通道关闭时退出
				log.Println("通道关闭")
				return
			}
			log.Printf("\ncontent: %s\n reasoning_content: %s", event.Content, event.ReasoningContent)
		}
	}
}
