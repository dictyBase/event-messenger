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
	S "github.com/IBM/fp-go/v2/string"
	"github.com/dictyBase/event-messenger/internal/datasource"
	issue "github.com/dictyBase/event-messenger/internal/issue-tracker"
	"github.com/dictyBase/event-messenger/internal/template"
	ioeutils "github.com/dictyBase/fp-go-loom/ioeitherutils"
	"github.com/dictyBase/fp-go-loom/strutils"
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
	Order   *order.Order
	Data    *allData
	Body    *bytes.Buffer
	Labels  []string
	Request *github.IssueRequest
}

// githubIssue includes all of the data needed to create an issue.
type githubIssue struct {
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

var (
	IssueParamsTokenLens = L.MakeLensRef(
		func(p *IssueParams) string { return p.Token },
		func(p *IssueParams, v string) *IssueParams {
			p.Token = v
			return p
		},
	)

	IssueParamsOwnerLens = L.MakeLensRef(
		func(p *IssueParams) string { return p.Owner },
		func(p *IssueParams, v string) *IssueParams {
			p.Owner = v
			return p
		},
	)

	IssueParamsRepoLens = L.MakeLensRef(
		func(p *IssueParams) string { return p.Repository },
		func(p *IssueParams, v string) *IssueParams {
			p.Repository = v
			return p
		},
	)

	IssueParamsOpenRouterLens = L.MakeLensRef(
		func(p *IssueParams) string { return p.OpenRouterToken },
		func(p *IssueParams, v string) *IssueParams {
			p.OpenRouterToken = v
			return p
		},
	)

	IssueParamsJevModelLens = L.MakeLensRef(
		func(p *IssueParams) string { return p.JevModel },
		func(p *IssueParams, v string) *IssueParams {
			p.JevModel = v
			return p
		},
	)

	IssueParamsDecisionsLens = L.MakeLensRef(
		func(p *IssueParams) DecisionsClient { return p.Decisions },
		func(p *IssueParams, v DecisionsClient) *IssueParams {
			p.Decisions = v
			return p
		},
	)

	IssueParamsStrainPriceLens = L.MakeLensRef(
		func(p *IssueParams) int { return p.StrainPrice },
		func(p *IssueParams, v int) *IssueParams {
			p.StrainPrice = v
			return p
		},
	)

	IssueParamsPlasmidPriceLens = L.MakeLensRef(
		func(p *IssueParams) int { return p.PlasmidPrice },
		func(p *IssueParams, v int) *IssueParams {
			p.PlasmidPrice = v
			return p
		},
	)

	IssueParamsLoggerLens = L.MakeLensRef(
		func(p *IssueParams) *logrus.Entry { return p.Logger },
		func(p *IssueParams, v *logrus.Entry) *IssueParams {
			p.Logger = v
			return p
		},
	)

	IssueParamsSourcesLens = L.MakeLensRef(
		func(p *IssueParams) *datasource.Sources { return p.Sources },
		func(p *IssueParams, v *datasource.Sources) *IssueParams {
			p.Sources = v
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

	createStateReqLens = L.MakeLens(
		func(s issueCreateState) *github.IssueRequest { return s.Request },
		func(s issueCreateState, v *github.IssueRequest) issueCreateState {
			s.Request = v
			return s
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

	formatOrderTitle = F.Pipe1(
		[]func(issueCreateState) string{
			F.Constant1[issueCreateState]("Order ID:"),
			F.Flow3(
				createStateOrderLens.Get,
				(*order.Order).GetData,
				(*order.Order_Data).GetId,
			),
			F.Constant1[issueCreateState](" "),
			F.Flow2(stateOrderAttrs, (*order.OrderAttributes).GetPurchaser),
		},
		strutils.ConcatFields[issueCreateState](S.Monoid),
	)
)

func stateOrderAttrs(s issueCreateState) *order.OrderAttributes {
	return F.Pipe3(
		s,
		createStateOrderLens.Get,
		(*order.Order).GetData,
		(*order.Order_Data).GetAttributes,
	)
}

// NewIssueCreator acts as a constructor for Github issue creation.
func NewIssueCreator(args *IssueParams) issue.Tracker {
	return F.Pipe11(
		githubIssue{},
		F.Pipe5(
			&oauth2.Token{},
			F.Pipe2(args, IssueParamsTokenLens.Get, oauthTokenAccessLens.Set),
			oauth2.StaticTokenSource,
			F.Bind1st(oauth2.NewClient, context.Background()),
			github.NewClient,
			ghClientLens.Set,
		),
		F.Pipe2(args, IssueParamsOwnerLens.Get, ghOwnerLens.Set),
		F.Pipe2(args, IssueParamsRepoLens.Get, ghRepoLens.Set),
		F.Pipe2(args, IssueParamsLoggerLens.Get, ghLoggerLens.Set),
		ghAnnoLens.Set(args.AnnoSource),
		ghStockLens.Set(args.StockSource),
		F.Pipe2(args, IssueParamsStrainPriceLens.Get, ghStrainPriceLens.Set),
		F.Pipe2(args, IssueParamsPlasmidPriceLens.Get, ghPlasmidPriceLens.Set),
		F.Pipe2(args, resolveParamsDecisions, ghDecisionsLens.Set),
		F.Pipe3(
			args,
			IssueParamsJevModelLens.Get,
			resolveJevModel,
			ghJevModelLens.Set,
		),
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
		IOE.Let[error](createStateReqLens.Set, prepareIssueRequest),
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
		IOE.Map[error](F.Ref[allData]),
	)
}

func (gh *githubIssue) renderIssueBody(
	s issueCreateState,
) IOE.IOEither[error, *bytes.Buffer] {
	return F.Pipe12(
		&template.IssueContent{},
		issueContentBaseLens.Set(&template.Content{}),
		F.Pipe3(
			s,
			stateOrderAttrs,
			(*order.OrderAttributes).GetConsumerInfo,
			issueContentShipperLens.Set,
		),
		F.Pipe3(
			s,
			stateOrderAttrs,
			(*order.OrderAttributes).GetPayerInfo,
			issueContentPayerLens.Set,
		),
		F.Pipe2(s, createStateOrderLens.Get, issueContentOrderLens.Set),
		F.Pipe2(*gh, ghStrainPriceLens.Get, issueContentStrainPriceLens.Set),
		F.Pipe2(*gh, ghPlasmidPriceLens.Get, issueContentPlasmidPriceLens.Set),
		F.Pipe3(
			s,
			createStateDataLens.Get,
			extractStrainInvRows,
			issueContentStrainInvLens.Set,
		),
		F.Pipe3(
			s,
			createStateDataLens.Get,
			extractPlasmidInvRows,
			issueContentPlasmidInvLens.Set,
		),
		F.Pipe3(
			s,
			createStateDataLens.Get,
			extractStrainInfoRows,
			issueContentStrainInfoLens.Set,
		),
		F.ToAny[*template.IssueContent],
		F.Pipe1(&template.OutputParams{}, F.Flip(outputParamsContentLens.Set)),
		IOE.Eitherize1(template.OutputText),
	)
}

func prepareIssueRequest(s issueCreateState) *github.IssueRequest {
	return F.Pipe3(
		&github.IssueRequest{},
		F.Pipe3(s, formatOrderTitle, F.Ref[string], issueReqTitleLens.Set),
		F.Pipe4(
			s,
			createStateBodyLens.Get,
			(*bytes.Buffer).String,
			F.Ref[string],
			issueReqBodyLens.Set,
		),
		F.Pipe3(
			s,
			createStateLabelsLens.Get,
			F.Ref[[]string],
			issueReqLabelsLens.Set,
		),
	)
}

func (gh *githubIssue) executePostIssue(
	s issueCreateState,
) IOE.IOEither[error, string] {
	return F.Pipe2(
		IOE.TryCatchError(func() (*github.Issue, error) {
			return gh.postIssue(createStateReqLens.Get(s))
		}),
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
		IOE.Map[error](F.Ref[strainData]),
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
		IOE.Map[error](F.Ref[plasmidData]),
	)
}
