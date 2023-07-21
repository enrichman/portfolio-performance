package security

import (
	"fmt"
	"sort"
	"time"

	"github.com/charmbracelet/log"
)

var Securities = make(map[string]QuoteLoader)

type QuoteLoader interface {
	Name() string
	ISIN() string
	LoadQuotes() ([]Quote, error)
}

type Quote struct {
	Date  time.Time `json:"date"`
	Close float32   `json:"close"`
}

func Register(fund QuoteLoader) {
	isin := fund.ISIN()
	if isin == "" {
		log.Fatal("security ISIN cannot be empty")
	}

	if _, found := Securities[isin]; found {
		log.Fatal(fmt.Sprintf("security '%s' already registered", isin))
	}

	Securities[isin] = fund
	log.Info(fmt.Sprintf("security '%s' registered", isin))
}

func Merge(quotes1 []Quote, quotes2 []Quote) []Quote {
	quotesMap := map[time.Time]Quote{}

	for _, q := range quotes1 {
		q.Date = q.Date.UTC()
		quotesMap[q.Date] = q
	}

	for _, q := range quotes2 {
		q.Date = q.Date.UTC()

		if oldQuote, found := quotesMap[q.Date]; found {
			if oldQuote.Close != q.Close {
				log.Warnf("quote for date '%v' already exists with different value [old: %v - new: %v]",
					q.Date, oldQuote.Close, q.Close,
				)
			}
		}
		quotesMap[q.Date] = q
	}

	mergedQuotes := []Quote{}
	for _, v := range quotesMap {
		mergedQuotes = append(mergedQuotes, v)
	}

	sort.Slice(mergedQuotes, func(i, j int) bool {
		return mergedQuotes[i].Date.Before(mergedQuotes[j].Date)
	})

	return mergedQuotes
}
