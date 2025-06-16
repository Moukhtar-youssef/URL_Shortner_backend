package handlers_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/Moukhtar-youssef/URL_Shortner.git/internl/handlers"
)

func TestShortner(t *testing.T) {
	baseURL := fmt.Sprintf("%s/", os.Getenv("BASE_URL"))
	alphabet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	alphabetMap := make(map[rune]bool)
	for _, ch := range alphabet {
		alphabetMap[ch] = true
	}
	tests := []struct {
		name        string
		longurl     string
		expectError bool
	}{
		{"Valid URL", "https://www.reddit.com/r/golang/comments/18pkfns/question_regarding_seeding_in_the_go_122", false},
		// making sure it doesn't accept empty long urls
		{"Empty URL", "", true},
		// make sure it isn't already short
		{"Short URL", "https://ex.co", true},
		// making sure it doesn't accept invalid urls
		{"Invalid URL", "http://exa mple.com", true},
		{"Invalid URL", "ht!tp://example.com", true},
		{"Invalid URL", "http://", true},
		{"Invalid URL", "http://?", true},
		{"Invalid URL", ":://invalid", true},
		{"Invalid URL", "http://example.com/[]{}", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := handlers.Shortner(tt.longurl)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for input %q, got none", tt.longurl)
				}
				return
			}
			if !tt.expectError {
				if err != nil {
					t.Errorf("Unexpected err for input %q, the error is: %q", tt.longurl, err)
					return
				}
			}
			code := strings.TrimPrefix(got, baseURL)
			if len(code) != (handlers.NumberOfChrs) {
				t.Errorf("Short code length = %d, want %d", len(code), handlers.NumberOfChrs)
			}

			for _, ch := range code {
				if !alphabetMap[ch] {
					t.Errorf("Short code contains invalid character: %v", string(ch))
				}
			}
		})
	}
}
