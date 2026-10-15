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
		t.Fatalf("主要 user lookup 錯誤未保留：%v", err)
	}
	if errors.Is(err, closeErr) {
		t.Fatalf("cleanup 錯誤不應覆蓋主要錯誤：%v", err)
	}
	if got, want := session.closeCount, 1; got != want {
		t.Fatalf("Discord session 關閉次數不正確：got %d, want %d", got, want)
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
