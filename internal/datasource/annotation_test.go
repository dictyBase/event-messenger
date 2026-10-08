package datasource

import (
	"testing"

	"github.com/dictyBase/event-messenger/internal/fake"
	"github.com/dictyBase/go-genproto/dictybaseapis/annotation"
	"github.com/dictyBase/go-genproto/dictybaseapis/stock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

const notFoundMsg = "annotation id DBS0391345 not found"

func strainWithID(id string) *stock.Strain {
	st := fake.Strain()
	st.Data.Id = id

	return st
}

func plasmidWithID(id string) *stock.Plasmid {
	pl := fake.Plasmid()
	pl.Data.Id = id

	return pl
}

func TestGetSysNameToleratesNotFound(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	mockedAnnoClient := new(TaggedAnnotationServiceClient)
	mockedAnnoClient.On(
		"GetEntryAnnotation",
		mock.Anything,
		mock.AnythingOfType("*annotation.EntryAnnotationRequest"),
	).Return((*annotation.TaggedAnnotation)(nil), status.Error(codes.NotFound, notFoundMsg))

	ann := &Annotation{Client: mockedAnnoClient}
	name, err := ann.getSysName("DBS0391345")
	require.NoError(t, err, "expect no error for missing annotation")
	assert.Empty(name, "should return empty systematic name")
}

func TestGetSysNameReturnsOtherErrors(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	mockedAnnoClient := new(TaggedAnnotationServiceClient)
	mockedAnnoClient.On(
		"GetEntryAnnotation",
		mock.Anything,
		mock.AnythingOfType("*annotation.EntryAnnotationRequest"),
	).Return((*annotation.TaggedAnnotation)(nil), status.Error(codes.Internal, "annotation backend down"))

	ann := &Annotation{Client: mockedAnnoClient}
	name, err := ann.getSysName("DBS0391345")
	require.Error(t, err, "expect error from non-notfound grpc failure")
	assert.Exactly(
		codes.Internal, status.Code(err),
		"should keep the grpc status code",
	)
	assert.Empty(name, "should not return any systematic name")
}

func TestGetBasicStrainInfoToleratesMissingSysName(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	mockedAnnoClient := new(TaggedAnnotationServiceClient)
	mockedAnnoClient.On(
		"GetEntryAnnotation",
		mock.Anything,
		mock.AnythingOfType("*annotation.EntryAnnotationRequest"),
	).Return((*annotation.TaggedAnnotation)(nil), status.Error(codes.NotFound, notFoundMsg)).Once()
	mockedAnnoClient.On(
		"GetEntryAnnotation",
		mock.Anything,
		mock.AnythingOfType("*annotation.EntryAnnotationRequest"),
	).Return(fake.SysNameAnno(), nil)
	mockedAnnoClient.On(
		"ListAnnotations",
		mock.Anything,
		mock.AnythingOfType("*annotation.ListParameters"),
	).Return(&annotation.TaggedAnnotationCollection{}, nil)

	ann := &Annotation{Client: mockedAnnoClient}
	missingName := strainWithID("DBS0391345")
	missingName.Data.Attributes.Label = "noSysName"
	rows, err := ann.GetBasicStrainInfo([]*stock.Strain{
		missingName,
		strainWithID("DBS0236926"),
	})
	require.NoError(t, err, "expect no error when one annotation is missing")
	require.Len(t, rows, 2, "should keep both strain rows")
	assert.Exactly("DBS0391345", rows[0][0], "should match first strain id")
	assert.Exactly(
		"noSysName", rows[0][3],
		"should fall back to strain label for missing annotation",
	)
	assert.Exactly("DBS0236926", rows[1][0], "should match second strain id")
	assert.Exactly("DBS0236922", rows[1][3], "should match systematic name of second strain")
}

func TestGetStrainInvContinuesAfterNotFound(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	mockedAnnoClient := new(TaggedAnnotationServiceClient)
	mockedAnnoClient.On(
		"ListAnnotationGroups",
		mock.Anything,
		mock.AnythingOfType("*annotation.ListGroupParameters"),
	).Return(
		(*annotation.TaggedAnnotationGroupCollection)(nil),
		status.Error(codes.NotFound, notFoundMsg),
	).Once()
	mockedAnnoClient.On(
		"ListAnnotationGroups",
		mock.Anything,
		mock.AnythingOfType("*annotation.ListGroupParameters"),
	).Return(fake.StrainInvAnno(), nil)

	ann := &Annotation{Client: mockedAnnoClient}
	invList, err := ann.GetStrainInv([]*stock.Strain{
		strainWithID("DBS0391345"),
		strainWithID("DBS0236926"),
	})
	require.NoError(t, err, "expect no error when one annotation is missing")
	assert.Len(invList, 4, "should collect inventory rows from later strains")
}

func TestGetPlasmidInvContinuesAfterNotFound(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	mockedAnnoClient := new(TaggedAnnotationServiceClient)
	mockedAnnoClient.On(
		"ListAnnotationGroups",
		mock.Anything,
		mock.AnythingOfType("*annotation.ListGroupParameters"),
	).Return(
		(*annotation.TaggedAnnotationGroupCollection)(nil),
		status.Error(codes.NotFound, notFoundMsg),
	).Once()
	mockedAnnoClient.On(
		"ListAnnotationGroups",
		mock.Anything,
		mock.AnythingOfType("*annotation.ListGroupParameters"),
	).Return(fake.PlasmidInvAnno(), nil)

	ann := &Annotation{Client: mockedAnnoClient}
	invList, err := ann.GetPlasmidInv([]*stock.Plasmid{
		plasmidWithID("DBP0000001"),
		plasmidWithID(fake.PlasmidID),
	})
	require.NoError(t, err, "expect no error when one annotation is missing")
	assert.Len(invList, 4, "should collect inventory rows from later plasmids")
}
