package main

import (
	"bitbucket.org/zombiezen/gopdf/pdf"
	"bitbucket.org/zombiezen/greyhound-scouting/barcode"
	"fmt"
)

const reportMargin = 0.5 * pdf.Inch

const (
	matchNumberFontName = pdf.HelveticaBold
	matchNumberFontSize = 18

	scoreFontName = pdf.Helvetica
	scoreFontSize = 14

	barcodeFontName = pdf.Courier
	barcodeFontSize = 12
)

const scoutFormsPerPage = 3

func renderMultipleScoutForms(doc *pdf.Document, pageWidth, pageHeight pdf.Unit, event *Event, matches []*Match) {
	n := 0
	sizeX, sizeY := pageWidth-reportMargin*2, (pageHeight-reportMargin*2)/scoutFormsPerPage

	var canvas *pdf.Canvas
	for _, match := range matches {
		// Retrieve the list of teams (in sorted order, red first, then blue)
		teamList := make([]TeamInfo, 0, len(match.Teams))
		teamList = append(teamList, match.AllianceInfo(Red).Teams...)
		teamList = append(teamList, match.AllianceInfo(Blue).Teams...)

		for _, info := range teamList {
			if canvas == nil {
				canvas = doc.NewPage(pageWidth, pageHeight)
				canvas.Translate(reportMargin, pageHeight-sizeY-reportMargin)
			}
			renderScoutForm(canvas, sizeX, sizeY, event, match, info.Team)
			if n%scoutFormsPerPage == scoutFormsPerPage-1 {
				canvas.Close()
				canvas = nil
			} else {
				// Page divider
				// TODO: set dash
				canvas.DrawLine(pdf.Point{0, 0}, pdf.Point{sizeX, 0})
				canvas.Translate(0, -pageHeight/scoutFormsPerPage)
			}
			n++
		}
	}

	// If we're in the middle of a page, insert page break.
	if canvas != nil {
		canvas.Close()
	}
}

const (
	scoutFormAllianceLine = 1.0 * pdf.Inch
)

// this will assume that both position and margins have already been transformed for.
func renderScoutForm(canvas *pdf.Canvas, w, h pdf.Unit, event *Event, match *Match, teamNum int) {
	// Determine alliance
	var alliance Alliance
	for _, teamInfo := range match.Teams {
		if teamInfo.Team == teamNum {
			alliance = teamInfo.Alliance
			break
		}
	}
	if alliance == "" {
		// TODO: log error?
		return
	}

	// Match number
	baseline := h - matchNumberFontSize
	canvas.Push()
	canvas.Translate(0, baseline)
	text := new(pdf.Text)
	text.SetFont(pdf.HelveticaBold, matchNumberFontSize)
	// TODO: Em dash
	text.Text(fmt.Sprintf("%s #%d - %s", match.Type.DisplayName(), match.Number, event.Location.Name))
	text.NextLine()
	text.Text(fmt.Sprintf("Team %d", teamNum))
	canvas.DrawText(text)
	canvas.Pop()

	// Barcode
	bc := &barcode.Image{
		Barcode: barcode.Encode(MatchTeamTag{MatchTag{event.Tag(), match.Type, uint(match.Number)}, uint(teamNum)}.String()),
		Scale:   1,
		Height:  24,
	}
	var bcRect pdf.Rectangle
	bcRect.Min.X = w - pdf.Unit(bc.Bounds().Dx())
	bcRect.Min.Y = h - pdf.Unit(bc.Bounds().Dy())
	bcRect.Max.X = bcRect.Min.X + pdf.Unit(bc.Bounds().Dx())
	bcRect.Max.Y = bcRect.Min.Y + pdf.Unit(bc.Bounds().Dy())
	canvas.DrawImage(bc, bcRect)
	// TODO: Text

	// Scores
	headingBaseline := baseline + text.Y() - 0.25*pdf.Inch - scoreFontSize
	baseline = headingBaseline - 0.1*pdf.Inch - scoreFontSize
	formPt1 := renderFields(canvas, pdf.Point{0, baseline}, pdf.Helvetica, scoreFontSize, 0.5*pdf.Inch, "High:", "Mid:", "Low:")
	formPt2 := renderFields(canvas, pdf.Point{formPt1.X + 0.25*pdf.Inch, baseline}, pdf.Helvetica, scoreFontSize, 0.5*pdf.Inch, "High", "Mid:", "Low:")
	formPt3 := renderFields(canvas, pdf.Point{formPt2.X + 0.5*pdf.Inch, baseline}, pdf.Helvetica, scoreFontSize, 0.5*pdf.Inch, "Coop Attempt:", "Bridge 1 Attempt:", "Bridge 2 Attempt:")
	formPt4 := renderFields(canvas, pdf.Point{formPt3.X + 0.25*pdf.Inch, baseline}, pdf.Helvetica, scoreFontSize, 0.5*pdf.Inch, "Success:", "Success:", "Success:")
	_ = formPt4

	canvas.Push()
	canvas.Translate(0, headingBaseline)
	text = new(pdf.Text)
	text.SetFont(pdf.HelveticaBold, scoreFontSize)
	text.Text("Autonomous")
	canvas.DrawText(text)
	canvas.Pop()

	canvas.Push()
	canvas.Translate(formPt1.X+0.25*pdf.Inch, headingBaseline)
	text = new(pdf.Text)
	text.SetFont(pdf.HelveticaBold, scoreFontSize)
	text.Text("Teleop")
	canvas.DrawText(text)
	canvas.Pop()

	// Scout name
	// TODO: don't assume formPt1.Y is the lowest
	baseline += formPt1.Y - (scoreFontSize + 0.4*pdf.Inch)
	namePt := renderFields(canvas, pdf.Point{0, baseline}, pdf.Helvetica, scoreFontSize, 3.0*pdf.Inch, "Scout Name:")

	// Comments
	baseline += namePt.Y - (scoreFontSize + 0.05*pdf.Inch)
	canvas.Push()
	canvas.Translate(0, baseline)
	text = new(pdf.Text)
	text.SetFont(pdf.Helvetica, scoreFontSize)
	text.Text("Comments:")
	canvas.DrawText(text)
	canvas.Pop()
}

const (
	fieldLeading     = 0.1 * pdf.Inch
	fieldLinePadding = 0.125 * pdf.Inch
)

func renderFields(canvas *pdf.Canvas, pt pdf.Point, fontName string, fontSize pdf.Unit, lineLength pdf.Unit, labels ...string) pdf.Point {
	canvas.Push()
	defer canvas.Pop()
	canvas.Translate(pt.X, pt.Y)

	text := new(pdf.Text)
	text.SetFont(fontName, fontSize)
	leading := fontSize + fieldLeading
	text.SetLeading(leading)

	// Draw labels
	var rightSide pdf.Unit
	for _, label := range labels {
		if label != "" {
			text.Text(label)
			if text.X() > rightSide {
				rightSide = text.X()
			}
		}
		text.NextLine()
	}
	canvas.DrawText(text)

	// Draw field lines
	var baseline pdf.Unit
	for i, label := range labels {
		if label == "" {
			continue
		}
		baseline = -leading * pdf.Unit(i)
		originX := rightSide + fieldLinePadding
		canvas.DrawLine(pdf.Point{originX, baseline}, pdf.Point{originX + lineLength, baseline})
	}

	return pdf.Point{pt.X + rightSide + fieldLinePadding + lineLength, baseline}
}

// renderMatchSheet creates a PDF document for a single match sheet.
func renderMatchSheet(doc *pdf.Document, pageWidth, pageHeight pdf.Unit, event *Event, match *Match) {
	const (
		entryHeight  = 3 * pdf.Inch
		numEntryRows = 3
	)

	canvas := doc.NewPage(pageWidth, pageHeight)
	defer canvas.Close()

	red := match.AllianceInfo(Red)
	blue := match.AllianceInfo(Blue)

	// Heading
	headingStyle := textStyle{matchNumberFontName, matchNumberFontSize, 0, 0, 0}
	top := pageHeight - reportMargin - matchNumberFontSize
	headingStyle.Drawf(canvas, pdf.Point{reportMargin, top}, "%s #%d - %s", match.Type.DisplayName(), match.Number, event.Location.Name)
	top -= 0.25 * pdf.Inch

	// Barcode
	bc := &barcode.Image{
		Barcode: barcode.Encode(MatchTag{event.Tag(), match.Type, uint(match.Number)}.String()),
		Scale:   1,
		Height:  24,
	}
	var bcRect pdf.Rectangle
	bcRect.Min.X = pageWidth - reportMargin - pdf.Unit(bc.Bounds().Dx())
	bcRect.Min.Y = pageHeight - reportMargin - pdf.Unit(bc.Bounds().Dy())
	bcRect.Max.X = bcRect.Min.X + pdf.Unit(bc.Bounds().Dx())
	bcRect.Max.Y = bcRect.Min.Y + pdf.Unit(bc.Bounds().Dy())
	canvas.DrawImage(bc, bcRect)
	top = bcRect.Min.Y - 0.25*pdf.Inch

	// Cell borders
	cellStyle := strokeStyle{1, 0, 0, 0}
	base := top - entryHeight*numEntryRows
	cellStyle.Rect(canvas, pdf.Rectangle{
		pdf.Point{reportMargin, base},
		pdf.Point{pageWidth - reportMargin, top},
	})
	cellStyle.Line(canvas,
		pdf.Point{pageWidth / 2, base},
		pdf.Point{pageWidth / 2, top},
	)
	for i := 1; i < numEntryRows; i++ {
		cellStyle.Line(canvas,
			pdf.Point{reportMargin, top - (pdf.Unit(i) * entryHeight)},
			pdf.Point{pageWidth - reportMargin, top - (pdf.Unit(i) * entryHeight)},
		)
	}

	// Teams
	for i, teamInfo := range red.Teams {
		if i >= numEntryRows {
			break
		}
		renderMatchSheetTeam(
			canvas,
			pdf.Rectangle{
				pdf.Point{reportMargin, top - (pdf.Unit(i+1) * entryHeight)},
				pdf.Point{pageWidth / 2, top - (pdf.Unit(i) * entryHeight)},
			},
			teamInfo,
			TeamStats{}, // TODO
		)
	}
	for i, teamInfo := range blue.Teams {
		if i >= numEntryRows {
			break
		}
		renderMatchSheetTeam(
			canvas,
			pdf.Rectangle{
				pdf.Point{pageWidth / 2, top - (pdf.Unit(i+1) * entryHeight)},
				pdf.Point{pageWidth - reportMargin, top - (pdf.Unit(i) * entryHeight)},
			},
			teamInfo,
			TeamStats{}, // TODO
		)
	}
}

// renderMatchSheetTeam renders a single team onto a match sheet.
func renderMatchSheetTeam(canvas *pdf.Canvas, rect pdf.Rectangle, info TeamInfo, stats TeamStats) {
	const (
		padding     = 0.125 * pdf.Inch
		statPadding = 0.0625 * pdf.Inch
	)

	rect.Min.X += padding
	rect.Min.Y += padding
	rect.Max.X -= padding
	rect.Max.Y -= padding

	// Team number
	teamNumberStyle := textStyle{FontName: pdf.HelveticaBold, FontSize: 16}
	if info.Alliance == Red {
		teamNumberStyle.R = 0.69
		teamNumberStyle.G = 0.08
		teamNumberStyle.B = 0.15
	} else {
		teamNumberStyle.R = 0.31
		teamNumberStyle.G = 0.34
		teamNumberStyle.B = 0.72
	}
	baseline := rect.Max.Y - teamNumberStyle.FontSize
	teamNumberStyle.Drawf(canvas, pdf.Point{rect.Min.X, baseline}, "%d", info.Team)

	// TODO: image

	// Stats
	statStyle := textStyle{pdf.Helvetica, 12, 0, 0, 0}
	var textObj pdf.Text
	textObj.SetFont(statStyle.FontName, statStyle.FontSize)
	textObj.Text(fmt.Sprintf("Matches Played: %d", stats.MatchCount))
	textObj.NextLine()
	if stats.MatchCount != 0 {
		textObj.Text(fmt.Sprintf("Average Score: %.1f", stats.AverageScore()))
		textObj.NextLine()
		textObj.Text(fmt.Sprintf("Average Teleop Hoops: %.1f", stats.AverageTeleoperatedHoops()))
		textObj.NextLine()
		textObj.Text(fmt.Sprintf("Average Auto Hoops: %.1f", stats.AverageAutonomousHoops()))
		textObj.NextLine()
	}

	baseline -= statStyle.FontSize + statPadding
	canvas.SetColor(statStyle.R, statStyle.G, statStyle.B)
	canvas.Push()
	canvas.Translate(rect.Min.X, baseline)
	canvas.DrawText(&textObj)
	canvas.Pop()
}

type textStyle struct {
	FontName string
	FontSize pdf.Unit
	R, G, B  float32
}

// Draw renders simple text at pt.
func (style textStyle) Draw(canvas *pdf.Canvas, pt pdf.Point, s string) {
	canvas.Push()
	canvas.Translate(pt.X, pt.Y)
	var text pdf.Text
	canvas.SetColor(style.R, style.G, style.B)
	text.SetFont(style.FontName, style.FontSize)
	text.Text(s)
	canvas.DrawText(&text)
	canvas.Pop()
}

// Drawf renders simple text at pt using Sprintf.
func (style textStyle) Drawf(canvas *pdf.Canvas, pt pdf.Point, format string, args ...interface{}) {
	style.Draw(canvas, pt, fmt.Sprintf(format, args...))
}

type strokeStyle struct {
	LineWidth pdf.Unit
	R, G, B   float32
}

// Rect draws a rectangle.
func (style strokeStyle) Rect(canvas *pdf.Canvas, rect pdf.Rectangle) {
	var path pdf.Path
	path.Rectangle(rect)

	canvas.SetLineWidth(style.LineWidth)
	canvas.SetStrokeColor(style.R, style.G, style.B)
	canvas.Stroke(&path)
}

// Line draws a line.
func (style strokeStyle) Line(canvas *pdf.Canvas, pt1, pt2 pdf.Point) {
	var path pdf.Path
	path.Move(pt1)
	path.Line(pt2)

	canvas.SetLineWidth(style.LineWidth)
	canvas.SetStrokeColor(style.R, style.G, style.B)
	canvas.Stroke(&path)
}
