package main

import (
	"bitbucket.org/zombiezen/gopdf/pdf"
	"fmt"
)

const reportMargin = 0.5 * pdf.Inch

const (
	matchNumberFontName = pdf.HelveticaBold
	matchNumberFontSize = 18

	allianceFontName = pdf.Helvetica
	allianceFontSize = 16

	scoreFontName = pdf.Helvetica
	scoreFontSize = 14

	barcodeFontName = pdf.Courier
	barcodeFontSize = 12
)

const scoutFormsPerPage = 3

func renderMultipleScoutForms(doc *pdf.Document, pageWidth, pageHeight float32, event *Event, matches []*NumberedMatch) {
	n := 0
	sizeX, sizeY := pageWidth-reportMargin*2, (pageHeight-reportMargin*2)/scoutFormsPerPage

	var canvas *pdf.Canvas
	for _, match := range matches {
		if canvas == nil {
			canvas = doc.NewPage(pageWidth, pageHeight)
			canvas.Translate(reportMargin, pageHeight-pageHeight/scoutFormsPerPage-reportMargin)
		}

		// Retrieve the list of teams (in sorted order, red first, then blue)
		teamList := make([]TeamInfo, 0, len(match.Teams))
		teamList = append(teamList, match.Alliance(Red)...)
		teamList = append(teamList, match.Alliance(Blue)...)

		for _, info := range teamList {
			renderScoutForm(canvas, sizeX, sizeY, event, match, info.Team)
			if n%scoutFormsPerPage == scoutFormsPerPage-1 {
				canvas.Close()
				canvas = nil
			} else {
				// Page divider
				// TODO: set dash
				canvas.DrawLine(0, 0, sizeX, 0)
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
func renderScoutForm(canvas *pdf.Canvas, w, h float32, event *Event, match *NumberedMatch, teamNum int) {
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

	// Alliance
	baseline += text.Y() - 0.25*pdf.Inch - allianceFontSize
	renderFields(
		canvas, 0, baseline,
		allianceFontName, allianceFontSize,
		scoutFormAllianceLine,
		fmt.Sprintf("%s Alliance Score:", alliance.DisplayName()),
	)

	// Scores
	baseline -= scoreFontSize + 0.25*pdf.Inch
	formx, formy1 := renderFields(canvas, 0, baseline, pdf.Helvetica, scoreFontSize, 1.0*pdf.Inch, "High:", "Middle:", "Low:")
	formx, formy2 := renderFields(canvas, formx+0.25*pdf.Inch, baseline, pdf.Helvetica, scoreFontSize, 0.75*pdf.Inch, "Ubertube (H/M/L/X):", "Minibot Rank:")
	formx, formy3 := renderFields(canvas, formx+0.5*pdf.Inch, baseline, pdf.Helvetica, scoreFontSize, 0.75*pdf.Inch, "Failure:", "No-Show:")
	_, _ = formy2, formy3

	// Scout name
	// TODO: don't assume formy1 is the lowest
	baseline += formy1 - (scoreFontSize + 0.4*pdf.Inch)
	renderFields(canvas, 0, baseline, pdf.Helvetica, scoreFontSize, 3.0*pdf.Inch, "Scout Name:")
}

const (
	fieldLeading     = 0.1 * pdf.Inch
	fieldLinePadding = 0.125 * pdf.Inch
)

func renderFields(canvas *pdf.Canvas, x, y float32, fontName pdf.Name, fontSize float32, lineLength float32, labels ...string) (right, bottom float32) {
	canvas.Push()
	defer canvas.Pop()
	canvas.Translate(x, y)

	text := new(pdf.Text)
	text.SetFont(fontName, fontSize)
	leading := fontSize + fieldLeading
	text.SetLeading(leading)

	// Draw labels
	rightSide := float32(0)
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
	var baseline float32
	for i, label := range labels {
		if label == "" {
			continue
		}
		baseline = -leading * float32(i)
		originX := rightSide + fieldLinePadding
		canvas.DrawLine(originX, baseline, originX+lineLength, baseline)
	}

	return x + rightSide + fieldLinePadding + lineLength, baseline
}
