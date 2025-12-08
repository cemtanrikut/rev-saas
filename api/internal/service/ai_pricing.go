package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"rev-saas-api/internal/model"
)

// Note: fmt is used in buildUserPrompt for error formatting

// Plan-based system prompts

const starterSystemPrompt = `You are a senior SaaS pricing strategist.

ABSOLUTE RULES (STRICT – DO NOT BREAK):
1. You MUST NOT invent any numbers (prices, medians, benchmarks, competitor metrics).
2. You MUST NOT mention "competitors", "market", "industry norms", or any external comparison UNLESS explicit competitor pricing is provided in the input.
3. If no competitor pricing exists, you MUST assume "competitor data not provided" and MUST focus ONLY on the company's own pricing structure.
4. You MUST NOT propose exact new prices. ONLY directional language is allowed:
   - "slightly reduce price"
   - "moderately increase"
   - "narrow plan gap"
   - "improve differentiation"
5. You MUST NOT claim that a plan is "higher/lower than competitors" unless actual competitor prices appear in the input.
6. Churn may be referenced directionally, but without inferring specific numerical relationships.
7. The output MUST NEVER contain invented facts or assumptions.
8. If no competitor pricing appears in the input, you MUST NOT use the words "competitor", "market", "benchmark", "industry", or any comparison phrase. Do not imply competitor existence in any way.

Your task:
- Refine the rule-based recommendation for a small SaaS.
- Provide ONE directional pricing scenario.
- Keep the output concise and actionable.

Output format (VALID JSON ONLY):
{
  "summary": "<2–3 short paragraphs without mentioning competitors unless real data is given>",
  "scenarios": [
    "Scenario A: <direction-only recommendation>"
  ]
}

Additional constraints:
- Max 300 words.
- No bullet characters, no headings, no markdown.
- NO invented competitor comparisons.`

const growthSystemPrompt = `You are a senior SaaS pricing strategist with experience in B2B SaaS.

ABSOLUTE RULES (STRICT – DO NOT BREAK):
1. You MUST NOT invent any numbers, medians, benchmarks, competitor prices, market data, industry averages, or any external facts not explicitly provided in the input.
2. If no competitor pricing is provided in the input:
   - DO NOT mention competitors at all.
   - DO NOT use words like "competitor", "market", "industry", "benchmark", or any comparison phrase.
   - MUST assume that no competitor data exists.
3. You MUST NOT propose specific numeric prices. ONLY directional language is allowed:
   - "slightly increase"
   - "moderately reduce"
   - "narrow the gap"
   - "improve differentiation"
   - "reposition more premium"
4. Churn or metrics may only be referenced directionally. NEVER infer exact numerical relationships.
5. The output MUST be strictly factual and based ONLY on the provided input data.

Your task:
- Produce a concise, executive-friendly summary.
- Propose 2–3 directional pricing scenarios.
- For each scenario, explain the trade-offs using directional language only.

Output format (VALID JSON ONLY):
{
  "summary": "<2–3 short paragraphs, no invented data>",
  "scenarios": [
    "Scenario A: <directional recommendation and trade-offs>",
    "Scenario B: <alternative direction and trade-offs>",
    "Scenario C: <optional, only if clearly distinct>"
  ]
}

Additional constraints:
- Max ~450 words total.
- No markdown, no headings, no bullet characters.
- NO invented competitor comparisons or market data.
- If no competitor data exists, focus ONLY on internal pricing structure.`

const enterpriseSystemPrompt = `You are a senior pricing strategist for mid-to-large B2B SaaS companies.

ABSOLUTE RULES (STRICT – DO NOT BREAK):
1. You MUST NOT invent any numbers, medians, benchmarks, competitor prices, market data, industry averages, or any external facts not explicitly provided in the input.
2. If no competitor pricing is provided in the input:
   - DO NOT mention competitors at all.
   - DO NOT use words like "competitor", "market", "industry", "benchmark", or any comparison phrase.
   - MUST assume that no competitor data exists and focus ONLY on internal pricing structure.
3. You MUST NOT propose specific numeric prices. ONLY directional language is allowed:
   - "slightly increase"
   - "moderately reduce"
   - "narrow the gap between tiers"
   - "improve differentiation"
   - "reposition toward premium"
   - "reposition toward volume"
4. Churn or metrics may only be referenced directionally. NEVER infer exact numerical relationships.
5. The output MUST be strictly factual and based ONLY on: currentPricing, businessMetrics, ruleBasedRecommendation, and competitors (ONLY if explicit data is present).

Your task:
- Write a concise, strategic, C-level summary of the current pricing situation.
- Propose 3 clearly distinct strategic pricing paths using directional language only.
- For each scenario, explain trade-offs and when it makes sense.

Output format (VALID JSON ONLY):
{
  "summary": "<2–4 short paragraphs, strategic tone, no invented data>",
  "scenarios": [
    "Scenario A: <strategic direction, rationale, trade-offs>",
    "Scenario B: <alternative strategic direction, trade-offs>",
    "Scenario C: <third distinct path, trade-offs>"
  ]
}

Additional constraints:
- Think like you are advising a VP Product or CFO.
- Be strategic but concrete; avoid vague consulting buzzwords.
- Max ~600 words total.
- No markdown, no headings, no bullet characters.
- NO invented competitor comparisons, market data, or external benchmarks.
- If no competitor data exists, all scenarios must focus on internal pricing optimization.`

// AIPricingReport contains AI-generated pricing insights.
type AIPricingReport struct {
	Summary   string   `json:"summary"`
	Scenarios []string `json:"scenarios"`
}

// AIPricingInput contains the data needed for AI analysis.
type AIPricingInput struct {
	Plans           []*model.Plan
	Competitors     []*model.Competitor
	BusinessMetrics *model.BusinessMetrics
	RuleBasedResult *model.Analysis
}

// AIPricingService handles AI-powered pricing analysis.
type AIPricingService struct {
	apiKey     string
	httpClient *http.Client
}

// NewAIPricingService creates a new AIPricingService.
func NewAIPricingService(apiKey string) *AIPricingService {
	return &AIPricingService{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second, // Increased timeout for AI calls
		},
	}
}

// IsEnabled returns true if the AI service is configured.
func (s *AIPricingService) IsEnabled() bool {
	return s.apiKey != ""
}

// ShouldRunAI returns true if AI should be used for the given user.
func (s *AIPricingService) ShouldRunAI(plan string) bool {
	if !s.IsEnabled() {
		return false
	}

	// Only run AI for paid plans
	switch plan {
	case model.PlanStarter, model.PlanGrowth, model.PlanEnterprise, model.PlanAdmin:
		return true
	default:
		// free or unknown → no AI
		return false
	}
}

// getSystemPrompt returns the appropriate system prompt based on user's plan.
func (s *AIPricingService) getSystemPrompt(user *model.User) string {
	plan := user.GetEffectivePlan()

	// Admin users get enterprise-level analysis
	if user.Role == model.RoleAdmin {
		return enterpriseSystemPrompt
	}

	switch plan {
	case model.PlanStarter:
		return starterSystemPrompt
	case model.PlanGrowth:
		return growthSystemPrompt
	case model.PlanEnterprise:
		return enterpriseSystemPrompt
	case model.PlanAdmin:
		return enterpriseSystemPrompt
	default:
		// free or unknown → empty (will be caught by ShouldRunAI)
		return ""
	}
}

// getMaxTokens returns the max tokens for the AI response based on plan.
func (s *AIPricingService) getMaxTokens(plan string) int {
	switch plan {
	case model.PlanStarter:
		return 500
	case model.PlanGrowth:
		return 700
	case model.PlanEnterprise, model.PlanAdmin:
		return 1000
	default:
		return 500
	}
}

// GenerateAIPricingReport generates AI-powered pricing insights based on user's plan.
func (s *AIPricingService) GenerateAIPricingReport(ctx context.Context, user *model.User, input AIPricingInput) (*AIPricingReport, error) {
	// Check if AI should run for this user's plan
	effectivePlan := user.GetEffectivePlan()
	if !s.ShouldRunAI(effectivePlan) {
		return nil, nil
	}

	// Get plan-specific system prompt
	systemPrompt := s.getSystemPrompt(user)
	if systemPrompt == "" {
		return nil, nil
	}

	// Build the user prompt with all context
	userPrompt := s.buildUserPrompt(user, input)

	// Get max tokens for this plan
	maxTokens := s.getMaxTokens(effectivePlan)

	// Call OpenAI API
	response, err := s.callOpenAI(ctx, systemPrompt, userPrompt, maxTokens)
	if err != nil {
		// Graceful fallback to rule-based analysis
		return nil, nil
	}

	// Parse the response
	report, err := s.parseResponse(response)
	if err != nil {
		return nil, nil
	}

	return report, nil
}

// buildUserPrompt creates the user prompt with all analysis context.
func (s *AIPricingService) buildUserPrompt(user *model.User, input AIPricingInput) string {
	// Build a structured JSON input for the AI
	data := map[string]interface{}{
		"tier": user.GetEffectivePlan(),
	}

	// Current pricing plans
	if len(input.Plans) > 0 {
		plans := make([]map[string]interface{}, len(input.Plans))
		for i, p := range input.Plans {
			plans[i] = map[string]interface{}{
				"name":  p.Name,
				"price": p.Price,
			}
		}
		data["currentPricing"] = plans
	}

	// Competitors - ONLY include if at least one has valid price data
	// This prevents the AI from assuming competitor context when none exists
	if len(input.Competitors) > 0 {
		validCompetitors := make([]map[string]interface{}, 0)
		for _, c := range input.Competitors {
			// Now competitors have multiple plans
			competitorPlans := make([]map[string]interface{}, 0)
			for _, p := range c.Plans {
				if p.Price > 0 {
					competitorPlans = append(competitorPlans, map[string]interface{}{
						"name":  p.Name,
						"price": p.Price,
					})
				}
			}
			if len(competitorPlans) > 0 {
				validCompetitors = append(validCompetitors, map[string]interface{}{
					"name":  c.Name,
					"plans": competitorPlans,
				})
			}
		}
		// Only add to data if we have at least one valid competitor with price
		if len(validCompetitors) > 0 {
			data["competitors"] = validCompetitors
		}
		// If no valid prices, completely omit the "competitors" field
	}

	// Business metrics
	if input.BusinessMetrics != nil {
		data["businessMetrics"] = map[string]interface{}{
			"mrr":              input.BusinessMetrics.MRR,
			"customers":        input.BusinessMetrics.Customers,
			"monthlyChurnRate": input.BusinessMetrics.MonthlyChurnRate,
			"currency":         input.BusinessMetrics.Currency,
		}
	}

	// Rule-based recommendation summary
	if input.RuleBasedResult != nil {
		recs := make([]map[string]interface{}, len(input.RuleBasedResult.Recommendations))
		for i, r := range input.RuleBasedResult.Recommendations {
			recs[i] = map[string]interface{}{
				"planName":          r.PlanName,
				"currentPrice":      r.CurrentPrice,
				"suggestedPrice":    r.SuggestedNewPrice,
				"suggestedAction":   r.SuggestedAction,
				"position":          r.Position,
				"rationale":         r.Rationale,
			}
		}
		data["ruleBasedRecommendation"] = map[string]interface{}{
			"summary":         input.RuleBasedResult.Summary,
			"recommendations": recs,
		}
	}

	// Convert to JSON string
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error building prompt: %v", err)
	}

	return string(jsonBytes)
}

// OpenAI API structures
type openAIRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	Temperature float64         `json:"temperature"`
	MaxTokens   int             `json:"max_tokens"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// callOpenAI makes the API call to OpenAI with separate system and user prompts.
func (s *AIPricingService) callOpenAI(ctx context.Context, systemPrompt, userPrompt string, maxTokens int) (string, error) {
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
		Temperature: 0.7,
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

// parseResponse parses the OpenAI response into an AIPricingReport.
func (s *AIPricingService) parseResponse(response string) (*AIPricingReport, error) {
	// Clean up the response (remove markdown code blocks if present)
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	var report AIPricingReport
	if err := json.Unmarshal([]byte(response), &report); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return &report, nil
}
