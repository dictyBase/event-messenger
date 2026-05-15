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

const invSize = 5

type Annotation struct {
	Client annotation.TaggedAnnotationServiceClient
}

func (an *Annotation) GetBasicStrainInfo(strains []*stock.Strain) ([][]string, error) {
	var allStrains [][]string

	for _, st := range strains {
		stdata, err := an.strainData(st)
		if err != nil {
			return allStrains, err
		}

		allStrains = append(allStrains, stdata)
	}

	return allStrains, nil
}

func (an *Annotation) GetStrainInfo(strains []*stock.Strain) ([][]string, error) {
	var allStrains [][]string

	for _, st := range strains {
		stChars, err := an.getAnnotations(
			fmt.Sprintf(
				"entry_id===%s;ontology===%s",
				st.GetData().GetId(), "strain_characteristics",
			))
		if err != nil {
			return allStrains,
				fmt.Errorf(
					"error in getting strain characteristics for strain %s %s",
					st.GetData().GetId(),
					err,
				)
		}

		stdata, err := an.strainData(st)
		if err != nil {
			return allStrains, err
		}

		stdata = append(stdata, strings.Join(an.annoColl2Tags(stChars), "<br/>"))
		allStrains = append(allStrains, stdata)
	}

	return allStrains, nil
}

func (an *Annotation) GetStrainInv(strains []*stock.Strain) ([][]string, error) {
	var allInv [][]string

	for _, st := range strains {
		gc, err := an.Client.ListAnnotationGroups(
			context.Background(),
			&annotation.ListGroupParameters{
				Filter: fmt.Sprintf(
					"entry_id===%s;tag===%s;ontology===%s",
					st.GetData().GetId(), registry.InvLocationTag, registry.StrainInvOnto,
				),
			},
		)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				return allInv, nil
			}

			return allInv, err
		}

		for _, gcd := range gc.GetData() {
			inv := make([]string, invSize)
			for _, gd := range gcd.GetGroup().GetData() {
				inv[0] = st.GetData().GetAttributes().GetLabel()
				switch gd.GetAttributes().GetTag() {
				case registry.InvStoredAsTag:
					inv[1] = gd.GetAttributes().GetValue()
				case registry.InvLocationTag:
					inv[2] = gd.GetAttributes().GetValue()
				case registry.InvVialCountTag:
					inv[3] = gd.GetAttributes().GetValue()
				case registry.InvVialColorTag:
					inv[4] = gd.GetAttributes().GetValue()
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
					pls.GetData().GetId(), registry.InvLocationTag, registry.PlasmidInvOntO,
				),
			},
		)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				return allInv, nil
			}

			return allInv, err
		}

		for _, gcd := range gc.GetData() {
			inv := make([]string, invSize)
			for _, gd := range gcd.GetGroup().GetData() {
				inv[0] = pls.GetData().GetId()

				inv[1] = pls.GetData().GetAttributes().GetName()
				switch gd.GetAttributes().GetTag() {
				case registry.InvStoredAsTag:
					inv[2] = gd.GetAttributes().GetValue()
				case registry.InvLocationTag:
					inv[3] = gd.GetAttributes().GetValue()
				case registry.InvVialColorTag:
					inv[4] = gd.GetAttributes().GetValue()
				}
			}

			allInv = append(allInv, inv)
		}
	}

	return allInv, nil
}

func (an *Annotation) strainData(st *stock.Strain) ([]string, error) {
	var stdata []string

	sysName, err := an.getSysName(st.GetData().GetId())
	if err != nil {
		return stdata,
			fmt.Errorf(
				"error in getting systematic name for strain %s %s",
				st.GetData().GetId(),
				err,
			)
	}

	stNames, err := an.getAnnotations(
		fmt.Sprintf(
			"entry_id===%s;tag===%s;ontology===%s",
			st.GetData().GetId(), registry.SynTag, registry.DictyAnnoOntology,
		))
	if err != nil {
		return stdata,
			fmt.Errorf("error in getting strain names for strain %s %s", st.GetData().GetId(), err)
	}

	return []string{
		st.GetData().GetId(),
		st.GetData().GetAttributes().GetLabel(),
		strings.Join(an.annoColl2Value(stNames), "<br/>"),
		sysName,
	}, nil
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

	for _, tad := range tac.GetData() {
		tags = append(tags, tad.GetAttributes().GetTag())
	}

	return tags
}

func (an *Annotation) annoColl2Value(tac *annotation.TaggedAnnotationCollection) []string {
	var values []string
	if tac == nil {
		return values
	}

	for _, tad := range tac.GetData() {
		values = append(values, tad.GetAttributes().GetValue())
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

	return ta.GetData().GetAttributes().GetValue(), nil
}
