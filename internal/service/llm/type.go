package llm

import (
	"context"
	"gitee.com/flycash/ai-gateway-demo/internal/domain"
)
//go:generate mockgen -source=./type.go -destination=../../mocks/llm.mock.go -package=llmmocks -typed=true Service
type Service interface {
	Stream(ctx context.Context, req domain.Request)(chan domain.StreamEvent,error)
}


