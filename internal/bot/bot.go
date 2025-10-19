package bot

import (
	"log"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	Client       *tgbotapi.BotAPI
	updateChan   tgbotapi.UpdatesChannel
	stopChan     chan struct{}
	name         string
	mu           sync.Mutex
	mainGroupID  int64
	adminGroupID int64
	adminUserIDs map[int64]bool
	adminMu      sync.RWMutex
}

// creates a new bot instance
func New(name, token string, mainGroupID, adminGroupID int64) (*Bot, error) {
	botClient, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60
	updateChan := botClient.GetUpdatesChan(updateConfig)

	return &Bot{
		Client:       botClient,
		updateChan:   updateChan,
		stopChan:     make(chan struct{}),
		name:         name,
		mainGroupID:  mainGroupID,
		adminGroupID: adminGroupID,
		adminUserIDs: make(map[int64]bool),
	}, nil
}

// HandlerSet contains handlers for a specific chat type
type HandlerSet struct {
	Commands  map[string]func(b *Bot, update tgbotapi.Update) error
	Messages  []func(b *Bot, update tgbotapi.Update) error
	Callbacks map[string]func(b *Bot, update tgbotapi.Update) error
}

// begins processing updates with handlers for different chat types
func (b *Bot) Start(
	mainGroupHandlers HandlerSet,
	adminGroupHandlers HandlerSet,
	privateHandlers HandlerSet,
) {
	log.Printf("[%s] authorized on account %s", b.name, b.Client.Self.UserName)

	// fetch admin list on startup
	b.refreshAdminList()

	for {
		select {
		case update := <-b.updateChan:
			go b.routeUpdate(update, mainGroupHandlers, adminGroupHandlers, privateHandlers)
		case <-b.stopChan:
			return
		}
	}
}

// refreshAdminList fetches current admin list from admin group
func (b *Bot) refreshAdminList() {
	config := tgbotapi.ChatAdministratorsConfig{
		ChatConfig: tgbotapi.ChatConfig{
			ChatID: b.adminGroupID,
		},
	}

	admins, err := b.Client.GetChatAdministrators(config)
	if err != nil {
		log.Printf("[%s] failed to get admin list: %v", b.name, err)
		return
	}

	b.adminMu.Lock()
	defer b.adminMu.Unlock()

	// clear and rebuild admin list
	b.adminUserIDs = make(map[int64]bool)
	for _, admin := range admins {
		b.adminUserIDs[admin.User.ID] = true
		log.Printf("[%s] registered admin: %d (%s)", b.name, admin.User.ID, admin.User.UserName)
	}
}

// isAdmin checks if user id is in admin list
func (b *Bot) isAdmin(userID int64) bool {
	b.adminMu.RLock()
	defer b.adminMu.RUnlock()
	return b.adminUserIDs[userID]
}

// routes updates to appropriate handler set based on chat type or user id
func (b *Bot) routeUpdate(
	update tgbotapi.Update,
	mainGroupHandlers HandlerSet,
	adminGroupHandlers HandlerSet,
	privateHandlers HandlerSet,
) {
	var chatID int64
	var userID int64

	// determine chat id and user id from update
	if update.Message != nil {
		chatID = update.Message.Chat.ID
		userID = update.Message.From.ID
	} else if update.CallbackQuery != nil && update.CallbackQuery.Message != nil {
		chatID = update.CallbackQuery.Message.Chat.ID
		userID = update.CallbackQuery.From.ID
	} else {
		log.Printf("[%s] update has no chat id", b.name)
		return
	}

	// route to appropriate handler set
	var handlers HandlerSet
	var chatType string

	switch {
	case chatID == b.mainGroupID:
		handlers = mainGroupHandlers
		chatType = "main group"
	case chatID == b.adminGroupID:
		handlers = adminGroupHandlers
		chatType = "admin group"
	case chatID > 0: // private chat
		// check if user is admin - route to admin handlers if true
		if b.isAdmin(userID) {
			handlers = adminGroupHandlers
			chatType = "private (admin)"
		} else {
			handlers = privateHandlers
			chatType = "private"
		}
	default:
		log.Printf("[%s] unrecognized chat id: %d", b.name, chatID)
		return
	}

	log.Printf("[%s] routing to %s handler", b.name, chatType)
	b.processUpdate(update, handlers)
}

// handles incoming updates with provided handler set
func (b *Bot) processUpdate(update tgbotapi.Update, handlers HandlerSet) {
	// handle command updates
	if update.Message != nil && update.Message.IsCommand() {
		command := update.Message.Command()
		if handler, exists := handlers.Commands[command]; exists {
			if err := handler(b, update); err != nil {
				log.Printf("[%s] command handler error for /%s: %v", b.name, command, err)
			}
			return
		}
		log.Printf("[%s] unhandled command: /%s", b.name, command)
		return
	}

	// handle callback queries
	if update.CallbackQuery != nil {
		data := update.CallbackQuery.Data
		// extract callback query identifier (before first colon if exists)
		query := data
		for i, c := range data {
			if c == ':' {
				query = data[:i]
				break
			}
		}

		if handler, exists := handlers.Callbacks[query]; exists {
			if err := handler(b, update); err != nil {
				log.Printf("[%s] callback handler error for %s: %v", b.name, query, err)
			}
			return
		}

		log.Printf("[%s] unhandled callback: %s", b.name, query)
		// send error message only for private chats
		if update.CallbackQuery.Message != nil && update.CallbackQuery.Message.Chat.ID > 0 {
			b.SendMessage(update.CallbackQuery.Message.Chat.ID, "команда не распознана")
		}
		return
	}

	// run generic message handlers
	for _, handler := range handlers.Messages {
		if err := handler(b, update); err != nil {
			log.Printf("[%s] message handler error: %v", b.name, err)
		}
	}
}

// halts the bot
func (b *Bot) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()
	close(b.stopChan)
}

func (b *Bot) SendMessage(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.DisableWebPagePreview = true
	_, err := b.Client.Send(msg)
	return err
}

func (b *Bot) SendMessageWithMarkdown(chatID int64, text string, disableLinks bool) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.DisableWebPagePreview = disableLinks
	_, err := b.Client.Send(msg)
	return err
}

func (b *Bot) SendMessageWithButtons(
	chatID int64,
	text string,
	keyboard tgbotapi.InlineKeyboardMarkup,
) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	msg.ParseMode = "Markdown"
	msg.DisableWebPagePreview = true

	_, err := b.Client.Send(msg)
	return err
}

func (b *Bot) SendMessageWithButtonsNoLinks(
	chatID int64,
	text string,
	keyboard tgbotapi.InlineKeyboardMarkup,
) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	msg.ParseMode = "Markdown"
	msg.DisableWebPagePreview = true

	_, err := b.Client.Send(msg)
	return err
}
