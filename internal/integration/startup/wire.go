//go:build wireinject

package startup

import (
	"gitee.com/flycash/ai-gateway-demo/internal/grpc"
	llmmocks "gitee.com/flycash/ai-gateway-demo/internal/mocks"
	"github.com/google/wire"
)

func InitServer(service *llmmocks.MockService) *grpc.LLMServiceServer {
	wire.Build(
		InitService,
		grpc.NewLLMServiceServer,
	)
	return new(grpc.LLMServiceServer)
}
