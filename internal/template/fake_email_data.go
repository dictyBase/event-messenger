package template

import (
	"github.com/dictyBase/event-messenger/internal/datasource"
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
	"github.com/dictyBase/go-genproto/dictybaseapis/user"
	"github.com/golang/protobuf/ptypes"
)

const (
	consumer = "pennpacker@dictycr.org"
	payer    = "varnsen@dictycr.org"
)

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
			Id:   "3894333",
			Attributes: &order.OrderAttributes{
				Payer:          payer,
				Consumer:       consumer,
				Payment:        "Credit card",
				Comments:       "Power",
				CourierAccount: "48934393",
				Courier:        "Fedex",
				CreatedAt:      ptypes.TimestampNow(),
				Items:          fakeStockItems(),
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
				CreatedAt:    ptypes.TimestampNow(),
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
				CreatedAt:    ptypes.TimestampNow(),
			},
		},
	}
}

func fakePub() []*datasource.PubInfo {
	return []*datasource.PubInfo{
		&datasource.PubInfo{
			AuthorStr: "Basu S et al. (2015)",
			PubmedURL: "https://pubmed.gov/26088819",
			DoiURL:    "https://doi.org/10.1002/dvg.22867",
		},
		&datasource.PubInfo{
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

func fakeStrainOnlyEmailContent() *EmailContent {
	return &EmailContent{
		Content: &Content{
			StrainPrice:  10,
			PlasmidPrice: 20,
			Order:        fakeOrder(),
			Shipper:      fakeConsumer(),
			Payer:        fakePayer(),
		},
		StrainData: fakeStrain(),
	}
}
