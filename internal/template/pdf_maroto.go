package template

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"strconv"
	"strings"
	"sync"

	"github.com/dictyBase/event-messenger/internal/datasource"
	"github.com/dictyBase/go-genproto/dictybaseapis/order"
	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	mimage "github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/line"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/extension"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/pagesize"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/core/entity"
	"github.com/johnfercher/maroto/v2/pkg/fontrepository"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

const (
	orderInfoURL        = "https://dictycr.org/stockcenter/information/order/"
	orderPaymentInfoURL = "https://dictycr.org/stockcenter/information/payment"

	itemColumn        = "Item"
	quantityColumn    = "Quantity"
	totalColumn       = "Total"
	systematicNameCol = "Systematic Name"
	descriptorColumn  = "Descriptor"

	fontFamilySans  = "dossier-sans"
	fontFamilySerif = "dossier-serif"
	fontFamilyMono  = "dossier-mono"

	bgWidthPx       = 840
	bgHeightPx      = 1188
	bgTopBannerPx   = 18
	bgSidebarWidth  = 280
	sidebarColUnits = 4
	mainColUnits    = 8
	halfMainCols    = 4
	topCardSlotMax  = 2

	leftSidebarPad = 10.0
	mainLeftPad    = 9.0
	rightEdgePad   = 12.0

	panelWidthPx       = 600
	panelHeightPx      = 118
	panelAccentPx      = 4
	panelDividerY      = 34
	panelDividerStart  = 24
	panelDividerEnd    = 576
	tableHeaderHeight  = 26
	sidebarRuleWidthPx = 200

	alphaOpaque = 255

	brandR, brandG, brandB             = 40, 72, 147
	darkInkR, darkInkG, darkInkB       = 26, 35, 50
	mutedR, mutedG, mutedB             = 72, 86, 106
	ruleR, ruleG, ruleB                = 206, 215, 228
	sideRuleR, sideRuleG, sideRuleB    = 212, 220, 233
	sidebarBgR, sidebarBgG, sidebarBgB = 234, 239, 247
	navyR, navyG, navyB                = 30, 43, 88
	cardFillR, cardFillG, cardFillB    = 239, 243, 250
	accentR, accentG, accentB          = 63, 90, 169
	thFillR, thFillG, thFillB          = 229, 236, 246
	thBorderR, thBorderG, thBorderB    = 196, 208, 226

	defaultFontSize     = 9.2
	brandTitleFontSize  = 12.0
	sideHeaderFontSize  = 15.5
	heroTitleFontSize   = 31.0
	sectionFontSize     = 16.0
	orderSubFontSize    = 15.5
	commentHdrFontSize  = 14.5
	orderMetaFontSize   = 10.2
	dateMetaFontSize    = 10.0
	totalLabelFontSize  = 9.5
	totalValFontSize    = 11.0
	linkFontSize        = 9.0
	addrLabelFontSize   = 8.8
	bodyLineFontSize    = 8.8
	emailFontSize       = 8.5
	tableHdrFontSize    = 8.5
	cardIDLabelFontSize = 8.2
	cardMetaHdrFontSize = 8.0
	citationFontSize    = 8.3

	topSpacerRowH     = 13.0
	brandHeaderRowH   = 6.0
	heroHeaderRowH    = 24.0
	metaRowH          = 8.0
	dividerRowH       = 5.0
	shipAndStocksRowH = 45.0
	billAndFirstRowH  = 68.0
	cardGapRowH       = 3.0
	sectionGapRowH    = 4.0
	sectionTitleRowH  = 9.0
	thanksRowH        = 8.0
	bottomRuleRowH    = 6.0
	linkRowH          = 6.0
	strainCardRowH    = 27.0
	plasmidCardRowH   = 25.0

	lineThickPrimary   = 0.35
	lineThickThin      = 0.25
	lineOffsetTopRule  = 45.0
	lineOffsetMid      = 50.0
	sidebarRulePercent = 72.0
	mainRulePercent    = 86.0

	sideTitleTop1   = 2.0
	sideTitleTop2   = 8.8
	heroTitleTop2   = 11.5
	metaRowTop      = 2.0
	sectionTitleTop = 1.0
	thanksTextTop   = 2.0
	linkTextTop     = 0.5
	rightTextPad    = 4.0

	shipNameTop       = 1.0
	shipLinesStartTop = 6.0
	addrLineStep      = 4.3
	emailLineStep     = 4.5
	shipRuleOffset    = 7.2

	billNameTop       = 6.5
	billLinesStartTop = 11.5
	billRuleOffset    = 6.5
	billCommentStep   = 11.0
	commentBodyOffset = 6.5

	tableHdrRectTop     = 8.8
	tableHdrTextTop     = 10.5
	tableBodyStartTop   = 17.5
	tableRowStep        = 6.2
	totalRuleOffset     = 1.2
	totalLabelOffset    = 1.0
	totalValOffset      = 0.5
	tableLeftPadOffset  = 3.0
	tableRightPadOffset = 2.0
	qtyHdrLeft          = 54.0
	unitPriceHdrLeft    = 86.0
	qtyValLeft          = 63.0
	unitPriceValLeft    = 101.0

	firstCardTop        = 10.0
	secondStrainCardTop = 39.0
	secondPlasmidTop    = 37.0
	standaloneCardTop   = 1.5
	cardZeroTop         = 0.0

	cardInnerPad        = 3.5
	cardIDValPad        = 12.5
	cardIDLabelTop      = 2.5
	cardIDValTop        = 2.2
	cardMetaHdrTop      = 9.0
	cardMetaValTop      = 13.5
	strainCiteTopOffset = 19.5
	plasmidCiteTopOff   = 18.5
	strainNamesLeft     = 50.0
	strainSysNameLeft   = 86.0
	citeAuthorPad       = 20.0
	citeFirstLinkLeft   = 74.0
	citeLinkStep        = 15.0
	citeRowStep         = 4.2
)

var (
	//go:embed fonts/sans.ttf
	sansRegularTTF []byte
	//go:embed fonts/sans-bold.ttf
	sansBoldTTF []byte
	//go:embed fonts/georgia.ttf
	serifRegularTTF []byte
	//go:embed fonts/georgia-bold.ttf
	serifBoldTTF []byte
	//go:embed fonts/mono.ttf
	monoRegularTTF []byte

	dossierAssetsOnce    sync.Once
	dossierBackgroundPNG []byte
	cardPanelPNG         []byte
	tableHeaderPNG       []byte
	thinRulePNG          []byte
	sidebarRulePNG       []byte
	dossierCustomFonts   []entity.CustomFont

	indigoBrandColor = &props.Color{Red: brandR, Green: brandG, Blue: brandB}
	darkInkColor     = &props.Color{Red: darkInkR, Green: darkInkG, Blue: darkInkB}
	mutedSlateColor  = &props.Color{Red: mutedR, Green: mutedG, Blue: mutedB}
	ruleSlateColor   = &props.Color{Red: ruleR, Green: ruleG, Blue: ruleB}
)

func initDossierAssets() {
	dossierAssetsOnce.Do(func() {
		dossierBackgroundPNG = encodePageBackgroundPNG()
		cardPanelPNG = encodeCardPanelPNG()
		tableHeaderPNG = encodeTableHeaderPNG()
		thinRulePNG = encodeRulePNG(
			panelWidthPx,
			color.RGBA{R: ruleR, G: ruleG, B: ruleB, A: alphaOpaque},
		)
		sidebarRulePNG = encodeRulePNG(
			sidebarRuleWidthPx,
			color.RGBA{R: sideRuleR, G: sideRuleG, B: sideRuleB, A: alphaOpaque},
		)

		fonts, err := fontrepository.New().
			AddUTF8FontFromBytes(fontFamilySans, fontstyle.Normal, sansRegularTTF).
			AddUTF8FontFromBytes(fontFamilySans, fontstyle.Bold, sansBoldTTF).
			AddUTF8FontFromBytes(fontFamilySerif, fontstyle.Normal, serifRegularTTF).
			AddUTF8FontFromBytes(fontFamilySerif, fontstyle.Bold, serifBoldTTF).
			AddUTF8FontFromBytes(fontFamilyMono, fontstyle.Normal, monoRegularTTF).
			AddUTF8FontFromBytes(fontFamilyMono, fontstyle.Bold, monoRegularTTF).
			Load()
		if err == nil {
			dossierCustomFonts = fonts
		}
	})
}

func encodePageBackgroundPNG() []byte {
	img := image.NewRGBA(image.Rect(0, 0, bgWidthPx, bgHeightPx))
	white := color.RGBA{R: alphaOpaque, G: alphaOpaque, B: alphaOpaque, A: alphaOpaque}
	sidebar := color.RGBA{R: sidebarBgR, G: sidebarBgG, B: sidebarBgB, A: alphaOpaque}
	navy := color.RGBA{R: navyR, G: navyG, B: navyB, A: alphaOpaque}

	for y := range bgHeightPx {
		for x := range bgWidthPx {
			switch {
			case y < bgTopBannerPx:
				img.SetRGBA(x, y, navy)
			case x < bgSidebarWidth:
				img.SetRGBA(x, y, sidebar)
			default:
				img.SetRGBA(x, y, white)
			}
		}
	}

	var buf bytes.Buffer

	_ = png.Encode(&buf, img)

	return buf.Bytes()
}

func encodeCardPanelPNG() []byte {
	img := image.NewRGBA(image.Rect(0, 0, panelWidthPx, panelHeightPx))
	fill := color.RGBA{R: cardFillR, G: cardFillG, B: cardFillB, A: alphaOpaque}
	accent := color.RGBA{R: accentR, G: accentG, B: accentB, A: alphaOpaque}
	divider := color.RGBA{R: ruleR, G: ruleG, B: ruleB, A: alphaOpaque}

	for y := range panelHeightPx {
		for x := range panelWidthPx {
			switch {
			case x < panelAccentPx:
				img.SetRGBA(x, y, accent)
			case y == panelDividerY && x >= panelDividerStart && x <= panelDividerEnd:
				img.SetRGBA(x, y, divider)
			default:
				img.SetRGBA(x, y, fill)
			}
		}
	}

	var buf bytes.Buffer

	_ = png.Encode(&buf, img)

	return buf.Bytes()
}

func encodeTableHeaderPNG() []byte {
	img := image.NewRGBA(image.Rect(0, 0, panelWidthPx, tableHeaderHeight))
	fill := color.RGBA{R: thFillR, G: thFillG, B: thFillB, A: alphaOpaque}
	borderBottom := color.RGBA{R: thBorderR, G: thBorderG, B: thBorderB, A: alphaOpaque}

	for y := range tableHeaderHeight {
		for x := range panelWidthPx {
			if y == tableHeaderHeight-1 {
				img.SetRGBA(x, y, borderBottom)
			} else {
				img.SetRGBA(x, y, fill)
			}
		}
	}

	var buf bytes.Buffer

	_ = png.Encode(&buf, img)

	return buf.Bytes()
}

func encodeRulePNG(widthPx int, c color.RGBA) []byte {
	img := image.NewRGBA(image.Rect(0, 0, widthPx, 1))
	for x := range widthPx {
		img.SetRGBA(x, 0, c)
	}

	var buf bytes.Buffer

	_ = png.Encode(&buf, img)

	return buf.Bytes()
}

// BuildOrderMaroto constructs the Indigo Dossier Maroto v2 document tree for an order.
func BuildOrderMaroto(ec *EmailContent) core.Maroto {
	initDossierAssets()

	cfg := config.NewBuilder().
		WithPageSize(pagesize.A4).
		WithLeftMargin(0).
		WithRightMargin(0).
		WithTopMargin(0).
		WithBottomMargin(0).
		WithCustomFonts(dossierCustomFonts).
		WithDefaultFont(&props.Font{
			Family: fontFamilySans,
			Size:   defaultFontSize,
			Color:  darkInkColor,
		}).
		WithBackgroundImage(dossierBackgroundPNG, extension.Png).
		Build()

	m := maroto.New(cfg)
	m.AddRows(row.New(topSpacerRowH))
	m.AddRows(buildTopSectionRows(ec)...)
	m.AddRows(buildBottomSectionRows(ec)...)

	return m
}

func buildTopSectionRows(ec *EmailContent) []core.Row {
	ordAttr := ec.Order.GetData().GetAttributes()
	ordID := ec.Order.GetData().GetId()

	return []core.Row{
		row.New(brandHeaderRowH).Add(
			line.NewCol(sidebarColUnits, props.Line{
				Color:         indigoBrandColor,
				Thickness:     lineThickPrimary,
				OffsetPercent: lineOffsetTopRule,
				SizePercent:   sidebarRulePercent,
			}),
			text.NewCol(mainColUnits, "Dicty Stock Center", props.Text{
				Size:  brandTitleFontSize,
				Style: fontstyle.Bold,
				Color: indigoBrandColor,
				Left:  mainLeftPad,
			}),
		),
		row.New(heroHeaderRowH).Add(
			col.New(sidebarColUnits).Add(
				text.New("Shipping and Billing", props.Text{
					Family: fontFamilySerif,
					Size:   sideHeaderFontSize,
					Color:  darkInkColor,
					Top:    sideTitleTop1,
					Left:   leftSidebarPad,
				}),
				text.New("Information", props.Text{
					Family: fontFamilySerif,
					Size:   sideHeaderFontSize,
					Color:  darkInkColor,
					Top:    sideTitleTop2,
					Left:   leftSidebarPad,
				}),
			),
			col.New(mainColUnits).Add(
				text.New("Order", props.Text{
					Family: fontFamilySerif,
					Size:   heroTitleFontSize,
					Color:  darkInkColor,
					Left:   mainLeftPad,
				}),
				text.New("Confirmation", props.Text{
					Family: fontFamilySerif,
					Size:   heroTitleFontSize,
					Color:  darkInkColor,
					Top:    heroTitleTop2,
					Left:   mainLeftPad,
				}),
			),
		),
		row.New(metaRowH).Add(
			text.NewCol(sidebarColUnits, "Shipping Address", props.Text{
				Size:  addrLabelFontSize,
				Style: fontstyle.Bold,
				Color: indigoBrandColor,
				Top:   metaRowTop,
				Left:  leftSidebarPad,
			}),
			text.NewCol(halfMainCols, fmt.Sprintf("Order #%s", ordID), props.Text{
				Size:  orderMetaFontSize,
				Style: fontstyle.Bold,
				Color: darkInkColor,
				Top:   metaRowTop,
				Left:  mainLeftPad,
			}),
			text.NewCol(halfMainCols, ec.OrderTimestamp(), props.Text{
				Size:  dateMetaFontSize,
				Color: mutedSlateColor,
				Align: align.Right,
				Top:   metaRowTop,
				Right: rightEdgePad,
			}),
		),
		row.New(dividerRowH).Add(
			col.New(sidebarColUnits),
			line.NewCol(mainColUnits, props.Line{
				Color:         indigoBrandColor,
				Thickness:     lineThickPrimary,
				OffsetPercent: lineOffsetMid,
				SizePercent:   mainRulePercent,
			}),
		),
		buildAddressAndStocksRow(ec, ordAttr),
		buildBillingAndCommentRow(ec, ordAttr),
	}
}

func buildAddressAndStocksRow(
	ec *EmailContent,
	ordAttr *order.OrderAttributes,
) core.Row {
	shipCol := buildAddressColumn(
		ec.Shipper,
		ordAttr.GetConsumer(),
		fmt.Sprintf("%s %s", ordAttr.GetCourier(), ordAttr.GetCourierAccount()),
	)
	rightCol := buildStocksOrderedColumn(ec)

	return row.New(shipAndStocksRowH).Add(shipCol, rightCol)
}

func buildAddressColumn(
	info *order.UserInfo,
	emailAddr string,
	footerLine string,
) core.Col {
	fullName := strings.TrimSpace(
		fmt.Sprintf("%s %s", info.GetFirstName(), info.GetLastName()),
	)
	cityLine := strings.TrimSpace(
		fmt.Sprintf(
			"%s %s %s",
			info.GetCity(),
			info.GetState(),
			info.GetZipcode(),
		),
	)

	lines := []string{
		info.GetOrganization(),
		info.GetFirstAddress(),
	}
	if second := strings.TrimSpace(info.GetSecondAddress()); second != "" {
		lines = append(lines, second)
	}

	lines = append(lines, cityLine, info.GetCountry(), info.GetPhone())

	components := []core.Component{
		text.New(fullName, props.Text{
			Size:  orderMetaFontSize,
			Style: fontstyle.Bold,
			Color: darkInkColor,
			Left:  leftSidebarPad,
			Top:   shipNameTop,
		}),
	}
	topOffset := shipLinesStartTop

	for _, lineText := range lines {
		if strings.TrimSpace(lineText) == "" {
			continue
		}

		components = append(components, text.New(lineText, props.Text{
			Size:  bodyLineFontSize,
			Color: darkInkColor,
			Top:   topOffset,
			Left:  leftSidebarPad,
			Right: rightTextPad,
		}))
		topOffset += addrLineStep
	}

	if trimmedEmail := strings.TrimSpace(emailAddr); trimmedEmail != "" {
		mailto := "mailto:" + trimmedEmail
		components = append(components, text.New(trimmedEmail, props.Text{
			Size:      emailFontSize,
			Top:       topOffset,
			Left:      leftSidebarPad,
			Right:     rightTextPad,
			Hyperlink: &mailto,
		}))
		topOffset += emailLineStep
	}

	if trimmedFooter := strings.TrimSpace(footerLine); trimmedFooter != "" {
		components = append(
			components,
			text.New(trimmedFooter, props.Text{
				Size:  bodyLineFontSize,
				Color: darkInkColor,
				Top:   topOffset,
				Left:  leftSidebarPad,
				Right: rightTextPad,
			}),
			mimage.NewFromBytes(sidebarRulePNG, extension.Png, props.Rect{
				Left:               leftSidebarPad,
				Top:                topOffset + shipRuleOffset,
				Percent:            sidebarRulePercent,
				JustReferenceWidth: true,
			}),
		)
	}

	return col.New(sidebarColUnits).Add(components...)
}

func buildStocksOrderedColumn(ec *EmailContent) core.Col {
	components := []core.Component{
		text.New("Stocks Ordered", props.Text{
			Family: fontFamilySerif,
			Size:   sectionFontSize,
			Color:  darkInkColor,
			Left:   mainLeftPad,
			Top:    sectionTitleTop,
		}),
		mimage.NewFromBytes(tableHeaderPNG, extension.Png, props.Rect{
			Left:               mainLeftPad,
			Top:                tableHdrRectTop,
			Percent:            mainRulePercent,
			JustReferenceWidth: true,
		}),
		text.New(itemColumn, props.Text{
			Size:  tableHdrFontSize,
			Style: fontstyle.Bold,
			Color: mutedSlateColor,
			Left:  mainLeftPad + tableLeftPadOffset,
			Top:   tableHdrTextTop,
		}),
		text.New(quantityColumn, props.Text{
			Size:  tableHdrFontSize,
			Style: fontstyle.Bold,
			Color: mutedSlateColor,
			Left:  qtyHdrLeft,
			Top:   tableHdrTextTop,
		}),
		text.New("Unit Price ($)", props.Text{
			Size:  tableHdrFontSize,
			Style: fontstyle.Bold,
			Color: mutedSlateColor,
			Left:  unitPriceHdrLeft,
			Top:   tableHdrTextTop,
		}),
		text.New("Total ($)", props.Text{
			Size:  tableHdrFontSize,
			Style: fontstyle.Bold,
			Color: mutedSlateColor,
			Align: align.Right,
			Right: rightEdgePad + tableRightPadOffset,
			Top:   tableHdrTextTop,
		}),
	}

	rowTop := tableBodyStartTop
	if len(ec.StrainData) > 0 {
		components = append(
			components,
			stockSummaryComponents(
				"Strain",
				ec.StrainItems(),
				ec.StrainPrice,
				ec.StrainCost(),
				rowTop,
			)...,
		)
		rowTop += tableRowStep
	}

	if len(ec.PlasmidData) > 0 {
		components = append(
			components,
			stockSummaryComponents(
				"Plasmid",
				ec.PlasmidItems(),
				ec.PlasmidPrice,
				ec.PlasmidCost(),
				rowTop,
			)...,
		)
		rowTop += tableRowStep
	}

	components = append(
		components,
		mimage.NewFromBytes(thinRulePNG, extension.Png, props.Rect{
			Left:               mainLeftPad,
			Top:                rowTop - totalRuleOffset,
			Percent:            mainRulePercent,
			JustReferenceWidth: true,
		}),
		text.New(totalColumn, props.Text{
			Size:  totalLabelFontSize,
			Style: fontstyle.Bold,
			Color: darkInkColor,
			Left:  mainLeftPad + tableLeftPadOffset,
			Top:   rowTop + totalLabelOffset,
		}),
		text.New(strconv.Itoa(ec.TotalCost()), props.Text{
			Size:  totalValFontSize,
			Style: fontstyle.Bold,
			Color: indigoBrandColor,
			Align: align.Right,
			Right: rightEdgePad + tableRightPadOffset,
			Top:   rowTop + totalValOffset,
		}),
	)

	return col.New(mainColUnits).Add(components...)
}

func stockSummaryComponents(
	label string,
	qty int,
	unitPrice int,
	cost int,
	top float64,
) []core.Component {
	return []core.Component{
		text.New(label, props.Text{
			Size:  defaultFontSize,
			Color: darkInkColor,
			Left:  mainLeftPad + tableLeftPadOffset,
			Top:   top,
		}),
		text.New(strconv.Itoa(qty), props.Text{
			Size:  defaultFontSize,
			Color: darkInkColor,
			Left:  qtyValLeft,
			Top:   top,
		}),
		text.New(strconv.Itoa(unitPrice), props.Text{
			Size:  defaultFontSize,
			Color: darkInkColor,
			Left:  unitPriceValLeft,
			Top:   top,
		}),
		text.New(strconv.Itoa(cost), props.Text{
			Size:  defaultFontSize,
			Color: darkInkColor,
			Align: align.Right,
			Right: rightEdgePad + tableRightPadOffset,
			Top:   top,
		}),
	}
}

func buildBillingAndCommentRow(
	ec *EmailContent,
	ordAttr *order.OrderAttributes,
) core.Row {
	leftCol := buildBillingAndCommentColumn(ec, ordAttr)
	rightCol := buildFirstStockSectionColumn(ec)

	return row.New(billAndFirstRowH).Add(leftCol, rightCol)
}

func buildBillingAndCommentColumn(
	ec *EmailContent,
	ordAttr *order.OrderAttributes,
) core.Col {
	info := ec.Payer
	fullName := strings.TrimSpace(
		fmt.Sprintf("%s %s", info.GetFirstName(), info.GetLastName()),
	)
	cityLine := strings.TrimSpace(
		fmt.Sprintf(
			"%s %s %s",
			info.GetCity(),
			info.GetState(),
			info.GetZipcode(),
		),
	)

	components := []core.Component{
		text.New("Billing Address", props.Text{
			Size:  addrLabelFontSize,
			Style: fontstyle.Bold,
			Color: indigoBrandColor,
			Top:   sectionTitleTop,
			Left:  leftSidebarPad,
		}),
		text.New(fullName, props.Text{
			Size:  orderMetaFontSize,
			Style: fontstyle.Bold,
			Color: darkInkColor,
			Top:   billNameTop,
			Left:  leftSidebarPad,
		}),
	}

	lines := []string{
		info.GetOrganization(),
		info.GetFirstAddress(),
	}
	if second := strings.TrimSpace(info.GetSecondAddress()); second != "" {
		lines = append(lines, second)
	}

	lines = append(lines, cityLine, info.GetCountry(), info.GetPhone())

	topOffset := billLinesStartTop

	for _, lineText := range lines {
		if strings.TrimSpace(lineText) == "" {
			continue
		}

		components = append(components, text.New(lineText, props.Text{
			Size:  bodyLineFontSize,
			Color: darkInkColor,
			Top:   topOffset,
			Left:  leftSidebarPad,
			Right: rightTextPad,
		}))
		topOffset += addrLineStep
	}

	if payerAddr := strings.TrimSpace(ordAttr.GetPayer()); payerAddr != "" {
		mailto := "mailto:" + payerAddr
		components = append(components, text.New(payerAddr, props.Text{
			Size:      emailFontSize,
			Top:       topOffset,
			Left:      leftSidebarPad,
			Right:     rightTextPad,
			Hyperlink: &mailto,
		}))
		topOffset += emailLineStep
	}

	payLine := strings.TrimSpace(
		fmt.Sprintf("%s %s", ordAttr.GetPayment(), ordAttr.GetPurchaseOrderNum()),
	)
	if payLine != "" {
		components = append(
			components,
			text.New(payLine, props.Text{
				Size:  bodyLineFontSize,
				Color: darkInkColor,
				Top:   topOffset,
				Left:  leftSidebarPad,
				Right: rightTextPad,
			}),
			mimage.NewFromBytes(sidebarRulePNG, extension.Png, props.Rect{
				Left:               leftSidebarPad,
				Top:                topOffset + billRuleOffset,
				Percent:            sidebarRulePercent,
				JustReferenceWidth: true,
			}),
		)
		topOffset += billCommentStep
	}

	components = append(
		components,
		text.New("Comment", props.Text{
			Family: fontFamilySerif,
			Size:   commentHdrFontSize,
			Color:  darkInkColor,
			Top:    topOffset,
			Left:   leftSidebarPad,
		}),
	)
	if comment := strings.TrimSpace(ordAttr.GetComments()); comment != "" {
		components = append(
			components,
			text.New(comment, props.Text{
				Size:  bodyLineFontSize,
				Color: darkInkColor,
				Top:   topOffset + commentBodyOffset,
				Left:  leftSidebarPad,
				Right: rightTextPad,
			}),
		)
	}

	return col.New(sidebarColUnits).Add(components...)
}

func buildFirstStockSectionColumn(ec *EmailContent) core.Col {
	if len(ec.StrainData) > 0 {
		comps := []core.Component{
			text.New("Strain Information", props.Text{
				Family: fontFamilySerif,
				Size:   sectionFontSize,
				Color:  darkInkColor,
				Left:   mainLeftPad,
				Top:    sectionTitleTop,
			}),
		}
		comps = append(comps, strainCardComponents(ec.StrainData[0], firstCardTop, true)...)

		if len(ec.StrainData) > 1 {
			comps = append(
				comps,
				strainCardComponents(ec.StrainData[1], secondStrainCardTop, true)...,
			)
		}

		return col.New(mainColUnits).Add(comps...)
	}

	if len(ec.PlasmidData) > 0 {
		comps := []core.Component{
			text.New("Plasmid Information", props.Text{
				Family: fontFamilySerif,
				Size:   sectionFontSize,
				Color:  darkInkColor,
				Left:   mainLeftPad,
				Top:    sectionTitleTop,
			}),
		}
		comps = append(comps, plasmidCardComponents(ec.PlasmidData[0], firstCardTop, true)...)

		if len(ec.PlasmidData) > 1 {
			comps = append(
				comps,
				plasmidCardComponents(ec.PlasmidData[1], secondPlasmidTop, true)...,
			)
		}

		return col.New(mainColUnits).Add(comps...)
	}

	return col.New(mainColUnits)
}

func strainCardComponents(
	sr *StrainRows,
	baseTop float64,
	withPanelImage bool,
) []core.Component {
	var comps []core.Component
	if withPanelImage {
		comps = append(comps, mimage.NewFromBytes(cardPanelPNG, extension.Png, props.Rect{
			Left:               mainLeftPad,
			Top:                baseTop,
			Percent:            mainRulePercent,
			JustReferenceWidth: true,
		}))
	}

	comps = append(
		comps,
		text.New("ID", props.Text{
			Size:  cardIDLabelFontSize,
			Style: fontstyle.Bold,
			Color: indigoBrandColor,
			Left:  mainLeftPad + cardInnerPad,
			Top:   baseTop + cardIDLabelTop,
		}),
		text.New(sr.ID, props.Text{
			Family: fontFamilyMono,
			Size:   dateMetaFontSize,
			Color:  darkInkColor,
			Left:   mainLeftPad + cardIDValPad,
			Top:    baseTop + cardIDValTop,
		}),
		text.New(descriptorColumn, props.Text{
			Size:  cardMetaHdrFontSize,
			Style: fontstyle.Bold,
			Color: mutedSlateColor,
			Left:  mainLeftPad + cardInnerPad,
			Top:   baseTop + cardMetaHdrTop,
		}),
		text.New("Name(s)", props.Text{
			Size:  cardMetaHdrFontSize,
			Style: fontstyle.Bold,
			Color: mutedSlateColor,
			Left:  strainNamesLeft,
			Top:   baseTop + cardMetaHdrTop,
		}),
		text.New(systematicNameCol, props.Text{
			Size:  cardMetaHdrFontSize,
			Style: fontstyle.Bold,
			Color: mutedSlateColor,
			Left:  strainSysNameLeft,
			Top:   baseTop + cardMetaHdrTop,
		}),
		text.New(cleanBreakTags(sr.Descriptor), props.Text{
			Size:  bodyLineFontSize,
			Color: darkInkColor,
			Left:  mainLeftPad + cardInnerPad,
			Top:   baseTop + cardMetaValTop,
		}),
		text.New(cleanBreakTags(sr.Names), props.Text{
			Size:  bodyLineFontSize,
			Color: darkInkColor,
			Left:  strainNamesLeft,
			Top:   baseTop + cardMetaValTop,
		}),
		text.New(cleanBreakTags(sr.SysName), props.Text{
			Size:  bodyLineFontSize,
			Color: darkInkColor,
			Left:  strainSysNameLeft,
			Top:   baseTop + cardMetaValTop,
		}),
	)

	return append(comps, citationComponents(sr.PubInfo, baseTop+strainCiteTopOffset)...)
}

func plasmidCardComponents(
	pr *PlasmidRows,
	baseTop float64,
	withPanelImage bool,
) []core.Component {
	var comps []core.Component
	if withPanelImage {
		comps = append(comps, mimage.NewFromBytes(cardPanelPNG, extension.Png, props.Rect{
			Left:               mainLeftPad,
			Top:                baseTop,
			Percent:            mainRulePercent,
			JustReferenceWidth: true,
		}))
	}

	comps = append(
		comps,
		text.New("ID", props.Text{
			Size:  cardIDLabelFontSize,
			Style: fontstyle.Bold,
			Color: indigoBrandColor,
			Left:  mainLeftPad + cardInnerPad,
			Top:   baseTop + cardIDLabelTop,
		}),
		text.New(pr.ID, props.Text{
			Family: fontFamilyMono,
			Size:   dateMetaFontSize,
			Color:  darkInkColor,
			Left:   mainLeftPad + cardIDValPad,
			Top:    baseTop + cardIDValTop,
		}),
		text.New("Plasmid Name", props.Text{
			Size:  cardMetaHdrFontSize,
			Style: fontstyle.Bold,
			Color: mutedSlateColor,
			Left:  mainLeftPad + cardInnerPad,
			Top:   baseTop + cardMetaHdrTop,
		}),
		text.New(cleanBreakTags(pr.Name), props.Text{
			Size:  bodyLineFontSize,
			Color: darkInkColor,
			Left:  mainLeftPad + cardInnerPad,
			Top:   baseTop + cardMetaValTop,
		}),
	)

	return append(comps, citationComponents(pr.PubInfo, baseTop+plasmidCiteTopOff)...)
}

func citationComponents(
	pubs []*datasource.PubInfo,
	baseTop float64,
) []core.Component {
	if len(pubs) == 0 {
		return nil
	}

	comps := []core.Component{
		text.New("Citation", props.Text{
			Size:  cardMetaHdrFontSize,
			Style: fontstyle.Bold,
			Color: mutedSlateColor,
			Left:  mainLeftPad + cardInnerPad,
			Top:   baseTop,
		}),
	}

	offset := baseTop

	for _, pub := range pubs {
		if pub == nil {
			continue
		}

		if author := strings.TrimSpace(pub.AuthorStr); author != "" {
			comps = append(comps, text.New(author, props.Text{
				Size:  citationFontSize,
				Color: darkInkColor,
				Left:  mainLeftPad + citeAuthorPad,
				Top:   offset,
			}))
		}

		linkLeft := citeFirstLinkLeft

		if pubmed := strings.TrimSpace(pub.PubmedURL); pubmed != "" {
			href := pubmed
			comps = append(comps, text.New("Pubmed", props.Text{
				Size:      citationFontSize,
				Left:      linkLeft,
				Top:       offset,
				Hyperlink: &href,
			}))
			linkLeft += citeLinkStep
		}

		if doi := strings.TrimSpace(pub.DoiURL); doi != "" {
			href := doi
			comps = append(comps, text.New("Full text", props.Text{
				Size:      citationFontSize,
				Left:      linkLeft,
				Top:       offset,
				Hyperlink: &href,
			}))
		}

		offset += citeRowStep
	}

	return comps
}

func buildBottomSectionRows(ec *EmailContent) []core.Row {
	var rows []core.Row

	if len(ec.StrainData) > topCardSlotMax {
		for _, sr := range ec.StrainData[topCardSlotMax:] {
			rows = append(rows, row.New(cardGapRowH), buildStrainCardRow(sr))
		}
	}

	if len(ec.PlasmidData) > 0 {
		startIdx := 0
		if len(ec.StrainData) == 0 {
			startIdx = topCardSlotMax
		} else {
			rows = append(
				rows,
				row.New(sectionGapRowH),
				row.New(sectionTitleRowH).Add(
					col.New(sidebarColUnits),
					text.NewCol(mainColUnits, "Plasmid Information", props.Text{
						Family: fontFamilySerif,
						Size:   sectionFontSize,
						Color:  darkInkColor,
						Left:   mainLeftPad,
						Top:    sectionTitleTop,
					}),
				),
			)
		}

		for _, pr := range ec.PlasmidData[startIdx:] {
			rows = append(rows, row.New(cardGapRowH), buildPlasmidCardRow(pr))
		}

		rows = append(rows, row.New(thanksRowH).Add(
			col.New(sidebarColUnits),
			text.NewCol(
				mainColUnits,
				"Thank you for citing the researchers who submitted the materials you ordered!",
				props.Text{
					Size:  bodyLineFontSize,
					Style: fontstyle.Bold,
					Color: darkInkColor,
					Left:  mainLeftPad,
					Top:   thanksTextTop,
				},
			),
		))
	}

	orderHref := orderInfoURL
	paymentHref := orderPaymentInfoURL

	rows = append(
		rows,
		row.New(bottomRuleRowH).Add(
			col.New(sidebarColUnits),
			line.NewCol(mainColUnits, props.Line{
				Color:         ruleSlateColor,
				Thickness:     lineThickThin,
				OffsetPercent: lineOffsetMid,
				SizePercent:   mainRulePercent,
			}),
		),
		row.New(sectionTitleRowH).Add(
			col.New(sidebarColUnits),
			text.NewCol(mainColUnits, "Order Information", props.Text{
				Family: fontFamilySerif,
				Size:   orderSubFontSize,
				Color:  darkInkColor,
				Left:   mainLeftPad,
				Top:    sectionTitleTop,
			}),
		),
		row.New(linkRowH).Add(
			col.New(sidebarColUnits),
			text.NewCol(mainColUnits, "Order information", props.Text{
				Size:      linkFontSize,
				Left:      mainLeftPad,
				Top:       linkTextTop,
				Hyperlink: &orderHref,
			}),
		),
		row.New(linkRowH).Add(
			col.New(sidebarColUnits),
			text.NewCol(mainColUnits, "Order payment information", props.Text{
				Size:      linkFontSize,
				Left:      mainLeftPad,
				Top:       linkTextTop,
				Hyperlink: &paymentHref,
			}),
		),
	)

	return rows
}

func buildStrainCardRow(sr *StrainRows) core.Row {
	return row.New(strainCardRowH).Add(
		col.New(sidebarColUnits),
		col.New(mainColUnits).Add(strainCardComponents(sr, standaloneCardTop, true)...),
	)
}

func buildPlasmidCardRow(pr *PlasmidRows) core.Row {
	return row.New(plasmidCardRowH).Add(
		col.New(sidebarColUnits),
		col.New(mainColUnits).Add(plasmidCardComponents(pr, cardZeroTop, true)...),
	)
}

func cleanBreakTags(val string) string {
	replaced := strings.ReplaceAll(val, "<br/>", ", ")
	replaced = strings.ReplaceAll(replaced, "<br>", ", ")

	return strings.TrimSpace(replaced)
}
