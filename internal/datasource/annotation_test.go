package datasource

import (
	"testing"

	"github.com/dictyBase/event-messenger/internal/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
	plasmids, err := stock.GetPlasmids(fake.PlasmidIds())
	assert.NoError(err, "expect no error from getting plasmids")
	ann := &Annotation{Client: mockedAnnoPlasmidClient()}
	invList, err := ann.GetPlasmidInv(plasmids)
	assert.NoError(err, "expect no error from getting strains")
	assert.Len(invList, 12, "should match no of groups in collection")
	for _, inv := range invList {
		assert.Len(inv, 5, "should have 5 entries for each inventory")
		assert.Exactly(inv[0], "DBP0000120", "should match the plasmid id")
		assert.Exactly(inv[1], "p123456", "should match plasmid name")
		assert.Exactly(inv[2], "DNA", "should match how plasmid is stored")
		assert.Exactly(inv[3], "17(21-22)", "should match storage location of plasmid")
		assert.Exactly(inv[4], "red", "should match color of vials")
	}
}

func TestGetStrainInv(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	stock := &Stock{Client: mockedStockClient()}
	strains, err := stock.GetStrains(fake.StrainIds())
	assert.NoError(err, "expect no error from getting strains")
	ann := &Annotation{Client: mockedAnnoClient()}
	invList, err := ann.GetStrainInv(strains)
	assert.NoError(err, "expect no error from getting strains")
	assert.Len(invList, 16, "should match no of groups in collection")
	for _, inv := range invList {
		assert.Len(inv, 5, "should have 5 entries for each inventory")
		assert.Exactly(inv[0], "yS13", "should match the strain lab4el")
		assert.Exactly(inv[1], "axenic cells", "should match how strain is stored")
		assert.Exactly(inv[2], "2-9(55-57)", "should match storage location of strain")
		assert.Exactly(inv[3], "9", "should match no of vials")
		assert.Exactly(inv[4], "blue", "should match the color of storage vial")
	}
}

func TestGetsysName(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	ann := &Annotation{Client: mockedAnnoClient()}
	name, err := ann.getSysName("DBS0236926")
	assert.NoError(err, "expect no error from getting systematic name")
	assert.Exactly(name, "DBS0236922", "should match systematic name")
}
