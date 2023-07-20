package prudente

import (
	"github.com/enrichman/portfolio-perfomance/pkg/fp/secondapensione"
	"github.com/enrichman/portfolio-perfomance/pkg/security"
)

func init() {
	security.Register(secondapensione.New("SecondaPensione Prudente ESG", "QS0000003560"))
}
