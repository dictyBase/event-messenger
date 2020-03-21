package template

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func checkSubstr(b fmt.Stringer, str []string, t *testing.T) {
	assert := assert.New(t)
	for _, s := range str {
		assert.Truef(
			strings.Contains(b.String(), s),
			"expect to have the pattern %s",
			s,
		)
	}
}

func emailSubstr() []string {
	return []string{
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
}

func issueSubstr() []string {
	return []string{
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
}

func TestReadFromBundle(t *testing.T) {
	b, err := ReadFromBundle("./../../assets", "email.tmpl")
	assert := assert.New(t)
	assert.NoError(err, "expect no error from reading email.tmpl template file")
	checkSubstr(b, emailSubstr(), t)
	b2, err := ReadFromBundle("./../../assets", "issue.tmpl")
	assert.NoError(err, "expect no error from reading issue.tmpl template file")
	checkSubstr(b2, issueSubstr(), t)
}
