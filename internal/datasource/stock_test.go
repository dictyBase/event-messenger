package datasource

import (
	"testing"

	"github.com/dictyBase/event-messenger/internal/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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
	ids := fake.StrainIDs()
	stock := &Stock{Client: mockedStockClient()}
	strains, err := stock.GetStrains(ids)
	require.NoError(t, err, "expect no error from getting strains")
	assert.Lenf(
		strains, len(ids),
		"expect %d received %d strains",
		len(ids), len(strains),
	)

	for _, st := range strains {
		assert.Exactly(fake.StrainID, st.GetData().GetId(), "should match the strain id")
		assert.Exactly(
			fake.Consumer,
			st.GetData().GetAttributes().GetCreatedBy(),
			"should match creator of the record",
		)
		assert.Exactly(
			fake.Depositor,
			st.GetData().GetAttributes().GetDepositor(),
			"should match depositor of the record",
		)
		assert.Exactly(
			"Radiation-sensitive mutant.", st.GetData().GetAttributes().GetSummary(),
			"should match creator of the record",
		)
		assert.ElementsMatch(
			st.GetData().GetAttributes().GetGenes(),
			[]string{"DDB_G0348394", "DDB_G098058933"},
			"should match list of genes",
		)
	}
}

func TestPlasmids(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	ids := fake.PlasmidIDs()
	stock := &Stock{Client: mockedStockClient()}
	plasmids, err := stock.GetPlasmids(ids)
	require.NoError(t, err, "expect no error from getting plasmids")
	assert.Lenf(
		plasmids, len(ids),
		"expect %d received %d plasmids",
		len(ids), len(plasmids),
	)

	for _, pl := range plasmids {
		assert.Exactly(
			fake.PlasmidID, pl.GetData().GetId(),
			"should match the plasmid id",
		)
		assert.Exactly(
			fake.Consumer, pl.GetData().GetAttributes().GetCreatedBy(),
			"should match creator of the record",
		)
		assert.Exactly(
			fake.Depositor, pl.GetData().GetAttributes().GetDepositor(),
			"should match depositor of the record",
		)
		assert.Exactly(
			"http://dictybase.org/data/plasmid/images/87.jpg",
			pl.GetData().GetAttributes().GetImageMap(),
			"should map the image map",
		)
		assert.Exactly(
			"p123456",
			pl.GetData().GetAttributes().GetName(),
			"should match the plasmid name",
		)
		assert.ElementsMatch(
			pl.GetData().GetAttributes().GetPublications(),
			[]string{"1348970", "48493483"},
			"should match the list of publications",
		)
	}
}
