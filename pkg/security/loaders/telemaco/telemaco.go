package telemaco

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/enrichman/portfolio-perfomance/pkg/security"
)

const (
	urlPath = "https://www.fondotelemaco.it/grafici/csv/valori quota %s.csv"
)

type Telemaco struct {
	name string
	isin string
}

func New(name, isin string) *Telemaco {
	return &Telemaco{
		name: name,
		isin: isin,
	}
}

func (e *Telemaco) Name() string { return e.name }

func (e *Telemaco) ISIN() string { return e.isin }

func (t *Telemaco) LoadQuotes() ([]security.Quote, error) {
	isinParts := strings.Split(t.isin, "-")
	if len(isinParts) != 3 {
		return nil, fmt.Errorf("invalid ISIN format for telemaco loader: expected 3 parts separated by '-', e.g. 'FP-Telemaco-dinamico', got %d", len(isinParts))
	}
	url := fmt.Sprintf(urlPath, isinParts[2])

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	reader := csv.NewReader(resp.Body)
	reader.Comma = ';'

	records, err := reader.ReadAll()
	if err != nil && err != io.EOF {
		return nil, err
	}

	quotes := []security.Quote{}
	for i, record := range records {
		if i == 0 || len(record) < 2 {
			continue
		}

		date, err := parseDate(record[0])
		if err != nil {
			return nil, err
		}

		value := strings.ReplaceAll(strings.TrimSpace(record[1]), ",", ".")
		closeQuote, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return nil, err
		}

		quotes = append(quotes, security.Quote{
			Date:  date,
			Close: float32(closeQuote),
		})
	}

	sort.Slice(quotes, func(i, j int) bool {
		return quotes[i].Date.Before(quotes[j].Date)
	})

	return quotes, nil
}

func parseDate(value string) (time.Time, error) {
	parts := strings.Split(strings.TrimSpace(value), "-")
	if len(parts) != 3 {
		return time.Time{}, fmt.Errorf("invalid date format: %s", value)
	}

	month, ok := monthMap[strings.ToLower(parts[1])]
	if !ok {
		return time.Time{}, fmt.Errorf("invalid month: %s", parts[1])
	}

	normalized := fmt.Sprintf("%s-%s-%s", parts[0], month, parts[2])
	return time.Parse("02-Jan-06", normalized)
}

var monthMap = map[string]string{
	"gen": "Jan",
	"feb": "Feb",
	"mar": "Mar",
	"apr": "Apr",
	"mag": "May",
	"giu": "Jun",
	"lug": "Jul",
	"ago": "Aug",
	"set": "Sep",
	"ott": "Oct",
	"nov": "Nov",
	"dic": "Dec",
}
