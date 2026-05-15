package datasource

import (
	"testing"

	"github.com/dictyBase/event-messenger/internal/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func mockedAnnoPlasmidClient() *TaggedAnnotationServiceClient {
	mockedAnnoClient := new(TaggedAnnotationServiceClient)
	mockedAnnoClient.On(
		"ListAnnotationGroups",
		mock.Anything,
		mock.AnythingOfType("*annotation.ListGroupParameters"),
	).Return(fake.PlasmidInvAnno(), nil)

	return mockedAnnoClient
}

func mockedAnnoClient() *TaggedAnnotationServiceClient {
	mockedAnnoClient := new(TaggedAnnotationServiceClient)
	mockedAnnoClient.On(
		"GetEntryAnnotation",
		mock.Anything,
		mock.AnythingOfType("*annotation.EntryAnnotationRequest"),
	).Return(fake.SysNameAnno(), nil).
		On(
			"ListAnnotationGroups",
			mock.Anything,
			mock.AnythingOfType("*annotation.ListGroupParameters"),
		).Return(fake.StrainInvAnno(), nil)

	return mockedAnnoClient
}

func TestGetPlasmidInv(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	stock := &Stock{Client: mockedStockClient()}
	plasmids, err := stock.GetPlasmids(fake.PlasmidIDs())
	require.NoError(t, err, "expect no error from getting plasmids")

	ann := &Annotation{Client: mockedAnnoPlasmidClient()}
	invList, err := ann.GetPlasmidInv(plasmids)
	require.NoError(t, err, "expect no error from getting strains")
	assert.Len(invList, 12, "should match no of groups in collection")

	for _, inv := range invList {
		assert.Len(inv, 5, "should have 5 entries for each inventory")
		assert.Exactly("DBP0000120", inv[0], "should match the plasmid id")
		assert.Exactly("p123456", inv[1], "should match plasmid name")
		assert.Exactly("DNA", inv[2], "should match how plasmid is stored")
		assert.Exactly("17(21-22)", inv[3], "should match storage location of plasmid")
		assert.Exactly("red", inv[4], "should match color of vials")
	}
}

func TestGetStrainInv(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	stock := &Stock{Client: mockedStockClient()}
	strains, err := stock.GetStrains(fake.StrainIDs())
	require.NoError(t, err, "expect no error from getting strains")

	ann := &Annotation{Client: mockedAnnoClient()}
	invList, err := ann.GetStrainInv(strains)
	require.NoError(t, err, "expect no error from getting strains")
	assert.Len(invList, 16, "should match no of groups in collection")

	for _, inv := range invList {
		assert.Len(inv, 5, "should have 5 entries for each inventory")
		assert.Exactly("yS13", inv[0], "should match the strain lab4el")
		assert.Exactly("axenic cells", inv[1], "should match how strain is stored")
		assert.Exactly("2-9(55-57)", inv[2], "should match storage location of strain")
		assert.Exactly("9", inv[3], "should match no of vials")
		assert.Exactly("blue", inv[4], "should match the color of storage vial")
	}
}

func TestGetsysName(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	ann := &Annotation{Client: mockedAnnoClient()}
	name, err := ann.getSysName("DBS0236926")
	require.NoError(t, err, "expect no error from getting systematic name")
	assert.Exactly("DBS0236922", name, "should match systematic name")
}
