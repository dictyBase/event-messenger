package template

import (
	"github.com/dictyBase/event-messenger/internal/datasource"
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
	"github.com/dictyBase/go-genproto/dictybaseapis/user"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	consumer = "pennpacker@dictycr.org"
	payer    = "varnsen@dictycr.org"
	orderID  = "3894333"
)

type templateData struct {
	Header  string
	Strains []struct {
		ID   string
		Name string
	}
}

func fakeStrainItems() []string {
	return []string{
		"DBS0236414",
		"DBS0236440",
		"DBS0236245",
		"DBS0235559",
	}
}

func fakePlasmidItems() []string {
	return []string{
		"DBP0000105",
		"DBP0000034",
		"DBP0000120",
	}
}

func fakeStockItems() []string {
	return append(fakeStrainItems(), fakePlasmidItems()...)
}

func fakeOrder() *order.Order {
	return &order.Order{
		Data: &order.Order_Data{
			Type: "stocks",
			Id:   orderID,
			Attributes: &order.OrderAttributes{
				Payer:            payer,
				Consumer:         consumer,
				Payment:          "Credit card",
				PurchaseOrderNum: "N/A",
				Comments:         "Power",
				CourierAccount:   "48934393",
				Courier:          "Fedex",
				CreatedAt:        timestamppb.Now(),
				Items:            fakeStockItems(),
			},
		},
	}
}

func fakePayer() *user.User {
	return &user.User{
		Data: &user.UserData{
			Type: "users",
			Id:   8448393,
			Attributes: &user.UserAttributes{
				FirstName:    "Kel",
				LastName:     "Varnsen",
				Email:        payer,
				Organization: "Cornell University",
				FirstAddress: "5 West 63 street",
				City:         "New York",
				State:        "NY",
				Zipcode:      "100009",
				Country:      "US",
				Phone:        "212-555-0109",
				CreatedAt:    timestamppb.Now(),
			},
		},
	}
}

func fakeConsumer() *user.User {
	return &user.User{
		Data: &user.UserData{
			Type: "users",
			Id:   8493438,
			Attributes: &user.UserAttributes{
				FirstName:    "Harrold",
				LastName:     "Pennypacker",
				Email:        consumer,
				Organization: "New York University",
				FirstAddress: "129 West 81 street",
				City:         "New York",
				State:        "NY",
				Zipcode:      "100001",
				Country:      "US",
				Phone:        "212-555-0171",
				CreatedAt:    timestamppb.Now(),
			},
		},
	}
}

func fakePub() []*datasource.PubInfo {
	return []*datasource.PubInfo{
		{
			AuthorStr: "Basu S et al. (2015)",
			PubmedURL: "https://pubmed.gov/26088819",
			DoiURL:    "https://doi.org/10.1002/dvg.22867",
		},
		{
			AuthorStr: "Tweedy L & Insall RH (2020)",
			PubmedURL: "https://pubmed.gov/32195256",
			DoiURL:    "https://doi.org/10.3389/fcell.2020.00133",
		},
	}
}

func fakeStrain() []*StrainRows {
	var rows []*StrainRows
	for _, s := range fakeStrainItems() {
		rows = append(rows,
			&StrainRows{
				ID:         s,
				SysName:    "JB10",
				Names:      "jb10ale<br/>jb10 ale<br/>jb10-ale",
				Descriptor: "gefA-",
				PubInfo:    fakePub(),
			},
		)
	}
	return rows
}

func fakePlasmid() []*PlasmidRows {
	var rows []*PlasmidRows
	for _, p := range fakePlasmidItems() {
		rows = append(rows,
			&PlasmidRows{
				ID:      p,
				PubInfo: fakePub(),
				Name:    "pDV-fAR1-CYFP",
			},
		)
	}
	return rows
}

func fakePlasmidInv() [][]string {
	var rows [][]string
	for _, p := range fakePlasmidItems() {
		rows = append(rows, []string{
			p,
			"pDV-fAR1-CYFP",
			"DH5α",
			"12(45,54)",
			"blue",
		})
	}
	return rows
}

func fakeStrainInv() [][]string {
	var rows [][]string
	for i := 0; i < len(fakeStrainItems()); i++ {
		rows = append(rows, []string{
			"talA-",
			"axenic cells",
			"1-34(76-78)",
			"9",
			"pink",
		})
	}
	return rows
}

func fakeStrainInfo() [][]string {
	var rows [][]string
	for _, s := range fakeStrainItems() {
		rows = append(rows, []string{
			s,
			"talA-",
			"talin-null talA-null",
			"HG1666",
			"blasticidin resistant<br/>neomycin resistant",
		})
	}
	return rows
}

func fakeContent() *Content {
	return &Content{
		Order:   fakeOrder(),
		Shipper: fakeConsumer(),
		Payer:   fakePayer(),
	}
}

func fakePlasmidOnlyEmailContent() *EmailContent {
	c := fakeContent()
	c.PlasmidPrice = 10
	return &EmailContent{
		PlasmidData: fakePlasmid(),
		Content:     c,
	}
}

func fakePlasmidOnlyIssueContent() *IssueContent {
	c := fakeContent()
	c.PlasmidPrice = 10
	return &IssueContent{
		Content:    c,
		PlasmidInv: fakePlasmidInv(),
	}
}

func fakeStrainOnlyIssueContent() *IssueContent {
	c := fakeContent()
	c.StrainPrice = 10
	return &IssueContent{
		Content:    c,
		StrainInfo: fakeStrainInfo(),
		StrainInv:  fakeStrainInv(),
	}
}

func fakeStockIssueContent() *IssueContent {
	c := fakeContent()
	c.StrainPrice = 10
	c.PlasmidPrice = 10
	return &IssueContent{
		Content:    c,
		StrainInfo: fakeStrainInfo(),
		StrainInv:  fakeStrainInv(),
		PlasmidInv: fakePlasmidInv(),
	}
}

func fakeStrainOnlyEmailContent() *EmailContent {
	c := fakeContent()
	c.StrainPrice = 10
	return &EmailContent{
		StrainData: fakeStrain(),
		Content:    c,
	}
}

func fakeTemplateData() templateData {
	st := []struct {
		ID   string
		Name string
	}{
		{"DBS0236831", "tori"},
		{"DBS0236415", "lori"},
	}
	return templateData{
		Header:  "Stock information",
		Strains: st,
	}
}
