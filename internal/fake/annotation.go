package fake

import (
	"github.com/dictyBase/event-messenger/internal/registry"
	"github.com/dictyBase/go-genproto/dictybaseapis/annotation"
	"github.com/golang/protobuf/ptypes"
)

func SysNameAnno() *annotation.TaggedAnnotation {
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
