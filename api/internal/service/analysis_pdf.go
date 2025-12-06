package service

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jung-kurt/gofpdf"

	"rev-saas-api/internal/model"
)

// PDFExportData contains all data needed to generate the analysis PDF.
type PDFExportData struct {
	Analysis *model.Analysis
	Metrics  *model.BusinessMetrics
}

// Color definitions
type pdfColor struct {
	R, G, B int
}

var (
	colorPrimary  = pdfColor{59, 130, 246}   // Blue-500
	colorDark     = pdfColor{30, 41, 59}     // Slate-800
	colorMedium   = pdfColor{100, 116, 139}  // Slate-500
	colorLight    = pdfColor{148, 163, 184}  // Slate-400
	colorLighter  = pdfColor{226, 232, 240}  // Slate-200
	colorBg       = pdfColor{248, 250, 252}  // Slate-50
	colorEmerald  = pdfColor{16, 185, 129}   // Emerald-500
	colorRed      = pdfColor{239, 68, 68}    // Red-500
	colorAmber    = pdfColor{245, 158, 11}   // Amber-500
)

// pdfBuilder wraps gofpdf with helper methods
type pdfBuilder struct {
	pdf          *gofpdf.Fpdf
	pageWidth    float64
	pageHeight   float64
	leftMargin   float64
	rightMargin  float64
	contentWidth float64
	logoPath     string
}

// Layout constants for consistent spacing
const (
	pdfMargin            = 18.0  // ~48pt - balanced margins
	pdfSectionSpaceBefore = 8.0  // ~22pt - space before section titles
	pdfSectionSpaceAfter  = 4.0  // ~11pt - space after section titles
	pdfParagraphSpace     = 3.0  // ~8pt - space between paragraphs
	pdfBodyFontSize       = 11.0 // Body text size
	pdfLineHeight         = 5.5  // ~1.3x line height for 11pt
	pdfBulletIndent       = 5.0  // Bullet indentation
	pdfBulletSpacing      = 2.5  // Space between bullets
)

// newPDFBuilder creates a new PDF builder
func newPDFBuilder() *pdfBuilder {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(pdfMargin, pdfMargin, pdfMargin)
	pdf.SetAutoPageBreak(true, 30) // Leave space for footer

	pageWidth, pageHeight := pdf.GetPageSize()
	leftMargin, _, rightMargin, _ := pdf.GetMargins()

	// Find logo path
	logoPath := findLogoPath()

	return &pdfBuilder{
		pdf:          pdf,
		pageWidth:    pageWidth,
		pageHeight:   pageHeight,
		leftMargin:   leftMargin,
		rightMargin:  rightMargin,
		contentWidth: pageWidth - leftMargin - rightMargin,
		logoPath:     logoPath,
	}
}

// findLogoPath tries to find the logo file
func findLogoPath() string {
	// Try multiple possible paths
	paths := []string{
		"assets/revalyze-logo.png",
		"../assets/revalyze-logo.png",
		"/home/cemtanrikut/Projects/rev-saas/api/assets/revalyze-logo.png",
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			absPath, _ := filepath.Abs(p)
			return absPath
		}
	}

	return ""
}

// setColor sets the text color
func (b *pdfBuilder) setColor(c pdfColor) {
	b.pdf.SetTextColor(c.R, c.G, c.B)
}

// setFillColor sets the fill color
func (b *pdfBuilder) setFillColor(c pdfColor) {
	b.pdf.SetFillColor(c.R, c.G, c.B)
}

// setDrawColor sets the draw color
func (b *pdfBuilder) setDrawColor(c pdfColor) {
	b.pdf.SetDrawColor(c.R, c.G, c.B)
}

// drawHeader draws the header bar with logo
func (b *pdfBuilder) drawHeader(date string) {
	headerHeight := 14.0

	// Draw header background
	b.setFillColor(colorBg)
	b.pdf.Rect(0, 0, b.pageWidth, headerHeight, "F")

	// Draw bottom border line
	b.setDrawColor(colorLighter)
	b.pdf.Line(0, headerHeight, b.pageWidth, headerHeight)

	// Logo (PNG image) or fallback to text
	if b.logoPath != "" {
		// Use PNG logo
		b.pdf.ImageOptions(b.logoPath, b.leftMargin, 2, 28, 0, false,
			gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}, 0, "")
	} else {
		// Fallback: Revalyze text
		b.pdf.SetXY(b.leftMargin, 4)
		b.setColor(colorPrimary)
		b.pdf.SetFont("Arial", "B", 14)
		b.pdf.CellFormat(60, 6, "Revalyze", "", 0, "L", false, 0, "")
	}

	// Date (right)
	b.setColor(colorMedium)
	b.pdf.SetFont("Arial", "", 10)
	b.pdf.SetXY(b.pageWidth-b.rightMargin-50, 4)
	b.pdf.CellFormat(50, 6, date, "", 0, "R", false, 0, "")

	// Move cursor below header
	b.pdf.SetY(headerHeight + 10)
}

// drawTitleBlock draws the main title section
func (b *pdfBuilder) drawTitleBlock() {
	// Main title
	b.setColor(colorDark)
	b.pdf.SetFont("Arial", "B", 22)
	b.pdf.CellFormat(b.contentWidth, 10, "Pricing Strategy Report", "", 1, "L", false, 0, "")

	b.pdf.Ln(2)

	// Subtitle
	b.setColor(colorMedium)
	b.pdf.SetFont("Arial", "", 11)
	b.pdf.CellFormat(b.contentWidth, 6, "Automated insight based on your metrics and market data.", "", 1, "L", false, 0, "")

	b.pdf.Ln(8)
}

// drawSectionTitle draws a section title with proper spacing
func (b *pdfBuilder) drawSectionTitle(title string) {
	// Space before section
	b.pdf.Ln(pdfSectionSpaceBefore)
	
	// Title text
	b.setColor(colorDark)
	b.pdf.SetFont("Arial", "B", 14)
	b.pdf.CellFormat(b.contentWidth, 7, title, "", 1, "L", false, 0, "")
	
	// Space after title
	b.pdf.Ln(pdfSectionSpaceAfter)
}

// drawSectionTitleWithDivider draws a section title with a divider line underneath
func (b *pdfBuilder) drawSectionTitleWithDivider(title string, titleColor pdfColor) {
	// Space before section
	b.pdf.Ln(pdfSectionSpaceBefore)
	
	// Title text
	b.setColor(titleColor)
	b.pdf.SetFont("Arial", "B", 15)
	b.pdf.CellFormat(b.contentWidth, 8, title, "", 1, "L", false, 0, "")
	
	// Divider line
	y := b.pdf.GetY() + 1
	b.setDrawColor(titleColor)
	b.pdf.SetLineWidth(0.4)
	b.pdf.Line(b.leftMargin, y, b.leftMargin+60, y) // Short divider under title
	
	// Space after divider
	b.pdf.Ln(pdfSectionSpaceAfter + 2)
}

// drawSnapshotBox draws the business snapshot in a bordered box
func (b *pdfBuilder) drawSnapshotBox(metrics *model.BusinessMetrics) {
	if metrics == nil {
		// No metrics - show placeholder
		b.setFillColor(colorBg)
		b.setDrawColor(colorLighter)
		b.pdf.SetLineWidth(0.3)

		boxX := b.leftMargin
		boxY := b.pdf.GetY()
		boxW := b.contentWidth
		boxH := 15.0

		b.pdf.RoundedRect(boxX, boxY, boxW, boxH, 2, "1234", "FD")

		b.pdf.SetXY(boxX+5, boxY+5)
		b.setColor(colorLight)
		b.pdf.SetFont("Arial", "I", 10)
		b.pdf.CellFormat(boxW-10, 5, "No business metrics configured. Add metrics in Settings for personalized insights.", "", 0, "L", false, 0, "")

		b.pdf.SetY(boxY + boxH + 8)
		return
	}

	// Draw rounded box
	b.setFillColor(colorBg)
	b.setDrawColor(colorLighter)
	b.pdf.SetLineWidth(0.3)

	boxX := b.leftMargin
	boxY := b.pdf.GetY()
	boxW := b.contentWidth
	boxH := 28.0

	b.pdf.RoundedRect(boxX, boxY, boxW, boxH, 2, "1234", "FD")

	// Left column
	colWidth := (boxW - 20) / 2
	leftX := boxX + 8
	rightX := boxX + colWidth + 12

	// Row 1 - MRR and Churn
	rowY := boxY + 6

	// MRR
	b.pdf.SetXY(leftX, rowY)
	b.setColor(colorMedium)
	b.pdf.SetFont("Arial", "", 9)
	b.pdf.CellFormat(25, 5, "MRR:", "", 0, "L", false, 0, "")
	b.setColor(colorDark)
	b.pdf.SetFont("Arial", "B", 10)
	b.pdf.CellFormat(colWidth-25, 5, formatCurrency(metrics.MRR, metrics.Currency), "", 0, "L", false, 0, "")

	// Monthly Churn
	b.pdf.SetXY(rightX, rowY)
	b.setColor(colorMedium)
	b.pdf.SetFont("Arial", "", 9)
	b.pdf.CellFormat(35, 5, "Monthly Churn:", "", 0, "L", false, 0, "")

	// Color based on churn rate
	if metrics.MonthlyChurnRate > 7 {
		b.setColor(colorRed)
	} else if metrics.MonthlyChurnRate < 3 {
		b.setColor(colorEmerald)
	} else {
		b.setColor(colorAmber)
	}
	b.pdf.SetFont("Arial", "B", 10)
	b.pdf.CellFormat(colWidth-35, 5, fmt.Sprintf("%.1f%%", metrics.MonthlyChurnRate), "", 0, "L", false, 0, "")

	// Row 2 - Customers and Currency
	rowY = boxY + 16

	// Customers
	b.pdf.SetXY(leftX, rowY)
	b.setColor(colorMedium)
	b.pdf.SetFont("Arial", "", 9)
	b.pdf.CellFormat(25, 5, "Customers:", "", 0, "L", false, 0, "")
	b.setColor(colorDark)
	b.pdf.SetFont("Arial", "B", 10)
	b.pdf.CellFormat(colWidth-25, 5, fmt.Sprintf("%d", metrics.Customers), "", 0, "L", false, 0, "")

	// Currency
	b.pdf.SetXY(rightX, rowY)
	b.setColor(colorMedium)
	b.pdf.SetFont("Arial", "", 9)
	b.pdf.CellFormat(35, 5, "Currency:", "", 0, "L", false, 0, "")
	b.setColor(colorDark)
	b.pdf.SetFont("Arial", "B", 10)
	b.pdf.CellFormat(colWidth-35, 5, metrics.Currency, "", 0, "L", false, 0, "")

	// Move cursor below box
	b.pdf.SetY(boxY + boxH + 8)
}

// drawSummary draws the executive summary section with proper paragraph formatting
func (b *pdfBuilder) drawSummary(summary string) {
	if summary == "" {
		summary = "No summary available for this analysis."
	}

	// Clean up summary text
	summary = cleanText(summary)

	b.setColor(colorDark)
	b.pdf.SetFont("Arial", "", pdfBodyFontSize)
	
	// Split into paragraphs and render each with spacing
	paragraphs := splitIntoParagraphs(summary)
	for i, para := range paragraphs {
		if para == "" {
			continue
		}
		b.pdf.MultiCell(b.contentWidth, pdfLineHeight, para, "", "L", false)
		if i < len(paragraphs)-1 {
			b.pdf.Ln(pdfParagraphSpace)
		}
	}

	b.pdf.Ln(pdfParagraphSpace)

	// Caption
	b.setColor(colorLight)
	b.pdf.SetFont("Arial", "I", 9)
	b.pdf.CellFormat(b.contentWidth, 5, "Based on your current business metrics.", "", 1, "L", false, 0, "")

	b.pdf.Ln(pdfSectionSpaceBefore)
}

// splitIntoParagraphs splits text into logical paragraphs
func splitIntoParagraphs(text string) []string {
	// First, try splitting by double newlines (already cleaned, but check anyway)
	// If text is one block, split by sentence groups (roughly every 2-3 sentences)
	
	text = strings.TrimSpace(text)
	
	// If short, return as single paragraph
	if len(text) < 200 {
		return []string{text}
	}
	
	// Split into sentences
	sentences := splitSentences(text)
	
	// Group sentences into paragraphs (2-3 sentences each)
	var paragraphs []string
	var current []string
	
	for _, sent := range sentences {
		current = append(current, sent)
		if len(current) >= 2 && len(strings.Join(current, " ")) > 150 {
			paragraphs = append(paragraphs, strings.Join(current, " "))
			current = nil
		}
	}
	
	// Add remaining sentences
	if len(current) > 0 {
		paragraphs = append(paragraphs, strings.Join(current, " "))
	}
	
	return paragraphs
}

// splitSentences splits text into sentences
func splitSentences(text string) []string {
	var sentences []string
	var current strings.Builder
	
	for i, r := range text {
		current.WriteRune(r)
		
		// Check for sentence ending
		if r == '.' || r == '!' || r == '?' {
			// Check if it's followed by space and capital letter (or end of string)
			if i == len(text)-1 || (i < len(text)-2 && text[i+1] == ' ') {
				sentences = append(sentences, strings.TrimSpace(current.String()))
				current.Reset()
			}
		}
	}
	
	// Add remaining text
	if current.Len() > 0 {
		sentences = append(sentences, strings.TrimSpace(current.String()))
	}
	
	return sentences
}

// drawRecommendationCards draws recommendations as individual cards with full rationale
func (b *pdfBuilder) drawRecommendationCards(recommendations []model.Recommendation) {
	if len(recommendations) == 0 {
		b.setColor(colorLight)
		b.pdf.SetFont("Arial", "I", 10)
		b.pdf.CellFormat(b.contentWidth, 6, "No recommendations available.", "", 1, "L", false, 0, "")
		return
	}

	for i, rec := range recommendations {
		// Check if we need a new page (leave space for card + footer)
		if b.pdf.GetY() > 210 {
			b.drawFooter()
			b.pdf.AddPage()
			b.drawHeader(rec.PlanName) // Use plan name as context
			b.drawSectionTitle("Pricing Recommendations (continued)")
		}

		startY := b.pdf.GetY()

		// Card background
		b.setFillColor(colorBg)
		b.setDrawColor(colorLighter)
		b.pdf.SetLineWidth(0.3)

		// Calculate card height based on rationale length
		// Layout breakdown:
		// - Header area (plan name, badge): 0-20mm from top
		// - Divider line: ~20mm
		// - Price row: 24-40mm
		// - Rationale label: 42mm
		// - Rationale text: starts at 47mm
		rationaleText := cleanText(rec.Rationale)
		b.pdf.SetFont("Arial", "", 9)
		rationaleLines := b.pdf.SplitText(rationaleText, b.contentWidth-20)
		rationaleHeight := float64(len(rationaleLines)) * 4.5
		if rationaleHeight < 5 {
			rationaleHeight = 5
		}

		// Base height = 50 (covers header + divider + price row + rationale label + padding)
		// Plus rationale text height + bottom padding
		cardHeight := 52.0 + rationaleHeight + 6.0

		// Draw card
		b.pdf.RoundedRect(b.leftMargin, startY, b.contentWidth, cardHeight, 2, "1234", "FD")

		// Card number badge
		b.pdf.SetXY(b.leftMargin+5, startY+5)
		b.setFillColor(colorPrimary)
		b.pdf.SetFillColor(colorPrimary.R, colorPrimary.G, colorPrimary.B)
		b.pdf.Circle(b.leftMargin+10, startY+10, 4, "F")
		b.pdf.SetXY(b.leftMargin+7, startY+7)
		b.pdf.SetTextColor(255, 255, 255)
		b.pdf.SetFont("Arial", "B", 8)
		b.pdf.CellFormat(6, 6, fmt.Sprintf("%d", i+1), "", 0, "C", false, 0, "")

		// Plan name
		b.pdf.SetXY(b.leftMargin+20, startY+6)
		b.setColor(colorDark)
		b.pdf.SetFont("Arial", "B", 12)
		b.pdf.CellFormat(80, 6, rec.PlanName, "", 0, "L", false, 0, "")

		// Action badge (right side)
		actionLabel, actionColor := getActionStyle(rec.SuggestedAction)
		b.pdf.SetXY(b.leftMargin+b.contentWidth-55, startY+6)
		b.setColor(actionColor)
		b.pdf.SetFont("Arial", "B", 9)
		b.pdf.CellFormat(50, 6, actionLabel, "", 0, "R", false, 0, "")

		// Position badge
		b.pdf.SetXY(b.leftMargin+20, startY+13)
		b.setColor(colorMedium)
		b.pdf.SetFont("Arial", "", 8)
		positionLabel := getPositionLabel(rec.Position)
		b.pdf.CellFormat(50, 4, positionLabel, "", 0, "L", false, 0, "")

		// Divider line
		b.setDrawColor(colorLighter)
		b.pdf.Line(b.leftMargin+10, startY+20, b.leftMargin+b.contentWidth-10, startY+20)

		// Price row
		priceRowY := startY + 24

		// Current price
		b.pdf.SetXY(b.leftMargin+15, priceRowY)
		b.setColor(colorMedium)
		b.pdf.SetFont("Arial", "", 8)
		b.pdf.CellFormat(30, 4, "CURRENT PRICE", "", 0, "L", false, 0, "")
		b.pdf.SetXY(b.leftMargin+15, priceRowY+5)
		b.setColor(colorDark)
		b.pdf.SetFont("Arial", "B", 14)
		b.pdf.CellFormat(40, 8, fmt.Sprintf("$%.2f", rec.CurrentPrice), "", 0, "L", false, 0, "")

		// Arrow
		b.pdf.SetXY(b.leftMargin+70, priceRowY+5)
		b.setColor(colorMedium)
		b.pdf.SetFont("Arial", "", 14)
		b.pdf.CellFormat(20, 8, "->", "", 0, "C", false, 0, "")

		// Suggested price
		b.pdf.SetXY(b.leftMargin+100, priceRowY)
		b.setColor(colorMedium)
		b.pdf.SetFont("Arial", "", 8)
		b.pdf.CellFormat(40, 4, "SUGGESTED PRICE", "", 0, "L", false, 0, "")
		b.pdf.SetXY(b.leftMargin+100, priceRowY+5)

		// Color suggested price based on change
		if rec.SuggestedNewPrice > rec.CurrentPrice {
			b.setColor(colorEmerald)
		} else if rec.SuggestedNewPrice < rec.CurrentPrice {
			b.setColor(colorAmber)
		} else {
			b.setColor(colorPrimary)
		}
		b.pdf.SetFont("Arial", "B", 14)
		b.pdf.CellFormat(40, 8, fmt.Sprintf("$%.2f", rec.SuggestedNewPrice), "", 0, "L", false, 0, "")

		// Change percentage
		if rec.CurrentPrice > 0 {
			changePercent := ((rec.SuggestedNewPrice - rec.CurrentPrice) / rec.CurrentPrice) * 100
			b.pdf.SetXY(b.leftMargin+150, priceRowY+5)
			b.pdf.SetFont("Arial", "", 10)
			if changePercent > 0 {
				b.setColor(colorEmerald)
				b.pdf.CellFormat(30, 8, fmt.Sprintf("+%.0f%%", changePercent), "", 0, "L", false, 0, "")
			} else if changePercent < 0 {
				b.setColor(colorAmber)
				b.pdf.CellFormat(30, 8, fmt.Sprintf("%.0f%%", changePercent), "", 0, "L", false, 0, "")
			} else {
				b.setColor(colorMedium)
				b.pdf.CellFormat(30, 8, "No change", "", 0, "L", false, 0, "")
			}
		}

		// Rationale section
		rationaleY := priceRowY + 18
		b.pdf.SetXY(b.leftMargin+10, rationaleY)
		b.setColor(colorMedium)
		b.pdf.SetFont("Arial", "B", 8)
		b.pdf.CellFormat(30, 4, "RATIONALE", "", 1, "L", false, 0, "")

		b.pdf.SetXY(b.leftMargin+10, rationaleY+5)
		b.setColor(colorDark)
		b.pdf.SetFont("Arial", "", 9)

		// Draw full rationale text (no truncation!)
		for _, line := range rationaleLines {
			b.pdf.SetX(b.leftMargin + 10)
			b.pdf.CellFormat(b.contentWidth-20, 4.5, line, "", 1, "L", false, 0, "")
		}

		// Move to next card
		b.pdf.SetY(startY + cardHeight + 6)
	}
}

// drawAIInsights draws the AI-generated insights section with improved formatting
func (b *pdfBuilder) drawAIInsights(aiSummary string, aiScenarios []string) {
	if aiSummary == "" && len(aiScenarios) == 0 {
		return
	}

	// Check if we need a new page
	if b.pdf.GetY() > 190 {
		b.drawFooter()
		b.pdf.AddPage()
		b.drawHeader("AI Insights")
	}

	// ─────────────────────────────────────────────────────────────────
	// SECTION TITLE: Bold, 15pt, with divider line
	// ─────────────────────────────────────────────────────────────────
	b.pdf.Ln(pdfSectionSpaceBefore)
	
	b.setColor(colorPrimary)
	b.pdf.SetFont("Arial", "B", 15)
	b.pdf.CellFormat(b.contentWidth, 8, "AI Pricing Insights", "", 1, "L", false, 0, "")
	
	// Thin divider line under title
	dividerY := b.pdf.GetY() + 1
	b.setDrawColor(colorPrimary)
	b.pdf.SetLineWidth(0.5)
	b.pdf.Line(b.leftMargin, dividerY, b.leftMargin+55, dividerY)
	
	b.pdf.Ln(pdfSectionSpaceAfter + 3)

	// ─────────────────────────────────────────────────────────────────
	// AI SUMMARY: Split into paragraphs with proper spacing
	// ─────────────────────────────────────────────────────────────────
	if aiSummary != "" {
		aiSummary = cleanText(aiSummary)

		// Draw summary in a subtle box
		b.setFillColor(pdfColor{245, 250, 255}) // Very light blue background
		b.setDrawColor(pdfColor{200, 220, 245}) // Subtle blue border
		b.pdf.SetLineWidth(0.3)

		// Calculate box height based on content
		b.pdf.SetFont("Arial", "", pdfBodyFontSize)
		paragraphs := splitIntoParagraphs(aiSummary)
		
		// Estimate height
		totalHeight := 12.0 // Padding
		for _, para := range paragraphs {
			lines := b.pdf.SplitText(para, b.contentWidth-20)
			totalHeight += float64(len(lines)) * pdfLineHeight
			totalHeight += pdfParagraphSpace
		}

		boxY := b.pdf.GetY()
		b.pdf.RoundedRect(b.leftMargin, boxY, b.contentWidth, totalHeight, 3, "1234", "FD")

		// Render paragraphs inside box
		b.pdf.SetXY(b.leftMargin+10, boxY+6)
		b.setColor(colorDark)
		b.pdf.SetFont("Arial", "", pdfBodyFontSize)

		for i, para := range paragraphs {
			if para == "" {
				continue
			}
			lines := b.pdf.SplitText(para, b.contentWidth-20)
			for _, line := range lines {
				b.pdf.SetX(b.leftMargin + 10)
				b.pdf.CellFormat(b.contentWidth-20, pdfLineHeight, line, "", 1, "L", false, 0, "")
			}
			if i < len(paragraphs)-1 {
				b.pdf.Ln(pdfParagraphSpace)
			}
		}

		b.pdf.SetY(boxY + totalHeight + 6)
	}

	// ─────────────────────────────────────────────────────────────────
	// AI SCENARIOS: Bullet points with proper formatting
	// ─────────────────────────────────────────────────────────────────
	if len(aiScenarios) > 0 {
		b.pdf.Ln(pdfParagraphSpace + 2)

		// Scenarios subtitle
		b.setColor(colorDark)
		b.pdf.SetFont("Arial", "B", 12)
		b.pdf.CellFormat(b.contentWidth, 6, "Recommended Scenarios", "", 1, "L", false, 0, "")
		b.pdf.Ln(pdfParagraphSpace)

		for i, scenario := range aiScenarios {
			if scenario == "" {
				continue
			}
			scenario = cleanText(scenario)

			startY := b.pdf.GetY()

			// Bullet point (filled circle)
			b.setFillColor(colorPrimary)
			bulletX := b.leftMargin + pdfBulletIndent
			bulletY := startY + 2.5
			b.pdf.Circle(bulletX, bulletY, 1.2, "F")

			// Scenario text with proper indentation
			textX := b.leftMargin + pdfBulletIndent + 5
			textWidth := b.contentWidth - pdfBulletIndent - 8

			b.pdf.SetXY(textX, startY)
			b.setColor(colorDark)
			b.pdf.SetFont("Arial", "", pdfBodyFontSize)
			
			// Word wrap the scenario text
			lines := b.pdf.SplitText(scenario, textWidth)
			for _, line := range lines {
				b.pdf.SetX(textX)
				b.pdf.CellFormat(textWidth, pdfLineHeight, line, "", 1, "L", false, 0, "")
			}

			// Space between bullets
			if i < len(aiScenarios)-1 {
				b.pdf.Ln(pdfBulletSpacing)
			}
		}
	}

	b.pdf.Ln(pdfSectionSpaceBefore)
}

// drawFooter draws the footer with disclaimer (at current position, after content)
func (b *pdfBuilder) drawFooter() {
	// Add some vertical space before footer
	b.pdf.Ln(10)

	// Draw separator line at current position
	currentY := b.pdf.GetY()
	b.setDrawColor(colorLighter)
	b.pdf.Line(b.leftMargin, currentY, b.pageWidth-b.rightMargin, currentY)

	// Move below the line
	b.pdf.SetY(currentY + 4)

	// Disclaimer text (small, gray, italic)
	b.setColor(colorLight)
	b.pdf.SetFont("Arial", "I", 7)
	b.pdf.CellFormat(b.contentWidth, 4, "These suggestions are based on the data you provided and do not constitute financial or legal advice.", "", 1, "C", false, 0, "")

	b.pdf.Ln(3)

	// Footer text
	b.pdf.SetFont("Arial", "", 8)
	b.pdf.CellFormat(b.contentWidth, 4, "Generated by Revalyze - AI-Powered Pricing Intelligence", "", 1, "C", false, 0, "")
	b.pdf.CellFormat(b.contentWidth, 4, "(c) 2025 Revalyze B.V.", "", 1, "C", false, 0, "")
}

// getActionStyle returns the label and color for an action
func getActionStyle(action string) (string, pdfColor) {
	switch action {
	case "raise_price":
		return "RAISE PRICE", colorEmerald
	case "raise_price_conservative":
		return "RAISE (CONSERVATIVE)", colorEmerald
	case "raise_price_aggressive":
		return "RAISE (AGGRESSIVE)", colorEmerald
	case "lower_price":
		return "LOWER PRICE", colorRed
	case "keep":
		return "KEEP CURRENT", colorMedium
	case "keep_for_growth":
		return "KEEP FOR GROWTH", colorMedium
	case "consider_increase":
		return "CONSIDER INCREASE", colorPrimary
	case "add_competitors":
		return "ADD DATA", colorLight
	case "restructure":
		return "RESTRUCTURE", colorAmber
	default:
		return strings.ToUpper(action), colorMedium
	}
}

// getPositionLabel returns a human-readable position label
func getPositionLabel(position string) string {
	labels := map[string]string{
		"lowest":       "Market Position: Lowest",
		"below_median": "Market Position: Below Median",
		"around_median": "Market Position: Around Median",
		"above_median": "Market Position: Above Median",
		"highest":      "Market Position: Highest",
		"unknown":      "Market Position: Unknown",
	}
	if label, ok := labels[position]; ok {
		return label
	}
	return "Market Position: " + position
}

// formatCurrency formats a currency value
func formatCurrency(amount float64, currency string) string {
	symbol := "$"
	switch currency {
	case "EUR":
		symbol = "EUR "
	case "GBP":
		symbol = "GBP "
	case "USD":
		symbol = "$"
	}

	if amount >= 1000 {
		return fmt.Sprintf("%s%.0f", symbol, amount)
	}
	return fmt.Sprintf("%s%.2f", symbol, amount)
}

// cleanText removes problematic characters that gofpdf can't render
func cleanText(s string) string {
	// Replace em dash with regular dash
	s = strings.ReplaceAll(s, "\u2014", "-") // em dash
	s = strings.ReplaceAll(s, "\u2013", "-") // en dash

	// Replace smart quotes with regular quotes
	s = strings.ReplaceAll(s, "\u201c", `"`) // left double quote
	s = strings.ReplaceAll(s, "\u201d", `"`) // right double quote
	s = strings.ReplaceAll(s, "\u2018", "'") // left single quote
	s = strings.ReplaceAll(s, "\u2019", "'") // right single quote

	// Replace copyright symbol
	s = strings.ReplaceAll(s, "\u00a9", "(c)")

	// Replace other special characters
	s = strings.ReplaceAll(s, "\u2026", "...") // ellipsis
	s = strings.ReplaceAll(s, "\u2022", "-")   // bullet
	s = strings.ReplaceAll(s, "\u2192", "->")  // right arrow
	s = strings.ReplaceAll(s, "\u2190", "<-")  // left arrow
	s = strings.ReplaceAll(s, "\u2713", "[OK]") // checkmark

	// Remove emoji-like characters (replace with text)
	s = strings.ReplaceAll(s, "\u26a0", "[!]") // warning sign

	// Normalize newlines - convert double newlines to single space
	// (paragraph splitting happens in splitIntoParagraphs)
	s = strings.ReplaceAll(s, "\n\n", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	
	// Clean up multiple spaces
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}

	return strings.TrimSpace(s)
}

// GenerateAnalysisPDF creates a professionally styled PDF report.
func GenerateAnalysisPDF(data PDFExportData) (*bytes.Buffer, error) {
	builder := newPDFBuilder()
	pdf := builder.pdf

	analysis := data.Analysis
	metrics := data.Metrics

	// Format date
	reportDate := analysis.CreatedAt.Format("Jan 02, 2006")

	// Add first page
	pdf.AddPage()

	// ═══════════════════════════════════════════════════════════════
	// HEADER BAR
	// ═══════════════════════════════════════════════════════════════
	builder.drawHeader(reportDate)

	// ═══════════════════════════════════════════════════════════════
	// TITLE BLOCK
	// ═══════════════════════════════════════════════════════════════
	builder.drawTitleBlock()

	// ═══════════════════════════════════════════════════════════════
	// BUSINESS SNAPSHOT
	// ═══════════════════════════════════════════════════════════════
	builder.drawSectionTitle("Business Snapshot")
	builder.drawSnapshotBox(metrics)

	// ═══════════════════════════════════════════════════════════════
	// EXECUTIVE SUMMARY
	// ═══════════════════════════════════════════════════════════════
	builder.drawSectionTitle("Executive Summary")
	builder.drawSummary(analysis.Summary)

	// ═══════════════════════════════════════════════════════════════
	// PRICING RECOMMENDATIONS (as cards with full rationale)
	// ═══════════════════════════════════════════════════════════════
	builder.drawSectionTitle("Pricing Recommendations")
	builder.drawRecommendationCards(analysis.Recommendations)

	// ═══════════════════════════════════════════════════════════════
	// AI PRICING INSIGHTS (if available)
	// ═══════════════════════════════════════════════════════════════
	if analysis.AISummary != "" || len(analysis.AIScenarios) > 0 {
		builder.drawAIInsights(analysis.AISummary, analysis.AIScenarios)
	}

	// ═══════════════════════════════════════════════════════════════
	// FOOTER
	// ═══════════════════════════════════════════════════════════════
	builder.drawFooter()

	// Output to buffer
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return &buf, nil
}
