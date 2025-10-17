package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

//Я сгенерил в бд курс валют на 2 года
//Логика ответа - на четные года вытаскивать из 24 года, нечет - 23

func RateHandler(c *gin.Context, dbConn *sql.DB) {
	date := c.Query("date_req")
	if date == "" {
		c.String(http.StatusBadRequest, "missing date_req")
		return
	}

	reqDate, err := time.Parse("02/01/2006", date)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid date format, use DD/MM/YYYY")
		return
	}

	var lookupYear int
	if reqDate.Year()%2 == 0 {
		lookupYear = 2024
	} else {
		lookupYear = 2023
	}

	normalized := time.Date(lookupYear, reqDate.Month(), reqDate.Day(), 0, 0, 0, 0, time.UTC)
	formatted := normalized.Format("02/01/2006")

	var xml string
	err = dbConn.QueryRow("SELECT XML FROM rates WHERE date = ?", formatted).Scan(&xml)
	if err == sql.ErrNoRows {
		c.String(http.StatusNotFound, "no data for date: %s", formatted)
		return
	} else if err != nil {
		c.String(http.StatusInternalServerError, "database error: %v", err)
		return
	}

	c.Header("Content-Type", "application/xml; charset=utf-8")
	c.String(http.StatusOK, xml)
}
