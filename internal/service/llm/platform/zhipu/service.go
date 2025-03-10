package zhipu

import (
	"gitee.com/flycash/ai-gateway-demo/internal/service/llm/platform/base"
)

type Service struct {
	*base.Service
}

const (
	baseUrl = "https://open.bigmodel.cn/api/paas/v4/"
)

func NewService(apikey string) *Service {
	return &Service{
		base.NewService(apikey, baseUrl),
	}
}
