package btp_italia_gn30

import (
	"github.com/enrichman/portfolio-perfomance/pkg/bond/btp"
	"github.com/enrichman/portfolio-perfomance/pkg/security"
)

func init() {
	security.Register(btp.New("BTP Italia Gn28", "IT0005497000"))
}
