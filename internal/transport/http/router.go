package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/ktsiligkos/xm_project/internal/transport/http/middleware"
)

// NewRouter sets up the gin engine with core middleware and routes.
func NewRouter(companiesHandler *CompaniesHandler, usersHandler *UsersHandler, authSecret []byte) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), middleware.RequestID())

	v1 := router.Group("/api/v1")
	v1.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1.GET("/companies/:uuid", companiesHandler.Get)
	v1.POST("/login", usersHandler.Login)

	secured := v1.Group("/")
	secured.Use(middleware.RequireAuth(authSecret))
	secured.POST("/companies", companiesHandler.Create)
	secured.DELETE("/companies/:uuid", companiesHandler.Delete)
	secured.PATCH("/companies/:uuid", companiesHandler.Patch)

	return router
}
