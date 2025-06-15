package dependency_injector

import (
	"webhost-go/webhost-go/internal/controller"
	"webhost-go/webhost-go/internal/middleware"
	"webhost-go/webhost-go/internal/services/user_service/authn/token"
)

type HandlerRegistry struct {
	UserHandler    *controller.UserHandler
	JWTManager     *token.JWTManager
	AuthMiddleware *middleware.AuthMiddleware
	HostingHandler *controller.HostingHandler
}
