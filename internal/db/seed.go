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

func seedRates(db *sql.DB) error {

	currencies := []currency{
		{"R01235", "840", "USD", 1, "Доллар США", 100.0},
		{"R01239", "978", "EUR", 1, "Евро", 110.0},
		{"R01820", "392", "JPY", 100, "Иена", 67.0},
		{"R01035", "826", "GBP", 1, "Фунт стерлингов", 130.0},
		{"R01350", "124", "CAD", 1, "Канадский доллар", 74.0},
	}

	start, _ := time.Parse("02/01/2006", "01/01/2023")
	end, _ := time.Parse("02/01/2006", "31/12/2024")

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO rates (date, xml) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	var count = 0
	randGen := rand.New(rand.NewSource(time.Now().UnixNano()))

	for d := start; d.Before(end); d = d.AddDate(0, 0, 1) {

		xml := fmt.Sprintf(`<ValCurs Date="%s" name="Foreign Currency Market">`, d.Format("02.01.2006"))

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

		_, err := stmt.Exec(d.Format("02/01/2006"), xml)
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to insert record: %v", err)
		}

		count++
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	log.Printf("Seeded %d records into database (from %s to %s)",
		count, start.Format("02.01.2006"), end.Format("02.01.2006"))
	return nil
}
