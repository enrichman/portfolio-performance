package btp_valore_gn27

import (
	"github.com/enrichman/portfolio-perfomance/pkg/bond/btp"
	"github.com/enrichman/portfolio-perfomance/pkg/security"
)

func init() {
	security.Register(btp.New("BTP Valore Gn27", "IT0005547408"))
}
