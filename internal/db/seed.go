package db

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"
)

type currency struct {
	ID       string
	NumCode  string
	CharCode string
	Nominal  int
	Name     string
	BaseRate float64
}

var currencies = []currency{
	{"R01235", "840", "USD", 1, "Доллар США", 100.0},
	{"R01239", "978", "EUR", 1, "Евро", 110.0},
	{"R01820", "392", "JPY", 100, "Иена", 67.0},
	{"R01035", "826", "GBP", 1, "Фунт стерлингов", 130.0},
	{"R01350", "124", "CAD", 1, "Канадский доллар", 74.0},
}

func generateXMLForDate(date time.Time, randGen *rand.Rand) string {
	xml := fmt.Sprintf(`<ValCurs Date="%s" name="Foreign Currency Market">`, date.Format("02.01.2006"))

	for _, c := range currencies {
		drift := 1 + (randGen.Float64()-0.5)*0.02
		val := c.BaseRate * drift
		vunit := val / float64(c.Nominal)

		xml += fmt.Sprintf(`
<Valute ID="%s">
    <NumCode>%s</NumCode>
    <CharCode>%s</CharCode>
    <Nominal>%d</Nominal>
    <Name>%s</Name>
    <Value>%.4f</Value>
    <VunitRate>%.6f</VunitRate>
</Valute>`, c.ID, c.NumCode, c.CharCode, c.Nominal, c.Name, val, vunit)
	}

	xml += "\n</ValCurs>"
	return xml
}

func SeedRates(db *sql.DB, startDate, endDate time.Time) error {
	if endDate.Before(startDate) {
		return fmt.Errorf("end date must be after start date")
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO rates (date, xml) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	randGen := rand.New(rand.NewSource(time.Now().UnixNano()))
	count := 0

	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		xml := generateXMLForDate(d, randGen)

		if _, err := stmt.Exec(d.Format("02/01/2006"), xml); err != nil {
			return fmt.Errorf("failed to insert record for %s: %v", d.Format("02.01.2006"), err)
		}

		count++
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	log.Printf("Seeded %d records into database (from %s to %s)",
		count, startDate.Format("02.01.2006"), endDate.Format("02.01.2006"))

	return nil
}
