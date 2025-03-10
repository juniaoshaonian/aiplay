package ali_deepseek

import (
	"gitee.com/flycash/ai-gateway-demo/internal/service/llm/platform/base"
)

type Service struct {
	*base.Service
}

const (
	baseUrl = "https://dashscope.aliyuncs.com/compatible-mode/v1/"
)

func NewService(apikey string) *Service {
	return &Service{
		base.NewService(apikey,baseUrl),
	}
}

