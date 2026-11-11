package template

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"strings"
	txt "text/template"

	"github.com/dictyBase/event-messenger/internal/datasource"
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
)

//go:embed assets/issue.tmpl
var issueTemplate string

type OutputParams struct {
	Content any
}

type StrainRows struct {
	ID         string
	SysName    string
	Names      string
	Descriptor string
	PubInfo    []*datasource.PubInfo
}

type PlasmidRows struct {
	ID      string
	Name    string
	PubInfo []*datasource.PubInfo
}

type Content struct {
	Order        *order.Order
	Shipper      *order.UserInfo
	Payer        *order.UserInfo
	StrainPrice  int
	PlasmidPrice int
}

func (c *Content) OrderTimestamp() string {
	return c.Order.GetData().GetAttributes().GetCreatedAt().AsTime().Format("Jan 02, 2006")
}

func (c *Content) IsPlasmid(str string) bool {
	return strings.Contains(str, "DBP")
}

func (c *Content) IsStrain(str string) bool {
	return strings.Contains(str, "DBS")
}

func (c *Content) PlasmidItems() int {
	count := 0

	for _, item := range c.Order.GetData().GetAttributes().GetItems() {
		if c.IsPlasmid(item) {
			count++
		}
	}

	return count
}

func (c *Content) StrainItems() int {
	count := 0

	for _, item := range c.Order.GetData().GetAttributes().GetItems() {
		if c.IsStrain(item) {
			count++
		}
	}

	return count
}

func (c *Content) PlasmidCost() int {
	return c.PlasmidItems() * c.PlasmidPrice
}

func (c *Content) StrainCost() int {
	return c.StrainItems() * c.StrainPrice
}

func (c *Content) TotalCost() int {
	return c.StrainCost() + c.PlasmidCost()
}

func OutputText(args *OutputParams) (*bytes.Buffer, error) {
	out := bytes.NewBuffer([]byte(""))
	if args == nil {
		return out, errors.New("empty output params")
	}

	t, err := txt.New("stock-invoice").Parse(issueTemplate)
	if err != nil {
		return out, fmt.Errorf("error in parsing template %s", err)
	}

	if err := t.Execute(out, args.Content); err != nil {
		return out, fmt.Errorf("error in executing template %s", err)
	}

	return out, nil
}

func OutputPDF(args *OutputParams) (*bytes.Buffer, error) {
	if args == nil {
		return new(bytes.Buffer), errors.New("empty output params")
	}

	ec, ok := args.Content.(*EmailContent)
	if !ok || ec == nil {
		return new(bytes.Buffer), fmt.Errorf(
			"expected *EmailContent for pdf generation, got %T",
			args.Content,
		)
	}

	doc, err := BuildOrderMaroto(ec).Generate()
	if err != nil {
		return new(bytes.Buffer), fmt.Errorf("error in generating pdf %w", err)
	}

	return bytes.NewBuffer(doc.GetBytes()), nil
}
