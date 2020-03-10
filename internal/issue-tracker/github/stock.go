package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/dictyBase/go-genproto/dictybaseapis/annotation"
	"github.com/dictyBase/go-genproto/dictybaseapis/stock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type annoParams struct {
	query   string
	aclient annotation.TaggedAnnotationServiceClient
}

func getStrainInfo(strains []*stock.Strain, aclient annotation.TaggedAnnotationServiceClient) ([][]string, error) {
	var allStrains [][]string
	for _, st := range strains {
		sysName, err := getSysName(st.Data.Id, aclient)
		if err != nil {
			return allStrains,
				fmt.Errorf("error in getting systematic name for strain %s %s", st.Data.Id, err)
		}
		stChars, err := getAnnotations(&annoParams{
			aclient: aclient,
			query: fmt.Sprintf(
				"entry_id===%s;ontology===%s",
				st.Data.Id, "strain_characteristics",
			)})
		if err != nil {
			return allStrains,
				fmt.Errorf("error in getting strain characteristics for strain %s %s", st.Data.Id, err)
		}
		stNames, err := getAnnotations(&annoParams{
			aclient: aclient,
			query: fmt.Sprintf(
				"entry_id===%s;tag===%s;ontology===%s",
				st.Data.Id, SynTag, DictyAnnoOntology,
			)})
		if err != nil {
			return allStrains,
				fmt.Errorf("error in getting strain names for strain %s %s", st.Data.Id, err)
		}
		allStrains = append(allStrains, []string{
			st.Data.Id,
			st.Data.Attributes.Label,
			strings.Join(annoColl2Value(stNames), "<br/>"),
			sysName,
			strings.Join(annoColl2Tags(stChars), "<br/>"),
		})
	}
	return allStrains, nil
}

func getStrains(ids []string, sclient stock.StockServiceClient) ([]*stock.Strain, error) {
	var strains []*stock.Strain
	for _, id := range ids {
		str, err := sclient.GetStrain(context.Background(), &stock.StockId{Id: id})
		if err != nil {
			return strains, err
		}
		strains = append(strains, str)
	}
	return strains, nil
}

func getPlasmids(ids []string, sclient stock.StockServiceClient) ([]*stock.Plasmid, error) {
	var plasmids []*stock.Plasmid
	for _, id := range ids {
		str, err := sclient.GetPlasmid(context.Background(), &stock.StockId{Id: id})
		if err != nil {
			return plasmids, err
		}
		plasmids = append(plasmids, str)
	}
	return plasmids, nil
}

func getAnnotations(args *annoParams) (*annotation.TaggedAnnotationCollection, error) {
	tac, err := args.aclient.ListAnnotations(
		context.Background(),
		&annotation.ListParameters{Filter: args.query},
	)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return tac, nil
		}
		return tac, err
	}
	return tac, nil
}

func annoColl2Tags(tac *annotation.TaggedAnnotationCollection) []string {
	var tags []string
	if tac == nil {
		return tags
	}
	for _, tad := range tac.Data {
		tags = append(tags, tad.Attributes.Tag)
	}
	return tags
}

func annoColl2Value(tac *annotation.TaggedAnnotationCollection) []string {
	var values []string
	if tac == nil {
		return values
	}
	for _, tad := range tac.Data {
		values = append(values, tad.Attributes.Value)
	}
	return values
}

func getSysName(id string, aclient annotation.TaggedAnnotationServiceClient) (string, error) {
	ta, err := aclient.GetEntryAnnotation(
		context.Background(),
		&annotation.EntryAnnotationRequest{
			Tag:      SysnameTag,
			Ontology: DictyAnnoOntology,
			EntryId:  id,
		})
	if err != nil {
		return "", err
	}
	return ta.Data.Attributes.Value, nil
}

func getStrainInv(strains []*stock.Strain, aclient annotation.TaggedAnnotationServiceClient) ([][]string, error) {
	var allInv [][]string
	for _, st := range strains {
		gc, err := aclient.ListAnnotationGroups(
			context.Background(),
			&annotation.ListGroupParameters{
				Filter: fmt.Sprintf(
					"entry_id===%s;tag===%s;ontology===%s",
					st.Data.Id, InvLocationTag, StrainInvOnto,
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
				case InvStoredAsTag:
					inv[1] = gd.Attributes.Value
				case InvLocationTag:
					inv[2] = gd.Attributes.Value
				case InvVialCountTag:
					inv[3] = gd.Attributes.Value
				case InvVialColorTag:
					inv[4] = gd.Attributes.Value
				}
			}
			allInv = append(allInv, inv)
		}
	}
	return allInv, nil
}

func getPlasmidInv(plasmids []*stock.Plasmid, aclient annotation.TaggedAnnotationServiceClient) ([][]string, error) {
	var allInv [][]string
	for _, pls := range plasmids {
		gc, err := aclient.ListAnnotationGroups(
			context.Background(),
			&annotation.ListGroupParameters{
				Filter: fmt.Sprintf(
					"entry_id===%s;tag===%s;ontology===%s",
					pls.Data.Id, InvLocationTag, PlasmidInvOntO,
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
				case InvStoredAsTag:
					inv[2] = gd.Attributes.Value
				case InvLocationTag:
					inv[3] = gd.Attributes.Value
				case InvVialColorTag:
					inv[4] = gd.Attributes.Value
				}
			}
			allInv = append(allInv, inv)
		}
	}
	return allInv, nil
}
