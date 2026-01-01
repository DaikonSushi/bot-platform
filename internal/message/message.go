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

// ImageSegmentWithURL creates an image segment with URL and optional cache control
func ImageSegmentWithURL(url string, cache bool) Segment {
	data := map[string]interface{}{"file": url}
	if !cache {
		data["cache"] = 0
	}
	return Segment{
		Type: "image",
		Data: data,
	}
}

// AtSegment creates an at message segment
func AtSegment(qq int64) Segment {
	return Segment{
		Type: "at",
		Data: map[string]interface{}{"qq": qq},
	}
}

// AtAllSegment creates an at-all message segment
func AtAllSegment() Segment {
	return Segment{
		Type: "at",
		Data: map[string]interface{}{"qq": "all"},
	}
}

// ReplySegment creates a reply message segment
func ReplySegment(messageID int64) Segment {
	return Segment{
		Type: "reply",
		Data: map[string]interface{}{"id": messageID},
	}
}

// FaceSegment creates a QQ face emoji segment
func FaceSegment(faceID int) Segment {
	return Segment{
		Type: "face",
		Data: map[string]interface{}{"id": faceID},
	}
}

// RecordSegment creates a voice/audio message segment
func RecordSegment(file string) Segment {
	return Segment{
		Type: "record",
		Data: map[string]interface{}{"file": file},
	}
}

// VideoSegment creates a video message segment
func VideoSegment(file string) Segment {
	return Segment{
		Type: "video",
		Data: map[string]interface{}{"file": file},
	}
}

// FileSegment creates a file message segment (for group file upload)
func FileSegment(file string, name string) Segment {
	data := map[string]interface{}{"file": file}
	if name != "" {
		data["name"] = name
	}
	return Segment{
		Type: "file",
		Data: data,
	}
}

// JsonSegment creates a JSON card message segment
func JsonSegment(jsonData string) Segment {
	return Segment{
		Type: "json",
		Data: map[string]interface{}{"data": jsonData},
	}
}

// ForwardSegment creates a forward message segment (for merged forwarding)
func ForwardSegment(id string) Segment {
	return Segment{
		Type: "forward",
		Data: map[string]interface{}{"id": id},
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

// ImageURL adds image from URL to the message
func (m *Message) ImageURL(url string, cache bool) *Message {
	m.Segments = append(m.Segments, ImageSegmentWithURL(url, cache))
	return m
}

// At adds at to the message
func (m *Message) At(qq int64) *Message {
	m.Segments = append(m.Segments, AtSegment(qq))
	return m
}

// AtAll adds at-all to the message
func (m *Message) AtAll() *Message {
	m.Segments = append(m.Segments, AtAllSegment())
	return m
}

// Reply adds reply to the message
func (m *Message) Reply(messageID int64) *Message {
	m.Segments = append(m.Segments, ReplySegment(messageID))
	return m
}

// Face adds QQ face emoji to the message
func (m *Message) Face(faceID int) *Message {
	m.Segments = append(m.Segments, FaceSegment(faceID))
	return m
}

// Record adds voice/audio to the message
func (m *Message) Record(file string) *Message {
	m.Segments = append(m.Segments, RecordSegment(file))
	return m
}

// Video adds video to the message
func (m *Message) Video(file string) *Message {
	m.Segments = append(m.Segments, VideoSegment(file))
	return m
}

// File adds file to the message (for group file upload)
func (m *Message) File(file string, name string) *Message {
	m.Segments = append(m.Segments, FileSegment(file, name))
	return m
}

// Json adds JSON card to the message
func (m *Message) Json(jsonData string) *Message {
	m.Segments = append(m.Segments, JsonSegment(jsonData))
	return m
}

// Forward adds forward message reference
func (m *Message) Forward(id string) *Message {
	m.Segments = append(m.Segments, ForwardSegment(id))
	return m
}

// AddSegment adds a raw segment to the message
func (m *Message) AddSegment(seg Segment) *Message {
	m.Segments = append(m.Segments, seg)
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

// GetSegments returns all message segments
func (e *Event) GetSegments() []Segment {
	return e.Message
}

// GetSegmentsByType returns segments of a specific type
func (e *Event) GetSegmentsByType(segType string) []Segment {
	result := make([]Segment, 0)
	for _, seg := range e.Message {
		if seg.Type == segType {
			result = append(result, seg)
		}
	}
	return result
}

// HasSegmentType checks if the message contains a segment of the given type
func (e *Event) HasSegmentType(segType string) bool {
	for _, seg := range e.Message {
		if seg.Type == segType {
			return true
		}
	}
	return false
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
