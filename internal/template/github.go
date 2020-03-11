package template

const IssueTmpl = `
 Shipping and billing information   

|	Shipping address |	  | Billing address	 |
: -------------------|----|------------------:
|  {{- with .Shipper.Data.Attributes }} {{.FirstName}} {{.LastaName}} <br/> {{.Organization}}	<br/> {{.FirstAddress}} <br/> {{.SecondAddress}} <br/> {{.City}} {{.State}} {{.ZipCode}} <br/> {{.Country}} <br/> Phone: {{.Phone}} <br/> {{.Email}} <br/> {{.Courier}} {{.CourierAccount}} {{- end }} |          | {{- with .Shipper.Data.Attributes }} {{.FirstName}} {{.LastaName}} <br/> {{.Organization}}	<br/> {{.FirstAddress}} <br/> {{.SecondAddress}} <br/> {{.City}} {{.State}} {{.ZipCode}} <br/> {{.Country}} <br/> Phone: {{.Phone}} <br/> {{.Email}} <br/> {{.Courier}} {{.CourierAccount}} {{- end }} |          | {{- with .Shipper.Data.Attributes }} {{.FirstName}} {{.LastaName}} <br/> {{.Organization}}	<br/> {{.FirstAddress}} <br/> {{.SecondAddress}} <br/> {{.City}} {{.State}} {{.ZipCode}} <br/> {{.Country}} <br/> Phone: {{.Phone}} <br/> {{.Email}} <br/> {{.Courier}} {{.CourierAccount}} {{- end }} |          | {{- with .Shipper.Data.Attributes }} {{.FirstName}} {{.LastaName}} <br/> {{.Organization}}	<br/> {{.FirstAddress}} <br/> {{.SecondAddress}} <br/> {{.City}} {{.State}} {{.ZipCode}} <br/> {{.Country}} <br/> Phone: {{.Phone}} <br/> {{.Email}} <br/> {{.Courier}} {{.CourierAccount}} {{- end }} |   | {{- with .Payer.Data.Attributes }} {{.FirstName}} {{.LastaName}} <br/> {{.Organization}}	<br/> {{.FirstAddress}} <br/> {{.SecondAddress}} <br/> {{.City}} {{.State}} {{.ZipCode}} <br/> {{.Country}} <br/> Phone: {{.Phone}} <br/> {{.Email}} <br/> {{.Courier}} {{.CourierAccount}} {{- end }} |          | {{- with .Shipper.Data.Attributes }} {{.FirstName}} {{.LastaName}} <br/> {{.Organization}}	<br/> {{.FirstAddress}} <br/> {{.SecondAddress}} <br/> {{.City}} {{.State}} {{.ZipCode}} <br/> {{.Country}} <br/> Phone: {{.Phone}} <br/> {{.Email}} <br/> {{.Courier}} {{.CourierAccount}} {{- end }} |          | {{- with .Shipper.Data.Attributes }} {{.FirstName}} {{.LastaName}} <br/> {{.Organization}}	<br/> {{.FirstAddress}} <br/> {{.SecondAddress}} <br/> {{.City}} {{.State}} {{.ZipCode}} <br/> {{.Country}} <br/> Phone: {{.Phone}} <br/> {{.Email}} <br/> {{.Courier}} {{.CourierAccount}} {{- end }} |   | {{- with .Shipper.Data.Attributes }} {{.FirstName}} {{.LastaName}} <br/> {{.Organization}}	<br/> {{.FirstAddress}} <br/> {{.SecondAddress}} <br/> {{.City}} {{.State}} {{.ZipCode}} <br/> {{.Country}} <br/> Phone: {{.Phone}} <br/> {{.Email}} <br/> {{.Payment}} {{- end }} |

{{if .Strains}}
# Stocks ordered

|	Item	|	Quantity 	     |	Unit price($)	  |	Total($)	       |
|-----------|--------------------|--------------------|--------------------|
|	Strain	| {{.StrainItems}}   | {{.StrainPrice}}   |  {{.StrainCost}}   |
|	Plasmid	| {{.PlasmidItems}}  | {{.PlasmidPrice}}  |  {{.PlasmidCost}}  |
|			|					 |					  |	 {{.TotalCost}}    |

# Strain information 

|  ID	|  Descriptor   |	Name(s)  |	Systematic Name   |	Characteristics | 
|-------|---------------|--- --------|--------------------|-----------------|
{{- range $idx,$e := .StrainInfo}}
| {{index $e 0}}  | {{index $e 1}}  | {{index $e 2}} | {{index $e 3}} | {{index $e 4}} |
{{- end}}
	

# Strain storage
{{if .StrainInv}}

|	Name |	Stored as |	Location |	No. of vials |	Color   |
|--------|------------|----------|---------------|----------|
{{- range $idx,$e := .StrainInv}}
| {{index $e 0}} | {{index $e 1}} | {{index $e 2}} |  {{index $e 3}} | {{index $e 4}} |
{{- end}}
{{else}}
## No strain inventories, POP,CRAKLE,BOOM!!!!!
{{end}}

{{end}}


{{if and .Plasmids .PlasmidInv}}
# Plasmid information and storage

| ID  |	Name  |	Stored as |	Location |	Color |
|-----|-------|-----------|----------|--------|
{{- range $idx,$e := .PlasmidInv}}
| {{index $e 0}} | {{index $e 1}} | {{index $e 2}} | {{index $e 3}} | {{index $e 4}} |
{{- end}}
{{else}}
# No plasmid inventories no order
{{end}}

{{if .Order.Data.Attributes.Comments}}
# Comment
{{.Order.Data.Attributes.Comments}}
{{end}}

`
