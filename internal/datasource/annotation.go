package datasource

import (
	"context"
	"fmt"
	"strings"

	"github.com/dictyBase/event-messenger/internal/registry"
	"github.com/dictyBase/go-genproto/dictybaseapis/annotation"
	"github.com/dictyBase/go-genproto/dictybaseapis/stock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Annotation struct {
	Client annotation.TaggedAnnotationServiceClient
}

func (an *Annotation) GetBasicStrainInfo(strains []*stock.Strain) ([][]string, error) {
	var allStrains [][]string
	for _, st := range strains {
		sysName, err := an.getSysName(st.Data.Id)
		if err != nil {
			return allStrains,
				fmt.Errorf("error in getting systematic name for strain %s %s", st.Data.Id, err)
		}
		stNames, err := an.getAnnotations(
			fmt.Sprintf(
				"entry_id===%s;tag===%s;ontology===%s",
				st.Data.Id, registry.SynTag, registry.DictyAnnoOntology,
			))
		if err != nil {
			return allStrains,
				fmt.Errorf("error in getting strain names for strain %s %s", st.Data.Id, err)
		}
		allStrains = append(allStrains, []string{
			st.Data.Id,
			st.Data.Attributes.Label,
			strings.Join(an.annoColl2Value(stNames), "<br/>"),
			sysName,
		})
	}
	return allStrains, nil
}

func (an *Annotation) GetStrainInfo(strains []*stock.Strain) ([][]string, error) {
	var allStrains [][]string
	for _, st := range strains {
		sysName, err := an.getSysName(st.Data.Id)
		if err != nil {
			return allStrains,
				fmt.Errorf("error in getting systematic name for strain %s %s", st.Data.Id, err)
		}
		stChars, err := an.getAnnotations(
			fmt.Sprintf(
				"entry_id===%s;ontology===%s",
				st.Data.Id, "strain_characteristics",
			))
		if err != nil {
			return allStrains,
				fmt.Errorf("error in getting strain characteristics for strain %s %s", st.Data.Id, err)
		}
		stNames, err := an.getAnnotations(
			fmt.Sprintf(
				"entry_id===%s;tag===%s;ontology===%s",
				st.Data.Id, registry.SynTag, registry.DictyAnnoOntology,
			))
		if err != nil {
			return allStrains,
				fmt.Errorf("error in getting strain names for strain %s %s", st.Data.Id, err)
		}
		allStrains = append(allStrains, []string{
			st.Data.Id,
			st.Data.Attributes.Label,
			strings.Join(an.annoColl2Value(stNames), "<br/>"),
			sysName,
			strings.Join(an.annoColl2Tags(stChars), "<br/>"),
		})
	}
	return allStrains, nil
}

func (an *Annotation) getAnnotations(query string) (*annotation.TaggedAnnotationCollection, error) {
	tac, err := an.Client.ListAnnotations(
		context.Background(),
		&annotation.ListParameters{Filter: query},
	)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return tac, nil
		}
		return tac, err
	}
	return tac, nil
}

func (an *Annotation) annoColl2Tags(tac *annotation.TaggedAnnotationCollection) []string {
	var tags []string
	if tac == nil {
		return tags
	}
	for _, tad := range tac.Data {
		tags = append(tags, tad.Attributes.Tag)
	}
	return tags
}

func (an *Annotation) annoColl2Value(tac *annotation.TaggedAnnotationCollection) []string {
	var values []string
	if tac == nil {
		return values
	}
	for _, tad := range tac.Data {
		values = append(values, tad.Attributes.Value)
	}
	return values
}

func (an *Annotation) getSysName(id string) (string, error) {
	ta, err := an.Client.GetEntryAnnotation(
		context.Background(),
		&annotation.EntryAnnotationRequest{
			Tag:      registry.SysnameTag,
			Ontology: registry.DictyAnnoOntology,
			EntryId:  id,
		})
	if err != nil {
		return "", err
	}
	return ta.Data.Attributes.Value, nil
}

func (an *Annotation) GetStrainInv(strains []*stock.Strain) ([][]string, error) {
	var allInv [][]string
	for _, st := range strains {
		gc, err := an.Client.ListAnnotationGroups(
			context.Background(),
			&annotation.ListGroupParameters{
				Filter: fmt.Sprintf(
					"entry_id===%s;tag===%s;ontology===%s",
					st.Data.Id, registry.InvLocationTag, registry.StrainInvOnto,
				)},
		)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				return allInv, nil
			}
			return allInv, err
		}
		for _, gcd := range gc.Data {
			inv := make([]string, 5)
			for _, gd := range gcd.Group.Data {
				inv[0] = st.Data.Attributes.Label
				switch gd.Attributes.Tag {
				case registry.InvStoredAsTag:
					inv[1] = gd.Attributes.Value
				case registry.InvLocationTag:
					inv[2] = gd.Attributes.Value
				case registry.InvVialCountTag:
					inv[3] = gd.Attributes.Value
				case registry.InvVialColorTag:
					inv[4] = gd.Attributes.Value
				}
			}
			allInv = append(allInv, inv)
		}
	}
	return allInv, nil
}

func (an *Annotation) GetPlasmidInv(plasmids []*stock.Plasmid) ([][]string, error) {
	var allInv [][]string
	for _, pls := range plasmids {
		gc, err := an.Client.ListAnnotationGroups(
			context.Background(),
			&annotation.ListGroupParameters{
				Filter: fmt.Sprintf(
					"entry_id===%s;tag===%s;ontology===%s",
					pls.Data.Id, registry.InvLocationTag, registry.PlasmidInvOntO,
				)},
		)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				return allInv, nil
			}
			return allInv, err
		}
		for _, gcd := range gc.Data {
			inv := make([]string, 5)
			for _, gd := range gcd.Group.Data {
				inv[0] = pls.Data.Id
				inv[1] = pls.Data.Attributes.Name
				switch gd.Attributes.Tag {
				case registry.InvStoredAsTag:
					inv[2] = gd.Attributes.Value
				case registry.InvLocationTag:
					inv[3] = gd.Attributes.Value
				case registry.InvVialColorTag:
					inv[4] = gd.Attributes.Value
				}
			}
			allInv = append(allInv, inv)
		}
	}
	return allInv, nil
}
