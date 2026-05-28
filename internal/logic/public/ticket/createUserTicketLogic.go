package ticket

import (
	"context"
	"fmt"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/perfect-panel/server/pkg/constant"

	"github.com/perfect-panel/server/internal/model/ticket"
	"github.com/perfect-panel/server/internal/model/user"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"

	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
)

type CreateUserTicketLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Create ticket
func NewCreateUserTicketLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateUserTicketLogic {
	return &CreateUserTicketLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateUserTicketLogic) CreateUserTicket(req *types.CreateUserTicketRequest) error {
	u, ok := l.ctx.Value(constant.CtxKeyUser).(*user.User)
	if !ok {
		logger.Error("current user is not found in context")
		return errors.Wrapf(xerr.NewErrCode(xerr.InvalidAccess), "Invalid Access")
	}
	tk := &ticket.Ticket{
		Title:       req.Title,
		Description: req.Description,
		UserId:      u.Id,
		Status:      ticket.Pending,
	}
	err := l.svcCtx.Store.Ticket().Insert(l.ctx, tk)
	if err != nil {
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseInsertError), "insert ticket error: %v", err.Error())
	}

	l.notifyAdminsNewSupportChat(u.Id, tk.Id, req.Title, req.Description)
	return nil
}

func (l *CreateUserTicketLogic) notifyAdminsNewSupportChat(userID, ticketID int64, title, description string) {
	if l.svcCtx.TelegramBot == nil {
		return
	}
	admins, err := l.svcCtx.Store.User().QueryAdminUsers(l.ctx)
	if err != nil {
		l.Errorw("[SupportChatNotify] query admin users failed", logger.Field("error", err.Error()))
		return
	}

	msgText := fmt.Sprintf("📨 *Support Chat 新会话*\n\n用户ID: `%d`\nTicket ID: `%d`\n标题: %s\n内容: %s\n时间: %s", userID, ticketID, title, description, time.Now().Format("2006-01-02 15:04:05"))

	for _, admin := range admins {
		for _, method := range admin.AuthMethods {
			if method.AuthType != "telegram" || method.AuthIdentifier == "" {
				continue
			}
			chatID, convErr := strconv.ParseInt(method.AuthIdentifier, 10, 64)
			if convErr != nil {
				continue
			}
			msg := tgbotapi.NewMessage(chatID, msgText)
			msg.ParseMode = "Markdown"
			_, sendErr := l.svcCtx.TelegramBot.Send(msg)
			if sendErr != nil {
				l.Errorw("[SupportChatNotify] send telegram failed", logger.Field("error", sendErr.Error()), logger.Field("ticketId", ticketID), logger.Field("chatId", chatID))
			}
		}
	}
}
