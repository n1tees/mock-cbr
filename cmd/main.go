package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	version = "0.1.0"
	port    = ":8080"
)

func main() {

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "working"})
	})

	router.GET("/scripts/XML_daily.asp", func(c *gin.Context) {
		date := c.Query("date_req")
		if date == "" {
			c.String(http.StatusBadRequest, "missing date_req")
			return
		} else {
			c.JSON(http.StatusOK, gin.H{"status": "There will be a sequel here"})
		}
	})

}
