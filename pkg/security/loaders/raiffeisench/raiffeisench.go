package raiffeisench

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/enrichman/portfolio-perfomance/pkg/security"
)

const (
	EPFCHFExchangeID = 3233
	CHFCurrencyID    = 1
)

type RaiffeisenchQuoteLoader struct {
	name string
	isin string
}

func New(name, isin string) *RaiffeisenchQuoteLoader {
	return &RaiffeisenchQuoteLoader{
		name: name,
		isin: isin,
	}
}

// getValorFromISIN returns the valoren number (https://en.wikipedia.org/wiki/Valoren_number) for a given ISIN (https://www.isin.org/isin-format/).
// i.e. "CH0025417491" will return "2541749"
func getValorFromISIN(isin string) (int, error) {
	if !strings.HasPrefix(isin, "CH") {
		return 0, fmt.Errorf("valor number can only be extracted for Swiss financial instruments")
	}

	val := strings.TrimPrefix(isin, "CH")
	val = strings.TrimLeft(val, "0")
	val = val[:len(val)-1]

	valor, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("failed to read valor number from ISIN %s: %w", isin, err)
	}
	return valor, nil
}

func (r *RaiffeisenchQuoteLoader) Name() string {
	return r.name
}

func (r *RaiffeisenchQuoteLoader) ISIN() string {
	return r.isin
}

type HistoryQuotesRequest struct {
	Valor      int       `json:"valor"`
	ExchangeId int       `json:"exchangeId"`
	CurrencyId int       `json:"currencyId"`
	From       time.Time `json:"from"`
	To         time.Time `json:"to"`
}

type HistoryQuotesResponse struct {
	HistoryQuotes []struct {
		Date  string  `json:"date"`
		Close float32 `json:"close"`
		Open  float32 `json:"open"`
		High  float32 `json:"high"`
		Low   float32 `json:"low"`
	} `json:"historyQuotes"`
}

func (r *RaiffeisenchQuoteLoader) LoadQuotes() ([]security.Quote, error) {
	// Translate ISIN to Valoren number
	valor, err := getValorFromISIN(r.isin)
	if err != nil {
		return nil, err
	}

	endDate := time.Now()
	startDate := endDate.AddDate(-1, 0, 0) // one year ago

	payload := HistoryQuotesRequest{
		Valor:      valor,
		ExchangeId: EPFCHFExchangeID,
		CurrencyId: CHFCurrencyID,
		From:       startDate.UTC(),
		To:         endDate.UTC(),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request body: %w", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		"https://boerse.raiffeisen.ch/api/HistoryQuotes",
		bytes.NewBuffer(payloadBytes),
	)
	if err != nil {
		return nil, fmt.Errorf("error loading raiffeisen history quotes: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-type", "application/json")
	// We need to directly set the key in the header map,
	// because Go's stdlib canonicalizes all HTTP headers,
	// but the remote API requires this header to be lowercase.
	req.Header["customer"] = []string{"raiffeisen-prod"}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error during get request: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server replied with unexpected status code '%s'", res.Status)
	}

	var result HistoryQuotesResponse
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling body: %w", err)
	}

	// Convert into the desired return format
	quotes := make([]security.Quote, 0, len(result.HistoryQuotes))
	for _, quote := range result.HistoryQuotes {
		date, err := time.Parse("2006-01-02T15:04:05", quote.Date)
		if err != nil {
			return nil, fmt.Errorf("failed to parse date '%s' of quote: %w", quote.Date, err)
		}
		quotes = append(quotes, security.Quote{
			Date:  date,
			Close: float32(quote.Close),
		})
	}

	return quotes, nil
}
