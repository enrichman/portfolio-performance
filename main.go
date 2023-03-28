package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/log"
	"github.com/enrichman/portfolio-perfomance/pkg/security"
)

func main() {
	log.Infof("loaded %d securities", len(security.Securities))

	for isin, f := range security.Securities {
		start := time.Now().In(time.UTC)

		log.Infof("loading quotes for '%s' [%s]", f.Name(), f.ISIN())

		quotes, err := f.LoadQuotes()
		if err != nil {
			log.Errorf("error loading quotes: %w", err)
			continue
		}
		if len(quotes) == 0 {
			log.Warn("no quotes found")
			continue
		}

		log.Debug("found new quotes",
			"from", quotes[0].Date,
			"to", quotes[len(quotes)-1].Date,
		)

		filename := fmt.Sprintf("out/%s.json", isin)
		log.Debug("loading OLD quotes for %s [%s] from '%s'", isin, f.ISIN(), filename)

		oldQuotes, err := loadQuotesFromFile(filename)
		if err != nil {
			log.Errorf("error loading quotes: %w", err)
			continue
		}

		if len(oldQuotes) == 0 {
			log.Warn("no OLD quotes found")
		} else {
			log.Debug("found OLD quotes",
				"from", oldQuotes[0].Date,
				"to", oldQuotes[len(oldQuotes)-1].Date,
			)

			quotes = security.Merge(oldQuotes, quotes)
			log.Debug("merged quotes",
				"from", quotes[0].Date,
				"to", quotes[len(quotes)-1].Date,
			)
		}

		err = writeQuotesToFile(filename, quotes)
		if err != nil {
			log.Errorf("error writing quotes: %w", err)
			continue
		}

		log.Infof("quotes loaded in %s", time.Now().Sub(start))
	}
}

func loadQuotesFromFile(filename string) ([]security.Quote, error) {
	oldQuotesByte, err := os.ReadFile(filename)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error reading file [%s]: %w", filename, err)
	}

	var oldQuotes []security.Quote
	err = json.Unmarshal(oldQuotesByte, &oldQuotes)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling file [%s]: %w", filename, err)
	}

	return oldQuotes, nil
}

func writeQuotesToFile(filename string, quotes []security.Quote) error {
	jsonOutput, err := json.MarshalIndent(quotes, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling file [%s]: %w", filename, err)
	}

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("error opening file [%s]: %w", filename, err)
	}
	file.Truncate(0)

	if _, err = file.Write(jsonOutput); err != nil {
		return fmt.Errorf("error writing to file [%s]: %w", filename, err)
	}

	return nil
}
