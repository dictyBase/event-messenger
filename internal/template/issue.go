package template

import "github.com/dictyBase/go-genproto/dictybaseapis/stock"

type IssueContent struct {
	Strains    []*stock.Strain
	Plasmids   []*stock.Plasmid
	StrainInv  [][]string
	PlasmidInv [][]string
	StrainInfo [][]string
	*Content
}
