package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/enrichman/portfolio-perfomance/pkg/security"
	"github.com/enrichman/portfolio-perfomance/pkg/security/loaders/borsaitaliana"
	"github.com/enrichman/portfolio-perfomance/pkg/security/loaders/fondidoc"
	"github.com/enrichman/portfolio-perfomance/pkg/security/loaders/fonte"
	"github.com/enrichman/portfolio-perfomance/pkg/security/loaders/morganstanley"
	"github.com/enrichman/portfolio-perfomance/pkg/security/loaders/secondapensione"
)

func main() {
	if strings.ToLower(os.Getenv("LOG_LEVEL")) == "debug" {
		log.SetLevel(log.DebugLevel)
	}

	err := loadSecuritiesFromCSV("securities.csv")
	if err != nil {
		log.Errorf("loading securities from CSV: %s", err)
		os.Exit(1)
	}

	log.Infof("loaded %d securities", len(security.Securities))

	for isin, f := range security.Securities {
		start := time.Now().In(time.UTC)

		log.Infof("[%s] loading quotes for '%s'", f.ISIN(), f.Name())

		newQuotes, err := f.LoadQuotes()
		if err != nil {
			log.Errorf("error loading quotes: %s", err)
			continue
		}
		if len(newQuotes) == 0 {
			log.Warn("no quotes found")
			continue
		}

		log.Debug("new quotes loaded",
			"from", newQuotes[0].Date,
			"to", newQuotes[len(newQuotes)-1].Date,
		)

		filename := fmt.Sprintf("out/json/%s.json", isin)
		log.Debugf("loading OLD quotes from '%s'", filename)

		oldQuotes, err := loadQuotesFromFile(filename)
		if err != nil {
			log.Errorf("error loading quotes: %s", err.Error())
			continue
		}

		if len(oldQuotes) == 0 {
			log.Warn("no OLD quotes found")
		} else {
			log.Debug("found OLD quotes",
				"from", oldQuotes[0].Date,
				"to", oldQuotes[len(oldQuotes)-1].Date,
			)
		}

		mergedQuotes := security.Merge(oldQuotes, newQuotes)
		log.Debug("merged quotes",
			"from", mergedQuotes[0].Date,
			"to", mergedQuotes[len(mergedQuotes)-1].Date,
		)

		err = writeQuotesToFile(filename, mergedQuotes)
		if err != nil {
			log.Errorf("error writing quotes: %s", err.Error())
			continue
		}

		addedQuotes := len(mergedQuotes) - len(oldQuotes)
		if addedQuotes == 0 {
			log.Infof("[%s] no new quotes added", f.ISIN())
		} else {
			log.Infof(
				"[%s] new quotes added [%d] - old [%d] - new [%d]",
				f.ISIN(), addedQuotes, len(oldQuotes), len(newQuotes),
			)
		}

		log.Infof("[%s] quotes loaded in %s", f.ISIN(), time.Since(start))
	}
}

func loadQuotesFromFile(filename string) ([]security.Quote, error) {
	oldQuotesByte, err := os.ReadFile(filename)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error reading file [%s]: %s", filename, err.Error())
	}

	var oldQuotes []security.Quote
	err = json.Unmarshal(oldQuotesByte, &oldQuotes)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling file [%s]: %s", filename, err.Error())
	}

	return oldQuotes, nil
}

func writeQuotesToFile(filename string, quotes []security.Quote) error {
	jsonOutput, err := json.MarshalIndent(quotes, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling file [%s]: %s", filename, err.Error())
	}

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("error opening file [%s]: %s", filename, err.Error())
	}
	err = file.Truncate(0)
	if err != nil {
		return fmt.Errorf("error truncating file [%s]: %s", filename, err.Error())
	}

	if _, err = file.Write(jsonOutput); err != nil {
		return fmt.Errorf("error writing to file [%s]: %s", filename, err.Error())
	}

	return nil
}

func loadSecuritiesFromCSV(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("opening file [%s]: %w", path, err)
	}
	defer f.Close()

	// read csv values using csv.Reader
	csvReader := csv.NewReader(f)
	csvReader.Comment = '#'

	data, err := csvReader.ReadAll()
	if err != nil {
		return fmt.Errorf("reading csv: %w", err)
	}

	for _, line := range data {
		isin := line[0]
		name := line[1]
		loader := line[2]

		var quoteLoader security.QuoteLoader
		switch loader {
		case "borsaitaliana":
			quoteLoader = borsaitaliana.New(name, isin)
		case "fonte":
			quoteLoader = fonte.New(name, isin)
		case "secondapensione":
			quoteLoader = secondapensione.New(name, isin)
		case "fondidoc":
			quoteLoader = fondidoc.New(name, isin)
		case "morganstanley":
			quoteLoader = morganstanley.New(name, isin)
		}

		if quoteLoader == nil {
			log.Warnf("quoteLoader [%s] not found for ISIN %s (%s)", loader, isin, name)
			continue
		}

		security.Register(quoteLoader)
	}

	return nil
}
