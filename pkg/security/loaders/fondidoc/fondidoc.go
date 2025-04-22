package fondidoc

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/enrichman/portfolio-perfomance/pkg/security"
	"github.com/gocolly/colly/v2"
)

type FondiDoc struct {
	name string
	isin string
}

func New(name, isin string) *FondiDoc {
	return &FondiDoc{
		name: name,
		isin: isin,
	}
}

type FondiDocData struct {
	Description string       `json:"desc"`
	Currency    string       `json:"cur"`
	Data        [][2]float32 `json:"data"`
}

func (f *FondiDoc) Name() string {
	return f.name
}

func (f *FondiDoc) ISIN() string {
	return f.isin
}

// LoadQuotes implements security.Fund
func (f *FondiDoc) LoadQuotes() ([]security.Quote, error) {
	c := colly.NewCollector()

	var fidacode string
	c.OnHTML("a[fidacode]", func(e *colly.HTMLElement) {
		fidacode = e.Attr("fidacode")
	})

	if err := c.Visit("https://www.fondidoc.it/Ricerca/Res?txt=" + f.isin); err != nil {
		return nil, err
	}

	res, err := http.Get("https://www.fondidoc.it/Chart/ChartData?ids=" + fidacode + "&cur=EUR")
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

	fondiDocResp := map[string]FondiDocData{}
	if err = json.Unmarshal(b, &fondiDocResp); err != nil {
		return nil, fmt.Errorf("error unmarshaling body: %w", err)
	}

	fondiDocData, ok := fondiDocResp[fidacode]
	if !ok {
		return nil, fmt.Errorf("error getting FondidocData from fidacode: %s", fidacode)
	}

	quotes := []security.Quote{}
	for _, quote := range fondiDocData.Data {
		quotes = append(quotes, security.Quote{
			Date:  time.Unix(int64(quote[0])*100, 0),
			Close: quote[1],
		})
	}

	return quotes, nil
}
