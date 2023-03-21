package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/enrichman/portfolio-perfomance/pkg/security"

	_ "github.com/enrichman/portfolio-perfomance/pkg/equity/eurizon"
)

func main() {
	for name, f := range security.Securities {
		log.Info("logging", "security", name)

		filename := fmt.Sprintf("out/%s.json", name)
		oldQuotes := loadQuotesFromFile(filename)

		log.Info("loading new quotes")
		quotes := f.LoadQuotes()

		if len(quotes) > 0 {
			log.Info("found new quotes",
				"from", quotes[0].Date, "to",
				quotes[len(quotes)-1].Date,
			)
		}

		quotes = security.Merge(oldQuotes, quotes)
		log.Info("merged quotes",
			"from", quotes[0].Date, "to",
			quotes[len(quotes)-1].Date,
		)

		writeQuotesToFile(filename, quotes)
	}
}

func loadQuotesFromFile(filename string) []security.Quote {
	log.Info("loading old quotes")

	oldQuotesByte, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal("reading file: " + err.Error())
	}

	var oldQuotes []security.Quote
	err = json.Unmarshal(oldQuotesByte, &oldQuotes)
	if err != nil {
		log.Fatal("unmarshaling file: " + err.Error())
	}

	if len(oldQuotes) > 0 {
		log.Info("found old quotes",
			"from", oldQuotes[0].Date, "to",
			oldQuotes[len(oldQuotes)-1].Date,
		)
	}

	return oldQuotes
}

func writeQuotesToFile(filename string, quotes []security.Quote) {
	jsonOutput, err := json.MarshalIndent(quotes, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	log.Info("opening file")

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		log.Fatal(err)
	}
	file.Truncate(0)

	log.Info("writing quotes")

	if _, err = file.Write(jsonOutput); err != nil {
		log.Fatal(err)
	}
}
