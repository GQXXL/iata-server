package system

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/perfect-panel/server/internal/logic/admin/system"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/internal/types"
	"github.com/perfect-panel/server/pkg/hertzx"
	"github.com/perfect-panel/server/pkg/result"
)

func tokenSha256(v string) string {
	sum := sha256.Sum256([]byte(v))
	return hex.EncodeToString(sum[:])
}

func fixedToken(serverId int64, secret string) string {
	base := fmt.Sprintf("%d:%s", serverId, strings.TrimSpace(secret))
	sum := sha256.Sum256([]byte(base))
	// 固定且唯一（同一 server_id + secret 恒定），长度控制为 24 hex
	return fmt.Sprintf("pa_%d_%s", serverId, hex.EncodeToString(sum[:12]))
}

func GenerateProbeAgentTokenHandler(svcCtx *svc.ServiceContext) func(c *hertzx.Context) {
	return func(c *hertzx.Context) {
		var req types.GenerateProbeAgentTokenRequest
		_ = c.ShouldBind(&req)
		if err := svcCtx.Validate(&req); err != nil {
			result.ParamErrorResult(c, err)
			return
		}
		if strings.TrimSpace(req.Name) == "" {
			req.Name = "probe-agent"
		}
		secret := strings.TrimSpace(svcCtx.Config.Node.NodeSecret)
		if secret == "" {
			secret = strings.TrimSpace(svcCtx.Config.JwtAuth.AccessSecret)
		}
		_ = secret
		token := fixedToken(req.ServerId, secret)
		hash := tokenSha256(token)
		l := system.NewGenerateProbeAgentTokenLogic(c.Request.Context(), svcCtx)
		resp, err := l.GenerateProbeAgentToken(&req, hash, token)
		result.HttpResult(c, resp, err)
	}
}
