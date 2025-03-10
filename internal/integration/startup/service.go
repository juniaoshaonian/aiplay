package startup

import (
	llmmocks "gitee.com/flycash/ai-gateway-demo/internal/mocks"
	"gitee.com/flycash/ai-gateway-demo/internal/service/llm"
)

func InitService(service *llmmocks.MockService) llm.Service {
	return service
}
