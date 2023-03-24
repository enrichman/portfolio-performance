package eurizon

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/enrichman/portfolio-perfomance/pkg/security"
)

type Eurizon struct {
	name string
	isin string
}

func New(name, isin string) *Eurizon {
	return &Eurizon{
		name: name,
		isin: isin,
	}
}

type FondiDoc struct {
	Sazinte Sazinte `json:"SAZINTE"`
}

type Sazinte struct {
	Data [][2]float32 `json:"data"`
}

func (e *Eurizon) Name() string {
	return e.name
}

func (e *Eurizon) ISIN() string {
	return e.isin
}

// LoadQuotes implements security.Fund
func (*Eurizon) LoadQuotes() ([]security.Quote, error) {
	res, err := http.Get("https://www.fondidoc.it/Chart/ChartData?ids=SAZINTE&cur=EUR")
	if err != nil {
		return nil, fmt.Errorf("error getting quotes: %w", err)
	}
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("error from request", "status_code", res.StatusCode)
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading body: %w", err)
	}

	var fondiDoc FondiDoc
	err = json.Unmarshal(b, &fondiDoc)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling body: %w", err)
	}

	quotes := []security.Quote{}
	for _, quote := range fondiDoc.Sazinte.Data {
		quotes = append(quotes, security.Quote{
			Date:  time.Unix(int64(quote[0])*100, 0),
			Close: quote[1],
		})
	}

	return quotes, nil
}

func init() {
	security.Register(New("Azioni Internazionali ESG", "IT0001083424"))
}
