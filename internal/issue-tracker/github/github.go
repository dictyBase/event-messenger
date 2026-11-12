package github

import (
	"bytes"
	"context"
	"fmt"

	E "github.com/IBM/fp-go/v2/either"
	F "github.com/IBM/fp-go/v2/function"
	IO "github.com/IBM/fp-go/v2/io"
	IOE "github.com/IBM/fp-go/v2/ioeither"
	L "github.com/IBM/fp-go/v2/optics/lens"
	P "github.com/IBM/fp-go/v2/predicate"
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

type issueCreateState struct {
	Order  *order.Order
	Data   *allData
	Body   *bytes.Buffer
	Labels []string
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

	oauthTokenAccessLens = L.MakeLensRef(
		func(t *oauth2.Token) string { return t.AccessToken },
		func(t *oauth2.Token, v string) *oauth2.Token {
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

	createStateLabelsLens = L.MakeLens(
		func(s issueCreateState) []string { return s.Labels },
		func(s issueCreateState, v []string) issueCreateState {
			s.Labels = v
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

	issueReqTitleLens = L.MakeLensRef(
		func(r *github.IssueRequest) *string { return r.Title },
		func(r *github.IssueRequest, v *string) *github.IssueRequest {
			r.Title = v
			return r
		},
	)

	issueReqBodyLens = L.MakeLensRef(
		func(r *github.IssueRequest) *string { return r.Body },
		func(r *github.IssueRequest, v *string) *github.IssueRequest {
			r.Body = v
			return r
		},
	)

	issueReqLabelsLens = L.MakeLensRef(
		func(r *github.IssueRequest) *[]string { return r.Labels },
		func(r *github.IssueRequest, v *[]string) *github.IssueRequest {
			r.Labels = v
			return r
		},
	)

	contentShipperLens = L.MakeLensRef(
		func(c *template.Content) *order.UserInfo { return c.Shipper },
		func(c *template.Content, v *order.UserInfo) *template.Content {
			c.Shipper = v
			return c
		},
	)

	contentPayerLens = L.MakeLensRef(
		func(c *template.Content) *order.UserInfo { return c.Payer },
		func(c *template.Content, v *order.UserInfo) *template.Content {
			c.Payer = v
			return c
		},
	)

	contentOrderLens = L.MakeLensRef(
		func(c *template.Content) *order.Order { return c.Order },
		func(c *template.Content, v *order.Order) *template.Content {
			c.Order = v
			return c
		},
	)

	contentStrainPriceLens = L.MakeLensRef(
		func(c *template.Content) int { return c.StrainPrice },
		func(c *template.Content, v int) *template.Content {
			c.StrainPrice = v
			return c
		},
	)

	contentPlasmidPriceLens = L.MakeLensRef(
		func(c *template.Content) int { return c.PlasmidPrice },
		func(c *template.Content, v int) *template.Content {
			c.PlasmidPrice = v
			return c
		},
	)

	issueContentBaseLens = L.MakeLensRef(
		func(ic *template.IssueContent) *template.Content { return ic.Content },
		func(
			ic *template.IssueContent,
			v *template.Content,
		) *template.IssueContent {
			ic.Content = v
			return ic
		},
	)

	issueContentShipperLens = F.Pipe1(
		issueContentBaseLens,
		L.Compose[*template.IssueContent](contentShipperLens),
	)

	issueContentPayerLens = F.Pipe1(
		issueContentBaseLens,
		L.Compose[*template.IssueContent](contentPayerLens),
	)

	issueContentOrderLens = F.Pipe1(
		issueContentBaseLens,
		L.Compose[*template.IssueContent](contentOrderLens),
	)

	issueContentStrainPriceLens = F.Pipe1(
		issueContentBaseLens,
		L.Compose[*template.IssueContent](contentStrainPriceLens),
	)

	issueContentPlasmidPriceLens = F.Pipe1(
		issueContentBaseLens,
		L.Compose[*template.IssueContent](contentPlasmidPriceLens),
	)

	issueContentStrainInvLens = L.MakeLensRef(
		func(ic *template.IssueContent) [][]string { return ic.StrainInv },
		func(ic *template.IssueContent, v [][]string) *template.IssueContent {
			ic.StrainInv = v
			return ic
		},
	)

	issueContentPlasmidInvLens = L.MakeLensRef(
		func(ic *template.IssueContent) [][]string { return ic.PlasmidInv },
		func(ic *template.IssueContent, v [][]string) *template.IssueContent {
			ic.PlasmidInv = v
			return ic
		},
	)

	issueContentStrainInfoLens = L.MakeLensRef(
		func(ic *template.IssueContent) [][]string { return ic.StrainInfo },
		func(ic *template.IssueContent, v [][]string) *template.IssueContent {
			ic.StrainInfo = v
			return ic
		},
	)

	outputParamsContentLens = L.MakeLensRef(
		func(p *template.OutputParams) any { return p.Content },
		func(p *template.OutputParams, v any) *template.OutputParams {
			p.Content = v
			return p
		},
	)

	isNonNilDecisions = P.IsNonZero[DecisionsClient]()

	stateOrderAttrs = F.Flow3(
		createStateOrderLens.Get,
		(*order.Order).GetData,
		(*order.Order_Data).GetAttributes,
	)

	stateShipperInfo = F.Flow2(
		stateOrderAttrs,
		(*order.OrderAttributes).GetConsumerInfo,
	)

	statePayerInfo = F.Flow2(
		stateOrderAttrs,
		(*order.OrderAttributes).GetPayerInfo,
	)

	stateBodyText = F.Flow2(createStateBodyLens.Get, (*bytes.Buffer).String)

	stateStrainInvRows  = F.Flow2(createStateDataLens.Get, extractStrainInvRows)
	statePlasmidInvRows = F.Flow2(createStateDataLens.Get, extractPlasmidInvRows)
	stateStrainInfoRows = F.Flow2(createStateDataLens.Get, extractStrainInfoRows)
)

func stateOrderTitle(s issueCreateState) string {
	return fmt.Sprintf(
		"Order ID:%s %s",
		createStateOrderLens.Get(s).GetData().GetId(),
		stateOrderAttrs(s).GetPurchaser(),
	)
}

func newGitHubOAuthClient(args *IssueParams) *github.Client {
	return F.Pipe4(
		&oauth2.Token{},
		oauthTokenAccessLens.Set(issueParamsTokenLens.Get(args)),
		oauth2.StaticTokenSource,
		F.Bind1st(oauth2.NewClient, context.Background()),
		github.NewClient,
	)
}

// NewIssueCreator acts as a constructor for Github issue creation.
func NewIssueCreator(args *IssueParams) issue.Tracker {
	return F.Pipe13(
		githubIssue{},
		ghClientLens.Set(newGitHubOAuthClient(args)),
		ghTokenLens.Set(issueParamsTokenLens.Get(args)),
		ghOwnerLens.Set(issueParamsOwnerLens.Get(args)),
		ghRepoLens.Set(issueParamsRepoLens.Get(args)),
		ghLoggerLens.Set(issueParamsLoggerLens.Get(args)),
		ghAnnoLens.Set(args.AnnoSource),
		ghStockLens.Set(args.StockSource),
		ghStrainPriceLens.Set(issueParamsStrainPriceLens.Get(args)),
		ghPlasmidPriceLens.Set(issueParamsPlasmidPriceLens.Get(args)),
		ghDecisionsLens.Set(resolveParamsDecisions(args)),
		ghJevModelLens.Set(issueParamsJevModelLens.Get(args)),
		ghJevModelLens.Modify(resolveJevModel),
		func(g githubIssue) issue.Tracker { return new(g) },
	)
}

// CreateIssue creates a new GitHub issue based on order data.
func (gh *githubIssue) CreateIssue(ord *order.Order) error {
	return F.Pipe11(
		issueCreateState{},
		createStateOrderLens.Set(ord),
		IOE.Of[error],
		IOE.Bind(createStateDataLens.Set, gh.loadAllOrderData),
		IOE.Bind(createStateBodyLens.Set, gh.renderIssueBody),
		IOE.Bind(createStateLabelsLens.Set, gh.resolveLabels),
		IOE.Let[error](createStateParamsLens.Set, gh.preparePostParams),
		IOE.Chain(gh.executePostIssue),
		IOE.ChainFirstIOK[error](gh.logIssueCreated),
		IOE.TapLeft[string](F.Flow2(gh.logIssueError, IOE.FromIO[error, IO.Void])),
		ioeutils.ToEither[error, string],
		E.ToError[string],
	)
}

func (gh *githubIssue) loadAllOrderData(
	s issueCreateState,
) IOE.IOEither[error, *allData] {
	return F.Pipe3(
		IOE.Of[error](allData{}),
		IOE.ApS(allStrainDataLens.Set, gh.loadStrainData(s)),
		IOE.ApS(allPlasmidDataLens.Set, gh.loadPlasmidData(s)),
		IOE.Map[error](func(d allData) *allData { return new(d) }),
	)
}

func (gh *githubIssue) renderIssueBody(
	s issueCreateState,
) IOE.IOEither[error, *bytes.Buffer] {
	return F.Pipe2(
		&template.OutputParams{},
		outputParamsContentLens.Set(gh.buildIssueContent(s)),
		IOE.Eitherize1(template.OutputText),
	)
}

func (gh *githubIssue) preparePostParams(s issueCreateState) *postIssueParams {
	return F.Pipe4(
		postIssueParams{},
		postTitleLens.Set(stateOrderTitle(s)),
		postBodyLens.Set(stateBodyText(s)),
		postLabelsLens.Set(createStateLabelsLens.Get(s)),
		func(p postIssueParams) *postIssueParams { return new(p) },
	)
}

func (gh *githubIssue) executePostIssue(
	s issueCreateState,
) IOE.IOEither[error, string] {
	return F.Pipe6(
		&github.IssueRequest{},
		issueReqTitleLens.Set(
			new(postTitleLens.Get(*createStateParamsLens.Get(s))),
		),
		issueReqBodyLens.Set(
			new(postBodyLens.Get(*createStateParamsLens.Get(s))),
		),
		issueReqLabelsLens.Set(
			new(postLabelsLens.Get(*createStateParamsLens.Get(s))),
		),
		IOE.Eitherize1(gh.postIssue),
		IOE.MapLeft[*github.Issue](func(err error) error {
			return fmt.Errorf("error in posting issue to github %s", err)
		}),
		IOE.Map[error]((*github.Issue).GetHTMLURL),
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

func (gh *githubIssue) postIssue(
	req *github.IssueRequest,
) (*github.Issue, error) {
	iss, _, err := ghClientLens.Get(*gh).Issues.Create(
		context.Background(),
		ghOwnerLens.Get(*gh),
		ghRepoLens.Get(*gh),
		req,
	)

	return iss, err
}

func (gh *githubIssue) fetchIssueStrains(
	s issueCreateState,
) IOE.IOEither[error, []*stock.Strain] {
	return F.Pipe1(
		IOE.TryCatchError(func() ([]*stock.Strain, error) {
			return ghStockLens.Get(*gh).GetStrains(
				ghStockLens.Get(*gh).StocksFromItems(
					createStateOrderLens.Get(s),
					"DBS",
				),
			)
		}),
		IOE.MapLeft[[]*stock.Strain](func(err error) error {
			return fmt.Errorf("error in getting strains %s", err)
		}),
	)
}

func (gh *githubIssue) fetchIssueStrainInvs(
	d strainData,
) IOE.IOEither[error, [][]string] {
	return F.Pipe1(
		IOE.TryCatchError(func() ([][]string, error) {
			return ghAnnoLens.Get(*gh).GetStrainInv(
				strainDataItemsLens.Get(d),
			)
		}),
		IOE.MapLeft[[][]string](func(err error) error {
			return fmt.Errorf("error in getting strain inventories %s", err)
		}),
	)
}

func (gh *githubIssue) fetchIssueStrainInfo(
	d strainData,
) IOE.IOEither[error, [][]string] {
	return F.Pipe1(
		IOE.TryCatchError(func() ([][]string, error) {
			return ghAnnoLens.Get(*gh).GetStrainInfo(
				strainDataItemsLens.Get(d),
			)
		}),
		IOE.MapLeft[[][]string](func(err error) error {
			return fmt.Errorf("error in getting strain information %s", err)
		}),
	)
}

func (gh *githubIssue) loadStrainData(
	s issueCreateState,
) IOE.IOEither[error, *strainData] {
	return F.Pipe4(
		IOE.Of[error](strainData{}),
		IOE.Bind(
			strainDataItemsLens.Set,
			F.Constant1[strainData](gh.fetchIssueStrains(s)),
		),
		IOE.Bind(strainDataInvsLens.Set, gh.fetchIssueStrainInvs),
		IOE.Bind(strainDataInfoLens.Set, gh.fetchIssueStrainInfo),
		IOE.Map[error](func(d strainData) *strainData { return new(d) }),
	)
}

func (gh *githubIssue) fetchIssuePlasmids(
	s issueCreateState,
) IOE.IOEither[error, []*stock.Plasmid] {
	return F.Pipe1(
		IOE.TryCatchError(func() ([]*stock.Plasmid, error) {
			return ghStockLens.Get(*gh).GetPlasmids(
				ghStockLens.Get(*gh).StocksFromItems(
					createStateOrderLens.Get(s),
					"DBP",
				),
			)
		}),
		IOE.MapLeft[[]*stock.Plasmid](func(err error) error {
			return fmt.Errorf("error in getting plasmids %s", err)
		}),
	)
}

func (gh *githubIssue) fetchIssuePlasmidInvs(
	d plasmidData,
) IOE.IOEither[error, [][]string] {
	return F.Pipe1(
		IOE.TryCatchError(func() ([][]string, error) {
			return ghAnnoLens.Get(*gh).GetPlasmidInv(
				plasmidDataItemsLens.Get(d),
			)
		}),
		IOE.MapLeft[[][]string](func(err error) error {
			return fmt.Errorf("error in getting plasmid inventories %s", err)
		}),
	)
}

func (gh *githubIssue) loadPlasmidData(
	s issueCreateState,
) IOE.IOEither[error, *plasmidData] {
	return F.Pipe3(
		IOE.Of[error](plasmidData{}),
		IOE.Bind(
			plasmidDataItemsLens.Set,
			F.Constant1[plasmidData](gh.fetchIssuePlasmids(s)),
		),
		IOE.Bind(plasmidDataInvsLens.Set, gh.fetchIssuePlasmidInvs),
		IOE.Map[error](func(d plasmidData) *plasmidData { return new(d) }),
	)
}

func (gh *githubIssue) buildIssueContent(
	s issueCreateState,
) *template.IssueContent {
	return F.Pipe9(
		&template.IssueContent{},
		issueContentBaseLens.Set(&template.Content{}),
		issueContentShipperLens.Set(stateShipperInfo(s)),
		issueContentPayerLens.Set(statePayerInfo(s)),
		issueContentOrderLens.Set(createStateOrderLens.Get(s)),
		issueContentStrainPriceLens.Set(ghStrainPriceLens.Get(*gh)),
		issueContentPlasmidPriceLens.Set(ghPlasmidPriceLens.Get(*gh)),
		issueContentStrainInvLens.Set(stateStrainInvRows(s)),
		issueContentPlasmidInvLens.Set(statePlasmidInvRows(s)),
		issueContentStrainInfoLens.Set(stateStrainInfoRows(s)),
	)
}
