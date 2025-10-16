package db

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand/v2"
	"time"
)

type currency struct {
	ID       string
	NumCode  string
	CharCode string
	Nominal  int
	Name     string
}

func seedRates(db *sql.DB) error {

	currencies := []currency{
		{"R01235", "840", "USD", 1, "Доллар США"},
		{"R01239", "978", "EUR", 1, "Евро"},
		{"R01820", "392", "JPY", 100, "Иен"},
		{"R01035", "826", "GBP", 1, "Фунт стерлингов"},
		{"R01350", "124", "CAD", 1, "Канадский доллар"},
	}

	start, _ := time.Parse("02/01/2006", "01/01/2024")
	end := time.Now()

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO rates (date, xml) values (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	var count = 0
	for d := start; d.Before(end); d = d.AddDate(0, 0, 1) {

		xml := fmt.Sprintf(`<ValCurs Date="%s" name="Foreign Currency Market">`, d.Format("02.01.2006"))

		for _, c := range currencies {
			val := 50 + rand.Float64()*40
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
			return err
		}
		count++
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Printf("Seeded %d records into database (from %s to %s)", count, start.Format("02.01.2006"), end.Format("02.01.2006"))
	return nil
}
