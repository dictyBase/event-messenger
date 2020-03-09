package github

import (
	"strings"

	"github.com/dictyBase/go-genproto/dictybaseapis/order"
	"github.com/dictyBase/go-genproto/dictybaseapis/stock"
	"github.com/dictyBase/go-genproto/dictybaseapis/user"
)

type IssueContent struct {
	Strains      []*stock.Strain
	Plasmids     []*stock.Plasmid
	StrainInv    [][]string
	PlasmidInv   [][]string
	StrainInfo   [][]string
	Order        *order.Order
	Shipper      *user.User
	Payer        *user.User
	StrainPrice  int
	PlasmidPrice int
}

func (c *IssueContent) IsPlasmid(str string) bool {
	return strings.Contains(str, "DBP")
}

func (c *IssueContent) IsStrain(str string) bool {
	return strings.Contains(str, "DBS")
}

func (c *IssueContent) PlasmidItems() int {
	count := 0
	for _, item := range c.Order.Data.Attributes.Items {
		if c.IsPlasmid(item) {
			count++
		}
	}
	return count
}

func (c *IssueContent) StrainItems() int {
	count := 0
	for _, item := range c.Order.Data.Attributes.Items {
		if c.IsStrain(item) {
			count++
		}
	}
	return count
}

func (c *IssueContent) PlasmidCost() int {
	return c.PlasmidItems() * c.PlasmidPrice
}

func (c *IssueContent) StrainCost() int {
	return c.StrainItems() * c.StrainPrice
}

func (c *IssueContent) TotalCost() int {
	return c.StrainCost() + c.PlasmidCost()
}

func tjoin(elems []string, sep string) string {
	return strings.Join(elems, sep)
}

const tmpl = `
 Shipping and billing information   

|	Shipping address									 |		  | Billing address	    						 |
: -------------------------------------------------------|--------|----------------------------------------------:
|   {{- with .Shipper.Data.Attributes }} {{.FirstName}} {{.LastaName}} <br/> {{.Organization}}	<br/> {{.FirstAddress}} <br/> {{.SecondAddress}} <br/> {{.City}} {{.State}} {{.ZipCode}} <br/> {{.Country}} <br/> Phone: {{.Phone}} <br/> {{.Email}} <br/> {{.Courier}} {{.CourierAccount}} {{- end }} |          | {{- with .Shipper.Data.Attributes }} {{.FirstName}} {{.LastaName}} <br/> {{.Organization}}	<br/> {{.FirstAddress}} <br/> {{.SecondAddress}} <br/> {{.City}} {{.State}} {{.ZipCode}} <br/> {{.Country}} <br/> Phone: {{.Phone}} <br/> {{.Email}} <br/> {{.Courier}} {{.CourierAccount}} {{- end }} |          | {{- with .Shipper.Data.Attributes }} {{.FirstName}} {{.LastaName}} <br/> {{.Organization}}	<br/> {{.FirstAddress}} <br/> {{.SecondAddress}} <br/> {{.City}} {{.State}} {{.ZipCode}} <br/> {{.Country}} <br/> Phone: {{.Phone}} <br/> {{.Email}} <br/> {{.Courier}} {{.CourierAccount}} {{- end }} |          | {{- with .Shipper.Data.Attributes }} {{.FirstName}} {{.LastaName}} <br/> {{.Organization}}	<br/> {{.FirstAddress}} <br/> {{.SecondAddress}} <br/> {{.City}} {{.State}} {{.ZipCode}} <br/> {{.Country}} <br/> Phone: {{.Phone}} <br/> {{.Email}} <br/> {{.Courier}} {{.CourierAccount}} {{- end }} |   | {{- with .Payer.Data.Attributes }} {{.FirstName}} {{.LastaName}} <br/> {{.Organization}}	<br/> {{.FirstAddress}} <br/> {{.SecondAddress}} <br/> {{.City}} {{.State}} {{.ZipCode}} <br/> {{.Country}} <br/> Phone: {{.Phone}} <br/> {{.Email}} <br/> {{.Courier}} {{.CourierAccount}} {{- end }} |          | {{- with .Shipper.Data.Attributes }} {{.FirstName}} {{.LastaName}} <br/> {{.Organization}}	<br/> {{.FirstAddress}} <br/> {{.SecondAddress}} <br/> {{.City}} {{.State}} {{.ZipCode}} <br/> {{.Country}} <br/> Phone: {{.Phone}} <br/> {{.Email}} <br/> {{.Courier}} {{.CourierAccount}} {{- end }} |          | {{- with .Shipper.Data.Attributes }} {{.FirstName}} {{.LastaName}} <br/> {{.Organization}}	<br/> {{.FirstAddress}} <br/> {{.SecondAddress}} <br/> {{.City}} {{.State}} {{.ZipCode}} <br/> {{.Country}} <br/> Phone: {{.Phone}} <br/> {{.Email}} <br/> {{.Courier}} {{.CourierAccount}} {{- end }} |   | {{- with .Shipper.Data.Attributes }} {{.FirstName}} {{.LastaName}} <br/> {{.Organization}}	<br/> {{.FirstAddress}} <br/> {{.SecondAddress}} <br/> {{.City}} {{.State}} {{.ZipCode}} <br/> {{.Country}} <br/> Phone: {{.Phone}} <br/> {{.Email}} <br/> {{.Payment}} {{- end }} |

{{if .Strains}}
# Stocks ordered

|	Item	|	Quantity 	     |	Unit price($)	  |	Total($)	       |
|-----------|--------------------|--------------------|--------------------|
|	Strain	| {{.StrainItems}}   | {{.StrainPrice}}   |  {{.StrainCost}}   |
|	Plasmid	| {{.PlasmidItems}}  | {{.PlasmidPrice}}  |  {{.PlasmidCost}}  |
|			|					 |					  |	 {{.TotalCost}}    |

# Strain information 

|  ID			|  Descriptor      |	Name(s)		            |	Systematic Name	|	Characteristics			                      |
|---------------|------------------|----------------------------|-------------------|----------------------------------------------- -|
{{- range .Strains}}
  {{- $attr := .Data.StrainAttributes -}}
| {{.Data.Id}}  | {{$attr.Label}}  | {{$attr.Names | join ","}} | {{.Data.Id}}     | {{characteristics (index .StrainChar .Data.Id)}} |
{{- end}}
	

# Strain storage

|	Name		  |	Stored as       |	Location	 |	No. of vials    |	Color	|
|-----------------|-----------------|----------------|------------------|-----------|
{{- range $strain := .Strain}}
  {{- $attr := $strain.Data.Attributes -}}
| {{$attr.Label}} | {{range $tag := .StrainInvTags}} {{inventory $tag (index .StrainInv $strain.Data.Id)}} | {{end}}
{{- end}}
{{end}}

{{if .Plasmids}}
# Plasmid information and storage

|	ID				   |	Name		     |	Stored as	     |	Location         |	Color	     |
|----------------------|---------------------|-------------------|-------------------|---------------|
{{- range $plasmid := .Plasmids}}
  {{- $attr := $plasmid.Data.Attributes -}}
| {{$plasmid.Data.Id}} | {{$attr.Name}}      | {{range $tag := .PlasmidInvTags}} {{inventory $tag (index .PlasmidInv $plasmid.Data.Id)}} | {{end}}
{{- end}}
{{end}}
`
