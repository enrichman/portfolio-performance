package btp_italia_mz28

import (
	"github.com/enrichman/portfolio-perfomance/pkg/bond/btp"
	"github.com/enrichman/portfolio-perfomance/pkg/security"
)

func init() {
	security.Register(btp.New("BTP Italia Mz28", "IT0005532723"))
}
