package handler

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	serverHandler "github.com/perfect-panel/server/internal/handler/server"
	"github.com/perfect-panel/server/internal/svc"
	"github.com/perfect-panel/server/pkg/hertzx"
)

func RegisterNativeHandlers(router *server.Hertz, serverCtx *svc.ServiceContext) {
	subscribePath := serverCtx.Config.Subscribe.SubscribePath
	if subscribePath == "" {
		subscribePath = "/v1/subscribe/config"
	}
	router.GET(subscribePath, SubscribeHandler(serverCtx))
	if serverCtx.Config.Subscribe.PanDomain {
		router.GET("/", PanDomainSubscribeHandler(serverCtx))
	}

	serverGroup := router.Group("/v1/server", serverHandler.ServerMiddleware(serverCtx))
	serverGroup.GET("/config", serverHandler.GetServerConfigHandler(serverCtx))
	serverGroup.POST("/online", serverHandler.PushOnlineUsersHandler(serverCtx))
	serverGroup.POST("/push", serverHandler.ServerPushUserTrafficHandler(serverCtx))
	serverGroup.POST("/push_network_activity", hertzx.Wrap(serverHandler.ServerPushUserNetworkActivityHandler(serverCtx)))
	serverGroup.POST("/status", serverHandler.ServerPushStatusHandler(serverCtx))
	serverGroup.GET("/user", serverHandler.GetServerUserListHandler(serverCtx))

	// Probe agent endpoints (token-based, no server middleware)
	router.GET("/v1/probe_agent/config", hertzx.Wrap(serverHandler.ProbeAgentGetConfigHandler(serverCtx)))
	router.POST("/v1/probe_agent/heartbeat", hertzx.Wrap(serverHandler.ProbeAgentHeartbeatHandler(serverCtx)))
	router.POST("/v1/probe_agent/report", hertzx.Wrap(serverHandler.ProbeAgentReportHandler(serverCtx)))

	router.GET("/v2/server/:server_id", serverHandler.QueryServerProtocolConfigHandler(serverCtx))
}
