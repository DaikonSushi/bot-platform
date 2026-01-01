// Package botservice implements the BotService gRPC server
// This service is called by external plugins to send messages and get info
package botservice

import (
	"context"
	"log"

	pb "github.com/DaikonSushi/bot-platform/api/proto"
	"github.com/DaikonSushi/bot-platform/internal/message"
)

// MessageSender is the interface for sending messages
type MessageSender interface {
	SendPrivateText(userID int64, text string) error
	SendGroupText(groupID int64, text string) error
	// SendPrivateMessage sends a message with segments to a user
	SendPrivateMessage(userID int64, msg *message.Message) error
	// SendGroupMessage sends a message with segments to a group
	SendGroupMessage(groupID int64, msg *message.Message) error
}

// Service implements pb.BotServiceServer
type Service struct {
	pb.UnimplementedBotServiceServer
	sender MessageSender
}

// NewService creates a new BotService
func NewService(sender MessageSender) *Service {
	return &Service{
		sender: sender,
	}
}

// SendMessage sends a message
func (s *Service) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	log.Printf("[BotService] SendMessage: type=%s, userId=%d, groupId=%d, segments=%d",
		req.MessageType, req.UserId, req.GroupId, len(req.Segments))

	// Convert protobuf segments to internal message
	msg := message.NewMessage()
	for _, seg := range req.Segments {
		switch seg.Type {
		case "text":
			if text, ok := seg.Data["text"]; ok {
				msg.Text(text)
			}
		case "image":
			if file, ok := seg.Data["file"]; ok {
				msg.Image(file)
			}
		case "at":
			if qq, ok := seg.Data["qq"]; ok {
				// Parse qq as int64
				var qqID int64
				if _, err := parseIntFromString(qq, &qqID); err == nil {
					msg.At(qqID)
				}
			}
		case "reply":
			if id, ok := seg.Data["id"]; ok {
				var msgID int64
				if _, err := parseIntFromString(id, &msgID); err == nil {
					msg.Reply(msgID)
				}
			}
		case "face":
			// Face segment - add to message segments directly
			if faceID, ok := seg.Data["id"]; ok {
				msg.Segments = append(msg.Segments, message.Segment{
					Type: "face",
					Data: map[string]interface{}{"id": faceID},
				})
			}
		case "record":
			// Voice/audio segment
			if file, ok := seg.Data["file"]; ok {
				msg.Segments = append(msg.Segments, message.Segment{
					Type: "record",
					Data: map[string]interface{}{"file": file},
				})
			}
		case "video":
			// Video segment
			if file, ok := seg.Data["file"]; ok {
				msg.Segments = append(msg.Segments, message.Segment{
					Type: "video",
					Data: map[string]interface{}{"file": file},
				})
			}
		case "file":
			// File segment
			data := make(map[string]interface{})
			for k, v := range seg.Data {
				data[k] = v
			}
			msg.Segments = append(msg.Segments, message.Segment{
				Type: "file",
				Data: data,
			})
		default:
			// For unknown types, pass through as-is
			data := make(map[string]interface{})
			for k, v := range seg.Data {
				data[k] = v
			}
			msg.Segments = append(msg.Segments, message.Segment{
				Type: seg.Type,
				Data: data,
			})
		}
	}

	var err error
	if req.MessageType == "private" {
		err = s.sender.SendPrivateMessage(req.UserId, msg)
	} else {
		err = s.sender.SendGroupMessage(req.GroupId, msg)
	}

	if err != nil {
		return &pb.SendMessageResponse{
			Error: err.Error(),
		}, nil
	}

	return &pb.SendMessageResponse{
		MessageId: 0, // We don't track message IDs for now
	}, nil
}

// parseIntFromString parses an int64 from string
func parseIntFromString(s string, result *int64) (bool, error) {
	var n int64
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int64(c-'0')
		} else {
			return false, nil
		}
	}
	*result = n
	return true, nil
}

// GetUserInfo gets user information
func (s *Service) GetUserInfo(ctx context.Context, req *pb.GetUserInfoRequest) (*pb.UserInfo, error) {
	log.Printf("[BotService] GetUserInfo: userId=%d", req.UserId)
	// Return placeholder for now
	return &pb.UserInfo{
		UserId:   req.UserId,
		Nickname: "Unknown",
	}, nil
}

// GetGroupInfo gets group information
func (s *Service) GetGroupInfo(ctx context.Context, req *pb.GetGroupInfoRequest) (*pb.GroupInfo, error) {
	log.Printf("[BotService] GetGroupInfo: groupId=%d", req.GroupId)
	// Return placeholder for now
	return &pb.GroupInfo{
		GroupId:   req.GroupId,
		GroupName: "Unknown",
	}, nil
}

// Log handles log requests from plugins
func (s *Service) Log(ctx context.Context, req *pb.LogRequest) (*pb.Empty, error) {
	log.Printf("[Plugin:Log:%s] %s", req.Level, req.Message)
	return &pb.Empty{}, nil
}
