package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"mock-cbr/internal/handlers"
	"mock-cbr/internal/router"

	"github.com/gin-gonic/gin"
	_ "modernc.org/sqlite"
)

var shuttingDown bool

func performRequest(r http.Handler, method, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestRateHandler(t *testing.T) {

	db := SetupTestDB(t)
	defer db.Close()

	router := gin.Default()
	router.GET("/scripts/XML_daily.asp", func(c *gin.Context) {
		handlers.RateHandler(c, db)
	})

	tests := []struct {
		name       string
		dateParam  string
		wantStatus int
		wantXML    bool
	}{
		{
			name:       "Data OK",
			dateParam:  "18/10/2025",
			wantStatus: http.StatusOK,
			wantXML:    true,
		},
		{
			name:       "Data in future",
			dateParam:  time.Now().AddDate(0, 0, 100).Format("02/01/2006"),
			wantStatus: http.StatusNotFound,
			wantXML:    false,
		},
		{
			name:       "Data in past",
			dateParam:  "01/07/1980",
			wantStatus: http.StatusNotFound,
			wantXML:    false,
		},
		{
			name:       "Data not found",
			dateParam:  "10/10/2010",
			wantStatus: http.StatusNotFound,
			wantXML:    false,
		},
		{
			name:       "Uncorrect data",
			dateParam:  "10.10.2025",
			wantStatus: http.StatusBadRequest,
			wantXML:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := "/scripts/XML_daily.asp"
			if tt.dateParam != "" {
				path += "?date_req=" + tt.dateParam
			}

			w := performRequest(router, "GET", path)

			if w.Code != tt.wantStatus {
				t.Fatalf("[%s] expected status %d, got %d", tt.name, tt.wantStatus, w.Code)
			}

			if tt.wantXML && !strings.Contains(w.Body.String(), "ValCurs") {
				t.Fatalf("[%s] expected XML response, got: %s", tt.name, w.Body.String())
			}

			if !tt.wantXML && strings.Contains(w.Body.String(), "ValCurs") {
				t.Fatalf("[%s] expected no XML, got: %s", tt.name, w.Body.String())
			}
		})
	}
}

func TestRateHandler_Parallel_TableDriven(t *testing.T) {
	db := SetupTestDB(t)
	defer db.Close()

	router := gin.Default()
	router.GET("/scripts/XML_daily.asp", func(c *gin.Context) {
		handlers.RateHandler(c, db)
	})

	cases := []struct {
		name       string
		dateParam  string
		wantStatus int
		wantXML    bool
	}{
		{
			name:       "Data OK",
			dateParam:  "18/10/2025",
			wantStatus: http.StatusOK,
			wantXML:    true,
		},
		{
			name:       "Data in future",
			dateParam:  time.Now().AddDate(0, 0, 100).Format("02/01/2006"),
			wantStatus: http.StatusNotFound,
			wantXML:    false,
		},
		{
			name:       "Data in past",
			dateParam:  "01/07/1980",
			wantStatus: http.StatusNotFound,
			wantXML:    false,
		},
		{
			name:       "Data not found",
			dateParam:  "10/10/2010",
			wantStatus: http.StatusNotFound,
			wantXML:    false,
		},
		{
			name:       "Uncorrect data",
			dateParam:  "10.10.2025",
			wantStatus: http.StatusBadRequest,
			wantXML:    false,
		},
	}

	const parallelRequests = 100

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var wg sync.WaitGroup
			wg.Add(parallelRequests)

			for i := 0; i < parallelRequests; i++ {
				go func(i int) {
					defer wg.Done()
					w := performRequest(router, "GET", "/scripts/XML_daily.asp?date_req="+tc.dateParam)
					if w.Code != tc.wantStatus {
						t.Errorf("[%s] goroutine %d: expected %d, got %d", tc.name, i, tc.wantStatus, w.Code)
					}
					if tc.wantXML && !strings.Contains(w.Body.String(), "ValCurs") {
						t.Errorf("[%s] goroutine %d: expected XML, got: %s", tc.name, i, w.Body.String())
					}
				}(i)
			}

			wg.Wait()
		})
	}
}

func TestRateHandler_ServiceUnavailableDuringShutdown(t *testing.T) {
	db := SetupTestDB(t)
	defer db.Close()

	r := router.SetupRouter(db)

	router.ShuttDown(true)
	defer func() { router.ShuttDown(false) }()

	w := performRequest(r, "GET", "/scripts/XML_daily.asp?date_req=02/03/2023")

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}

func TestRateHandler_ParallelRequestsTable(t *testing.T) {
	db := SetupTestDB(t)
	defer db.Close()

	router := gin.Default()
	router.GET("/scripts/XML_daily.asp", func(c *gin.Context) {
		handlers.RateHandler(c, db)
	})

	testDate := time.Now().AddDate(0, 0, -5).Format("02/01/2006")

	cases := []struct {
		name        string
		dateParam   string
		numRequests int
		wantStatus  int
	}{
		{"10 requests valid date", testDate, 10, http.StatusOK},
		{"50 requests valid date", testDate, 50, http.StatusOK},
		{"1 request valid date", testDate, 1, http.StatusOK},
		{"5 requests future date", time.Now().AddDate(0, 0, 5).Format("02/01/2006"), 5, http.StatusNotFound},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var wg sync.WaitGroup
			wg.Add(tc.numRequests)

			for i := 0; i < tc.numRequests; i++ {
				go func(i int) {
					defer wg.Done()
					w := performRequest(router, "GET", "/scripts/XML_daily.asp?date_req="+tc.dateParam)
					if w.Code != tc.wantStatus {
						t.Errorf("goroutine %d: expected %d, got %d", i, tc.wantStatus, w.Code)
					}
				}(i)
			}

			wg.Wait()
		})
	}
}
