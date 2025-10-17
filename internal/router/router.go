package router

import (
	"database/sql"
	"net/http"

	"mock-cbr/internal/handlers"

	"github.com/gin-gonic/gin"
)

var shuttingDown = false

func ShuttDown(sd bool) {
	shuttingDown = sd
}

func SetupRouter(db *sql.DB) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Middleware
	r.Use(func(c *gin.Context) {
		if shuttingDown {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error": "server is shutting down",
			})
			return
		}
		c.Next()
	})

	// Routes
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "working"})
	})

	r.GET("/scripts/XML_daily.asp", func(c *gin.Context) {
		handlers.RateHandler(c, db)
	})

	return r
}
