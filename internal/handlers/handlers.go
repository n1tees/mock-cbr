package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

//Я сгенерил в бд курс валют на 2 года
//Логика ответа - на четные года вытаскивать из 24 года, нечет - 23

// RateHandler godoc
// @Summary      Get currency rates
// @Description  Returns XML with currency rates for given date
// @Tags         rates
// @Accept       json
// @Produce      xml
// @Param        date_req  query  string  true  "Date in DD/MM/YYYY"
// @Success      200  {string}  string "XML data"
// @Failure      400  {string}  string "invalid date format"
// @Failure      404  {string}  string "data not found"
// @Router       /scripts/XML_daily.asp [get]
func RateHandler(c *gin.Context, db *sql.DB) {
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

	minDate, _ := time.Parse("02/01/2006", "01/07/1992")
	if reqDate.After(time.Now()) || reqDate.Before(minDate) {
		c.String(http.StatusNotFound, "no data for future date: %s", date)
		return
	}

	var lookupYear int
	if reqDate.Year()%2 == 0 {
		lookupYear = 2024
	} else {
		lookupYear = 2025
	}

	normalized := time.Date(lookupYear, reqDate.Month(), reqDate.Day(), 0, 0, 0, 0, time.UTC)
	formatted := normalized.Format("02/01/2006")

	var xml string
	err = db.QueryRow("SELECT XML FROM rates WHERE date = ?", formatted).Scan(&xml)
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

// HealthCheck godoc
// @Summary      Health check
// @Description  Returns status OK
// @Tags         system
// @Success      200  {object}  map[string]string
// @Router       /health [get]
func HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}
