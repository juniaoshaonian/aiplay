package zhipu

import (
	"context"
	"gitee.com/flycash/ai-gateway-demo/internal/domain"
	"github.com/stretchr/testify/require"
	"log"
	"testing"
)

func TestService(t *testing.T) {
	apiKey := "04600cbf5247d663e4a4efb1592a4459.SgPJAK2idKO7n87I"
	service := NewService(apiKey)
	msgChan, err := service.Stream(context.Background(), domain.Request{
		Messages: []domain.Msg{
			{
				Type:    domain.ChatMessageTypeUser,
				Content: "请简述一下苏州天气",
			},
		},
		Model: "glm-4",
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
