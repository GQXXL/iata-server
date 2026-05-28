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
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/logger"
	"github.com/perfect-panel/server/pkg/xerr"
	"github.com/pkg/errors"
)

type CreateUserTicketFollowLogic struct {
	logger.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Create ticket follow
func NewCreateUserTicketFollowLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateUserTicketFollowLogic {
	return &CreateUserTicketFollowLogic{
		Logger: logger.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateUserTicketFollowLogic) CreateUserTicketFollow(req *types.CreateUserTicketFollowRequest) error {
	u, ok := l.ctx.Value(constant.CtxKeyUser).(*user.User)
	if !ok {
		logger.Error("current user is not found in context")
		return errors.Wrapf(xerr.NewErrCode(xerr.InvalidAccess), "Invalid Access")
	}
	// query ticket
	t, err := l.svcCtx.Store.Ticket().FindOne(l.ctx, req.TicketId)
	if err != nil {
		l.Errorw("[CreateUserTicketFollow] Database query error", logger.Field("error", err.Error()), logger.Field("request", req))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseQueryError), "query ticket failed: %v", err.Error())
	}
	// check access
	if u.Id != t.UserId {
		l.Errorw("[CreateUserTicketFollow] Invalid access", logger.Field("user_id", u.Id), logger.Field("ticket_user_id", t.UserId))
		return errors.Wrapf(xerr.NewErrCode(xerr.InvalidAccess), "invalid access")
	}
	// insert follow
	err = l.svcCtx.Store.Ticket().InsertTicketFollow(l.ctx, &ticket.Follow{
		TicketId: req.TicketId,
		From:     req.From,
		Type:     req.Type,
		Content:  req.Content,
	})
	if err != nil {
		l.Errorw("[CreateUserTicketFollow] Database insert error", logger.Field("error", err.Error()), logger.Field("request", req))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseInsertError), "create ticket follow failed: %v", err.Error())
	}
	err = l.svcCtx.Store.Ticket().UpdateTicketStatus(l.ctx, req.TicketId, u.Id, ticket.Pending)
	if err != nil {
		l.Errorw("[CreateUserTicketFollow] Database update error", logger.Field("error", err.Error()), logger.Field("status", ticket.Pending))
		return errors.Wrapf(xerr.NewErrCode(xerr.DatabaseUpdateError), "update ticket status failed: %v", err.Error())
	}

	l.notifyAdminsSupportChatFollow(u.Id, req.TicketId, req.Content)
	return nil
}

func (l *CreateUserTicketFollowLogic) notifyAdminsSupportChatFollow(userID, ticketID int64, content string) {
	if l.svcCtx.TelegramBot == nil {
		return
	}
	admins, err := l.svcCtx.Store.User().QueryAdminUsers(l.ctx)
	if err != nil {
		l.Errorw("[SupportChatNotify] query admin users failed", logger.Field("error", err.Error()))
		return
	}

	msgText := fmt.Sprintf("💬 *Support Chat 新消息*\n\n用户ID: `%d`\nTicket ID: `%d`\n消息: %s\n时间: %s", userID, ticketID, content, time.Now().Format("2006-01-02 15:04:05"))

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
