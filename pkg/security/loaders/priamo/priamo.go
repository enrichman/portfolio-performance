package priamo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/enrichman/portfolio-perfomance/pkg/security"
)

const (
	PriamoBilanciatoAvanzatoURL = "https://www.fondopriamo.it/grafici/tabella.php?c=330"
)

type Priamo struct {
	name string
	isin string
}

func New(name, isin string) *Priamo {
	return &Priamo{
		name: name,
		isin: isin,
	}
}

func (e *Priamo) Name() string {
	return e.name
}

func (e *Priamo) ISIN() string {
	return e.isin
}

type PriamoData struct {
	Data   string `json:"data"`
	Anno   string `json:"anno"`
	Valore string `json:"valore"`
}

// {
// 	"data": "settembre 2025",
// 	"anno": "2025",
// 	"valore": "22,762"
// }

func (f *Priamo) LoadQuotes() ([]security.Quote, error) {
	resp, err := http.Get(PriamoBilanciatoAvanzatoURL)
	if err != nil {
		return nil, err
	}

	var data []PriamoData
	if err = json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	// reverse(years)

	quotes := []security.Quote{}

	for _, d := range data {
		fields := strings.Fields(d.Data)

		dateString := fmt.Sprintf("%s %s", d.Anno, convertMonth(fields[0]))
		tt, err := time.Parse("2006 January", dateString)
		if err != nil {
			panic(err)
		}
		tt = tt.AddDate(0, 1, -1)

		value := strings.ReplaceAll(d.Valore, ",", ".")
		closeQuote, err := strconv.ParseFloat(value, 32)
		if err != nil {
			panic(err)
		}

		quotes = append(quotes, security.Quote{
			Date:  tt,
			Close: float32(closeQuote),
		})
	}

	sort.Slice(quotes, func(i, j int) bool {
		return quotes[i].Date.Before(quotes[j].Date)
	})

	return quotes, nil
}

func convertMonth(month string) time.Month {
	switch month {
	case "gennaio":
		return time.January
	case "febbraio":
		return time.February
	case "marzo":
		return time.March
	case "aprile":
		return time.April
	case "maggio":
		return time.May
	case "giugno":
		return time.June
	case "luglio":
		return time.July
	case "agosto":
		return time.August
	case "settembre":
		return time.September
	case "ottobre":
		return time.October
	case "novembre":
		return time.November
	case "dicembre":
		return time.December
	}
	return time.January
}
