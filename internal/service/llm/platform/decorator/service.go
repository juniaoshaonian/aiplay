package decorator

import (
	"context"
	"fmt"
	"gitee.com/flycash/ai-gateway-demo/internal/domain"
	"gitee.com/flycash/ai-gateway-demo/internal/service/llm"
)

type Service struct {
	llms map[string]llm.Service
}

func NewService(llms map[string]llm.Service) llm.Service {
	return &Service{
		llms: llms,
	}
}

func (s *Service) Stream(ctx context.Context, req domain.Request) (chan domain.StreamEvent, error) {
	svc, ok := s.llms[req.Model]
	if !ok {
		return nil, fmt.Errorf("未知的大模型")
	}
	return svc.Stream(ctx, req)
}
