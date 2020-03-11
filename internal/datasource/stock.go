package datasource

import (
	"context"
	"strings"

	"github.com/dictyBase/go-genproto/dictybaseapis/order"
	"github.com/dictyBase/go-genproto/dictybaseapis/stock"
)

type Stock struct {
	Client stock.StockServiceClient
}

func (st *Stock) StocksFromItems(ord *order.Order, pattern string) []string {
	var stocks []string
	for _, item := range ord.Data.Attributes.Items {
		if strings.Contains(item, pattern) {
			stocks = append(stocks, item)
		}
	}
	return stocks
}

func (st *Stock) GetStrains(ids []string) ([]*stock.Strain, error) {
	var strains []*stock.Strain
	for _, id := range ids {
		str, err := st.Client.GetStrain(context.Background(), &stock.StockId{Id: id})
		if err != nil {
			return strains, err
		}
		strains = append(strains, str)
	}
	return strains, nil
}

func (st *Stock) GetPlasmids(ids []string) ([]*stock.Plasmid, error) {
	var plasmids []*stock.Plasmid
	for _, id := range ids {
		str, err := st.Client.GetPlasmid(context.Background(), &stock.StockId{Id: id})
		if err != nil {
			return plasmids, err
		}
		plasmids = append(plasmids, str)
	}
	return plasmids, nil
}
