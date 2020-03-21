package template

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func checkSubstr(b *bytes.Buffer, str []string, t *testing.T) {
	assert := assert.New(t)
	for _, s := range str {
		assert.Truef(
			strings.Contains(b.String(), s),
			"expect to have the pattern %s",
			s,
		)
	}
}

func TestReadFromBundle(t *testing.T) {
	b, err := ReadFromBundle("./../../assets", "email.tmpl")
	assert := assert.New(t)
	assert.NoError(err, "expect no error from reading email.tmpl template file")
	str := []string{
		"dsc-header",
		"shipping-row",
		"Order Confirmation",
		"Order #",
		"Shipping Address",
		"Billing Address",
		"Item",
		"Quantity",
		"Total",
		"Systematic Name",
		"Descriptor",
		"Plasmid Name",
		"Comment",
		"Payment information",
	}
	checkSubstr(b, str, t)
	b2, err := ReadFromBundle("./../../assets", "issue.tmpl")
	assert.NoError(err, "expect no error from reading issue.tmpl template file")
	str2 := []string{
		"Shipping address",
		"Billing address",
		"Item",
		"Quantity",
		"Total",
		"Systematic Name",
		"Descriptor",
		"Characteristics",
		"Strain storage",
		"Location",
		"Color",
		"Plasmid information and storage",
	}
	checkSubstr(b2, str2, t)
}
