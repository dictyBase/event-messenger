package template

type EmailContent struct {
	*Content
	StrainData  []*StrainRows
	PlasmidData []*PlasmidRows
}
