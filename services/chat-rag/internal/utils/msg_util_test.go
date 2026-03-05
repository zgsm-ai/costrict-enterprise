package utils

import (
	"testing"

	"github.com/zgsm-ai/chat-rag/internal/types"
)

func TestGetContentAsString(t *testing.T) {
	tests := []struct {
		name     string
		content  interface{}
		expected string
	}{
		{
			name:     "simple string content",
			content:  "test content",
			expected: "test content",
		},
		{
			name: "content list with valid text",
			content: []interface{}{
				map[string]interface{}{
					"type": ContentTypeText,
					"text": "part1 ",
				},
				map[string]interface{}{
					"type": ContentTypeText,
					"text": "part2",
				},
			},
			expected: "part1 part2",
		},
		{
			name:     "invalid content type",
			content:  123,
			expected: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetContentAsString(tt.content); got != tt.expected {
				t.Errorf("GetContentAsString() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGetUserMsgs(t *testing.T) {
	tests := []struct {
		name     string
		messages []types.Message
		want     []types.Message
	}{
		{
			name: "filter system messages",
			messages: []types.Message{
				{Role: types.RoleSystem, Content: "system msg"},
				{Role: types.RoleUser, Content: "user msg1"},
				{Role: types.RoleAssistant, Content: "assistant msg"},
				{Role: types.RoleUser, Content: "user msg2"},
			},
			want: []types.Message{
				{Role: types.RoleUser, Content: "user msg1"},
				{Role: types.RoleAssistant, Content: "assistant msg"},
				{Role: types.RoleUser, Content: "user msg2"},
			},
		},
		{
			name:     "empty messages",
			messages: []types.Message{},
			want:     []types.Message{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetUserMsgs(tt.messages); len(got) != len(tt.want) {
				t.Errorf("GetUserMsgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSystemMsg(t *testing.T) {
	tests := []struct {
		name     string
		messages []types.Message
		want     types.Message
	}{
		{
			name: "with system message",
			messages: []types.Message{
				{Role: types.RoleSystem, Content: "system msg"},
				{Role: types.RoleUser, Content: "user msg"},
			},
			want: types.Message{Role: types.RoleSystem, Content: "system msg"},
		},
		{
			name:     "no system message",
			messages: []types.Message{},
			want:     types.Message{Role: types.RoleSystem, Content: ""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetSystemMsg(tt.messages); got.Role != tt.want.Role || got.Content != tt.want.Content {
				t.Errorf("GetSystemMsg() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTruncateContent(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		maxLength int
		want      string
	}{
		{
			name:      "content shorter than max",
			content:   "short",
			maxLength: 10,
			want:      "short",
		},
		{
			name:      "content longer than max",
			content:   "long content that needs truncation",
			maxLength: 10,
			want:      "long conte...",
		},
		{
			name:      "empty content",
			content:   "",
			maxLength: 10,
			want:      "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TruncateContent(tt.content, tt.maxLength); got != tt.want {
				t.Errorf("TruncateContent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetLatestUserMsg(t *testing.T) {
	tests := []struct {
		name     string
		messages []types.Message
		want     string
		wantErr  bool
	}{
		{
			name: "with user messages",
			messages: []types.Message{
				{Role: types.RoleUser, Content: "user msg1"},
				{Role: types.RoleUser, Content: "user msg2"},
			},
			want:    "user msg2",
			wantErr: false,
		},
		{
			name: "no user messages",
			messages: []types.Message{
				{Role: types.RoleSystem, Content: "system msg"},
				{Role: types.RoleAssistant, Content: "assistant msg"},
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetLastUserMsgContent(tt.messages)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLatestUserMsg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetLatestUserMsg() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetOldUserMsgsWithNum(t *testing.T) {
	tests := []struct {
		name     string
		messages []types.Message
		num      int
		want     []types.Message
	}{
		{
			name: "valid case with system and 2 user messages",
			messages: []types.Message{
				{Role: types.RoleSystem, Content: "system msg"},
				{Role: types.RoleUser, Content: "user msg1"},
				{Role: types.RoleAssistant, Content: "assistant msg"},
				{Role: types.RoleUser, Content: "user msg2"},
			},
			num: 1,
			want: []types.Message{
				{Role: types.RoleUser, Content: "user msg1"},
				{Role: types.RoleAssistant, Content: "assistant msg"},
			},
		},
		{
			name: "num larger than user messages count",
			messages: []types.Message{
				{Role: types.RoleSystem, Content: "system msg"},
				{Role: types.RoleUser, Content: "user msg1"},
				{Role: types.RoleUser, Content: "user msg2"},
				{Role: types.RoleUser, Content: "user msg3"},
			},
			num:  3,
			want: []types.Message{},
		},
		{
			name: "num larger than user messages count",
			messages: []types.Message{
				{Role: types.RoleSystem, Content: "system msg"},
				{Role: types.RoleUser, Content: "user msg1"},
				{Role: types.RoleUser, Content: "user msg2"},
				{Role: types.RoleUser, Content: "user msg3"},
			},
			num: 2,
			want: []types.Message{
				{Role: types.RoleUser, Content: "user msg1"},
			},
		},
		{
			name: "num larger than user messages count",
			messages: []types.Message{
				{Role: types.RoleSystem, Content: "system msg"},
				{Role: types.RoleUser, Content: "user msg1"},
				{Role: types.RoleUser, Content: "user msg2"},
				{Role: types.RoleUser, Content: "user msg3"},
			},
			num:  4,
			want: []types.Message{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetOldUserMsgsWithNum(tt.messages, tt.num); len(got) != len(tt.want) {
				t.Errorf("GetOldUserMsgsWithNum() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetRecentUserMsgsWithNum(t *testing.T) {
	tests := []struct {
		name     string
		messages []types.Message
		num      int
		want     []types.Message
	}{
		{
			name: "valid case with 2 user messages",
			messages: []types.Message{
				{Role: types.RoleSystem, Content: "system msg"},
				{Role: types.RoleUser, Content: "user msg1"},
				{Role: types.RoleUser, Content: "user msg2"},
			},
			num: 2,
			want: []types.Message{
				{Role: types.RoleUser, Content: "user msg1"},
				{Role: types.RoleUser, Content: "user msg2"},
			},
		},
		{
			name: "num larger than user messages count",
			messages: []types.Message{
				{Role: types.RoleSystem, Content: "system msg"},
				{Role: types.RoleUser, Content: "user msg1"},
			},
			num:  3,
			want: []types.Message{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetRecentUserMsgsWithNum(tt.messages, tt.num); len(got) != len(tt.want) {
				t.Errorf("GetRecentUserMsgsWithNum() = %v, want %v", got, tt.want)
			}
		})
	}
}
