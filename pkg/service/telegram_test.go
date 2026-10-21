package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"alert-webhooks/config"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func TestNewTelegramService(t *testing.T) {
	t.Run("拒絕空 token", func(t *testing.T) {
		_, err := newTelegramService("", telegramDependenciesForTest(config.Config{}, &telegramBotForTest{}))
		if err == nil || !strings.Contains(err.Error(), "telegram token is required") {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("重試後建立服務並略過無效 chat ID", func(t *testing.T) {
		client := &telegramBotForTest{me: &models.User{ID: 42, Username: "alert_bot", FirstName: "Alert"}}
		cfg := config.Config{Telegram: config.TelegramConf{
			ChatIDs0: "100",
			ChatIDs1: "invalid",
			ChatIDs2: "200",
		}}
		attempts := 0
		var sleeps []time.Duration
		dependencies := telegramDependenciesForTest(cfg, client)
		dependencies.newBot = func(string, ...bot.Option) (telegramBotClient, error) {
			attempts++
			if attempts < 3 {
				return nil, errors.New("temporary failure")
			}
			return client, nil
		}
		dependencies.sleep = func(duration time.Duration) { sleeps = append(sleeps, duration) }

		service, err := newTelegramService("test-token", dependencies)
		if err != nil {
			t.Fatalf("newTelegramService() error = %v", err)
		}
		if attempts != 3 {
			t.Fatalf("attempts = %d, want 3", attempts)
		}
		if len(sleeps) != 2 || sleeps[0] != 2*time.Second || sleeps[1] != 4*time.Second {
			t.Fatalf("sleeps = %v, want [2s 4s]", sleeps)
		}
		if chatID, ok := service.GetChatID(0); !ok || chatID != 100 {
			t.Fatalf("level 0 chat ID = %d, %t", chatID, ok)
		}
		if _, ok := service.GetChatID(1); ok {
			t.Fatal("invalid chat ID must not be loaded")
		}
	})

	t.Run("沒有有效設定時使用預設 chat ID", func(t *testing.T) {
		client := &telegramBotForTest{me: &models.User{ID: 42}}
		service, err := newTelegramService("test-token", telegramDependenciesForTest(config.Config{}, client))
		if err != nil {
			t.Fatalf("newTelegramService() error = %v", err)
		}
		if chatID, ok := service.GetChatID(4); !ok || chatID == 0 {
			t.Fatalf("default level 4 chat ID = %d, %t", chatID, ok)
		}
	})

	t.Run("建立 bot 持續失敗", func(t *testing.T) {
		dependencies := telegramDependenciesForTest(config.Config{}, nil)
		dependencies.newBot = func(string, ...bot.Option) (telegramBotClient, error) {
			return nil, errors.New("permanent failure")
		}
		dependencies.sleep = func(time.Duration) {}

		_, err := newTelegramService("test-token", dependencies)
		if err == nil || !strings.Contains(err.Error(), "after 3 attempts") {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("連線測試失敗", func(t *testing.T) {
		client := &telegramBotForTest{getMeErr: errors.New("get me failed")}
		_, err := newTelegramService("test-token", telegramDependenciesForTest(config.Config{}, client))
		if err == nil || !strings.Contains(err.Error(), "failed to test bot connection") {
			t.Fatalf("error = %v", err)
		}
	})
}

func TestTelegramServiceSendMessage(t *testing.T) {
	t.Run("降級模式", func(t *testing.T) {
		service := &TelegramService{chatIDs: map[int]int64{0: 100}}
		if err := service.SendMessage(context.Background(), 0, "message"); err == nil {
			t.Fatal("expected degraded mode error")
		}
	})

	t.Run("缺少 level mapping", func(t *testing.T) {
		service := &TelegramService{client: &telegramBotForTest{}, chatIDs: map[int]int64{0: 100}}
		if err := service.SendMessage(context.Background(), 1, "message"); err == nil {
			t.Fatal("expected missing chat ID error")
		}
	})

	t.Run("傳送成功", func(t *testing.T) {
		client := &telegramBotForTest{message: &models.Message{ID: 99}}
		service := &TelegramService{client: client, chatIDs: map[int]int64{2: 200}}

		if err := service.SendMessage(context.Background(), 2, "<b>alert</b>"); err != nil {
			t.Fatalf("SendMessage() error = %v", err)
		}
		if client.sent == nil || client.sent.ChatID != int64(200) || client.sent.Text != "<b>alert</b>" {
			t.Fatalf("sent params = %#v", client.sent)
		}
		if client.sent.ParseMode != models.ParseModeHTML {
			t.Fatalf("parse mode = %q", client.sent.ParseMode)
		}
	})

	t.Run("傳送失敗", func(t *testing.T) {
		sendErr := errors.New("send failed")
		client := &telegramBotForTest{sendErr: sendErr}
		service := &TelegramService{client: client, chatIDs: map[int]int64{2: 200}}

		if err := service.SendMessage(context.Background(), 2, "alert"); !errors.Is(err, sendErr) {
			t.Fatalf("error = %v", err)
		}
	})
}

func TestTelegramServiceAccessors(t *testing.T) {
	client := &telegramBotForTest{me: &models.User{ID: 42, Username: "alert_bot"}}
	service := &TelegramService{client: client, chatIDs: map[int]int64{}}

	service.SetChatID(3, 300)
	if chatID, ok := service.GetChatID(3); !ok || chatID != 300 {
		t.Fatalf("chat ID = %d, %t", chatID, ok)
	}
	if service.GetBot() != nil {
		t.Fatal("test service must not expose a concrete bot")
	}
	info, err := service.GetBotInfo()
	if err != nil {
		t.Fatalf("GetBotInfo() error = %v", err)
	}
	if info.(*models.User).Username != "alert_bot" {
		t.Fatalf("bot info = %#v", info)
	}

	short := "short message"
	if got := getMessagePreview(short); got != short {
		t.Fatalf("short preview = %q", got)
	}
	long := strings.Repeat("a", 101)
	if got := getMessagePreview(long); len(got) != 103 || !strings.HasSuffix(got, "...") {
		t.Fatalf("long preview = %q", got)
	}
}

func telegramDependenciesForTest(cfg config.Config, client telegramBotClient) telegramServiceDependencies {
	return telegramServiceDependencies{
		currentConfig: func() config.Config { return cfg },
		newBot: func(string, ...bot.Option) (telegramBotClient, error) {
			return client, nil
		},
		sleep: func(time.Duration) {},
	}
}

type telegramBotForTest struct {
	me       *models.User
	message  *models.Message
	getMeErr error
	sendErr  error
	sent     *bot.SendMessageParams
	context  context.Context
}

func (b *telegramBotForTest) ID() int64 {
	if b.me == nil {
		return 0
	}
	return b.me.ID
}

func (b *telegramBotForTest) GetMe(context.Context) (*models.User, error) {
	if b.getMeErr != nil {
		return nil, b.getMeErr
	}
	if b.me == nil {
		return &models.User{}, nil
	}
	return b.me, nil
}

func (b *telegramBotForTest) SendMessage(ctx context.Context, params *bot.SendMessageParams) (*models.Message, error) {
	b.context = ctx
	b.sent = params
	if b.sendErr != nil {
		return nil, b.sendErr
	}
	if b.message == nil {
		return &models.Message{}, nil
	}
	return b.message, nil
}
