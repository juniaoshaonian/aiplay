package domain

type Request struct {
	Messages []Msg
	Model    string
}

type ChatMsgType string
type EventType string

func (e EventType)ToString()string{
	return string(e)
}

type Msg struct {
	Type    ChatMsgType
	Content string
}

type StreamEvent struct {
	Err              error
	Type             EventType
	Content          string
	// 推理文本
	ReasoningContent string
}

const (
	// ChatMessageTypeUser is a message sent by a human.
	ChatMessageTypeUser ChatMsgType = "user"
	// ChatMessageTypeSystem is a message sent by the system.
	ChatMessageTypeSystem ChatMsgType = "system"

)

const (
	ErrorStreamEvent EventType = "error"
	MessageStreamEvent EventType = "message"
)