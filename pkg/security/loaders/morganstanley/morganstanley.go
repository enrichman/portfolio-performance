package morganstanley

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/enrichman/portfolio-perfomance/pkg/security"
	"golang.org/x/exp/slices"
)

const (
	MorganStanleyUrlTemplate = "https://www.morganstanley.com/pub/content/dam/im/json/imwebdata/im/data/product/OF/%s/chart/historicalNav.json"
)

type MorganStanleyQuoteLoader struct {
	name         string
	isin         string
	fundId       string
	shareClassId string
}

func New(name, isin string) *MorganStanleyQuoteLoader {
	isinIdShareId := strings.Split(isin, ".")

	if len(isinIdShareId) != 3 {
		log.Warnf("Wrong ISIN format for MorganStanleyQuoteLoader: \"%s\" - should be \"ISIN.fundId.shareClassId\"", isin)
		return nil
	}

	return &MorganStanleyQuoteLoader{
		name:         name,
		isin:         isinIdShareId[0],
		fundId:       isinIdShareId[1],
		shareClassId: isinIdShareId[2],
	}
}

type HistoricalNav struct {
	En En `json:"en"`
}

type En struct {
	ShareClasses []ShareClass `json:"shareClasses"`
}

type ShareClass struct {
	Id         string     `json:"id"`
	Ccy        string     `json:"ccy"`
	Currencies []Currency `json:"currencies"`
}

type Currency struct {
	Id     string `json:"id"`
	Series Series `json:"series"`
}

type Series struct {
	Name     string   `json:"name"`
	Category []string `json:"category"`
	Data     []string `json:"data"`
}

func (m *MorganStanleyQuoteLoader) Name() string {
	return m.name
}

func (m *MorganStanleyQuoteLoader) ISIN() string {
	return m.isin
}

func (m *MorganStanleyQuoteLoader) FundId() string {
	return m.fundId
}

func (m *MorganStanleyQuoteLoader) ShareClassId() string {
	return m.shareClassId
}

func (m *MorganStanleyQuoteLoader) LoadQuotes() ([]security.Quote, error) {
	url := fmt.Sprintf(MorganStanleyUrlTemplate, m.fundId)

	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error getting quotes: %w", err)
	}
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("error from request: status_code %d", res.StatusCode)
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading body: %w", err)
	}

	var historicalNav HistoricalNav
	err = json.Unmarshal(b, &historicalNav)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling body: %w", err)
	}

	shareIdx := slices.IndexFunc(historicalNav.En.ShareClasses, func(s ShareClass) bool {
		return s.Id == m.shareClassId
	})
	if shareIdx == -1 {
		return nil, nil
	}

	shareClass := historicalNav.En.ShareClasses[shareIdx]

	eurIdx := slices.IndexFunc(shareClass.Currencies, func(c Currency) bool {
		return c.Id == "EUR"
	})
	if eurIdx == -1 {
		return nil, nil
	}

	currency := shareClass.Currencies[eurIdx]

	if len(currency.Series.Category) != len(currency.Series.Data) {
		log.Warn("Series Category and Data must be the same length")
		return nil, nil
	}

	quotes := []security.Quote{}

	for idx, dateString := range currency.Series.Category {
		valueString := currency.Series.Data[idx]

		if dateString == "" || valueString == "" {
			continue
		}

		date, err := time.Parse("01/02/2006", dateString)
		if err != nil {
			panic(err)
		}

		closeQuote, err := strconv.ParseFloat(valueString, 32)
		if err != nil {
			panic(err)
		}

		quotes = append(quotes, security.Quote{
			Date:  date,
			Close: float32(closeQuote),
		})
	}

	return quotes, nil
}
