package service

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"alert-webhooks/config"
	"alert-webhooks/pkg/logger"
	"alert-webhooks/pkg/notification"
	"alert-webhooks/pkg/template"
)

// Accessor 是路由層取得目前服務組合所需的最小介面。
type Accessor interface{ Acquire() *Lease }

// NewStaticAccessor 提供相容測試與不支援熱重載的呼叫端使用。
func NewStaticAccessor(discord *DiscordService) Accessor {
	return &staticAccessor{bundle: &serviceBundle{discord: discord, drained: make(chan struct{})}}
}

type staticAccessor struct{ bundle *serviceBundle }

func (a *staticAccessor) Acquire() *Lease {
	if !a.bundle.acquire() {
		return nil
	}
	return &Lease{bundle: a.bundle}
}

// Lease 保護 request 使用的服務組合，釋放前不得關閉其資源。
type Lease struct {
	bundle *serviceBundle
	once   sync.Once
}

// Telegram 取得 lease 期間可安全使用的 Telegram 服務。
func (l *Lease) Telegram() *TelegramService { return l.bundle.telegram }

// Slack 取得 lease 期間可安全使用的 Slack 服務。
func (l *Lease) Slack() *SlackService { return l.bundle.slack }

// Discord 取得 lease 期間可安全使用的 Discord 服務。
func (l *Lease) Discord() *DiscordService { return l.bundle.discord }

// Release 釋放 lease，使已退休服務在最後一個請求結束後可關閉。
func (l *Lease) Release() {
	if l != nil {
		l.once.Do(l.bundle.release)
	}
}

type serviceBundle struct {
	telegram  *TelegramService
	slack     *SlackService
	discord   *DiscordService
	mu        sync.Mutex
	refs      int
	retired   bool
	drained   chan struct{}
	closeOnce sync.Once
	closeFn   func() error
}

func (b *serviceBundle) acquire() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.retired {
		return false
	}
	b.refs++
	return true
}
func (b *serviceBundle) release() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.refs--
	if b.retired && b.refs == 0 {
		close(b.drained)
	}
}
func (b *serviceBundle) retire() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.retired {
		return
	}
	b.retired = true
	if b.refs == 0 {
		close(b.drained)
	}
}
func (b *serviceBundle) close() error {
	var err error
	b.closeOnce.Do(func() {
		if b.closeFn != nil {
			err = b.closeFn()
			return
		}
		if b.discord != nil {
			err = b.discord.Close()
		}
	})
	return err
}

// ServiceManager 管理可交換的服務組合。
type ServiceManager struct {
	mu             sync.RWMutex
	current        *serviceBundle
	templateEngine *template.TemplateEngine
	retired        sync.WaitGroup
}

var (
	instance *ServiceManager
	once     sync.Once
)

// GetServiceManager 回傳管理目前服務組合的 singleton。
func GetServiceManager() *ServiceManager {
	once.Do(func() { instance = &ServiceManager{} })
	return instance
}

// InitServices 建立初始服務組合，維持啟動時的降級行為。
func (sm *ServiceManager) InitServices() error {
	sm.mu.Lock()
	if sm.templateEngine == nil {
		sm.templateEngine = newTemplateEngine()
	}
	sm.mu.Unlock()
	b, err := buildServiceBundle(config.Current(), false)
	if err != nil {
		return err
	}
	sm.initializeNotificationManager(b)
	sm.swap(b)
	return nil
}

// Reload 只在候選服務全部可用時交換目前組合。
func (sm *ServiceManager) Reload(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("重載服務設定已取消: %w", err)
	}
	b, err := buildServiceBundle(config.Current(), true)
	if err != nil {
		return err
	}
	sm.initializeNotificationManager(b)
	sm.swap(b)
	return nil
}

func (sm *ServiceManager) reload(build func() (*serviceBundle, error)) error {
	b, err := build()
	if err != nil {
		return err
	}
	sm.swap(b)
	return nil
}
func (sm *ServiceManager) swap(next *serviceBundle) {
	sm.mu.Lock()
	old := sm.current
	sm.current = next
	sm.mu.Unlock()
	if old != nil {
		sm.retire(old)
	}
}
func (sm *ServiceManager) retire(b *serviceBundle) {
	b.retire()
	sm.retired.Add(1)
	go func() {
		defer sm.retired.Done()
		<-b.drained
		if err := b.close(); err != nil {
			logger.Error("Failed to close retired service bundle", "service_manager", logger.Err(err))
		}
	}()
}

// Acquire 取得目前服務組合的 lease；沒有可用服務時回傳 nil。
func (sm *ServiceManager) Acquire() *Lease {
	sm.mu.RLock()
	b := sm.current
	if b == nil || !b.acquire() {
		sm.mu.RUnlock()
		return nil
	}
	sm.mu.RUnlock()
	return &Lease{bundle: b}
}

// Close 停止新 lease 並等待現有 request 釋放。
func (sm *ServiceManager) Close(ctx context.Context) error {
	sm.mu.Lock()
	b := sm.current
	sm.current = nil
	sm.mu.Unlock()
	if b != nil {
		sm.retire(b)
	}
	done := make(chan struct{})
	go func() { sm.retired.Wait(); close(done) }()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("等待服務組合關閉逾時: %w", ctx.Err())
	}
}

// GetTelegramService 回傳目前 Telegram 服務，僅適用於不需跨呼叫保留 lease 的相容呼叫端。
func (sm *ServiceManager) GetTelegramService() *TelegramService {
	l := sm.Acquire()
	if l == nil {
		return nil
	}
	defer l.Release()
	return l.Telegram()
}

// GetSlackService 回傳目前 Slack 服務，僅適用於不需跨呼叫保留 lease 的相容呼叫端。
func (sm *ServiceManager) GetSlackService() *SlackService {
	l := sm.Acquire()
	if l == nil {
		return nil
	}
	defer l.Release()
	return l.Slack()
}

// GetDiscordService 回傳目前 Discord 服務，僅適用於不需跨呼叫保留 lease 的相容呼叫端。
func (sm *ServiceManager) GetDiscordService() *DiscordService {
	l := sm.Acquire()
	if l == nil {
		return nil
	}
	defer l.Release()
	return l.Discord()
}

// IsTelegramServiceReady 回報目前 Telegram 服務是否可用。
func (sm *ServiceManager) IsTelegramServiceReady() bool {
	l := sm.Acquire()
	if l == nil {
		return false
	}
	defer l.Release()
	return l.Telegram() != nil
}

// IsSlackServiceReady 回報目前 Slack 服務是否可用。
func (sm *ServiceManager) IsSlackServiceReady() bool {
	l := sm.Acquire()
	if l == nil {
		return false
	}
	defer l.Release()
	return l.Slack() != nil
}

// IsDiscordServiceReady 回報目前 Discord 服務是否可用。
func (sm *ServiceManager) IsDiscordServiceReady() bool {
	l := sm.Acquire()
	if l == nil {
		return false
	}
	defer l.Release()
	return l.Discord() != nil
}

// GetTemplateEngine 回傳目前模板引擎。
func (sm *ServiceManager) GetTemplateEngine() *template.TemplateEngine {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.templateEngine
}
func (sm *ServiceManager) initializeNotificationManager(b *serviceBundle) {
	if err := notification.GetNotificationManager().Initialize(sm.GetTemplateEngine(), b.telegram, b.slack, b.discord); err != nil {
		logger.Error("Failed to initialize notification manager", "service_manager", logger.Err(err))
	}
}

func buildServiceBundle(cfg config.Config, strict bool) (*serviceBundle, error) {
	b := &serviceBundle{drained: make(chan struct{})}
	if cfg.Telegram.Enable && cfg.Telegram.Token != "" {
		s, e := NewTelegramService(cfg.Telegram.Token)
		if e != nil {
			if strict {
				return nil, fmt.Errorf("建立 Telegram 服務失敗: %w", e)
			}
			b.telegram = &TelegramService{chatIDs: make(map[int]int64)}
		} else {
			b.telegram = s
		}
	}
	if cfg.Slack.Enable && cfg.Slack.Token != "" {
		s, e := NewSlackService(cfg.Slack.Token)
		if e != nil {
			if strict {
				return nil, fmt.Errorf("建立 Slack 服務失敗: %w", e)
			}
		} else {
			b.slack = s
		}
	}
	if cfg.Discord.Enable && cfg.Discord.Token != "" {
		s, e := NewDiscordService(cfg.Discord)
		if e != nil {
			if strict {
				return nil, fmt.Errorf("建立 Discord 服務失敗: %w", e)
			}
		} else {
			b.discord = s
		}
	}
	return b, nil
}
func newTemplateEngine() *template.TemplateEngine {
	e := template.NewTemplateEngine()
	if err := e.LoadConfigFromConfigs(); err != nil {
		logger.Warn("Failed to load template config, using defaults", "service_manager", logger.Err(err))
	}
	for _, p := range []string{"templates/alerts", "./templates/alerts", "../templates/alerts"} {
		if e.LoadTemplates(p) == nil {
			return e
		}
	}
	logger.Warn("Failed to load templates from all paths", "service_manager")
	return e
}

// ReloadTemplateEngine 以新載入的模板引擎原子替換目前實例。
func (sm *ServiceManager) ReloadTemplateEngine() {
	sm.mu.Lock()
	sm.templateEngine = newTemplateEngine()
	sm.mu.Unlock()
}

var _ = strings.Join
