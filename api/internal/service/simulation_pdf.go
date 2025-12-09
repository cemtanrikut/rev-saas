package service

import (
	"bytes"
	"fmt"
	"math"
	"strings"

	"rev-saas-api/internal/model"
)

// SimulationPDFData contains all data needed to generate a simulation PDF.
type SimulationPDFData struct {
	Simulation *model.SimulationResult
}

// GenerateSimulationPDF creates a professionally styled PDF report for a pricing simulation.
func GenerateSimulationPDF(data SimulationPDFData) (*bytes.Buffer, error) {
	builder := newPDFBuilder()
	pdf := builder.pdf

	sim := data.Simulation
	reportDate := sim.CreatedAt.Format("Jan 02, 2006")

	// Add first page
	pdf.AddPage()

	// ═══════════════════════════════════════════════════════════════
	// HEADER BAR
	// ═══════════════════════════════════════════════════════════════
	builder.drawHeader(reportDate)

	// ═══════════════════════════════════════════════════════════════
	// TITLE BLOCK
	// ═══════════════════════════════════════════════════════════════
	builder.drawSimulationTitleBlock()

	// ═══════════════════════════════════════════════════════════════
	// SECTION 1: PLAN & PRICE CHANGE
	// ═══════════════════════════════════════════════════════════════
	builder.drawSectionTitleWithDivider("Plan & Price Change", colorPrimary)
	builder.drawPriceChangeBox(sim)

	// ═══════════════════════════════════════════════════════════════
	// SECTION 2: SIMULATION SUMMARY
	// ═══════════════════════════════════════════════════════════════
	builder.drawSectionTitleWithDivider("Simulation Summary", colorDark)
	builder.drawSimulationSummary(sim)

	// ═══════════════════════════════════════════════════════════════
	// SECTION 3: SCENARIO COMPARISON
	// ═══════════════════════════════════════════════════════════════
	builder.drawSectionTitleWithDivider("Scenario Comparison", colorDark)
	builder.drawScenarioTable(sim)

	// ═══════════════════════════════════════════════════════════════
	// SECTION 4: AI PRICING INSIGHTS
	// ═══════════════════════════════════════════════════════════════
	if sim.AINarrative != "" {
		builder.drawSimulationAIInsights(sim.AINarrative)
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

// drawSimulationTitleBlock draws the main title section for simulation PDF
func (b *pdfBuilder) drawSimulationTitleBlock() {
	// Main title
	b.setColor(colorDark)
	b.pdf.SetFont("Arial", "B", 22)
	b.pdf.CellFormat(b.contentWidth, 10, "Pricing Simulation Report", "", 1, "L", false, 0, "")

	b.pdf.Ln(2)

	// Subtitle
	b.setColor(colorMedium)
	b.pdf.SetFont("Arial", "", 11)
	b.pdf.CellFormat(b.contentWidth, 6, "Projected impact analysis for your pricing change.", "", 1, "L", false, 0, "")

	b.pdf.Ln(8)
}

// drawPriceChangeBox draws the plan and price change information
func (b *pdfBuilder) drawPriceChangeBox(sim *model.SimulationResult) {
	// Draw rounded box
	b.setFillColor(colorBg)
	b.setDrawColor(colorLighter)
	b.pdf.SetLineWidth(0.3)

	boxX := b.leftMargin
	boxY := b.pdf.GetY()
	boxW := b.contentWidth
	boxH := 42.0

	b.pdf.RoundedRect(boxX, boxY, boxW, boxH, 2, "1234", "FD")

	// Plan name (large)
	b.pdf.SetXY(boxX+8, boxY+6)
	b.setColor(colorDark)
	b.pdf.SetFont("Arial", "B", 14)
	b.pdf.CellFormat(boxW-16, 7, sim.PlanName, "", 0, "L", false, 0, "")

	// Price change row
	priceY := boxY + 16

	// Current price
	b.pdf.SetXY(boxX+8, priceY)
	b.setColor(colorMedium)
	b.pdf.SetFont("Arial", "", 9)
	b.pdf.CellFormat(40, 4, "CURRENT PRICE", "", 0, "L", false, 0, "")
	b.pdf.SetXY(boxX+8, priceY+5)
	b.setColor(colorDark)
	b.pdf.SetFont("Arial", "B", 16)
	b.pdf.CellFormat(50, 8, formatCurrency(sim.CurrentPrice, sim.Currency), "", 0, "L", false, 0, "")

	// Arrow
	b.pdf.SetXY(boxX+65, priceY+5)
	b.setColor(colorMedium)
	b.pdf.SetFont("Arial", "", 18)
	b.pdf.CellFormat(20, 8, "->", "", 0, "C", false, 0, "")

	// New price
	b.pdf.SetXY(boxX+90, priceY)
	b.setColor(colorMedium)
	b.pdf.SetFont("Arial", "", 9)
	b.pdf.CellFormat(40, 4, "NEW PRICE", "", 0, "L", false, 0, "")
	b.pdf.SetXY(boxX+90, priceY+5)

	// Color based on direction
	if sim.PriceChangePct > 0 {
		b.setColor(colorEmerald)
	} else if sim.PriceChangePct < 0 {
		b.setColor(colorPrimary)
	} else {
		b.setColor(colorDark)
	}
	b.pdf.SetFont("Arial", "B", 16)
	b.pdf.CellFormat(50, 8, formatCurrency(sim.NewPrice, sim.Currency), "", 0, "L", false, 0, "")

	// Change percentage badge
	b.pdf.SetXY(boxX+145, priceY+4)
	changeText := fmt.Sprintf("%.1f%%", sim.PriceChangePct)
	if sim.PriceChangePct > 0 {
		changeText = "+" + changeText
		b.setColor(colorEmerald)
	} else if sim.PriceChangePct < 0 {
		b.setColor(colorRed)
	} else {
		b.setColor(colorMedium)
		changeText = "No change"
	}
	b.pdf.SetFont("Arial", "B", 12)
	b.pdf.CellFormat(30, 8, changeText, "", 0, "L", false, 0, "")

	// Bottom row - metrics
	metricsY := boxY + 32
	colWidth := (boxW - 16) / 4

	// Pricing Goal
	b.pdf.SetXY(boxX+8, metricsY)
	b.setColor(colorMedium)
	b.pdf.SetFont("Arial", "", 8)
	b.pdf.CellFormat(colWidth, 4, "PRICING GOAL", "", 0, "L", false, 0, "")
	b.pdf.SetXY(boxX+8, metricsY+4)
	b.setColor(colorDark)
	b.pdf.SetFont("Arial", "B", 9)
	b.pdf.CellFormat(colWidth, 4, strings.Title(sim.PricingGoal), "", 0, "L", false, 0, "")

	// Active Customers
	b.pdf.SetXY(boxX+8+colWidth, metricsY)
	b.setColor(colorMedium)
	b.pdf.SetFont("Arial", "", 8)
	b.pdf.CellFormat(colWidth, 4, "ACTIVE CUSTOMERS", "", 0, "L", false, 0, "")
	b.pdf.SetXY(boxX+8+colWidth, metricsY+4)
	b.setColor(colorDark)
	b.pdf.SetFont("Arial", "B", 9)
	b.pdf.CellFormat(colWidth, 4, fmt.Sprintf("%d", sim.ActiveCustomersOnPlan), "", 0, "L", false, 0, "")

	// Global MRR
	b.pdf.SetXY(boxX+8+colWidth*2, metricsY)
	b.setColor(colorMedium)
	b.pdf.SetFont("Arial", "", 8)
	b.pdf.CellFormat(colWidth, 4, "GLOBAL MRR", "", 0, "L", false, 0, "")
	b.pdf.SetXY(boxX+8+colWidth*2, metricsY+4)
	b.setColor(colorDark)
	b.pdf.SetFont("Arial", "B", 9)
	b.pdf.CellFormat(colWidth, 4, formatCurrency(sim.GlobalMRR, sim.Currency), "", 0, "L", false, 0, "")

	// Churn Rate
	b.pdf.SetXY(boxX+8+colWidth*3, metricsY)
	b.setColor(colorMedium)
	b.pdf.SetFont("Arial", "", 8)
	b.pdf.CellFormat(colWidth, 4, "CHURN RATE", "", 0, "L", false, 0, "")
	b.pdf.SetXY(boxX+8+colWidth*3, metricsY+4)
	if sim.GlobalChurnRate > 7 {
		b.setColor(colorRed)
	} else if sim.GlobalChurnRate < 3 {
		b.setColor(colorEmerald)
	} else {
		b.setColor(colorAmber)
	}
	b.pdf.SetFont("Arial", "B", 9)
	b.pdf.CellFormat(colWidth, 4, fmt.Sprintf("%.1f%%", sim.GlobalChurnRate), "", 0, "L", false, 0, "")

	// Move cursor below box
	b.pdf.SetY(boxY + boxH + 8)
}

// drawSimulationSummary draws a textual summary of what was simulated
func (b *pdfBuilder) drawSimulationSummary(sim *model.SimulationResult) {
	direction := "increase"
	if sim.PriceChangePct < 0 {
		direction = "decrease"
	}

	absChange := math.Abs(sim.PriceChangePct)
	magnitude := "small"
	if absChange >= 20 {
		magnitude = "large"
	} else if absChange >= 10 {
		magnitude = "moderate"
	}

	summary := fmt.Sprintf(
		"This simulation models a %s price %s from %s to %s for the %s plan. "+
			"With %d active customers on this plan and a current global churn rate of %.1f%%, "+
			"three scenarios (Conservative, Base, and Aggressive) were generated to estimate the potential impact on customer retention, MRR, and ARR.",
		magnitude, direction,
		formatCurrency(sim.CurrentPrice, sim.Currency),
		formatCurrency(sim.NewPrice, sim.Currency),
		sim.PlanName,
		sim.ActiveCustomersOnPlan,
		sim.GlobalChurnRate,
	)

	b.setColor(colorDark)
	b.pdf.SetFont("Arial", "", pdfBodyFontSize)
	b.pdf.MultiCell(b.contentWidth, pdfLineHeight, cleanText(summary), "", "L", false)

	b.pdf.Ln(pdfSectionSpaceBefore)
}

// drawScenarioTable draws the three scenarios in a table format
func (b *pdfBuilder) drawScenarioTable(sim *model.SimulationResult) {
	if len(sim.Scenarios) == 0 {
		return
	}

	isPriceIncrease := sim.PriceChangePct >= 0

	// Table dimensions
	tableX := b.leftMargin
	tableW := b.contentWidth
	colW := tableW / 4 // 4 columns: metric name + 3 scenarios
	rowH := 8.0
	headerH := 10.0

	startY := b.pdf.GetY()

	// Table header
	b.setFillColor(colorDark)
	b.pdf.SetFillColor(colorDark.R, colorDark.G, colorDark.B)
	b.pdf.Rect(tableX, startY, tableW, headerH, "F")

	b.pdf.SetXY(tableX, startY+2)
	b.pdf.SetTextColor(255, 255, 255)
	b.pdf.SetFont("Arial", "B", 9)
	b.pdf.CellFormat(colW, headerH-4, "Metric", "", 0, "C", false, 0, "")

	// Scenario headers with colors
	scenarioColors := []pdfColor{colorEmerald, colorPrimary, colorAmber}
	for i, sc := range sim.Scenarios {
		b.pdf.SetXY(tableX+colW*(float64(i)+1), startY+2)
		b.pdf.CellFormat(colW, headerH-4, sc.Name, "", 0, "C", false, 0, "")
	}

	// Table rows
	currentY := startY + headerH
	rows := []struct {
		Label string
		GetValue func(sc model.SimulationScenario) string
	}{
		{
			Label: func() string {
				if isPriceIncrease {
					return "Customer Loss"
				}
				return "Customer Gain"
			}(),
			GetValue: func(sc model.SimulationScenario) string {
				if isPriceIncrease {
					return fmt.Sprintf("%.1f%% - %.1f%%", sc.CustomerLossMinPct, sc.CustomerLossMaxPct)
				}
				return fmt.Sprintf("%.1f%% - %.1f%%", sc.CustomerGainMinPct, sc.CustomerGainMaxPct)
			},
		},
		{
			Label: "Projected Customers",
			GetValue: func(sc model.SimulationScenario) string {
				return fmt.Sprintf("%d - %d", sc.NewCustomerCountMin, sc.NewCustomerCountMax)
			},
		},
		{
			Label: "Projected MRR",
			GetValue: func(sc model.SimulationScenario) string {
				return fmt.Sprintf("%s - %s", formatCurrency(sc.NewMRRMin, sim.Currency), formatCurrency(sc.NewMRRMax, sim.Currency))
			},
		},
		{
			Label: "Projected ARR (12 mo)",
			GetValue: func(sc model.SimulationScenario) string {
				return fmt.Sprintf("%s - %s", formatCurrency(sc.NewARRMin, sim.Currency), formatCurrency(sc.NewARRMax, sim.Currency))
			},
		},
		{
			Label: "Estimated Churn",
			GetValue: func(sc model.SimulationScenario) string {
				return fmt.Sprintf("%.1f%% - %.1f%%", sc.EstimatedChurnMinPct, sc.EstimatedChurnMaxPct)
			},
		},
		{
			Label: "Risk Level",
			GetValue: func(sc model.SimulationScenario) string {
				return strings.Title(sc.RiskLevel)
			},
		},
	}

	for rowIdx, row := range rows {
		// Alternate row background
		if rowIdx%2 == 0 {
			b.setFillColor(colorBg)
			b.pdf.Rect(tableX, currentY, tableW, rowH, "F")
		}

		// Metric label
		b.pdf.SetXY(tableX+2, currentY+2)
		b.setColor(colorMedium)
		b.pdf.SetFont("Arial", "", 9)
		b.pdf.CellFormat(colW-4, rowH-4, row.Label, "", 0, "L", false, 0, "")

		// Values for each scenario
		for i, sc := range sim.Scenarios {
			b.pdf.SetXY(tableX+colW*(float64(i)+1), currentY+2)
			
			// Special coloring for risk level
			if row.Label == "Risk Level" {
				switch sc.RiskLevel {
				case "low":
					b.setColor(colorEmerald)
				case "medium":
					b.setColor(colorAmber)
				case "high":
					b.setColor(colorRed)
				default:
					b.setColor(colorDark)
				}
				b.pdf.SetFont("Arial", "B", 9)
			} else {
				b.setColor(colorDark)
				b.pdf.SetFont("Arial", "", 9)
			}
			
			b.pdf.CellFormat(colW, rowH-4, row.GetValue(sc), "", 0, "C", false, 0, "")
		}

		currentY += rowH
	}

	// Draw table borders
	b.setDrawColor(colorLighter)
	b.pdf.SetLineWidth(0.3)

	// Outer border
	b.pdf.Rect(tableX, startY, tableW, currentY-startY, "D")

	// Column dividers
	for i := 1; i < 4; i++ {
		x := tableX + colW*float64(i)
		b.pdf.Line(x, startY, x, currentY)
	}

	// Row dividers
	for rowY := startY + headerH; rowY < currentY; rowY += rowH {
		b.pdf.Line(tableX, rowY, tableX+tableW, rowY)
	}

	// Move cursor below table
	b.pdf.SetY(currentY + 8)

	// Scenario descriptions
	b.pdf.Ln(2)
	b.setColor(colorMedium)
	b.pdf.SetFont("Arial", "I", 9)

	descriptions := []string{
		"Conservative: Most cautious estimate with minimal risk assumptions.",
		"Base: Balanced projection using typical market responses.",
		"Aggressive: Optimistic estimate assuming strong market acceptance.",
	}

	for i, desc := range descriptions {
		b.setFillColor(scenarioColors[i])
		b.pdf.Circle(b.leftMargin+2, b.pdf.GetY()+2, 1.5, "F")
		b.pdf.SetX(b.leftMargin + 8)
		b.pdf.CellFormat(b.contentWidth-8, 5, desc, "", 1, "L", false, 0, "")
	}

	b.pdf.Ln(pdfSectionSpaceBefore)
}

// drawSimulationAIInsights draws the AI-generated narrative
func (b *pdfBuilder) drawSimulationAIInsights(narrative string) {
	// Check if we need a new page
	if b.pdf.GetY() > 180 {
		b.drawFooter()
		b.pdf.AddPage()
		b.drawHeader("AI Insights")
	}

	b.drawSectionTitleWithDivider("AI Pricing Insights", colorPrimary)

	// Clean the narrative
	narrative = cleanText(narrative)

	// Draw in a subtle box
	b.setFillColor(pdfColor{245, 250, 255}) // Very light blue background
	b.setDrawColor(pdfColor{200, 220, 245}) // Subtle blue border
	b.pdf.SetLineWidth(0.3)

	// Calculate box height
	b.pdf.SetFont("Arial", "", pdfBodyFontSize)
	paragraphs := splitIntoParagraphs(narrative)

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

	b.pdf.SetY(boxY + totalHeight + 8)
}

