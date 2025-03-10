package grpc

import (
	"fmt"
	llmv1 "gitee.com/flycash/ai-gateway-demo/internal/api/proto/gen/llm/v1"
	"gitee.com/flycash/ai-gateway-demo/internal/domain"
	"gitee.com/flycash/ai-gateway-demo/internal/service/llm"
	"github.com/ecodeclub/ekit/slice"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LLMServiceServer struct {
	// 正常我都会组合这个
	llmv1.UnimplementedLLMServiceServer
	svc llm.Service
}

func NewLLMServiceServer(svc llm.Service) *LLMServiceServer {
	return &LLMServiceServer{
		svc: svc,
	}
}

func (l *LLMServiceServer) Stream(request *llmv1.Request, g grpc.ServerStreamingServer[llmv1.StreamEvent]) error {
	// todo 校验token

	if len(request.Messages) == 0 {
		return status.Error(codes.InvalidArgument, "消息不能为空")
	}

	// 调用业务层获取生成器（假设业务层返回一个消息通道和错误通道）
	evtChan, err := l.svc.Stream(g.Context(), domain.Request{
		Messages: slice.Map(request.Messages, func(idx int, src *llmv1.Msg) domain.Msg {
			return domain.Msg{
				Type:    domain.ChatMsgType(src.GetType()),
				Content: src.GetContent(),
			}
		}),
		Model: request.GetModel(),
	})
	if err != nil {
		return err
	}

	// 流式发送响应
	for {
		select {
		case evt, ok := <-evtChan:
			if !ok { // 通道关闭，正常结束
				return nil
			}

			// 构建响应事件
			event := &llmv1.StreamEvent{
				Type:             evt.Type.ToString(),
				Content:          evt.Content,
				ReasoningContent: evt.ReasoningContent,
			}
			// 遇到错误
			if evt.Err != nil {
				_ = g.Send(&llmv1.StreamEvent{
					Type:  evt.Type.ToString(),
					Error: evt.Err.Error(),
				})
				return status.Error(codes.Internal, err.Error())
			}
			// 发送事件
			if err := g.Send(event); err != nil {
				return status.Error(codes.Internal, fmt.Sprintf("发送流数据失败: %v", err))
			}

		case <-g.Context().Done(): // 处理客户端取消
			return status.Error(codes.Canceled, "客户端请求取消")
		}
	}
}

func (l *LLMServiceServer) Register(server grpc.ServiceRegistrar) {
	llmv1.RegisterLLMServiceServer(server, l)
}
