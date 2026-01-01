package message

import "encoding/json"

// MessageType represents the type of message
type MessageType string

const (
	MessageTypePrivate MessageType = "private"
	MessageTypeGroup   MessageType = "group"
)

// Event represents a OneBot event from NapCat
type Event struct {
	Time        int64       `json:"time"`
	SelfID      int64       `json:"self_id"`
	PostType    string      `json:"post_type"`
	MessageType MessageType `json:"message_type"`
	SubType     string      `json:"sub_type"`
	MessageID   int64       `json:"message_id"`
	UserID      int64       `json:"user_id"`
	GroupID     int64       `json:"group_id,omitempty"`
	RawMessage  string      `json:"raw_message"`
	Message     []Segment   `json:"message"`
	Sender      Sender      `json:"sender"`
}

// Sender represents message sender info
type Sender struct {
	UserID   int64  `json:"user_id"`
	Nickname string `json:"nickname"`
	Card     string `json:"card,omitempty"`
	Role     string `json:"role,omitempty"`
}

// Segment represents a message segment (CQ code in array format)
type Segment struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

// TextSegment creates a text message segment
func TextSegment(text string) Segment {
	return Segment{
		Type: "text",
		Data: map[string]interface{}{"text": text},
	}
}

// ImageSegment creates an image message segment
func ImageSegment(file string) Segment {
	return Segment{
		Type: "image",
		Data: map[string]interface{}{"file": file},
	}
}

// AtSegment creates an at message segment
func AtSegment(qq int64) Segment {
	return Segment{
		Type: "at",
		Data: map[string]interface{}{"qq": qq},
	}
}

// ReplySegment creates a reply message segment
func ReplySegment(messageID int64) Segment {
	return Segment{
		Type: "reply",
		Data: map[string]interface{}{"id": messageID},
	}
}

// Message represents outgoing message
type Message struct {
	Segments []Segment
}

// NewMessage creates a new message
func NewMessage() *Message {
	return &Message{Segments: []Segment{}}
}

// Text adds text to the message
func (m *Message) Text(text string) *Message {
	m.Segments = append(m.Segments, TextSegment(text))
	return m
}

// Image adds image to the message
func (m *Message) Image(file string) *Message {
	m.Segments = append(m.Segments, ImageSegment(file))
	return m
}

// At adds at to the message
func (m *Message) At(qq int64) *Message {
	m.Segments = append(m.Segments, AtSegment(qq))
	return m
}

// Reply adds reply to the message
func (m *Message) Reply(messageID int64) *Message {
	m.Segments = append(m.Segments, ReplySegment(messageID))
	return m
}

// Build returns the segments for sending
func (m *Message) Build() []Segment {
	return m.Segments
}

// GetText extracts text content from event message
func (e *Event) GetText() string {
	return e.RawMessage
}

// IsPrivate checks if the event is a private message
func (e *Event) IsPrivate() bool {
	return e.MessageType == MessageTypePrivate
}

// IsGroup checks if the event is a group message
func (e *Event) IsGroup() bool {
	return e.MessageType == MessageTypeGroup
}

// ParseEvent parses JSON data into Event
func ParseEvent(data []byte) (*Event, error) {
	var event Event
	if err := json.Unmarshal(data, &event); err != nil {
		return nil, err
	}
	return &event, nil
}
