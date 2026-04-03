package config

import (
	"testing"
	"time"
)

func TestParseFlexibleTime(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Time
		wantErr bool
	}{
		{
			name:    "Complete format with seconds",
			input:   "2024-01-01T00:00:00+08:00",
			wantErr: false,
		},
		{
			name:    "Complete format without seconds",
			input:   "2024-01-01T00:00+08:00",
			wantErr: false,
		},
		{
			name:    "Format YYYY-MM-DDTHH:MM",
			input:   "2024-01-01T00:00",
			wantErr: false,
		},
		{
			name:    "Format YYYY-MM-DDTHH",
			input:   "2024-01-01T00",
			wantErr: false,
		},
		{
			name:    "Format YYYY-MM-DD",
			input:   "2024-01-01",
			wantErr: false,
		},
		{
			name:    "Empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "Invalid format",
			input:   "2024/01/01",
			wantErr: true,
		},
		{
			name:    "Format YYYY-MM-DD HH:MM (space separator)",
			input:   "2024-01-01 00:00",
			wantErr: false,
		},
		{
			name:    "Format YYYY-MM-DD HH:MM:SS (space separator with seconds)",
			input:   "2024-01-01 00:00:00",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFlexibleTime(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseFlexibleTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.IsZero() {
				t.Errorf("parseFlexibleTime() returned zero time, expected non-zero")
			}
			// For YYYY-MM-DD format, verify it's auto-completed to 00:00:00 in local timezone
			if tt.input == "2024-01-01" && err == nil {
				expected := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)
				if !got.Equal(expected) {
					t.Errorf("parseFlexibleTime() = %v, want %v", got, expected)
				}
			}
			// For YYYY-MM-DDTHH:MM format, verify it's auto-completed to 00:00 in local timezone
			if tt.input == "2024-01-01T00:00" && err == nil {
				expected := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)
				if !got.Equal(expected) {
					t.Errorf("parseFlexibleTime() = %v, want %v", got, expected)
				}
			}
		})
	}
}

func TestUnmarshalYAMLContent(t *testing.T) {
	yamlContent := `
enabled: true
signingKey: "test-key"
activities:
  - keyword: "代金券"
    creditAmount: 100
    voucherExpiryDays: 30
    startTime: "2024-01-01"
    endTime: "2024-12-31T23:59"
    totalQuota: 10000
    successTemplate: "test"
    alreadyRedeemedMessage: "test"
    expiredMessage: "test"
    quotaExhaustedMessage: "test"
`

	var config VoucherActivityConfig
	err := unmarshalYAMLContent(yamlContent, &config)
	if err != nil {
		t.Fatalf("unmarshalYAMLContent() error = %v", err)
	}

	// Verify configuration
	if !config.Enabled {
		t.Errorf("Enabled = false, want true")
	}
	if config.SigningKey != "test-key" {
		t.Errorf("SigningKey = %v, want test-key", config.SigningKey)
	}
	if len(config.Activities) != 1 {
		t.Fatalf("len(config.Activities) = %d, want 1", len(config.Activities))
	}

	// Verify start time is auto-completed to 00:00:00 in local timezone
	activity := config.Activities[0]
	expectedStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)
	if !activity.StartTime.Equal(expectedStart) {
		t.Errorf("StartTime = %v, want %v", activity.StartTime, expectedStart)
	}

	// Verify end time is auto-completed to 23:59:00 in local timezone
	expectedEnd := time.Date(2024, 12, 31, 23, 59, 0, 0, time.Local)
	if !activity.EndTime.Equal(expectedEnd) {
		t.Errorf("EndTime = %v, want %v", activity.EndTime, expectedEnd)
	}
}
