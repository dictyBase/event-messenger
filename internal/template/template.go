package template

import (
	"bytes"
	"fmt"
	html "html/template"
	"strings"
	txt "text/template"

	"github.com/dictyBase/event-messenger/internal/datasource"
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
	"github.com/dictyBase/go-genproto/dictybaseapis/user"
	"github.com/gobuffalo/packr/v2"
	"github.com/golang/protobuf/ptypes"
)

type OutputParams struct {
	File    string
	Path    string
	Content interface{}
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
	Shipper      *user.User
	Payer        *user.User
	StrainPrice  int
	PlasmidPrice int
}

func (c *Content) OrderTimestamp() string {
	t, _ := ptypes.Timestamp(c.Order.Data.Attributes.CreatedAt)
	return t.Format("Jan 02, 2006 15:04")
}

func (c *Content) IsPlasmid(str string) bool {
	return strings.Contains(str, "DBP")
}

func (c *Content) IsStrain(str string) bool {
	return strings.Contains(str, "DBS")
}

func (c *Content) PlasmidItems() int {
	count := 0
	for _, item := range c.Order.Data.Attributes.Items {
		if c.IsPlasmid(item) {
			count++
		}
	}
	return count
}

func (c *Content) StrainItems() int {
	count := 0
	for _, item := range c.Order.Data.Attributes.Items {
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
	b, err := ReadFromBundle(args.Path, args.File)
	if err != nil {
		return b, err
	}
	t, err := txt.New("stock-invoice").Parse(b.String())
	if err != nil {
		return b, fmt.Errorf("error in parsing template %s", err)
	}
	if err := t.Execute(b, args.Content); err != nil {
		return b, fmt.Errorf("error in executing template %s", err)
	}
	return b, nil
}

func OutputHTML(args *OutputParams) (*bytes.Buffer, error) {
	b, err := ReadFromBundle(args.Path, args.File)
	if err != nil {
		return b, err
	}
	t, err := html.New("stock-invoice").Parse(b.String())
	if err != nil {
		return b, fmt.Errorf("error in parsing template %s", err)
	}
	out := bytes.NewBuffer([]byte(""))
	if err := t.Execute(out, args.Content); err != nil {
		return out, fmt.Errorf("error in executing template %s", err)
	}
	return out, nil
}

func ReadFromBundle(path, file string) (*bytes.Buffer, error) {
	var b *bytes.Buffer
	box := packr.New("html", "|")
	box.ResolutionDir = path
	tb, err := box.Find(file)
	if err != nil {
		return b, fmt.Errorf("error in reading template file content from path %s %s", path, err)
	}
	return bytes.NewBuffer(tb), nil
}