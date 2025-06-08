package http

import (
	"auth/config"
	"auth/internal/usecase"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	authUC usecase.AuthUseCase
}

func NewHandler(router *gin.Engine, authUC usecase.AuthUseCase, cfg *config.Config) {
	h := &Handler{authUC: authUC}

	router.POST("/register", h.Register)
	router.POST("/login", h.Login)
	router.POST("/refresh", h.Refresh)

	auth := router.Group("/", JWTMiddleware(cfg))
	auth.GET("/me", h.Me)
	auth.POST("/logout", h.Logout)
	auth.DELETE("/users/:id", h.DeleteUser)
}
