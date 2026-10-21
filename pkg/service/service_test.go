package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"alert-webhooks/config"

	"github.com/bwmarrin/discordgo"
)

func TestServiceManagerRetiresBundleAfterLastLease(t *testing.T) {
	closed := make(chan struct{})
	old := testBundle(func() error {
		close(closed)
		return nil
	})
	manager := &ServiceManager{current: old}
	lease := manager.Acquire()
	if lease == nil {
		t.Fatal("expected old service lease to be available")
	}

	manager.swap(testBundle(nil))
	select {
	case <-closed:
		t.Fatal("old service was closed before the lease was released")
	case <-time.After(50 * time.Millisecond):
	}

	lease.Release()
	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("old service was not closed after the lease was released")
	}

	if current := manager.Acquire(); current == nil {
		t.Fatal("new service lease should be available after swap")
	} else {
		current.Release()
	}
	if err := manager.Close(testContext(t)); err != nil {
		t.Fatalf("close service manager failed: %v", err)
	}
}

func TestServiceManagerFailedReloadKeepsCurrentBundle(t *testing.T) {
	old := testBundle(nil)
	manager := &ServiceManager{current: old}
	buildErr := errors.New("候選服務建立失敗")
	if err := manager.reload(func() (*serviceBundle, error) {
		return nil, buildErr
	}); !errors.Is(err, buildErr) {
		t.Fatalf("expected candidate creation error to be preserved, got: %v", err)
	}

	manager.mu.RLock()
	current := manager.current
	manager.mu.RUnlock()
	if current != old {
		t.Fatal("candidate creation failure must not replace the current service")
	}
	if err := manager.Close(testContext(t)); err != nil {
		t.Fatalf("close service manager failed: %v", err)
	}
}

func TestServiceBundleCloseOnlyRunsOnce(t *testing.T) {
	var calls atomic.Int32
	bundle := testBundle(func() error {
		calls.Add(1)
		return nil
	})
	if err := bundle.close(); err != nil {
		t.Fatalf("first close failed: %v", err)
	}
	if err := bundle.close(); err != nil {
		t.Fatalf("second close failed: %v", err)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("close function ran %d times, want 1 times", got)
	}
}

func TestServiceManagerConcurrentAcquireAndSwap(t *testing.T) {
	manager := &ServiceManager{current: testBundle(nil)}
	var workers sync.WaitGroup
	stop := make(chan struct{})
	for range 8 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				lease := manager.Acquire()
				if lease != nil {
					lease.Release()
				}
			}
		}()
	}

	for range 100 {
		manager.swap(testBundle(nil))
	}
	close(stop)
	workers.Wait()
	if err := manager.Close(testContext(t)); err != nil {
		t.Fatalf("close service manager after concurrent swap failed: %v", err)
	}
}

func TestStaticAccessorAndLeaseAccessors(t *testing.T) {
	discord := &DiscordService{}
	accessor := NewStaticAccessor(discord)
	lease := accessor.Acquire()
	if lease == nil {
		t.Fatal("expected lease")
	}
	if lease.Telegram() != nil || lease.Slack() != nil || lease.Discord() != discord {
		t.Fatalf("lease services = telegram:%#v slack:%#v discord:%#v", lease.Telegram(), lease.Slack(), lease.Discord())
	}
	lease.Release()
	lease.Release()
	var nilLease *Lease
	nilLease.Release()

	static := accessor.(*staticAccessor)
	static.bundle.retire()
	static.bundle.retire()
	if got := accessor.Acquire(); got != nil {
		got.Release()
		t.Fatal("retired static accessor must not issue a lease")
	}
}

func TestServiceManagerGettersAndReadiness(t *testing.T) {
	telegram := &TelegramService{}
	slack := &SlackService{}
	discord := &DiscordService{}
	engine := newTemplateEngine()
	manager := &ServiceManager{
		current:        &serviceBundle{telegram: telegram, slack: slack, discord: discord, drained: make(chan struct{})},
		templateEngine: engine,
	}

	if manager.GetTelegramService() != telegram || manager.GetSlackService() != slack || manager.GetDiscordService() != discord {
		t.Fatal("service getter returned an unexpected instance")
	}
	if !manager.IsTelegramServiceReady() || !manager.IsSlackServiceReady() || !manager.IsDiscordServiceReady() {
		t.Fatal("configured services must be ready")
	}
	if manager.GetTemplateEngine() != engine {
		t.Fatal("template engine getter returned an unexpected instance")
	}

	empty := &ServiceManager{}
	if empty.GetTelegramService() != nil || empty.GetSlackService() != nil || empty.GetDiscordService() != nil {
		t.Fatal("empty manager returned a service")
	}
	if empty.IsTelegramServiceReady() || empty.IsSlackServiceReady() || empty.IsDiscordServiceReady() {
		t.Fatal("empty manager reported a ready service")
	}

	if err := manager.Close(testContext(t)); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}

func TestServiceManagerReloadSuccessAndCloseTimeout(t *testing.T) {
	manager := &ServiceManager{current: testBundle(nil)}
	next := testBundle(nil)
	if err := manager.reload(func() (*serviceBundle, error) { return next, nil }); err != nil {
		t.Fatalf("reload() error = %v", err)
	}
	lease := manager.Acquire()
	if lease == nil {
		t.Fatal("expected lease")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := manager.Close(ctx)
	if err == nil || !strings.Contains(err.Error(), "等待服務組合關閉逾時") {
		t.Fatalf("Close() error = %v", err)
	}
	lease.Release()
	if err := manager.Close(testContext(t)); err != nil {
		t.Fatalf("second Close() error = %v", err)
	}
}

func TestServiceBundleCloseDiscordAndError(t *testing.T) {
	closeErr := errors.New("close failed")
	session := &discordSessionForTest{closeErr: closeErr}
	bundle := &serviceBundle{
		discord: &DiscordService{session: session},
		drained: make(chan struct{}),
	}
	if err := bundle.close(); !errors.Is(err, closeErr) {
		t.Fatalf("close error = %v", err)
	}
}

func TestBuildServiceBundleWithFactories(t *testing.T) {
	cfg := config.Config{
		Telegram: config.TelegramConf{Enable: true, Token: "telegram"},
		Slack:    config.SlackConf{Enable: true, Token: "slack"},
		Discord:  config.DiscordConf{Enable: true, Token: "discord"},
	}
	telegram := &TelegramService{}
	slack := &SlackService{}
	discord := &DiscordService{}
	factories := serviceFactorySet{
		newTelegram: func(token string) (*TelegramService, error) {
			if token != "telegram" {
				t.Fatalf("Telegram token = %q", token)
			}
			return telegram, nil
		},
		newSlack: func(token string) (*SlackService, error) {
			if token != "slack" {
				t.Fatalf("Slack token = %q", token)
			}
			return slack, nil
		},
		newDiscord: func(value config.DiscordConf) (*DiscordService, error) {
			if value.Token != "discord" {
				t.Fatalf("Discord token = %q", value.Token)
			}
			return discord, nil
		},
	}
	bundle, err := buildServiceBundleWithFactories(cfg, true, factories)
	if err != nil {
		t.Fatalf("buildServiceBundleWithFactories() error = %v", err)
	}
	if bundle.telegram != telegram || bundle.slack != slack || bundle.discord != discord {
		t.Fatalf("bundle = %#v", bundle)
	}

	tests := []struct {
		name      string
		configure func(*serviceFactorySet)
		want      string
	}{
		{name: "Telegram", configure: func(value *serviceFactorySet) {
			value.newTelegram = func(string) (*TelegramService, error) { return nil, errors.New("telegram failed") }
		}, want: "建立 Telegram 服務失敗"},
		{name: "Slack", configure: func(value *serviceFactorySet) {
			value.newSlack = func(string) (*SlackService, error) { return nil, errors.New("slack failed") }
		}, want: "建立 Slack 服務失敗"},
		{name: "Discord", configure: func(value *serviceFactorySet) {
			value.newDiscord = func(config.DiscordConf) (*DiscordService, error) { return nil, errors.New("discord failed") }
		}, want: "建立 Discord 服務失敗"},
	}
	for _, tt := range tests {
		t.Run(tt.name+" strict failure", func(t *testing.T) {
			failing := factories
			tt.configure(&failing)
			_, err := buildServiceBundleWithFactories(cfg, true, failing)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want containing %q", err, tt.want)
			}
		})
		t.Run(tt.name+" non-strict degradation", func(t *testing.T) {
			failing := factories
			tt.configure(&failing)
			bundle, err := buildServiceBundleWithFactories(cfg, false, failing)
			if err != nil || bundle == nil {
				t.Fatalf("bundle = %#v, error = %v", bundle, err)
			}
		})
	}

	disabled, err := buildServiceBundle(config.Config{}, true)
	if err != nil || disabled == nil {
		t.Fatalf("disabled bundle = %#v, error = %v", disabled, err)
	}
}

func TestServiceManagerDefaultLifecycle(t *testing.T) {
	originalTelegram := config.Telegram
	originalSlack := config.Slack
	originalDiscord := config.Conf.Discord
	config.Telegram = config.TelegramConf{}
	config.Slack = config.SlackConf{}
	config.Conf.Discord = config.DiscordConf{}
	t.Cleanup(func() {
		config.Telegram = originalTelegram
		config.Slack = originalSlack
		config.Conf.Discord = originalDiscord
	})

	manager := &ServiceManager{}
	if err := manager.InitServices(); err != nil {
		t.Fatalf("InitServices() error = %v", err)
	}
	if err := manager.Reload(context.Background()); err != nil {
		t.Fatalf("Reload() error = %v", err)
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if err := manager.Reload(cancelled); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled Reload() error = %v", err)
	}
	manager.ReloadTemplateEngine()
	if manager.GetTemplateEngine() == nil {
		t.Fatal("ReloadTemplateEngine() did not install an engine")
	}
	if err := manager.Close(testContext(t)); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}

func TestGetServiceManagerReturnsSingleton(t *testing.T) {
	first := GetServiceManager()
	second := GetServiceManager()
	if first != second {
		t.Fatal("GetServiceManager() returned different instances")
	}
}

func TestNewTemplateEngineLoadsFromRepositoryRoot(t *testing.T) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	repositoryRoot := filepath.Clean(filepath.Join(workingDirectory, "../.."))
	if err := os.Chdir(repositoryRoot); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(workingDirectory); err != nil {
			t.Errorf("restore working directory error = %v", err)
		}
	})
	if engine := newTemplateEngine(); engine == nil {
		t.Fatal("newTemplateEngine() returned nil")
	}
}

func TestPublicConstructorsRejectSafeInvalidInput(t *testing.T) {
	if _, err := NewTelegramService(""); err == nil {
		t.Fatal("NewTelegramService must reject empty token")
	}
	if _, err := NewSlackService(""); err == nil {
		t.Fatal("NewSlackService must reject empty token")
	}
	if service, err := NewDiscordService(config.DiscordConf{}); err != nil || service == nil {
		t.Fatalf("NewDiscordService disabled result = %#v, error = %v", service, err)
	}
}

func TestDiscordSessionAdapterLocalState(t *testing.T) {
	session, err := discordgo.New("Bot token")
	if err != nil {
		t.Fatalf("discordgo.New() error = %v", err)
	}
	adapter := &discordSessionAdapter{Session: session}
	adapter.SetIntents(discordgo.IntentsGuilds)
	if session.Identify.Intents != discordgo.IntentsGuilds {
		t.Fatalf("intents = %v", session.Identify.Intents)
	}
	if got := adapter.CurrentUserID(); got != "" {
		t.Fatalf("empty current user ID = %q", got)
	}
	session.State.User = &discordgo.User{ID: "bot-id"}
	if got := adapter.CurrentUserID(); got != "bot-id" {
		t.Fatalf("current user ID = %q", got)
	}
}

func testBundle(closeFn func() error) *serviceBundle {
	return &serviceBundle{
		drained: make(chan struct{}),
		closeFn: closeFn,
	}
}

func testContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	t.Cleanup(cancel)
	return ctx
}
