package github

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IO "github.com/IBM/fp-go/v2/io"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	L "github.com/IBM/fp-go/v2/optics/lens"
	O "github.com/IBM/fp-go/v2/option"
	P "github.com/IBM/fp-go/v2/predicate"
	S "github.com/IBM/fp-go/v2/string"
	openrouter "github.com/OpenRouterTeam/go-sdk"
	"github.com/dictyBase/event-messenger/internal/datasource"
	issue "github.com/dictyBase/event-messenger/internal/issue-tracker"
	"github.com/dictyBase/event-messenger/internal/template"
	ioeutils "github.com/dictyBase/fp-go-loom/ioeitherutils"
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
	"github.com/dictyBase/go-genproto/dictybaseapis/stock"
	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type allData struct {
	*strainData
	*plasmidData
}

type strainData struct {
	strains []*stock.Strain
	invs    [][]string
	info    [][]string
}

type plasmidData struct {
	plasmids []*stock.Plasmid
	invs     [][]string
}

type issueStrainState struct {
	Order   *order.Order
	Strains []*stock.Strain
	Invs    [][]string
	Info    [][]string
}

type issuePlasmidState struct {
	Order    *order.Order
	Plasmids []*stock.Plasmid
	Invs     [][]string
}

type issueCreateState struct {
	Order  *order.Order
	Data   *allData
	Body   *bytes.Buffer
	Params *postIssueParams
}

// githubIssue includes all of the data needed to create an issue.
type githubIssue struct {
	token      string
	owner      string
	repository string
	client     *github.Client
	logger     *logrus.Entry
	anno       *datasource.Annotation
	stk        *datasource.Stock
	strprice   int
	plasprice  int
	decisions  DecisionsClient
	jevModel   string
}

type IssueParams struct {
	Token           string
	Owner           string
	Repository      string
	OpenRouterToken string
	JevModel        string
	Decisions       DecisionsClient
	StrainPrice     int
	PlasmidPrice    int
	Logger          *logrus.Entry
	*datasource.Sources
}

type postIssueParams struct {
	title  string
	body   string
	labels []string
}

var (
	issueParamsTokenLens = L.MakeLens(
		func(p *IssueParams) string { return p.Token },
		func(p *IssueParams, v string) *IssueParams {
			p.Token = v
			return p
		},
	)

	issueParamsOwnerLens = L.MakeLens(
		func(p *IssueParams) string { return p.Owner },
		func(p *IssueParams, v string) *IssueParams {
			p.Owner = v
			return p
		},
	)

	issueParamsRepoLens = L.MakeLens(
		func(p *IssueParams) string { return p.Repository },
		func(p *IssueParams, v string) *IssueParams {
			p.Repository = v
			return p
		},
	)

	issueParamsOpenRouterLens = L.MakeLens(
		func(p *IssueParams) string { return p.OpenRouterToken },
		func(p *IssueParams, v string) *IssueParams {
			p.OpenRouterToken = v
			return p
		},
	)

	issueParamsJevModelLens = L.MakeLens(
		func(p *IssueParams) string { return p.JevModel },
		func(p *IssueParams, v string) *IssueParams {
			p.JevModel = v
			return p
		},
	)

	issueParamsDecisionsLens = L.MakeLens(
		func(p *IssueParams) DecisionsClient { return p.Decisions },
		func(p *IssueParams, v DecisionsClient) *IssueParams {
			p.Decisions = v
			return p
		},
	)

	issueParamsStrainPriceLens = L.MakeLens(
		func(p *IssueParams) int { return p.StrainPrice },
		func(p *IssueParams, v int) *IssueParams {
			p.StrainPrice = v
			return p
		},
	)

	issueParamsPlasmidPriceLens = L.MakeLens(
		func(p *IssueParams) int { return p.PlasmidPrice },
		func(p *IssueParams, v int) *IssueParams {
			p.PlasmidPrice = v
			return p
		},
	)

	issueParamsLoggerLens = L.MakeLens(
		func(p *IssueParams) *logrus.Entry { return p.Logger },
		func(p *IssueParams, v *logrus.Entry) *IssueParams {
			p.Logger = v
			return p
		},
	)

	oauthTokenAccessLens = L.MakeLens(
		func(t oauth2.Token) string { return t.AccessToken },
		func(t oauth2.Token, v string) oauth2.Token {
			t.AccessToken = v
			return t
		},
	)

	ghClientLens = L.MakeLens(
		func(g githubIssue) *github.Client { return g.client },
		func(g githubIssue, v *github.Client) githubIssue {
			g.client = v
			return g
		},
	)

	ghTokenLens = L.MakeLens(
		func(g githubIssue) string { return g.token },
		func(g githubIssue, v string) githubIssue {
			g.token = v
			return g
		},
	)

	ghOwnerLens = L.MakeLens(
		func(g githubIssue) string { return g.owner },
		func(g githubIssue, v string) githubIssue {
			g.owner = v
			return g
		},
	)

	ghRepoLens = L.MakeLens(
		func(g githubIssue) string { return g.repository },
		func(g githubIssue, v string) githubIssue {
			g.repository = v
			return g
		},
	)

	ghLoggerLens = L.MakeLens(
		func(g githubIssue) *logrus.Entry { return g.logger },
		func(g githubIssue, v *logrus.Entry) githubIssue {
			g.logger = v
			return g
		},
	)

	ghAnnoLens = L.MakeLens(
		func(g githubIssue) *datasource.Annotation { return g.anno },
		func(g githubIssue, v *datasource.Annotation) githubIssue {
			g.anno = v
			return g
		},
	)

	ghStockLens = L.MakeLens(
		func(g githubIssue) *datasource.Stock { return g.stk },
		func(g githubIssue, v *datasource.Stock) githubIssue {
			g.stk = v
			return g
		},
	)

	ghStrainPriceLens = L.MakeLens(
		func(g githubIssue) int { return g.strprice },
		func(g githubIssue, v int) githubIssue {
			g.strprice = v
			return g
		},
	)

	ghPlasmidPriceLens = L.MakeLens(
		func(g githubIssue) int { return g.plasprice },
		func(g githubIssue, v int) githubIssue {
			g.plasprice = v
			return g
		},
	)

	ghDecisionsLens = L.MakeLens(
		func(g githubIssue) DecisionsClient { return g.decisions },
		func(g githubIssue, v DecisionsClient) githubIssue {
			g.decisions = v
			return g
		},
	)

	ghJevModelLens = L.MakeLens(
		func(g githubIssue) string { return g.jevModel },
		func(g githubIssue, v string) githubIssue {
			g.jevModel = v
			return g
		},
	)

	strainStateOrderLens = L.MakeLens(
		func(s issueStrainState) *order.Order { return s.Order },
		func(s issueStrainState, v *order.Order) issueStrainState {
			s.Order = v
			return s
		},
	)

	strainStateItemsLens = L.MakeLens(
		func(s issueStrainState) []*stock.Strain { return s.Strains },
		func(s issueStrainState, v []*stock.Strain) issueStrainState {
			s.Strains = v
			return s
		},
	)

	strainStateInvsLens = L.MakeLens(
		func(s issueStrainState) [][]string { return s.Invs },
		func(s issueStrainState, v [][]string) issueStrainState {
			s.Invs = v
			return s
		},
	)

	strainStateInfoLens = L.MakeLens(
		func(s issueStrainState) [][]string { return s.Info },
		func(s issueStrainState, v [][]string) issueStrainState {
			s.Info = v
			return s
		},
	)

	plasmidStateOrderLens = L.MakeLens(
		func(s issuePlasmidState) *order.Order { return s.Order },
		func(s issuePlasmidState, v *order.Order) issuePlasmidState {
			s.Order = v
			return s
		},
	)

	plasmidStateItemsLens = L.MakeLens(
		func(s issuePlasmidState) []*stock.Plasmid { return s.Plasmids },
		func(s issuePlasmidState, v []*stock.Plasmid) issuePlasmidState {
			s.Plasmids = v
			return s
		},
	)

	plasmidStateInvsLens = L.MakeLens(
		func(s issuePlasmidState) [][]string { return s.Invs },
		func(s issuePlasmidState, v [][]string) issuePlasmidState {
			s.Invs = v
			return s
		},
	)

	strainDataItemsLens = L.MakeLens(
		func(d strainData) []*stock.Strain { return d.strains },
		func(d strainData, v []*stock.Strain) strainData {
			d.strains = v
			return d
		},
	)

	strainDataInvsLens = L.MakeLens(
		func(d strainData) [][]string { return d.invs },
		func(d strainData, v [][]string) strainData {
			d.invs = v
			return d
		},
	)

	strainDataInfoLens = L.MakeLens(
		func(d strainData) [][]string { return d.info },
		func(d strainData, v [][]string) strainData {
			d.info = v
			return d
		},
	)

	plasmidDataItemsLens = L.MakeLens(
		func(d plasmidData) []*stock.Plasmid { return d.plasmids },
		func(d plasmidData, v []*stock.Plasmid) plasmidData {
			d.plasmids = v
			return d
		},
	)

	plasmidDataInvsLens = L.MakeLens(
		func(d plasmidData) [][]string { return d.invs },
		func(d plasmidData, v [][]string) plasmidData {
			d.invs = v
			return d
		},
	)

	allStrainDataLens = L.MakeLens(
		func(d allData) *strainData { return d.strainData },
		func(d allData, v *strainData) allData {
			d.strainData = v
			return d
		},
	)

	allPlasmidDataLens = L.MakeLens(
		func(d allData) *plasmidData { return d.plasmidData },
		func(d allData, v *plasmidData) allData {
			d.plasmidData = v
			return d
		},
	)

	createStateOrderLens = L.MakeLens(
		func(s issueCreateState) *order.Order { return s.Order },
		func(s issueCreateState, v *order.Order) issueCreateState {
			s.Order = v
			return s
		},
	)

	createStateDataLens = L.MakeLens(
		func(s issueCreateState) *allData { return s.Data },
		func(s issueCreateState, v *allData) issueCreateState {
			s.Data = v
			return s
		},
	)

	createStateBodyLens = L.MakeLens(
		func(s issueCreateState) *bytes.Buffer { return s.Body },
		func(s issueCreateState, v *bytes.Buffer) issueCreateState {
			s.Body = v
			return s
		},
	)

	createStateParamsLens = L.MakeLens(
		func(s issueCreateState) *postIssueParams { return s.Params },
		func(s issueCreateState, v *postIssueParams) issueCreateState {
			s.Params = v
			return s
		},
	)

	postTitleLens = L.MakeLens(
		func(p postIssueParams) string { return p.title },
		func(p postIssueParams, v string) postIssueParams {
			p.title = v
			return p
		},
	)

	postBodyLens = L.MakeLens(
		func(p postIssueParams) string { return p.body },
		func(p postIssueParams, v string) postIssueParams {
			p.body = v
			return p
		},
	)

	postLabelsLens = L.MakeLens(
		func(p postIssueParams) []string { return p.labels },
		func(p postIssueParams, v []string) postIssueParams {
			p.labels = v
			return p
		},
	)

	issueReqTitleLens = L.MakeLens(
		func(r github.IssueRequest) *string { return r.Title },
		func(r github.IssueRequest, v *string) github.IssueRequest {
			r.Title = v
			return r
		},
	)

	issueReqBodyLens = L.MakeLens(
		func(r github.IssueRequest) *string { return r.Body },
		func(r github.IssueRequest, v *string) github.IssueRequest {
			r.Body = v
			return r
		},
	)

	issueReqLabelsLens = L.MakeLens(
		func(r github.IssueRequest) *[]string { return r.Labels },
		func(r github.IssueRequest, v *[]string) github.IssueRequest {
			r.Labels = v
			return r
		},
	)

	contentShipperLens = L.MakeLens(
		func(c template.Content) *order.UserInfo { return c.Shipper },
		func(c template.Content, v *order.UserInfo) template.Content {
			c.Shipper = v
			return c
		},
	)

	contentPayerLens = L.MakeLens(
		func(c template.Content) *order.UserInfo { return c.Payer },
		func(c template.Content, v *order.UserInfo) template.Content {
			c.Payer = v
			return c
		},
	)

	contentOrderLens = L.MakeLens(
		func(c template.Content) *order.Order { return c.Order },
		func(c template.Content, v *order.Order) template.Content {
			c.Order = v
			return c
		},
	)

	contentStrainPriceLens = L.MakeLens(
		func(c template.Content) int { return c.StrainPrice },
		func(c template.Content, v int) template.Content {
			c.StrainPrice = v
			return c
		},
	)

	contentPlasmidPriceLens = L.MakeLens(
		func(c template.Content) int { return c.PlasmidPrice },
		func(c template.Content, v int) template.Content {
			c.PlasmidPrice = v
			return c
		},
	)

	issueContentBaseLens = L.MakeLens(
		func(ic template.IssueContent) *template.Content { return ic.Content },
		func(ic template.IssueContent, v *template.Content) template.IssueContent {
			ic.Content = v
			return ic
		},
	)

	issueContentStrainInvLens = L.MakeLens(
		func(ic template.IssueContent) [][]string { return ic.StrainInv },
		func(ic template.IssueContent, v [][]string) template.IssueContent {
			ic.StrainInv = v
			return ic
		},
	)

	issueContentPlasmidInvLens = L.MakeLens(
		func(ic template.IssueContent) [][]string { return ic.PlasmidInv },
		func(ic template.IssueContent, v [][]string) template.IssueContent {
			ic.PlasmidInv = v
			return ic
		},
	)

	issueContentStrainInfoLens = L.MakeLens(
		func(ic template.IssueContent) [][]string { return ic.StrainInfo },
		func(ic template.IssueContent, v [][]string) template.IssueContent {
			ic.StrainInfo = v
			return ic
		},
	)

	outputParamsContentLens = L.MakeLens(
		func(p template.OutputParams) any { return p.Content },
		func(p template.OutputParams, v any) template.OutputParams {
			p.Content = v
			return p
		},
	)

	isNonNilDecisions = P.IsNonZero[DecisionsClient]()
)

func resolveParamsDecisions(args *IssueParams) DecisionsClient {
	return F.Pipe3(
		args,
		issueParamsDecisionsLens.Get,
		O.FromPredicate(isNonNilDecisions),
		O.GetOrElse(func() DecisionsClient {
			return F.Pipe8(
				args,
				issueParamsOpenRouterLens.Get,
				strings.TrimSpace,
				O.FromPredicate(S.IsNonEmpty),
				O.Map(func(t string) *openrouter.OpenRouter {
					return openrouter.New(openrouter.WithSecurity(t))
				}),
				O.Chain(O.FromPredicate(isNonNilSDK)),
				O.Chain(sdkAlpha),
				O.Chain(alphaDecisions),
				O.GetOrElse(F.Constant(DecisionsClient(nil))),
			)
		}),
	)
}

func newGitHubOAuthClient(token string) *github.Client {
	tok := F.Pipe1(oauth2.Token{}, oauthTokenAccessLens.Set(token))
	tc := oauth2.NewClient(
		context.Background(),
		oauth2.StaticTokenSource(new(tok)),
	)

	return github.NewClient(tc)
}

// NewIssueCreator acts as a constructor for Github issue creation.
func NewIssueCreator(args *IssueParams) issue.Tracker {
	return F.Pipe12(
		githubIssue{},
		ghClientLens.Set(
			F.Pipe2(args, issueParamsTokenLens.Get, newGitHubOAuthClient),
		),
		ghTokenLens.Set(issueParamsTokenLens.Get(args)),
		ghOwnerLens.Set(issueParamsOwnerLens.Get(args)),
		ghRepoLens.Set(issueParamsRepoLens.Get(args)),
		ghLoggerLens.Set(issueParamsLoggerLens.Get(args)),
		ghAnnoLens.Set(args.AnnoSource),
		ghStockLens.Set(args.StockSource),
		ghStrainPriceLens.Set(issueParamsStrainPriceLens.Get(args)),
		ghPlasmidPriceLens.Set(issueParamsPlasmidPriceLens.Get(args)),
		ghDecisionsLens.Set(resolveParamsDecisions(args)),
		ghJevModelLens.Set(
			F.Pipe2(args, issueParamsJevModelLens.Get, resolveJevModel),
		),
		func(g githubIssue) issue.Tracker { return new(g) },
	)
}

// CreateIssue creates a new GitHub issue based on order data.
func (gh *githubIssue) CreateIssue(ord *order.Order) error {
	return F.Pipe8(
		F.Pipe1(issueCreateState{}, createStateOrderLens.Set(ord)),
		IOE.Of[error],
		IOE.Bind(createStateDataLens.Set, gh.loadAllOrderData),
		IOE.Bind(createStateBodyLens.Set, gh.renderIssueBody),
		IOE.Let[error](createStateParamsLens.Set, gh.preparePostParams),
		IOE.Chain(gh.executePostIssue),
		IOE.ChainFirstIOK[error](gh.logIssueCreated),
		IOE.TapLeft[string](F.Flow2(gh.logIssueError, IOE.FromIO[error, IO.Void])),
		F.Flow2(ioeutils.ToEither[error, string], E.ToError[string]),
	)
}

func (gh *githubIssue) loadAllOrderData(
	s issueCreateState,
) IOE.IOEither[error, *allData] {
	ord := createStateOrderLens.Get(s)

	return F.Pipe3(
		IOE.Of[error](allData{}),
		IOE.ApS(allStrainDataLens.Set, gh.loadStrainData(ord)),
		IOE.ApS(allPlasmidDataLens.Set, gh.loadPlasmidData(ord)),
		IOE.Map[error](func(d allData) *allData { return new(d) }),
	)
}

func (gh *githubIssue) renderIssueBody(
	s issueCreateState,
) IOE.IOEither[error, *bytes.Buffer] {
	ct := gh.buildIssueContent(
		createStateDataLens.Get(s),
		createStateOrderLens.Get(s),
	)
	params := F.Pipe1(template.OutputParams{}, outputParamsContentLens.Set(ct))

	return IOE.TryCatchError(func() (*bytes.Buffer, error) {
		return template.OutputText(new(params))
	})
}

func (gh *githubIssue) preparePostParams(s issueCreateState) *postIssueParams {
	base := gh.newPostIssueParams(
		createStateDataLens.Get(s),
		createStateOrderLens.Get(s),
	)
	withBody := F.Pipe1(
		*base,
		postBodyLens.Set(createStateBodyLens.Get(s).String()),
	)

	return new(withBody)
}

func (gh *githubIssue) executePostIssue(
	s issueCreateState,
) IOE.IOEither[error, string] {
	params := createStateParamsLens.Get(s)

	return F.Pipe2(
		IOE.TryCatchError(func() (*github.Issue, error) {
			return gh.postIssue(params)
		}),
		IOE.MapLeft[*github.Issue](func(err error) error {
			return fmt.Errorf("error in posting issue to github %s", err)
		}),
		IOE.Map[error](func(iss *github.Issue) string {
			return iss.GetHTMLURL()
		}),
	)
}

func (gh *githubIssue) logIssueCreated(url string) IO.IO[IO.Void] {
	return IO.FromImpure(func() {
		ghLoggerLens.Get(*gh).Infof("created a new issue with id %s", url)
	})
}

func (gh *githubIssue) logIssueError(err error) IO.IO[IO.Void] {
	return IO.FromImpure(func() {
		ghLoggerLens.Get(*gh).Error(err)
	})
}

func (gh *githubIssue) newPostIssueParams(
	all *allData,
	ord *order.Order,
) *postIssueParams {
	title := fmt.Sprintf(
		"Order ID:%s %s",
		ord.GetData().GetId(),
		ord.GetData().GetAttributes().GetPurchaser(),
	)
	labels := gh.resolveLabels(all, ord)
	params := F.Pipe2(
		postIssueParams{},
		postLabelsLens.Set(labels),
		postTitleLens.Set(title),
	)

	return new(params)
}

func (gh *githubIssue) postIssue(
	args *postIssueParams,
) (*github.Issue, error) {
	title := postTitleLens.Get(*args)
	body := postBodyLens.Get(*args)
	labels := postLabelsLens.Get(*args)
	req := F.Pipe3(
		github.IssueRequest{},
		issueReqTitleLens.Set(new(title)),
		issueReqBodyLens.Set(new(body)),
		issueReqLabelsLens.Set(new(labels)),
	)
	iss, _, err := ghClientLens.Get(*gh).Issues.Create(
		context.Background(),
		ghOwnerLens.Get(*gh),
		ghRepoLens.Get(*gh),
		new(req),
	)

	return iss, err
}

func (gh *githubIssue) fetchIssueStrains(
	s issueStrainState,
) IOE.IOEither[error, []*stock.Strain] {
	ids := ghStockLens.Get(*gh).StocksFromItems(
		strainStateOrderLens.Get(s),
		"DBS",
	)

	return F.Pipe1(
		IOE.TryCatchError(func() ([]*stock.Strain, error) {
			return ghStockLens.Get(*gh).GetStrains(ids)
		}),
		IOE.MapLeft[[]*stock.Strain](func(err error) error {
			return fmt.Errorf("error in getting strains %s", err)
		}),
	)
}

func (gh *githubIssue) fetchIssueStrainInvs(
	s issueStrainState,
) IOE.IOEither[error, [][]string] {
	strains := strainStateItemsLens.Get(s)

	return F.Pipe1(
		IOE.TryCatchError(func() ([][]string, error) {
			return ghAnnoLens.Get(*gh).GetStrainInv(strains)
		}),
		IOE.MapLeft[[][]string](func(err error) error {
			return fmt.Errorf("error in getting strain inventories %s", err)
		}),
	)
}

func (gh *githubIssue) fetchIssueStrainInfo(
	s issueStrainState,
) IOE.IOEither[error, [][]string] {
	strains := strainStateItemsLens.Get(s)

	return F.Pipe1(
		IOE.TryCatchError(func() ([][]string, error) {
			return ghAnnoLens.Get(*gh).GetStrainInfo(strains)
		}),
		IOE.MapLeft[[][]string](func(err error) error {
			return fmt.Errorf("error in getting strain information %s", err)
		}),
	)
}

func toStrainData(s issueStrainState) *strainData {
	d := F.Pipe3(
		strainData{},
		strainDataItemsLens.Set(strainStateItemsLens.Get(s)),
		strainDataInvsLens.Set(strainStateInvsLens.Get(s)),
		strainDataInfoLens.Set(strainStateInfoLens.Get(s)),
	)

	return new(d)
}

func (gh *githubIssue) loadStrainData(
	ord *order.Order,
) IOE.IOEither[error, *strainData] {
	return F.Pipe5(
		F.Pipe1(issueStrainState{}, strainStateOrderLens.Set(ord)),
		IOE.Of[error],
		IOE.Bind(strainStateItemsLens.Set, gh.fetchIssueStrains),
		IOE.Bind(strainStateInvsLens.Set, gh.fetchIssueStrainInvs),
		IOE.Bind(strainStateInfoLens.Set, gh.fetchIssueStrainInfo),
		IOE.Map[error](toStrainData),
	)
}

func (gh *githubIssue) fetchIssuePlasmids(
	s issuePlasmidState,
) IOE.IOEither[error, []*stock.Plasmid] {
	ids := ghStockLens.Get(*gh).StocksFromItems(
		plasmidStateOrderLens.Get(s),
		"DBP",
	)

	return F.Pipe1(
		IOE.TryCatchError(func() ([]*stock.Plasmid, error) {
			return ghStockLens.Get(*gh).GetPlasmids(ids)
		}),
		IOE.MapLeft[[]*stock.Plasmid](func(err error) error {
			return fmt.Errorf("error in getting plasmids %s", err)
		}),
	)
}

func (gh *githubIssue) fetchIssuePlasmidInvs(
	s issuePlasmidState,
) IOE.IOEither[error, [][]string] {
	plasmids := plasmidStateItemsLens.Get(s)

	return F.Pipe1(
		IOE.TryCatchError(func() ([][]string, error) {
			return ghAnnoLens.Get(*gh).GetPlasmidInv(plasmids)
		}),
		IOE.MapLeft[[][]string](func(err error) error {
			return fmt.Errorf("error in getting plasmid inventories %s", err)
		}),
	)
}

func toPlasmidData(s issuePlasmidState) *plasmidData {
	d := F.Pipe2(
		plasmidData{},
		plasmidDataItemsLens.Set(plasmidStateItemsLens.Get(s)),
		plasmidDataInvsLens.Set(plasmidStateInvsLens.Get(s)),
	)

	return new(d)
}

func (gh *githubIssue) loadPlasmidData(
	ord *order.Order,
) IOE.IOEither[error, *plasmidData] {
	return F.Pipe4(
		F.Pipe1(issuePlasmidState{}, plasmidStateOrderLens.Set(ord)),
		IOE.Of[error],
		IOE.Bind(plasmidStateItemsLens.Set, gh.fetchIssuePlasmids),
		IOE.Bind(plasmidStateInvsLens.Set, gh.fetchIssuePlasmidInvs),
		IOE.Map[error](toPlasmidData),
	)
}

func (gh *githubIssue) buildIssueContent(
	all *allData,
	ord *order.Order,
) *template.IssueContent {
	attr := ord.GetData().GetAttributes()
	sData := allStrainDataLens.Get(*all)
	pData := allPlasmidDataLens.Get(*all)

	baseContent := F.Pipe5(
		template.Content{},
		contentShipperLens.Set(attr.GetConsumerInfo()),
		contentPayerLens.Set(attr.GetPayerInfo()),
		contentOrderLens.Set(ord),
		contentStrainPriceLens.Set(ghStrainPriceLens.Get(*gh)),
		contentPlasmidPriceLens.Set(ghPlasmidPriceLens.Get(*gh)),
	)
	ic := F.Pipe4(
		template.IssueContent{},
		issueContentStrainInvLens.Set(strainDataInvsLens.Get(*sData)),
		issueContentPlasmidInvLens.Set(plasmidDataInvsLens.Get(*pData)),
		issueContentStrainInfoLens.Set(strainDataInfoLens.Get(*sData)),
		issueContentBaseLens.Set(new(baseContent)),
	)

	return new(ic)
}
