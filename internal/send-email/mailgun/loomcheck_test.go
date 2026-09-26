package mailgun

import (
	"testing"

	loomcheck "github.com/dictyBase/fp-go-loom/loomcheck"
)

func TestNoHandRolledPredicates(t *testing.T) {
	t.Parallel()

	loomcheck.Require(t, loomcheck.Config{Roots: []string{"."}})
}
