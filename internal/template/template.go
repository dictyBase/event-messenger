package template

import (
	"bytes"
	"fmt"
	html "html/template"
	"io/ioutil"
	"path/filepath"
	"strings"
	txt "text/template"

	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/dictyBase/event-messenger/internal/datasource"
	_ "github.com/dictyBase/event-messenger/internal/statik"
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
	"github.com/dictyBase/go-genproto/dictybaseapis/user"
	"github.com/rakyll/statik/fs"
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
	return c.Order.Data.Attributes.CreatedAt.AsTime().Format("Jan 02, 2006")
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
	out := bytes.NewBuffer([]byte(""))
	ct, err := ReadFromBundle(args.Path, args.File)
	if err != nil {
		return out, err
	}
	t, err := txt.New("stock-invoice").Parse(ct)
	if err != nil {
		return out, fmt.Errorf("error in parsing template %s", err)
	}
	if err := t.Execute(out, args.Content); err != nil {
		return out, fmt.Errorf("error in executing template %s", err)
	}
	return out, nil
}

func OutputPDF(args *OutputParams) (*bytes.Buffer, error) {
	input, err := OutputHTML(args)
	if err != nil {
		return new(bytes.Buffer), err
	}
	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		return new(
				bytes.Buffer,
			), fmt.Errorf(
				"error in creating pdf generator %s",
				err,
			)
	}
	page := wkhtmltopdf.NewPageReader(input)
	pdfg.AddPage(page)
	if err := pdfg.Create(); err != nil {
		return new(bytes.Buffer), fmt.Errorf("error in generating pdf %s", err)
	}
	return pdfg.Buffer(), nil
}

func OutputHTML(args *OutputParams) (*bytes.Buffer, error) {
	out := bytes.NewBuffer([]byte(""))
	ct, err := ReadFromBundle(args.Path, args.File)
	if err != nil {
		return out, err
	}
	t, err := html.New("stock-invoice").Parse(ct)
	if err != nil {
		return out, fmt.Errorf("error in parsing template %s", err)
	}
	if err := t.Execute(out, args.Content); err != nil {
		return out, fmt.Errorf("error in executing template %s", err)
	}
	return out, nil
}

func ReadFromBundle(path, file string) (string, error) {
	statikFs, err := fs.New()
	if err != nil {
		return "", fmt.Errorf("error in getting embedded filesystem %s", err)
	}
	input := filepath.Join(path, file)
	r, err := statikFs.Open(input)
	if err != nil {
		return "", fmt.Errorf(
			"error in reading template file from path %s %s",
			input,
			err,
		)
	}
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("error in reading content of file %s", err)
	}
	return string(b), nil
}
