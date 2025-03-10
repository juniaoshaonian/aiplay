package integration

import (
	"context"
	llmv1 "gitee.com/flycash/ai-gateway-demo/internal/api/proto/gen/llm/v1"
	"gitee.com/flycash/ai-gateway-demo/internal/domain"
	"gitee.com/flycash/ai-gateway-demo/internal/integration/startup"
	llmmocks "gitee.com/flycash/ai-gateway-demo/internal/mocks"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"io"
	"net"
	"testing"
	"time"
)

type TestSuite struct {
	suite.Suite
}

const bufSize = 1024 * 1024

func (t *TestSuite) TestStream() {
	// 正常响应
	testcases := []struct {
		name     string
		req      *llmv1.Request
		before   func(t *testing.T, ctrl *gomock.Controller) *llmmocks.MockService
		wantEvts []*llmv1.StreamEvent
	}{
		{
			name: "正常响应",
			req: &llmv1.Request{
				Messages: []*llmv1.Msg{
					{
						Type:    "user",
						Content: "请描述一下当前的天气",
					},
				},
				Model: "deepseek-r1",
			},
			before: func(t *testing.T, ctrl *gomock.Controller) *llmmocks.MockService {
				mockStreamService := llmmocks.NewMockService(ctrl)
				mockStreamService.EXPECT().Stream(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, req domain.Request) (chan domain.StreamEvent, error) {
					require.Equal(t, "deepseek-r1", req.Model)
					events := make(chan domain.StreamEvent, 4)
					go func() {
						defer close(events)
						events <- domain.StreamEvent{
							Type:             domain.MessageStreamEvent,
							ReasoningContent: "reasoning1"}
						time.Sleep(100 * time.Millisecond)
						events <- domain.StreamEvent{
							Type:             domain.MessageStreamEvent,
							ReasoningContent: "reasoning2"}
						time.Sleep(100 * time.Millisecond)
						events <- domain.StreamEvent{
							Type:    domain.MessageStreamEvent,
							Content: "msg3"}
						events <- domain.StreamEvent{
							Type:    domain.MessageStreamEvent,
							Content: "msg4"}
					}()
					return events, nil
				})
				return mockStreamService
			},
			wantEvts: []*llmv1.StreamEvent{
				{
					Type:             domain.MessageStreamEvent.ToString(),
					ReasoningContent: "reasoning1",
				},
				{
					Type:             domain.MessageStreamEvent.ToString(),
					ReasoningContent: "reasoning2",
				},
				{
					Type:    domain.MessageStreamEvent.ToString(),
					Content: "msg3",
				},
				{
					Type:    domain.MessageStreamEvent.ToString(),
					Content: "msg4",
				},
			},
		},
	}
	for _, tc := range testcases {
		tc := tc
		t.T().Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			svc := tc.before(t, ctrl)

			// 创建内存监听器
			lis := bufconn.Listen(bufSize)
			defer lis.Close()

			// 创建gRPC服务器
			server := grpc.NewServer()
			defer server.Stop()

			// 注册服务实现
			grpcServer := startup.InitServer(svc)
			llmv1.RegisterLLMServiceServer(server, grpcServer)

			// 启动服务器
			go func() {
				if err := server.Serve(lis); err != nil {
					t.Errorf("gRPC server failed: %v", err)
				}
			}()

			// 创建客户端连接
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			conn, err := grpc.DialContext(ctx, "bufnet",
				grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
					return lis.Dial()
				}),
				grpc.WithInsecure(),
			)
			require.NoError(t, err)
			defer conn.Close()

			// 创建客户端
			client := llmv1.NewLLMServiceClient(conn)

			// 调用Stream方法
			stream, err := client.Stream(ctx, tc.req)
			require.NoError(t, err)

			// 接收流事件
			var receivedEvts []*llmv1.StreamEvent
			for {
				evt, err := stream.Recv()
				if err == io.EOF {
					break
				}
				require.NoError(t, err)
				receivedEvts = append(receivedEvts, evt)
			}

			// 断言结果
			assertGrpcEvts(t, tc.wantEvts, receivedEvts)
		})
	}
}


func TestLLMServiceSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func assertGrpcEvts(t *testing.T, wantEvts []*llmv1.StreamEvent, actualEvts []*llmv1.StreamEvent) {
	require.True(t, len(wantEvts) == len(actualEvts))
	for idx := range wantEvts {
		want := wantEvts[idx]
		actual := actualEvts[idx]
		require.True(t, want.Type == actual.Type &&
			want.Content == actual.Content &&
			want.Error == actual.Error &&
			want.ReasoningContent == actual.ReasoningContent)
	}

}
