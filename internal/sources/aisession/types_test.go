package aisession

import (
	"testing"
	"time"
)

func TestSessionSummary_DurationMinutes(t *testing.T) {
	tests := []struct {
		name     string
		start    time.Time
		end      time.Time
		expected int
	}{
		{
			name:     "exact minutes",
			start:    time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC),
			end:      time.Date(2026, 1, 1, 10, 47, 0, 0, time.UTC),
			expected: 47,
		},
		{
			name:     "rounds up from 30 seconds",
			start:    time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC),
			end:      time.Date(2026, 1, 1, 10, 5, 30, 0, time.UTC),
			expected: 6,
		},
		{
			name:     "rounds down below 30 seconds",
			start:    time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC),
			end:      time.Date(2026, 1, 1, 10, 5, 20, 0, time.UTC),
			expected: 5,
		},
		{
			name:     "zero duration",
			start:    time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC),
			end:      time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC),
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := SessionSummary{StartTime: tt.start, EndTime: tt.end}
			got := s.DurationMinutes()
			if got != tt.expected {
				t.Errorf("DurationMinutes() = %d, want %d", got, tt.expected)
			}
		})
	}
}
