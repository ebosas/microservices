package timeutil

import (
	"testing"
	"time"
)

// TestFormatDuration tests duration formatting
func TestFormatDuration(t *testing.T) {
	now := time.Now().Unix() * 1000
	var tests = []struct {
		input int64
		want  string
	}{
		{now, "less than a minute"},
		{now - 1.4*60*1000, "1 minute"},
		{now - 1.5*60*1000, "2 minutes"},
		{now - 44*60*1000, "44 minutes"},
		{now - 2*60*60*1000, "about 2 hours"},
		{now - 1439*60*1000, "about 24 hours"},
		{now - 1440*60*1000, "1 day"},
		{now - 3*43200*60*1000, "3 months"},
		{now - 24*43200*60*1000, "about 2 years"},
	}
	for _, test := range tests {
		if got := FormatDuration(test.input); got != test.want {
			t.Errorf("FormatDuration(%d) = %q, want %q", test.input, got, test.want)
		}
	}
}
