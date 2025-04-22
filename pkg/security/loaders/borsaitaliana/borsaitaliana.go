package borsaitaliana

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/enrichman/portfolio-perfomance/pkg/security"
)

type BorsaItalianaQuoteLoader struct {
	name             string
	isin             string
	market           string
	alphanumericCode string
}

func New(name, isin string) *BorsaItalianaQuoteLoader {
	isinMarketCode := strings.Split(isin, ".")

	loader := &BorsaItalianaQuoteLoader{
		name: name,
		isin: isinMarketCode[0],
	}

	if len(isinMarketCode) > 1 {
		loader.market = isinMarketCode[1]
	}

	if len(isinMarketCode) > 2 {
		loader.alphanumericCode = isinMarketCode[2]
	}

	return loader
}

func (b *BorsaItalianaQuoteLoader) Name() string {
	return b.name
}

func (b *BorsaItalianaQuoteLoader) ISIN() string {
	return b.isin
}

type Data struct {
	Data [][5]float32 `json:"d"`
}

type RequestPayload struct {
	SampleTime           string
	TimeFrame            string
	RequestedDataSetType string
	ChartPriceType       string
	Key                  string
	OffSet               int
	FromDate             string `json:",omitempty"`
	ToDate               string `json:",omitempty"`
	UseDelay             bool
	KeyType              string
	KeyType2             string
	Language             string
}

func (b *BorsaItalianaQuoteLoader) LoadQuotes() ([]security.Quote, error) {
	payload := RequestPayload{
		SampleTime:           "1d",
		TimeFrame:            "5y",
		RequestedDataSetType: "ohlc",
		ChartPriceType:       "price",
		Key:                  fmt.Sprintf("%s.%s", b.isin, b.market),
		KeyType:              "Topic",
		KeyType2:             "Topic",
		Language:             "en-US",
	}

	// if the alphanumeric code is not empty, use it as the key (CW and Certificates)
	if b.alphanumericCode != "" && b.market == "MCW" {
		payload.Key = fmt.Sprintf("%s.%s", b.alphanumericCode, b.market)
	}

	payloadBytes, err := json.Marshal(struct {
		Request RequestPayload `json:"request"`
	}{Request: payload})
	if err != nil {
		return nil, fmt.Errorf("error marshaling request body")
	}

	res, err := http.Post(
		"https://charts.borsaitaliana.it/charts/services/ChartWService.asmx/GetPricesWithVolume",
		"application/json",
		bytes.NewBuffer(payloadBytes),
	)
	if err != nil {
		return nil, fmt.Errorf("error during post request: %w", err)
	}

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading body: %w", err)
	}

	var result Data
	err = json.Unmarshal(bodyBytes, &result)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling body: %w", err)
	}

	quotes := []security.Quote{}
	for _, quote := range result.Data {
		quotes = append(quotes, security.Quote{
			Date:  time.Unix(int64(quote[0]/1000), 0).In(time.UTC),
			Close: quote[1],
		})
	}

	return quotes, nil
}
