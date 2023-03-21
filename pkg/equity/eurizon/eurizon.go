package eurizon

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/enrichman/portfolio-perfomance/pkg/security"
)

type Eurizon struct{}

type FondiDoc struct {
	Sazinte Sazinte `json:"SAZINTE"`
}

type Sazinte struct {
	Data [][2]float32 `json:"data"`
}

// Isin implements security.Fund
func (e *Eurizon) Name() string {
	return "IT0001083424"
}

// LoadQuotes implements security.Fund
func (*Eurizon) LoadQuotes() []security.Quote {
	res, err := http.Get("https://www.fondidoc.it/Chart/ChartData?ids=SAZINTE&cur=EUR")
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("status code: %d", res.StatusCode)

	b, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var fondiDoc FondiDoc
	err = json.Unmarshal(b, &fondiDoc)
	if err != nil {
		log.Fatal(err)
	}

	quotes := []security.Quote{}
	for _, quote := range fondiDoc.Sazinte.Data {
		quotes = append(quotes, security.Quote{
			Date:  time.Unix(int64(quote[0])*100, 0),
			Close: quote[1],
		})
	}

	return quotes
}

func init() {
	security.Register(&Eurizon{})
}
