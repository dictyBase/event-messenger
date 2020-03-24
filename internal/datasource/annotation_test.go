package datasource

import (
	"testing"

	"github.com/dictyBase/event-messenger/internal/registry"
	"github.com/dictyBase/go-genproto/dictybaseapis/annotation"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func mockSysNameAnno() *annotation.TaggedAnnotation {
	return &annotation.TaggedAnnotation{
		Data: &annotation.TaggedAnnotation_Data{
			Type: "annotation",
			Id:   "123456",
			Attributes: &annotation.TaggedAnnotationAttributes{
				Value:     "DBS0236922",
				EntryId:   "DBS0236922",
				CreatedBy: "dsc@dictycr.org",
				CreatedAt: ptypes.TimestampNow(),
				Tag:       registry.SysnameTag,
				Ontology:  registry.DictyAnnoOntology,
				Version:   1,
			},
		},
	}
}

func mockedClient() *TaggedAnnotationServiceClient {
	mockedAnnoClient := new(TaggedAnnotationServiceClient)
	mockedAnnoClient.On(
		"GetEntryAnnotation",
		mock.Anything,
		mock.AnythingOfType("*annotation.EntryAnnotationRequest"),
	).Return(mockSysNameAnno(), nil)
	return mockedAnnoClient
}

func TestGetsysName(t *testing.T) {
	assert := assert.New(t)
	ann := &Annotation{Client: mockedClient()}
	name, err := ann.getSysName("DBS0236926")
	assert.NoError(err, "expect no error from getting systematic name")
	assert.Exactly(name, "DBS0236922", "should match systematic name")
}
