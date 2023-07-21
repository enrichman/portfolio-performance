package fonte

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/enrichman/portfolio-perfomance/pkg/security"
	"github.com/gocolly/colly/v2"
)

const (
	FonTeDinamicoURL = "https://www.fondofonte.it/gestione-finanziaria/i-valori-quota-dei-comparti/comparto-dinamico/"
)

type Fonte struct {
	name string
	isin string
}

func New(name, isin string) *Fonte {
	return &Fonte{
		name: name,
		isin: isin,
	}
}

func (e *Fonte) Name() string {
	return e.name
}

func (e *Fonte) ISIN() string {
	return e.isin
}

func (f *Fonte) LoadQuotes() ([]security.Quote, error) {
	c := colly.NewCollector()

	type yearContent struct {
		year   string
		months []string
		values []string
	}
	years := []yearContent{}

	c.OnHTML("article.content-text-page", func(e *colly.HTMLElement) {
		e.ForEach("h5.toggle-acf", func(i int, e *colly.HTMLElement) {
			years = append(years, yearContent{year: e.Text})
		})

		e.ForEach("div.toggle-content-acf", func(i int, e *colly.HTMLElement) {
			year := years[i]

			e.ForEach("span", func(spanIndex int, e *colly.HTMLElement) {
				textValue := strings.TrimSpace(e.Text)

				if spanIndex > 1 {
					if spanIndex%2 == 0 {
						year.months = append(year.months, textValue)
					} else {
						year.values = append(year.values, textValue)
					}
				}
			})

			reverse(year.months)
			reverse(year.values)

			years[i] = year
		})
	})

	c.Visit(FonTeDinamicoURL)

	reverse(years)

	quotes := []security.Quote{}

	for _, y := range years {
		for i := range y.months {
			dateString := fmt.Sprintf("%s %s", y.year, convertMonth(y.months[i]))
			tt, err := time.Parse("2006 January", dateString)
			if err != nil {
				panic(err)
			}
			tt = tt.AddDate(0, 1, -1)

			y.values[i] = strings.ReplaceAll(y.values[i], ",", ".")
			closeQuote, err := strconv.ParseFloat(y.values[i], 32)
			if err != nil {
				panic(err)
			}

			quotes = append(quotes, security.Quote{
				Date:  tt,
				Close: float32(closeQuote),
			})
		}
	}

	return quotes, nil
}

func reverse[S ~[]E, E any](s S) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

func convertMonth(month string) time.Month {
	switch month {
	case "Gennaio":
		return time.January
	case "Febbraio":
		return time.February
	case "Marzo":
		return time.March
	case "Aprile":
		return time.April
	case "Maggio":
		return time.May
	case "Giugno":
		return time.June
	case "Luglio":
		return time.July
	case "Agosto":
		return time.August
	case "Settembre":
		return time.September
	case "Ottobre":
		return time.October
	case "Novembre":
		return time.November
	case "Dicembre":
		return time.December
	}
	return time.January
}
