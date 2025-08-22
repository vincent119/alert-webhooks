package service

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"alert-webhooks/config"
	"alert-webhooks/pkg/logger"
)

// DiscordService provides Discord messaging functionality
type DiscordService struct {
	session  *discordgo.Session
	config   config.DiscordConf
	guildID  string
	channels map[string]string // level -> channel ID mapping
}

// NewDiscordService creates a new Discord service instance
func NewDiscordService(cfg config.DiscordConf) (*DiscordService, error) {
	if !cfg.Enable {
		return &DiscordService{config: cfg}, nil
	}

	if cfg.Token == "" {
		return nil, fmt.Errorf("Discord token is required when Discord is enabled")
	}

	// Create Discord session
	session, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to create Discord session: %w", err)
	}

	// Test connection
	session.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages

	err = session.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open Discord connection: %w", err)
	}

	// Verify bot permissions
	user, err := session.User("@me")
	if err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to get bot user info: %w", err)
	}

	logger.Info("Discord service connected successfully", 
		"DiscordService", 
		logger.String("bot_id", user.ID),
		logger.String("bot_username", user.Username))

	ds := &DiscordService{
		session:  session,
		config:   cfg,
		guildID:  cfg.GuildID,
		channels: cfg.Channels,
	}

	return ds, nil
}

// SendMessage send message to specified level chat
func (ds *DiscordService) SendMessage(level string, message string) error {
	if !ds.config.Enable {
		return fmt.Errorf("Discord service is disabled")
	}

	if ds.session == nil {
		return fmt.Errorf("Discord session not initialized")
	}

	channelID, err := ds.getChannelForLevel(level)
	if err != nil {
		return err
	}

	return ds.SendMessageToChannel(channelID, message)
}

// SendMessageToChannel sends a message to a specific Discord channel
func (ds *DiscordService) SendMessageToChannel(channelID, message string) error {
	if !ds.config.Enable {
		return fmt.Errorf("Discord service is disabled")
	}

	if ds.session == nil {
		return fmt.Errorf("Discord session not initialized")
	}

	// Discord message length limit is 2000 characters
	if len(message) > 2000 {
		return ds.sendLongMessage(channelID, message)
	}

	_, err := ds.session.ChannelMessageSend(channelID, message)
	if err != nil {
		return ds.handleDiscordError(err, channelID)
	}

	logger.Info("Discord message sent successfully", 
		"DiscordService", 
		logger.String("channel_id", channelID))

	return nil
}

// SendMessageToLevel sends message to specified level channel
func (ds *DiscordService) SendMessageToLevel(level string, message string) error {
	return ds.SendMessage(level, message)
}

// TestConnection tests the Discord connection and bot permissions
func (ds *DiscordService) TestConnection() error {
	if !ds.config.Enable {
		return fmt.Errorf("Discord service is disabled")
	}

	if ds.session == nil {
		return fmt.Errorf("Discord session not initialized")
	}

	// Test getting bot user info
	user, err := ds.session.User("@me")
	if err != nil {
		return fmt.Errorf("failed to get bot user info: %w", err)
	}

	logger.Info("Discord connection test successful", 
		"DiscordService", 
		logger.String("bot_id", user.ID),
		logger.String("bot_username", user.Username))

	return nil
}

// ValidateChannel validates if the bot can send messages to a channel
func (ds *DiscordService) ValidateChannel(channelID string) error {
	if !ds.config.Enable {
		return fmt.Errorf("Discord service is disabled")
	}

	if ds.session == nil {
		return fmt.Errorf("Discord session not initialized")
	}

	// Try to get channel info
	channel, err := ds.session.Channel(channelID)
	if err != nil {
		return fmt.Errorf("channel not accessible: %w", err)
	}

	// Check if it's a text channel
	if channel.Type != discordgo.ChannelTypeGuildText {
		return fmt.Errorf("channel %s is not a text channel", channelID)
	}

	// Check bot permissions in the channel
	permissions, err := ds.session.UserChannelPermissions(ds.session.State.User.ID, channelID)
	if err != nil {
		return fmt.Errorf("failed to check bot permissions: %w", err)
	}

	// Check if bot has send messages permission
	if permissions&discordgo.PermissionSendMessages == 0 {
		return fmt.Errorf("bot lacks Send Messages permission in channel %s", channelID)
	}

	logger.Info("Discord channel validation successful", 
		"DiscordService", 
		logger.String("channel_id", channelID),
		logger.String("channel_name", channel.Name))

	return nil
}

// GetBotInfo returns Discord bot information
func (ds *DiscordService) GetBotInfo() (interface{}, error) {
	if !ds.config.Enable {
		return nil, fmt.Errorf("Discord service is disabled")
	}

	if ds.session == nil {
		return nil, fmt.Errorf("Discord session not initialized")
	}

	user, err := ds.session.User("@me")
	if err != nil {
		return nil, fmt.Errorf("failed to get bot user info: %w", err)
	}

	return map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
		"bot":      user.Bot,
		"avatar":   user.Avatar,
	}, nil
}

// Close closes the Discord session
func (ds *DiscordService) Close() error {
	if ds.session != nil {
		return ds.session.Close()
	}
	return nil
}

// getChannelForLevel returns the channel ID for a given level
func (ds *DiscordService) getChannelForLevel(level string) (string, error) {
	// Convert level format from "L0" to "0" if needed
	levelKey := level
	if strings.HasPrefix(level, "L") {
		levelKey = strings.TrimPrefix(level, "L")
	}
	
	// Try with chat_ids prefix (e.g., "chat_ids0")
	if channelID, exists := ds.channels["chat_ids"+levelKey]; exists && channelID != "" {
		return channelID, nil
	}

	// Try lowercase match for YAML parser compatibility
	if channelID, exists := ds.channels["chat_ids"+strings.ToLower(levelKey)]; exists && channelID != "" {
		return channelID, nil
	}

	return "", fmt.Errorf("no channel configured for level %s (looking for chat_ids%s)", level, levelKey)
}

// sendLongMessage splits and sends long messages that exceed Discord's 2000 character limit
func (ds *DiscordService) sendLongMessage(channelID, message string) error {
	const maxLength = 2000
	
	// Split message into chunks
	chunks := make([]string, 0)
	for len(message) > maxLength {
		splitIndex := maxLength
		// Try to split at last newline to keep formatting
		if lastNewline := strings.LastIndex(message[:maxLength], "\n"); lastNewline > maxLength/2 {
			splitIndex = lastNewline + 1
		}
		
		chunks = append(chunks, message[:splitIndex])
		message = message[splitIndex:]
	}
	
	if len(message) > 0 {
		chunks = append(chunks, message)
	}

	// Send each chunk
	for i, chunk := range chunks {
		if i > 0 {
			chunk = fmt.Sprintf("(Part %d/%d)\n%s", i+1, len(chunks), chunk)
		} else if len(chunks) > 1 {
			chunk = fmt.Sprintf("(Part 1/%d)\n%s", len(chunks), chunk)
		}
		
		_, err := ds.session.ChannelMessageSend(channelID, chunk)
		if err != nil {
			return fmt.Errorf("failed to send message chunk %d: %w", i+1, err)
		}
	}

	logger.Info("Discord long message sent successfully", 
		"DiscordService", 
		logger.String("channel_id", channelID),
		logger.Int("chunks", len(chunks)))

	return nil
}

// handleDiscordError provides user-friendly error messages for common Discord API errors
func (ds *DiscordService) handleDiscordError(err error, channelID string) error {
	errStr := err.Error()
	
	switch {
	case strings.Contains(errStr, "Missing Permissions"):
		return fmt.Errorf("bot lacks necessary permissions in channel %s. Please ensure the bot has 'Send Messages' permission", channelID)
	case strings.Contains(errStr, "Unknown Channel"):
		return fmt.Errorf("channel %s does not exist or bot cannot access it", channelID)
	case strings.Contains(errStr, "Missing Access"):
		return fmt.Errorf("bot does not have access to channel %s. Please invite the bot to the server/channel", channelID)
	case strings.Contains(errStr, "Invalid Form Body"):
		return fmt.Errorf("message content is invalid or too long")
	case strings.Contains(errStr, "Unauthorized"):
		return fmt.Errorf("invalid Discord token. Please check if the token in configuration is correct")
	default:
		return fmt.Errorf("Discord API error: %w", err)
	}
}
