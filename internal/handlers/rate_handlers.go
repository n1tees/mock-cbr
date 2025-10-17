package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RateHandler(c *gin.Context, dbConn *sql.DB) {

	date := c.Query("date_req")

	if date == "" {
		c.String(http.StatusBadRequest, "missing date_req")
		return
	}

	var xml string
	err := dbConn.QueryRow("SELECT XML FROM rates WHERE date = ?", date).Scan(&xml)
	if err == sql.ErrNoRows {
		c.String(http.StatusNotFound, "no data for date: %s", date)
		return
	} else if err != nil {
		c.String(http.StatusInternalServerError, "database error: %v", err)
		return
	}

	c.Header("Content-Type", "application/xml; charset=utf-8")
	c.String(http.StatusOK, xml)
}
