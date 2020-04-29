package fake

import (
	"github.com/dictyBase/event-messenger/internal/registry"
	"github.com/dictyBase/go-genproto/dictybaseapis/annotation"
	"github.com/golang/protobuf/ptypes"
)

type groupDataFn func() []*annotation.TaggedAnnotationGroup_Data

type propValue struct {
	Tag   string
	Value string
}

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

func StrainInvAnno() *annotation.TaggedAnnotationGroupCollection {
	return &annotation.TaggedAnnotationGroupCollection{
		Data: stockGroupCollData(strainGroupData),
	}
}

func PlasmidInvAnno() *annotation.TaggedAnnotationGroupCollection {
	return &annotation.TaggedAnnotationGroupCollection{
		Data: stockGroupCollData(plasmidGroupData),
	}
}

func stockGroupCollData(gfn groupDataFn) []*annotation.TaggedAnnotationGroupCollection_Data {
	var gcd []*annotation.TaggedAnnotationGroupCollection_Data
	for i := 0; i <= 3; i++ {
		gcd = append(gcd, &annotation.TaggedAnnotationGroupCollection_Data{
			Type: "annotation_group",
			Group: &annotation.TaggedAnnotationGroup{
				GroupId:   "4924132",
				CreatedAt: ptypes.TimestampNow(),
				UpdatedAt: ptypes.TimestampNow(),
				Data:      gfn(),
			},
		})
	}
	return gcd
}

func groupData(all []propValue, id string) []*annotation.TaggedAnnotationGroup_Data {
	var gd []*annotation.TaggedAnnotationGroup_Data
	for _, a := range all {
		gd = append(gd, &annotation.TaggedAnnotationGroup_Data{
			Type: "annotation",
			Id:   "489483843",
			Attributes: &annotation.TaggedAnnotationAttributes{
				Version:   1,
				EntryId:   id,
				CreatedBy: Consumer,
				CreatedAt: ptypes.TimestampNow(),
				Ontology:  registry.DictyAnnoOntology,
				Tag:       a.Tag,
				Value:     a.Value,
			},
		})
	}
	return gd
}

func strainGroupData() []*annotation.TaggedAnnotationGroup_Data {
	allData := []propValue{
		{registry.InvStoredAsTag, "axenic cells"},
		{registry.InvLocationTag, "2-9(55-57)"},
		{registry.InvVialCountTag, "9"},
		{registry.InvVialColorTag, "blue"},
	}
	return groupData(allData, StrainID)
}

func plasmidGroupData() []*annotation.TaggedAnnotationGroup_Data {
	allData := []propValue{
		{registry.InvStoredAsTag, "DNA"},
		{registry.InvLocationTag, "17(21-22)"},
		{registry.InvVialColorTag, "red"},
	}
	return groupData(allData, PlasmidID)
}
