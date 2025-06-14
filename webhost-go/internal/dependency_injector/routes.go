package dependency_injector

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine, h *HandlerRegistry) {
	r.POST("/register", h.UserHandler.Register)
	r.POST("/login", h.UserHandler.Login)

	userProtected := r.Group("/users", h.AuthMiddleware.RequireUser(), h.AuthMiddleware.RequireSelfOrAdmin())
	{
		userProtected.GET("/:username", h.UserHandler.GetUserInfo)
		userProtected.DELETE("/:username", h.UserHandler.DelteUser)
		userProtected.PUT("/:username", h.UserHandler.UpdateUserHandler)
	}

	adminProtected := r.Group("/users", h.AuthMiddleware.RequireAdmin())
	{
		adminProtected.GET("", h.UserHandler.ListUsers)
	}
}
