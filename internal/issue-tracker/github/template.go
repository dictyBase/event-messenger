package github

import (
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
	"github.com/dictyBase/go-genproto/dictybaseapis/stock"
	"github.com/dictyBase/go-genproto/dictybaseapis/user"
)

type IssueContent struct {
	Strains      []*stock.Strain
	Plasmids     []*stock.Plasmid
	StrainInv    [][]string
	PlasmidInv   [][]string
	StrainInfo   [][]string
	Order        *order.Order
	Shipper      *user.User
	Payer        *user.User
	StrainPrice  int
	PlasmidPrice int
}
