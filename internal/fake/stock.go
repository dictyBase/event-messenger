package fake

import (
	"github.com/dictyBase/go-genproto/dictybaseapis/stock"
	"github.com/golang/protobuf/ptypes"
)

const (
	StrainId  = "DBS0235559"
	PlasmidId = "DBP0000120"
)

func StrainIds() []string {
	return []string{
		"DBS0236414",
		"DBS0236440",
		"DBS0236245",
		"DBS0235559",
	}
}

func PlasmidIds() []string {
	return []string{
		"DBP0000105",
		"DBP0000034",
		"DBP0000120",
	}
}

func Strain() *stock.Strain {
	return &stock.Strain{
		Data: &stock.Strain_Data{
			Type: "strain",
			Id:   StrainId,
			Attributes: &stock.StrainAttributes{
				CreatedAt:       ptypes.TimestampNow(),
				UpdatedAt:       ptypes.TimestampNow(),
				CreatedBy:       Consumer,
				UpdatedBy:       Consumer,
				Depositor:       Depositor,
				Summary:         "Radiation-sensitive mutant.",
				EditableSummary: "Radiation-sensitive mutant.",
				Dbxrefs:         []string{"5466867", "4536935", "d2578", "d0319", "d2020/1033268", "d2580"},
				Genes:           []string{"DDB_G0348394", "DDB_G098058933"},
				Publications:    []string{"4849343943", "48394394"},
				Label:           "yS13",
				Species:         "Dictyostelium discoideum",
				Plasmid:         "DBP0000027",
				Names:           []string{"gammaS13", "gammaS-13", "γS-13"},
			},
		},
	}

}

func Plasmid() *stock.Plasmid {
	return &stock.Plasmid{
		Data: &stock.Plasmid_Data{
			Type: "plasmid",
			Id:   PlasmidId,
			Attributes: &stock.PlasmidAttributes{
				CreatedBy:       Consumer,
				UpdatedBy:       Consumer,
				Depositor:       Depositor,
				Summary:         "update this plasmid",
				EditableSummary: "update this plasmid",
				Publications:    []string{"1348970", "48493483"},
				Dbxrefs:         []string{"5466867", "4536935", "d2578"},
				Genes:           []string{"DDB_G0348394", "DDB_G098058933"},
				ImageMap:        "http://dictybase.org/data/plasmid/images/87.jpg",
				Sequence:        "tttttyyyyjkausadaaaavvvvvv",
				Name:            "p123456",
			},
		},
	}
}
