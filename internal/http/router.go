package router

import (
	"net/http"
	"time"

	handlers "github.com/url_shortener/internal/handlers"

	"github.com/gin-gonic/gin"
)

func New(userHandler *handlers.UserHandler, domainHandler *handlers.DomainHandler, linkHandler *handlers.LinkHandler) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(requestLogger())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "ts": time.Now().UTC()})
	})

	v1 := router.Group("/v1")
	{
		users := v1.Group("/users")
		{
			users.GET("", userHandler.GetUsers)
			users.GET("/:id", userHandler.GetUserById)
			users.POST("", userHandler.Create)
			users.DELETE("/:id", userHandler.DeleteUser)
			users.PUT("", userHandler.UpdateUser)
		}
		domains := v1.Group("/domains")
		{
			domains.GET("", domainHandler.GetDomains)
			domains.POST("", domainHandler.Create)
			domains.GET("/:id", domainHandler.GetDomainById)
		}
		links := v1.Group("/links")
		{
			links.GET("/:code", linkHandler.RedirectLink)
			links.POST("", linkHandler.Create)
		}
	}

	return router
}

func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		lat := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path
		c.Writer.Header().Set("X-Response-Time", lat.String())

		if status >= 500 {
			_, _ = gin.DefaultErrorWriter.Write([]byte(
				time.Now().Format(time.RFC3339) + " " + method + " " + path + " -> " + lat.String() + "\n"))
		} else {
			// stdout
			_, _ = gin.DefaultWriter.Write([]byte(
				time.Now().Format(time.RFC3339) + " " + method + " " + path + " -> " + lat.String() + "\n"))
		}
	}
}
