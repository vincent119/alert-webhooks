package slack

import (
	"errors"
	"testing"
)

func TestSlackHandlerLogFieldsDoesNotLogAlertContent(t *testing.T) {
	fields := slackHandlerLogFields("channel", "#alerts", errors.New("Slack request failed"))
	fields = append(fields, slackTemplateLogFields("eng", []string{"eng", "zhtw"})...)
	keys := make(map[string]struct{}, len(fields))
	for _, field := range fields {
		keys[field.Key] = struct{}{}
		switch field.Key {
		case "message", "title", "message_preview", "text":
			t.Fatalf("handler log fields contain alert content key %q", field.Key)
		}
	}
	for _, key := range []string{"category", "channel", "error", "language", "available_languages"} {
		if _, ok := keys[key]; !ok {
			t.Errorf("handler log fields do not retain metadata key %q", key)
		}
	}
}
