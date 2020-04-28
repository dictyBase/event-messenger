package datasource

import (
	"testing"

	"github.com/dictyBase/event-messenger/internal/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

func TestGetsysName(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	ann := &Annotation{Client: mockedAnnoClient()}
	name, err := ann.getSysName("DBS0236926")
	assert.NoError(err, "expect no error from getting systematic name")
	assert.Exactly(name, "DBS0236922", "should match systematic name")
}
