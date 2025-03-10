package base

import (
	"context"
	"encoding/json"
	"gitee.com/flycash/ai-gateway-demo/internal/domain"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/ssestream"
	"time"
)

type Service struct {
	client *openai.Client
}

func NewService(apikey, url string) *Service {
	client := openai.NewClient(
		option.WithBaseURL(url),
		option.WithAPIKey(apikey),
	)
	return &Service{
		client: client,
	}
}
type Delta struct {
	Content          string `json:"content"`
	ReasoningContent string `json:"reasoning_content"`
}

func (a *Service) Stream(ctx context.Context, req domain.Request) (chan domain.StreamEvent, error) {
	eventCh := make(chan domain.StreamEvent, 10)
	params := openai.ChatCompletionNewParams{
		Messages: openai.F(a.buildMsgs(req.Messages)),
		Model:    openai.F(req.Model),
		StreamOptions: openai.F(openai.ChatCompletionStreamOptionsParam{
			IncludeUsage: openai.F(true),
		}),
	}
	go func() {
		newCtx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
		defer cancel()
		stream := a.client.Chat.Completions.NewStreaming(newCtx, params)
		a.recv(eventCh, stream)
	}()
	return eventCh, nil
}

func (a *Service) buildMsgs(msgs []domain.Msg) []openai.ChatCompletionMessageParamUnion {
	ans := make([]openai.ChatCompletionMessageParamUnion, 0, len(msgs))
	for _, msg := range msgs {
		switch msg.Type {
		case domain.ChatMessageTypeUser:
			ans = append(ans, openai.UserMessage(msg.Content))
		case domain.ChatMessageTypeSystem:
			ans = append(ans, openai.SystemMessage(msg.Content))
		}
	}
	return ans
}

func (a *Service) recv(eventCh chan domain.StreamEvent,
	stream *ssestream.Stream[openai.ChatCompletionChunk]) {
	defer close(eventCh)
	acc := openai.ChatCompletionAccumulator{}
	for stream.Next() {
		chunk := stream.Current()
		acc.AddChunk(chunk)
		// 建议在处理完 JustFinished 事件后使用数据块
		if len(chunk.Choices) > 0 {
			// 说明没结束
			if chunk.Choices[0].FinishReason == "" {
				var delta Delta
				err := json.Unmarshal([]byte(chunk.Choices[0].Delta.JSON.RawJSON()), &delta)
				if err != nil {
					eventCh <- domain.StreamEvent{
						Type: domain.ErrorStreamEvent,
						Err:  err,
					}
					return
				}
				eventCh <- domain.StreamEvent{
					Type:             domain.MessageStreamEvent,
					Content:          delta.Content,
					ReasoningContent: delta.ReasoningContent,
				}
			}
		}
	}
	if stream.Err() != nil {
		eventCh <- domain.StreamEvent{
			Err: stream.Err(),
		}
		return
	}
}
