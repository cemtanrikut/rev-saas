package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"rev-saas-api/internal/model"
)

const simulationSystemPrompt = `You are an senior SaaS pricing strategist and advisor.

You will be given a Pricing Simulation Result that contains:
- The old price, the new price, and the pricing goal.
- Three scenario outputs: Conservative, Base, and Aggressive.
- Each scenario includes ranges for customer impact, revenue impact, and churn sensitivity.

IMPORTANT:
You must NOT include or reference ANY numbers in your response.
You must NOT calculate anything.
You must NOT restate, transform, re-interpret, or re-explain numerical data.
You must NOT produce ranges, percentages, estimates, or any numeric details.
You must NOT perform math of any kind.

STRICT RULES:
1. You may ONLY speak in qualitative, descriptive terms.
2. Use abstract comparative language such as:
   - "higher risk"
   - "more stable"
   - "stronger revenue potential"
   - "greater sensitivity to churn"
   - "more conservative outcome"
   - "more aggressive tradeoff"
   - "a balanced middle path"
3. Never mention numbers, ranges, percentages, or dollar values.
4. Never summarize results using numeric expressions.
5. Never generate new figures or projections.

YOUR TASK:
Write an executive-level narrative summarizing:
- The strategic implications of moving from the old price to the new price.
- The tradeoffs between the Conservative, Base, and Aggressive scenarios.
- The retention vs revenue tension for this price change.
- The type of company each scenario may appeal to (revenue-first, retention-first, balanced).
- 2–3 qualitative recommendations that help the user evaluate the change.

TONE:
- Professional
- Neutral
- Concise
- Decision-maker focused
- Avoid hype, fluff, or marketing tone

OUTPUT FORMAT:
Plain text paragraphs only. No numbers. No lists with numbers. No bullet points with numeric details.

Additional constraints:
- Max 400 words.
- Write as if advising a VP Product or CFO.`

// GenerateSimulationNarrative generates an AI narrative for a simulation result.
func (s *AIPricingService) GenerateSimulationNarrative(ctx context.Context, result *model.SimulationResult) (string, error) {
	if !s.IsEnabled() {
		return "", nil
	}

	// Build user prompt with simulation data
	userPrompt := s.buildSimulationPrompt(result)

	// Call OpenAI with temperature=0 for deterministic output
	response, err := s.callOpenAIForSimulation(ctx, simulationSystemPrompt, userPrompt, 600, 0.0)
	if err != nil {
		return "", nil // Graceful fallback
	}

	// Parse response
	narrative, err := s.parseSimulationResponse(response)
	if err != nil {
		return "", nil
	}

	return narrative, nil
}

// buildSimulationPrompt creates the user prompt with simulation data.
func (s *AIPricingService) buildSimulationPrompt(result *model.SimulationResult) string {
	data := map[string]interface{}{
		"planName":        result.PlanName,
		"currentPrice":    result.CurrentPrice,
		"newPrice":        result.NewPrice,
		"currency":        result.Currency,
		"priceChangePct":  result.PriceChangePct,
		"activeCustomers": result.ActiveCustomersOnPlan,
		"globalMRR":       result.GlobalMRR,
		"globalChurnRate": result.GlobalChurnRate,
		"pricingGoal":     result.PricingGoal,
	}

	// Add scenarios
	scenarios := make([]map[string]interface{}, len(result.Scenarios))
	for i, sc := range result.Scenarios {
		scenario := map[string]interface{}{
			"name":                 sc.Name,
			"newCustomerCountMin":  sc.NewCustomerCountMin,
			"newCustomerCountMax":  sc.NewCustomerCountMax,
			"newMRRMin":            sc.NewMRRMin,
			"newMRRMax":            sc.NewMRRMax,
			"newARRMin":            sc.NewARRMin,
			"newARRMax":            sc.NewARRMax,
			"estimatedChurnMinPct": sc.EstimatedChurnMinPct,
			"estimatedChurnMaxPct": sc.EstimatedChurnMaxPct,
			"riskLevel":            sc.RiskLevel,
		}

		// Include loss or gain percentages based on price direction
		if result.PriceChangePct >= 0 {
			scenario["customerLossMinPct"] = sc.CustomerLossMinPct
			scenario["customerLossMaxPct"] = sc.CustomerLossMaxPct
		} else {
			scenario["customerGainMinPct"] = sc.CustomerGainMinPct
			scenario["customerGainMaxPct"] = sc.CustomerGainMaxPct
		}

		scenarios[i] = scenario
	}
	data["scenarios"] = scenarios

	// Determine price direction for context
	if result.PriceChangePct >= 0 {
		data["priceDirection"] = "increase"
	} else {
		data["priceDirection"] = "decrease"
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error building prompt: %v", err)
	}

	return string(jsonBytes)
}

// callOpenAIForSimulation calls OpenAI API with a custom temperature setting.
func (s *AIPricingService) callOpenAIForSimulation(ctx context.Context, systemPrompt, userPrompt string, maxTokens int, temperature float64) (string, error) {
	reqBody := openAIRequest{
		Model: "gpt-4o-mini",
		Messages: []openAIMessage{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: userPrompt,
			},
		},
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var apiResp openAIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if apiResp.Error != nil {
		return "", fmt.Errorf("OpenAI error: %s", apiResp.Error.Message)
	}

	if len(apiResp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return apiResp.Choices[0].Message.Content, nil
}

// parseSimulationResponse cleans up and returns the AI response as plain text.
func (s *AIPricingService) parseSimulationResponse(response string) (string, error) {
	// Clean up the response - remove any potential markdown formatting
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	// Validate that we got a meaningful response
	if len(response) < 50 {
		return "", fmt.Errorf("AI response too short")
	}

	return response, nil
}
