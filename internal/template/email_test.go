package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmailStrainHtml(t *testing.T) {
	_, err := OutputHTML(&OutputParams{
		File:    "email.tmpl",
		Path:    "./../../assets",
		Content: fakeStrainOnlyEmailContent(),
	})
	assert := assert.New(t)
	assert.NoError(err, "expect no error from rendering email template with strain data")
}
