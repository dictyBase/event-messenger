package mailgun

import (
	"errors"
	"sync"
	"testing"

	E "github.com/IBM/fp-go/v2/either"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	"github.com/dictyBase/event-messenger/internal/datasource"
	ioeutils "github.com/dictyBase/fp-go-loom/ioeitherutils"
	"github.com/stretchr/testify/require"
)

// fakePublicationSource is a lazy publicationSource stub that records
// every id it is asked for.
type fakePublicationSource struct {
	mu    sync.Mutex
	infos map[string]*datasource.PubInfo
	errs  map[string]error
	calls []string
}

func (f *fakePublicationSource) ParsedInfo(
	id string,
) IOE.IOEither[error, *datasource.PubInfo] {
	return func() E.Either[error, *datasource.PubInfo] {
		f.mu.Lock()
		f.calls = append(f.calls, id)
		err := f.errs[id]
		info := f.infos[id]
		f.mu.Unlock()

		if err != nil {
			return E.Left[*datasource.PubInfo](err)
		}

		return E.Right[error](info)
	}
}

// recordedCalls returns a copy of the ids asked for so far.
func (f *fakePublicationSource) recordedCalls() []string {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]string(nil), f.calls...)
}

// pubInfoFor builds a PubInfo whose author string tags its origin.
func pubInfoFor(author string) *datasource.PubInfo {
	return &datasource.PubInfo{AuthorStr: author}
}

// emailerWithPublications builds an emailer with only the publication
// source wired, enough for the pubInfo traversal tests.
func emailerWithPublications(fake *fakePublicationSource) *mailgunEmailer {
	return &mailgunEmailer{pub: fake}
}

func TestPubInfoFiltersAndTrimsIDs(t *testing.T) {
	t.Parallel()

	fake := &fakePublicationSource{
		infos: map[string]*datasource.PubInfo{
			"123": pubInfoFor("first"),
			"456": pubInfoFor("second"),
		},
		errs: map[string]error{},
	}

	res, err := E.UnwrapError(ioeutils.ToEither(
		emailerWithPublications(fake).pubInfo(
			[]string{"  ", "", " 123 ", "456", " \t "},
		),
	))
	require.NoError(t, err)
	require.Equal(t, []string{"123", "456"}, fake.recordedCalls())
	require.Len(t, res, 2)
	require.Equal(t, "first", res[0].AuthorStr)
	require.Equal(t, "second", res[1].AuthorStr)
}

func TestPubInfoPreservesOrder(t *testing.T) {
	t.Parallel()

	fake := &fakePublicationSource{
		infos: map[string]*datasource.PubInfo{
			"p1": pubInfoFor("first"),
			"p2": pubInfoFor("second"),
			"p3": pubInfoFor("third"),
		},
		errs: map[string]error{},
	}

	res, err := E.UnwrapError(ioeutils.ToEither(
		emailerWithPublications(fake).pubInfo([]string{"p1", "p2", "p3"}),
	))
	require.NoError(t, err)
	require.Equal(t, []string{"p1", "p2", "p3"}, fake.recordedCalls())

	authors := make([]string, 0, len(res))
	for _, info := range res {
		authors = append(authors, info.AuthorStr)
	}

	require.Equal(t, []string{"first", "second", "third"}, authors)
}

func TestPubInfoPropagatesFailure(t *testing.T) {
	t.Parallel()

	fake := &fakePublicationSource{
		infos: map[string]*datasource.PubInfo{
			"p1": pubInfoFor("first"),
			"p3": pubInfoFor("third"),
		},
		errs: map[string]error{"p2": errors.New("boom")},
	}

	_, err := E.UnwrapError(ioeutils.ToEither(
		emailerWithPublications(fake).pubInfo([]string{"p1", "p2", "p3"}),
	))
	require.Error(t, err)
	require.ErrorContains(t, err, "boom")
	require.Equal(t, []string{"p1", "p2", "p3"}, fake.recordedCalls())
}

func TestNormalizePublicationIDs(t *testing.T) {
	t.Parallel()

	require.Equal(
		t,
		[]string{"123", "456"},
		normalizePublicationIDs([]string{" 123 ", "", "456", " \t "}),
	)
	require.Empty(t, normalizePublicationIDs(nil))
}
