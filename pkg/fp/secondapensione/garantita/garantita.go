package garantita

import (
	"github.com/enrichman/portfolio-perfomance/pkg/fp/secondapensione"
	"github.com/enrichman/portfolio-perfomance/pkg/security"
)

func init() {
	security.Register(secondapensione.New("SecondaPensione Garantita ESG", "QS0000013033"))
}
