package datasource

import (
	"testing"

	"github.com/dictyBase/event-messenger/internal/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func mockedStockClient() *StockServiceClient {
	client := new(StockServiceClient)
	client.On(
		"GetStrain",
		mock.Anything,
		mock.AnythingOfType("*stock.StockId"),
	).Return(fake.Strain(), nil).
		On(
			"GetPlasmid",
			mock.Anything,
			mock.AnythingOfType("*stock.StockId"),
		).Return(fake.Plasmid(), nil)
	return client
}

func TestGetStrains(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	ids := fake.StrainIds()
	stock := &Stock{Client: mockedStockClient()}
	strains, err := stock.GetStrains(ids)
	assert.NoError(err, "expect no error from getting strains")
	assert.Lenf(
		strains, len(ids),
		"expect %d received %d strains",
		len(ids), len(strains),
	)
	for _, st := range strains {
		assert.Exactly(st.Data.Id, fake.StrainId, "should match the strain id")
		assert.Exactly(st.Data.Attributes.CreatedBy, fake.Consumer, "should match creator of the record")
		assert.Exactly(st.Data.Attributes.Depositor, fake.Depositor, "should match depositor of the record")
		assert.Exactly(
			st.Data.Attributes.Summary,
			"Radiation-sensitive mutant.",
			"should match creator of the record",
		)
		assert.ElementsMatch(
			st.Data.Attributes.Genes,
			[]string{"DDB_G0348394", "DDB_G098058933"},
			"should match list of genes",
		)
	}
}
