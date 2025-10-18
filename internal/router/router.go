package router

import (
	"database/sql"
	"net/http"

	"mock-cbr/internal/handlers"

	_ "mock-cbr/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
		handlers.HealthCheck(c)
	})

	r.GET("/scripts/XML_daily.asp", func(c *gin.Context) {
		handlers.RateHandler(c, db)
	})

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}
