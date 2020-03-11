package github

import (
	"strings"

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

func (c *IssueContent) IsPlasmid(str string) bool {
	return strings.Contains(str, "DBP")
}

func (c *IssueContent) IsStrain(str string) bool {
	return strings.Contains(str, "DBS")
}

func (c *IssueContent) PlasmidItems() int {
	count := 0
	for _, item := range c.Order.Data.Attributes.Items {
		if c.IsPlasmid(item) {
			count++
		}
	}
	return count
}

func (c *IssueContent) StrainItems() int {
	count := 0
	for _, item := range c.Order.Data.Attributes.Items {
		if c.IsStrain(item) {
			count++
		}
	}
	return count
}

func (c *IssueContent) PlasmidCost() int {
	return c.PlasmidItems() * c.PlasmidPrice
}

func (c *IssueContent) StrainCost() int {
	return c.StrainItems() * c.StrainPrice
}

func (c *IssueContent) TotalCost() int {
	return c.StrainCost() + c.PlasmidCost()
}
