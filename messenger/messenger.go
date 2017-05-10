package messenger

import (
	"context"
	"net/http"
	"notifier/logging"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
)

const (
	userLeftStatus   = "left"
	userKickedStatus = "kicked"
)

var (
	gLogger = logging.WithPackage("messenger")
)

type Messenger interface {
	SendText(ctx context.Context, chatID int, text string) error
	SendForward(ctx context.Context, toChatID, fromChatID, msgID int) error
	SendForwardWithText(ctx context.Context, toChatID, fromChatID, msgID int, text string) error
	IsUserInChat(ctx context.Context, userID, chatID int) (bool, error)
}

type telegram struct {
	bot *tgbotapi.BotAPI
}

func NewTelegram(apiToken string, httpTimeout int) (Messenger, error) {
	timeout := time.Duration(httpTimeout) * time.Second
	bot, err := tgbotapi.NewBotAPIWithClient(apiToken, &http.Client{Timeout: timeout})
	if err != nil {
		return nil, errors.Wrap(err, "telegram api failed")
	}
	return &telegram{bot}, nil
}

func (ts *telegram) SendText(ctx context.Context, chatID int, text string) error {
	logger := logging.FromContextAndBase(ctx, gLogger).WithFields(log.Fields{"chat_id": chatID, "text": text})
	msg := tgbotapi.NewMessage(int64(chatID), text)
	logger.Debug("Calling send text API method")
	_, err := ts.bot.Send(msg)
	if err != nil {
		return errors.Wrap(err, "cannot send text to telegram API")
	}
	return nil
}

func (ts *telegram) SendForward(ctx context.Context, toChatID, fromChatID, msgID int) error {
	logger := logging.FromContextAndBase(ctx, gLogger).WithField("msg_id", msgID)
	forward := tgbotapi.NewForward(int64(toChatID), int64(fromChatID), msgID)
	logger.Debug("Calling forward message API method")
	_, err := ts.bot.Send(forward)
	if err != nil {
		return errors.Wrap(err, "cannot forward msg to telegram API")
	}
	return nil
}

func (ts *telegram) SendForwardWithText(ctx context.Context, toChatID, fromChatID, msgID int, text string) error {
	err := ts.SendText(ctx, toChatID, text)
	if err != nil {
		return err
	}
	return ts.SendForward(ctx, toChatID, fromChatID, msgID)
}

func (ts *telegram) IsUserInChat(ctx context.Context, userID, chatID int) (bool, error) {
	logger := logging.FromContextAndBase(ctx, gLogger)
	args := tgbotapi.ChatConfigWithUser{ChatID: int64(chatID), UserID: userID}
	logger.Debugf("Calling get chat member info API method, args: %+v", args)
	memberInfo, err := ts.bot.GetChatMember(args)
	if err != nil {
		if isUserNotInChatErr(err) {
			return false, nil
		}
		return false, errors.Wrap(err, "cannot request info about user in chat")
	}
	userStatus := memberInfo.Status
	return userStatus != userLeftStatus && userStatus != userKickedStatus, nil
}

func isUserNotInChatErr(err error) bool {
	return http.StatusText(http.StatusBadRequest) == err.Error()
}
