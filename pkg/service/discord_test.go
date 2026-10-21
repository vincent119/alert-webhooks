package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"alert-webhooks/config"

	"github.com/bwmarrin/discordgo"
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

func TestNewDiscordService(t *testing.T) {
	t.Run("停用時不建立 session", func(t *testing.T) {
		service, err := newDiscordService(config.DiscordConf{}, func(string) (discordSessionClient, error) {
			t.Fatal("session factory must not be called")
			return nil, nil
		})
		if err != nil || service == nil {
			t.Fatalf("service = %#v, error = %v", service, err)
		}
	})

	t.Run("啟用但缺少 token", func(t *testing.T) {
		_, err := newDiscordService(config.DiscordConf{Enable: true}, func(string) (discordSessionClient, error) {
			return nil, nil
		})
		if err == nil {
			t.Fatal("expected missing token error")
		}
	})

	t.Run("session factory 失敗", func(t *testing.T) {
		_, err := newDiscordService(config.DiscordConf{Enable: true, Token: "token"}, func(string) (discordSessionClient, error) {
			return nil, errors.New("factory failed")
		})
		if err == nil || !strings.Contains(err.Error(), "failed to create Discord session") {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("開啟連線失敗", func(t *testing.T) {
		session := &discordSessionForTest{openErr: errors.New("open failed")}
		_, err := newDiscordService(config.DiscordConf{Enable: true, Token: "token"}, func(token string) (discordSessionClient, error) {
			if token != "Bot token" {
				t.Fatalf("token = %q", token)
			}
			return session, nil
		})
		if err == nil || !strings.Contains(err.Error(), "failed to open Discord connection") {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("取得 bot 資訊失敗時關閉 session", func(t *testing.T) {
		session := &discordSessionForTest{userErr: errors.New("user failed")}
		_, err := newDiscordService(config.DiscordConf{Enable: true, Token: "token"}, func(string) (discordSessionClient, error) {
			return session, nil
		})
		if err == nil || session.closeCount != 1 {
			t.Fatalf("error = %v, close count = %d", err, session.closeCount)
		}
	})

	t.Run("成功", func(t *testing.T) {
		session := &discordSessionForTest{user: &discordgo.User{ID: "bot-id", Username: "alert-bot"}}
		cfg := config.DiscordConf{Enable: true, Token: "token", GuildID: "guild", Channels: map[string]string{"chat_ids0": "channel"}}
		service, err := newDiscordService(cfg, func(string) (discordSessionClient, error) { return session, nil })
		if err != nil {
			t.Fatalf("newDiscordService() error = %v", err)
		}
		wantIntents := discordgo.IntentsGuilds | discordgo.IntentsGuildMessages
		if session.intents != wantIntents {
			t.Fatalf("intents = %v, want %v", session.intents, wantIntents)
		}
		if service.guildID != "guild" || service.channels["chat_ids0"] != "channel" {
			t.Fatalf("service = %#v", service)
		}
	})
}

func TestDiscordServiceSendMessage(t *testing.T) {
	t.Run("停用", func(t *testing.T) {
		service := &DiscordService{}
		if err := service.SendMessage(context.Background(), "L0", "message"); err == nil {
			t.Fatal("expected disabled error")
		}
		if err := service.SendMessageToChannel("channel", "message"); err == nil {
			t.Fatal("expected disabled direct send error")
		}
	})

	t.Run("session 未初始化", func(t *testing.T) {
		service := &DiscordService{config: config.DiscordConf{Enable: true}}
		if err := service.SendMessage(context.Background(), "L0", "message"); err == nil {
			t.Fatal("expected session error")
		}
		if err := service.SendMessageToChannel("channel", "message"); err == nil {
			t.Fatal("expected direct send session error")
		}
	})

	t.Run("缺少 channel mapping", func(t *testing.T) {
		service := newDiscordServiceForTest(&discordSessionForTest{})
		if err := service.SendMessage(context.Background(), "L9", "message"); err == nil {
			t.Fatal("expected mapping error")
		}
	})

	t.Run("傳送成功與 SendMessageToLevel", func(t *testing.T) {
		session := &discordSessionForTest{}
		service := newDiscordServiceForTest(session)
		service.channels["chat_ids0"] = "channel-0"
		if err := service.SendMessageToLevel(context.Background(), "L0", "message"); err != nil {
			t.Fatalf("SendMessageToLevel() error = %v", err)
		}
		if len(session.sent) != 1 || session.sent[0].channel != "channel-0" {
			t.Fatalf("sent = %#v", session.sent)
		}
	})

	t.Run("API error 轉換", func(t *testing.T) {
		session := &discordSessionForTest{sendErr: errors.New("Missing Permissions")}
		service := newDiscordServiceForTest(session)
		service.channels["chat_ids0"] = "channel-0"
		err := service.SendMessage(context.Background(), "0", "message")
		if err == nil || !strings.Contains(err.Error(), "lacks necessary permissions") {
			t.Fatalf("error = %v", err)
		}
	})
}

func TestDiscordServiceConnectionAndMetadata(t *testing.T) {
	t.Run("停用與未初始化", func(t *testing.T) {
		for _, service := range []*DiscordService{{}, {config: config.DiscordConf{Enable: true}}} {
			if err := service.TestConnection(); err == nil {
				t.Fatal("expected TestConnection error")
			}
			if _, err := service.GetBotInfo(); err == nil {
				t.Fatal("expected GetBotInfo error")
			}
			if err := service.ValidateChannel("channel"); err == nil {
				t.Fatal("expected ValidateChannel error")
			}
		}
	})

	t.Run("成功", func(t *testing.T) {
		session := &discordSessionForTest{
			user:        &discordgo.User{ID: "bot-id", Username: "alert-bot", Bot: true, Avatar: "avatar"},
			channel:     &discordgo.Channel{ID: "channel", Name: "alerts", Type: discordgo.ChannelTypeGuildText},
			permissions: discordgo.PermissionSendMessages,
			currentUser: "bot-id",
		}
		service := newDiscordServiceForTest(session)
		if err := service.TestConnection(); err != nil {
			t.Fatalf("TestConnection() error = %v", err)
		}
		if err := service.ValidateChannel("channel"); err != nil {
			t.Fatalf("ValidateChannel() error = %v", err)
		}
		info, err := service.GetBotInfo()
		if err != nil {
			t.Fatalf("GetBotInfo() error = %v", err)
		}
		if info.(map[string]interface{})["username"] != "alert-bot" {
			t.Fatalf("bot info = %#v", info)
		}
	})

	t.Run("API failures", func(t *testing.T) {
		tests := []struct {
			name    string
			session *discordSessionForTest
			want    string
		}{
			{name: "user", session: &discordSessionForTest{userErr: errors.New("user failed")}, want: "failed to get bot user info"},
			{name: "channel", session: &discordSessionForTest{channelErr: errors.New("channel failed")}, want: "channel not accessible"},
			{name: "非文字頻道", session: &discordSessionForTest{channel: &discordgo.Channel{Type: discordgo.ChannelTypeGuildVoice}}, want: "is not a text channel"},
			{name: "權限查詢", session: &discordSessionForTest{channel: &discordgo.Channel{Type: discordgo.ChannelTypeGuildText}, permissionErr: errors.New("permission failed")}, want: "failed to check bot permissions"},
			{name: "缺少權限", session: &discordSessionForTest{channel: &discordgo.Channel{Type: discordgo.ChannelTypeGuildText}}, want: "lacks Send Messages permission"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				service := newDiscordServiceForTest(tt.session)
				var err error
				if tt.name == "user" {
					err = service.TestConnection()
				} else {
					err = service.ValidateChannel("channel")
				}
				if err == nil || !strings.Contains(err.Error(), tt.want) {
					t.Fatalf("error = %v, want containing %q", err, tt.want)
				}
			})
		}
	})
}

func TestDiscordServiceLongMessagesAndErrors(t *testing.T) {
	t.Run("依換行分段", func(t *testing.T) {
		session := &discordSessionForTest{}
		service := newDiscordServiceForTest(session)
		message := strings.Repeat("a", 1100) + "\n" + strings.Repeat("b", 1100)
		if err := service.SendMessageToChannel("channel", message); err != nil {
			t.Fatalf("SendMessageToChannel() error = %v", err)
		}
		if len(session.sent) != 2 || !strings.HasPrefix(session.sent[0].message, "(Part 1/2)") || !strings.HasPrefix(session.sent[1].message, "(Part 2/2)") {
			t.Fatalf("sent chunks = %#v", session.sent)
		}
	})

	t.Run("分段傳送失敗", func(t *testing.T) {
		session := &discordSessionForTest{sendErr: errors.New("send failed")}
		service := newDiscordServiceForTest(session)
		err := service.SendMessageToChannel("channel", strings.Repeat("a", 2001))
		if err == nil || !strings.Contains(err.Error(), "failed to send message chunk 1") {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("錯誤映射", func(t *testing.T) {
		service := newDiscordServiceForTest(&discordSessionForTest{})
		tests := []struct {
			apiError string
			want     string
		}{
			{apiError: "Unknown Channel", want: "does not exist"},
			{apiError: "Missing Access", want: "does not have access"},
			{apiError: "Invalid Form Body", want: "content is invalid"},
			{apiError: "Unauthorized", want: "invalid Discord token"},
			{apiError: "other", want: "Discord API error"},
		}
		for _, tt := range tests {
			if err := service.handleDiscordError(errors.New(tt.apiError), "channel"); !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want containing %q", err, tt.want)
			}
		}
	})
}

func TestDiscordServiceChannelLookupAndClose(t *testing.T) {
	service := newDiscordServiceForTest(&discordSessionForTest{})
	service.channels = map[string]string{"chat_ids1": "channel-1", "chat_idserror": ""}
	if channel, err := service.getChannelForLevel("L1"); err != nil || channel != "channel-1" {
		t.Fatalf("channel = %q, error = %v", channel, err)
	}
	if _, err := service.getChannelForLevel("error"); err == nil {
		t.Fatal("expected empty channel mapping error")
	}
	if err := service.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if service.session.(*discordSessionForTest).closeCount != 1 {
		t.Fatal("session was not closed")
	}
	if err := (&DiscordService{}).Close(); err != nil {
		t.Fatalf("nil session Close() error = %v", err)
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

type discordSentMessage struct {
	channel string
	message string
}

type discordSessionForTest struct {
	intents       discordgo.Intent
	openErr       error
	user          *discordgo.User
	userErr       error
	channel       *discordgo.Channel
	channelErr    error
	permissions   int64
	permissionErr error
	sendErr       error
	closeErr      error
	closeCount    int
	currentUser   string
	sent          []discordSentMessage
}

func (s *discordSessionForTest) SetIntents(intents discordgo.Intent) { s.intents = intents }
func (s *discordSessionForTest) Open() error                         { return s.openErr }
func (s *discordSessionForTest) User(string, ...discordgo.RequestOption) (*discordgo.User, error) {
	if s.userErr != nil {
		return nil, s.userErr
	}
	if s.user == nil {
		return &discordgo.User{}, nil
	}
	return s.user, nil
}
func (s *discordSessionForTest) Channel(string, ...discordgo.RequestOption) (*discordgo.Channel, error) {
	if s.channelErr != nil {
		return nil, s.channelErr
	}
	if s.channel == nil {
		return &discordgo.Channel{}, nil
	}
	return s.channel, nil
}
func (s *discordSessionForTest) UserChannelPermissions(string, string, ...discordgo.RequestOption) (int64, error) {
	return s.permissions, s.permissionErr
}
func (s *discordSessionForTest) ChannelMessageSend(channel, message string, _ ...discordgo.RequestOption) (*discordgo.Message, error) {
	s.sent = append(s.sent, discordSentMessage{channel: channel, message: message})
	if s.sendErr != nil {
		return nil, s.sendErr
	}
	return &discordgo.Message{}, nil
}
func (s *discordSessionForTest) CurrentUserID() string { return s.currentUser }
func (s *discordSessionForTest) Close() error {
	s.closeCount++
	return s.closeErr
}

func newDiscordServiceForTest(session discordSessionClient) *DiscordService {
	return &DiscordService{
		session:  session,
		config:   config.DiscordConf{Enable: true},
		channels: map[string]string{},
	}
}
