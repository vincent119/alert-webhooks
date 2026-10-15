package service

import (
	"errors"
	"testing"
)

func TestCloseDiscordSessionAfterUserLookupFailurePreservesPrimaryError(t *testing.T) {
	userErr := errors.New("user lookup failed")
	closeErr := errors.New("session close failed")
	session := &discordSessionCloserForTest{closeErr: closeErr}

	err := closeDiscordSessionAfterUserLookupFailure(session, userErr)
	if !errors.Is(err, userErr) {
		t.Fatalf("primary user lookup error was not preserved: %v", err)
	}
	if errors.Is(err, closeErr) {
		t.Fatalf("cleanup error must not replace the primary error: %v", err)
	}
	if got, want := session.closeCount, 1; got != want {
		t.Fatalf("Discord session close count is incorrect: got %d, want %d", got, want)
	}
}

type discordSessionCloserForTest struct {
	closeErr   error
	closeCount int
}

func (s *discordSessionCloserForTest) Close() error {
	s.closeCount++
	return s.closeErr
}
