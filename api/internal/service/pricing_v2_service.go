package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/net/html"

	"rev-saas-api/internal/model"
	mongorepo "rev-saas-api/internal/repository/mongo"
)

const (
	maxResponseSize = 5 * 1024 * 1024 // 5MB
	httpTimeout     = 30 * time.Second
	defaultWebsite  = "https://www.usemotion.com/"
)

// Common pricing page paths to try
var commonPricingPaths = []string{
	"/pricing",
	"/plans",
	"/billing",
	"/upgrade",
	"/subscribe",
	"/pro",
	"/premium",
}

// Keywords to look for in links
var pricingKeywords = []string{
	"pricing", "price", "plan", "plans", "billing",
	"upgrade", "subscribe", "signup", "membership",
	"pro", "premium", "enterprise",
}

// Toggle indicators - patterns that suggest a billing toggle exists
var toggleIndicators = []string{
	"pay monthly", "pay annually", "monthly", "yearly", "annual",
	"billed monthly", "billed annually", "billed yearly",
	"save", "per month", "per year", "/mo", "/yr",
	"switch to annual", "switch to monthly",
}

// Monthly keyword synonyms for tab detection
var monthlyKeywords = []string{
	"monthly", "month", "/mo", "per month", "mo", "pay monthly", "billed monthly",
}

// Yearly keyword synonyms for tab detection  
var yearlyKeywords = []string{
	"yearly", "annual", "annually", "year", "/yr", "per year", 
	"pay annually", "billed annually", "save", "pay yearly",
}

// PricingV2Service handles pricing v2 operations
type PricingV2Service struct {
	repo       *mongorepo.PricingV2Repository
	openAIKey  string
	httpClient *http.Client
}

// NewPricingV2Service creates a new PricingV2Service
func NewPricingV2Service(repo *mongorepo.PricingV2Repository, openAIKey string) *PricingV2Service {
	return &PricingV2Service{
		repo:      repo,
		openAIKey: openAIKey,
		httpClient: &http.Client{
			Timeout: httpTimeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
	}
}

// DiscoverPricingPage finds potential pricing page URLs for a website
func (s *PricingV2Service) DiscoverPricingPage(ctx context.Context, websiteURL string) (*model.PricingDiscoverResponse, error) {
	// Normalize and validate URL
	websiteURL = s.normalizeURL(websiteURL)
	if websiteURL == "" {
		websiteURL = defaultWebsite
	}

	if err := s.validateURL(websiteURL); err != nil {
		return &model.PricingDiscoverResponse{
			Error: fmt.Sprintf("invalid URL: %v", err),
		}, nil
	}

	baseURL, err := url.Parse(websiteURL)
	if err != nil {
		return &model.PricingDiscoverResponse{
			Error: "failed to parse URL",
		}, nil
	}

	candidates := make([]string, 0)
	candidateScores := make(map[string]int)

	// Try common pricing paths
	for i, path := range commonPricingPaths {
		testURL := fmt.Sprintf("%s://%s%s", baseURL.Scheme, baseURL.Host, path)
		if s.urlExists(ctx, testURL) {
			candidates = append(candidates, testURL)
			candidateScores[testURL] = 100 - i*10 // Higher score for earlier paths
		}
	}

	// Fetch homepage and extract links
	homepageLinks, err := s.extractLinksFromPage(ctx, websiteURL)
	if err == nil {
		for _, link := range homepageLinks {
			// Check if link contains pricing keywords
			linkLower := strings.ToLower(link)
			for _, keyword := range pricingKeywords {
				if strings.Contains(linkLower, keyword) {
					// Resolve relative URLs
					fullURL := s.resolveURL(baseURL, link)
					if fullURL != "" && !s.containsURL(candidates, fullURL) {
						candidates = append(candidates, fullURL)
						candidateScores[fullURL] = 50
					}
					break
				}
			}
		}
	}

	// Sort by score
	sort.Slice(candidates, func(i, j int) bool {
		return candidateScores[candidates[i]] > candidateScores[candidates[j]]
	})

	// Limit to top 5
	if len(candidates) > 5 {
		candidates = candidates[:5]
	}

	// Select first as primary if available
	var selected *string
	if len(candidates) > 0 {
		selected = &candidates[0]
	}

	return &model.PricingDiscoverResponse{
		PricingCandidates:  candidates,
		SelectedPricingURL: selected,
	}, nil
}

// ExtractPricing extracts pricing information from a URL with 3-stage strategy
func (s *PricingV2Service) ExtractPricing(ctx context.Context, pricingURL string) (*model.PricingExtractResponse, error) {
	// Validate URL
	if err := s.validateURL(pricingURL); err != nil {
		return &model.PricingExtractResponse{
			Error: fmt.Sprintf("invalid URL: %v", err),
		}, nil
	}

	// Stage 1: Static HTML parse
	visibleText, rawHTML, err := s.fetchPageContent(ctx, pricingURL)
	if err != nil {
		return &model.PricingExtractResponse{
			Error: fmt.Sprintf("failed to fetch page: %v", err),
		}, nil
	}

	if len(visibleText) < 100 {
		return &model.PricingExtractResponse{
			Error:    "page content too short or empty",
			Warnings: []string{"page_content_minimal"},
		}, nil
	}

	// Extract hidden content (aria-hidden, display:none, etc.)
	hiddenContent := s.extractHiddenContent(rawHTML)

	// Extract script JSON candidates (__NEXT_DATA__, ld+json, window.__NUXT__)
	scriptJSON := s.extractScriptJSON(rawHTML)

	// Combine all content for LLM
	combinedContent := visibleText
	if hiddenContent != "" {
		combinedContent += "\n\n--- HIDDEN CONTENT (may contain alternate billing) ---\n" + hiddenContent
	}
	if scriptJSON != "" {
		combinedContent += "\n\n--- SCRIPT DATA ---\n" + scriptJSON
	}

	// Stage 2: Detection - check if toggle exists
	hasToggle := s.detectBillingToggle(visibleText, rawHTML)
	
	// First extraction attempt with static content
	plans, warnings, err := s.extractWithLLM(ctx, combinedContent, rawHTML, pricingURL)
	if err != nil {
		return &model.PricingExtractResponse{
			Error:    fmt.Sprintf("extraction failed: %v", err),
			Warnings: warnings,
		}, nil
	}

	// Deduplicate plans
	plans = s.deduplicatePlans(plans)

	// Detect billing periods from extracted plans
	periods := s.detectBillingPeriods(plans)
	
	// Check if we need browser rendering
	needsRender := false
	if hasToggle && len(periods) <= 1 {
		needsRender = true
		warnings = append(warnings, "toggle_detected_single_period")
	}

	// Stage 3: Browser render if needed
	if needsRender && s.shouldUseBrowserRender() {
		log.Printf("[pricing-v2] toggle detected, attempting browser render for: %s", pricingURL)
		
		browserPlans, browserPeriods, browserWarnings, err := s.extractWithBrowserRender(ctx, pricingURL)
		if err != nil {
			log.Printf("[pricing-v2] browser render failed: %v", err)
			warnings = append(warnings, "browser_render_failed")
		} else {
			// Deduplicate browser plans
			browserPlans = s.deduplicatePlans(browserPlans)
			
			// Use browser results if better
			if len(browserPlans) > len(plans) || len(browserPeriods) > len(periods) {
				plans = browserPlans
				periods = browserPeriods
				warnings = append(warnings, browserWarnings...)
				return &model.PricingExtractResponse{
					Plans:           plans,
					SourceURL:       pricingURL,
					DetectedPeriods: periods,
					NeedsRender:     false,
					RenderUsed:      true,
					Warnings:        warnings,
				}, nil
			}
		}
	}

	return &model.PricingExtractResponse{
		Plans:           plans,
		SourceURL:       pricingURL,
		DetectedPeriods: periods,
		NeedsRender:     needsRender,
		RenderUsed:      false,
		Warnings:        warnings,
	}, nil
}

// shouldUseBrowserRender checks if browser rendering is available
func (s *PricingV2Service) shouldUseBrowserRender() bool {
	// Always try browser render when needed
	return true
}

// tabCandidate represents a potential billing toggle tab element
type tabCandidate struct {
	selector string
	text     string
	score    int
	isMonthly bool
	isYearly  bool
}

// scoreTabText scores text against monthly/yearly keywords
func (s *PricingV2Service) scoreTabText(text string, keywords []string) int {
	normalized := strings.ToLower(strings.TrimSpace(text))
	normalized = regexp.MustCompile(`\s+`).ReplaceAllString(normalized, " ")
	
	score := 0
	for _, kw := range keywords {
		if strings.Contains(normalized, kw) {
			score += 10
			// Bonus for exact match
			if normalized == kw {
				score += 20
			}
		}
	}
	return score
}

// extractWithBrowserRender uses chromedp to render page and capture toggle states
func (s *PricingV2Service) extractWithBrowserRender(ctx context.Context, pricingURL string) ([]model.ExtractedPlan, []string, []string, error) {
	// Create browser context with timeout
	allocCtx, cancel := chromedp.NewExecAllocator(ctx,
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("no-sandbox", true),
			chromedp.Flag("disable-dev-shm-usage", true),
			chromedp.Flag("disable-setuid-sandbox", true),
			chromedp.Flag("disable-web-security", false),
			chromedp.Flag("disable-background-networking", true),
			chromedp.Flag("disable-default-apps", true),
			chromedp.Flag("disable-extensions", true),
			chromedp.Flag("disable-sync", true),
			chromedp.Flag("disable-translate", true),
			chromedp.Flag("mute-audio", true),
			chromedp.Flag("hide-scrollbars", true),
		)...,
	)
	defer cancel()

	browserCtx, cancelBrowser := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancelBrowser()

	// Set timeout for browser operations
	browserCtx, cancelTimeout := context.WithTimeout(browserCtx, 60*time.Second)
	defer cancelTimeout()

	var warnings []string

	// Load the page
	var defaultHTML string
	var defaultText string
	
	log.Printf("[pricing-v2] loading page in browser: %s", pricingURL)
	
	err := chromedp.Run(browserCtx,
		chromedp.Navigate(pricingURL),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Sleep(3*time.Second), // Wait for JS to load
		chromedp.InnerHTML("html", &defaultHTML, chromedp.ByQuery),
		chromedp.Text("body", &defaultText, chromedp.ByQuery),
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load page: %w", err)
	}

	log.Printf("[pricing-v2] captured default state, text length: %d", len(defaultText))

	// Find tab candidates using improved heuristics
	monthlyTab, yearlyTab := s.findBillingTabs(browserCtx, defaultHTML)
	
	var monthlyHTML, monthlyText string
	var yearlyHTML, yearlyText string
	monthlyClicked := false
	yearlyClicked := false

	// Click monthly tab with state verification
	if monthlyTab != "" {
		log.Printf("[pricing-v2] attempting to click monthly tab: %s", monthlyTab)
		monthlyClicked, monthlyHTML, monthlyText = s.clickTabWithVerification(browserCtx, monthlyTab, "monthly", defaultText)
		if !monthlyClicked {
			warnings = append(warnings, "monthly_toggle_failed")
		}
	} else {
		warnings = append(warnings, "monthly_toggle_not_found")
	}

	// Click yearly tab with state verification
	if yearlyTab != "" {
		log.Printf("[pricing-v2] attempting to click yearly tab: %s", yearlyTab)
		yearlyClicked, yearlyHTML, yearlyText = s.clickTabWithVerification(browserCtx, yearlyTab, "yearly", defaultText)
		if !yearlyClicked {
			warnings = append(warnings, "yearly_toggle_failed")
		}
	} else {
		warnings = append(warnings, "yearly_toggle_not_found")
	}

	// Build combined content for LLM with clear section markers
	var combinedContent strings.Builder
	
	if monthlyClicked && monthlyText != "" {
		combinedContent.WriteString("=== MONTHLY BILLING STATE (after clicking monthly tab) ===\n")
		combinedContent.WriteString(monthlyText)
		combinedContent.WriteString("\n\n")
	}
	
	if yearlyClicked && yearlyText != "" {
		combinedContent.WriteString("=== YEARLY/ANNUAL BILLING STATE (after clicking yearly tab) ===\n")
		combinedContent.WriteString(yearlyText)
		combinedContent.WriteString("\n\n")
	}
	
	// Use default state if nothing was clicked
	if !monthlyClicked && !yearlyClicked {
		combinedContent.WriteString("=== DEFAULT STATE (no tabs clicked) ===\n")
		combinedContent.WriteString(defaultText)
		warnings = append(warnings, "no_toggle_clicked")
	}

	// Add script JSON from rendered HTML
	combinedHTML := defaultHTML
	if yearlyHTML != "" {
		combinedHTML += yearlyHTML
	}
	if monthlyHTML != "" {
		combinedHTML += monthlyHTML
	}
	scriptJSON := s.extractScriptJSON(combinedHTML)
	if scriptJSON != "" {
		combinedContent.WriteString("\n\n--- SCRIPT DATA ---\n")
		combinedContent.WriteString(scriptJSON)
	}

	// Extract with LLM
	plans, llmWarnings, err := s.extractWithLLM(ctx, combinedContent.String(), combinedHTML, pricingURL)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("LLM extraction failed: %w", err)
	}

	// Deduplicate plans
	plans = s.deduplicatePlans(plans)

	warnings = append(warnings, llmWarnings...)
	periods := s.detectBillingPeriods(plans)

	return plans, periods, warnings, nil
}

// findBillingTabs finds the best monthly and yearly tab selectors using scoring
func (s *PricingV2Service) findBillingTabs(ctx context.Context, pageHTML string) (monthlySelector, yearlySelector string) {
	// JavaScript to find all potential tab elements with their text and attributes
	var tabsJSON string
	
	err := chromedp.Run(ctx,
		chromedp.Evaluate(`
			(() => {
				const tabs = [];
				
				// Find role="tab" elements (highest priority)
				document.querySelectorAll('[role="tab"]').forEach((el, i) => {
					tabs.push({
						selector: '[role="tab"]:nth-of-type(' + (i+1) + ')',
						text: el.textContent.trim(),
						ariaSelected: el.getAttribute('aria-selected'),
						ariaControls: el.getAttribute('aria-controls'),
						type: 'role-tab'
					});
				});
				
				// Find tablist children
				document.querySelectorAll('[role="tablist"] > *').forEach((el, i) => {
					if (!el.hasAttribute('role') || el.getAttribute('role') !== 'tab') {
						tabs.push({
							selector: '[role="tablist"] > *:nth-child(' + (i+1) + ')',
							text: el.textContent.trim(),
							ariaSelected: el.getAttribute('aria-selected'),
							type: 'tablist-child'
						});
					}
				});
				
				// Find button elements with billing keywords
				document.querySelectorAll('button').forEach((el, i) => {
					const text = el.textContent.toLowerCase();
					if (text.includes('month') || text.includes('year') || text.includes('annual') || 
					    text.includes('/mo') || text.includes('/yr') || text.includes('save')) {
						tabs.push({
							selector: 'button:nth-of-type(' + (i+1) + ')',
							text: el.textContent.trim(),
							ariaSelected: el.getAttribute('aria-selected'),
							type: 'button'
						});
					}
				});
				
				// Find label elements (for toggle switches)
				document.querySelectorAll('label').forEach((el, i) => {
					const text = el.textContent.toLowerCase();
					if (text.includes('month') || text.includes('year') || text.includes('annual')) {
						tabs.push({
							selector: 'label:nth-of-type(' + (i+1) + ')',
							text: el.textContent.trim(),
							type: 'label'
						});
					}
				});
				
				return JSON.stringify(tabs);
			})()
		`, &tabsJSON),
	)
	
	if err != nil {
		log.Printf("[pricing-v2] failed to find tabs: %v", err)
		return "", ""
	}

	var tabs []struct {
		Selector     string `json:"selector"`
		Text         string `json:"text"`
		AriaSelected string `json:"ariaSelected"`
		AriaControls string `json:"ariaControls"`
		Type         string `json:"type"`
	}
	
	if err := json.Unmarshal([]byte(tabsJSON), &tabs); err != nil {
		log.Printf("[pricing-v2] failed to parse tabs JSON: %v", err)
		return "", ""
	}

	// Score each tab for monthly/yearly
	var monthlyBest, yearlyBest struct {
		selector string
		score    int
	}

	for _, tab := range tabs {
		monthlyScore := s.scoreTabText(tab.Text, monthlyKeywords)
		yearlyScore := s.scoreTabText(tab.Text, yearlyKeywords)
		
		// Bonus for role="tab" elements
		if tab.Type == "role-tab" {
			monthlyScore += 5
			yearlyScore += 5
		}
		
		// Bonus for aria-selected attribute (indicates it's a real tab)
		if tab.AriaSelected != "" {
			monthlyScore += 3
			yearlyScore += 3
		}
		
		// Penalty if text contains both (ambiguous)
		if monthlyScore > 0 && yearlyScore > 0 {
			// Keep only the higher score
			if monthlyScore > yearlyScore {
				yearlyScore = 0
			} else {
				monthlyScore = 0
			}
		}
		
		if monthlyScore > monthlyBest.score {
			monthlyBest.selector = tab.Selector
			monthlyBest.score = monthlyScore
		}
		
		if yearlyScore > yearlyBest.score {
			yearlyBest.selector = tab.Selector
			yearlyBest.score = yearlyScore
		}
	}

	log.Printf("[pricing-v2] found monthly tab: %s (score=%d), yearly tab: %s (score=%d)", 
		monthlyBest.selector, monthlyBest.score, yearlyBest.selector, yearlyBest.score)

	return monthlyBest.selector, yearlyBest.selector
}

// clickTabWithVerification clicks a tab and verifies the state changed
func (s *PricingV2Service) clickTabWithVerification(ctx context.Context, selector, billingType, previousText string) (success bool, newHTML, newText string) {
	maxRetries := 2
	
	for attempt := 0; attempt < maxRetries; attempt++ {
		// Try to click the tab
		err := chromedp.Run(ctx,
			chromedp.Click(selector, chromedp.ByQuery),
			chromedp.Sleep(1500*time.Millisecond),
		)
		
		if err != nil {
			log.Printf("[pricing-v2] click attempt %d failed for %s: %v", attempt+1, billingType, err)
			continue
		}

		// Capture new state
		var capturedHTML, capturedText string
		err = chromedp.Run(ctx,
			chromedp.InnerHTML("html", &capturedHTML, chromedp.ByQuery),
			chromedp.Text("body", &capturedText, chromedp.ByQuery),
		)
		
		if err != nil {
			log.Printf("[pricing-v2] failed to capture state after clicking %s: %v", billingType, err)
			continue
		}

		// Verify state changed using multiple indicators
		stateChanged := s.verifyStateChange(ctx, capturedText, previousText, billingType)
		
		if stateChanged {
			log.Printf("[pricing-v2] successfully clicked %s tab (attempt %d)", billingType, attempt+1)
			return true, capturedHTML, capturedText
		}
		
		log.Printf("[pricing-v2] state did not change after clicking %s (attempt %d)", billingType, attempt+1)
	}
	
	return false, "", ""
}

// verifyStateChange checks if the page state actually changed after clicking a tab
func (s *PricingV2Service) verifyStateChange(ctx context.Context, newText, previousText, billingType string) bool {
	// Check 1: Text content changed significantly
	if newText != previousText && len(newText) > 100 {
		// Calculate similarity - if texts are very different, state changed
		similarity := s.textSimilarity(previousText, newText)
		if similarity < 0.95 { // Less than 95% similar means significant change
			log.Printf("[pricing-v2] text changed (similarity: %.2f)", similarity)
			return true
		}
	}

	// Check 2: Look for billing-specific indicators in new text
	newTextLower := strings.ToLower(newText)
	
	if billingType == "monthly" {
		// Should see monthly-specific text
		monthlyIndicators := []string{"billed monthly", "/mo", "per month", "monthly billing"}
		for _, indicator := range monthlyIndicators {
			if strings.Contains(newTextLower, indicator) {
				return true
			}
		}
	} else if billingType == "yearly" {
		// Should see yearly-specific text
		yearlyIndicators := []string{"billed annually", "billed yearly", "/yr", "per year", "save", "annually"}
		for _, indicator := range yearlyIndicators {
			if strings.Contains(newTextLower, indicator) {
				return true
			}
		}
	}

	// Check 3: Check for aria-selected change via JavaScript
	var ariaChanged bool
	chromedp.Run(ctx,
		chromedp.Evaluate(`
			(() => {
				const tabs = document.querySelectorAll('[role="tab"]');
				for (const tab of tabs) {
					if (tab.getAttribute('aria-selected') === 'true') {
						const text = tab.textContent.toLowerCase();
						const billingType = '`+billingType+`';
						if (billingType === 'monthly' && (text.includes('month') || text.includes('/mo'))) {
							return true;
						}
						if (billingType === 'yearly' && (text.includes('year') || text.includes('annual'))) {
							return true;
						}
					}
				}
				return false;
			})()
		`, &ariaChanged),
	)
	
	if ariaChanged {
		return true
	}

	// Check 4: URL query parameter changed
	var currentURL string
	chromedp.Run(ctx, chromedp.Location(&currentURL))
	if strings.Contains(strings.ToLower(currentURL), billingType) {
		return true
	}

	return false
}

// textSimilarity calculates a simple similarity ratio between two texts
func (s *PricingV2Service) textSimilarity(text1, text2 string) float64 {
	if text1 == text2 {
		return 1.0
	}
	
	// Simple word-based similarity
	words1 := strings.Fields(strings.ToLower(text1))
	words2 := strings.Fields(strings.ToLower(text2))
	
	if len(words1) == 0 || len(words2) == 0 {
		return 0.0
	}
	
	wordSet := make(map[string]bool)
	for _, w := range words1 {
		wordSet[w] = true
	}
	
	matches := 0
	for _, w := range words2 {
		if wordSet[w] {
			matches++
		}
	}
	
	// Jaccard-like similarity
	return float64(matches) / float64(len(words1)+len(words2)-matches)
}

// deduplicatePlans removes duplicate plans using canonical key
func (s *PricingV2Service) deduplicatePlans(plans []model.ExtractedPlan) []model.ExtractedPlan {
	if len(plans) <= 1 {
		return plans
	}

	// Map to store deduplicated plans by canonical key
	deduped := make(map[string]model.ExtractedPlan)
	
	for _, plan := range plans {
		key := s.canonicalPlanKey(plan)
		
		existing, exists := deduped[key]
		if !exists {
			deduped[key] = plan
			continue
		}
		
		// Merge: prefer plan with more features/units/evidence
		merged := s.mergePlans(existing, plan)
		deduped[key] = merged
	}

	// Convert back to slice
	result := make([]model.ExtractedPlan, 0, len(deduped))
	for _, plan := range deduped {
		result = append(result, plan)
	}
	
	// Sort by name and billing period for consistent ordering
	sort.Slice(result, func(i, j int) bool {
		if result[i].Name != result[j].Name {
			return result[i].Name < result[j].Name
		}
		return result[i].BillingPeriod < result[j].BillingPeriod
	})

	log.Printf("[pricing-v2] deduplicated %d plans to %d unique plans", len(plans), len(result))
	return result
}

// canonicalPlanKey generates a unique key for deduplication
// Key format: normalize(name) + "|" + billing_period + "|" + normalize(price)
func (s *PricingV2Service) canonicalPlanKey(plan model.ExtractedPlan) string {
	// Normalize name: lowercase, remove extra spaces, remove common suffixes
	name := strings.ToLower(strings.TrimSpace(plan.Name))
	name = regexp.MustCompile(`\s+`).ReplaceAllString(name, " ")
	name = strings.TrimSuffix(name, " plan")
	name = strings.TrimSuffix(name, " tier")
	
	// Normalize billing period
	billing := strings.ToLower(plan.BillingPeriod)
	if billing == "" {
		billing = "unknown"
	}
	
	// Normalize price: use price_amount or extract number from price_string
	var priceKey string
	if plan.PriceAmount > 0 {
		priceKey = fmt.Sprintf("%.2f", plan.PriceAmount)
	} else if plan.PriceString != "" {
		// Extract first number from price string
		re := regexp.MustCompile(`[\d,]+\.?\d*`)
		if match := re.FindString(plan.PriceString); match != "" {
			priceKey = strings.ReplaceAll(match, ",", "")
		}
	}
	
	// Don't use monthly_equivalent in key (as per spec)
	return fmt.Sprintf("%s|%s|%s", name, billing, priceKey)
}

// mergePlans merges two plans, preferring the one with more data
func (s *PricingV2Service) mergePlans(existing, new model.ExtractedPlan) model.ExtractedPlan {
	result := existing
	
	// Prefer more features
	if len(new.Features) > len(result.Features) {
		result.Features = new.Features
	}
	
	// Prefer more included units
	if len(new.IncludedUnits) > len(result.IncludedUnits) {
		result.IncludedUnits = new.IncludedUnits
	}
	
	// Prefer evidence with more content
	if len(new.Evidence.PriceSnippet) > len(result.Evidence.PriceSnippet) {
		result.Evidence = new.Evidence
	}
	
	// Fill in missing monthly equivalent
	if result.MonthlyEquivalentAmount == 0 && new.MonthlyEquivalentAmount > 0 {
		result.MonthlyEquivalentAmount = new.MonthlyEquivalentAmount
	}
	
	// Fill in missing annual amount
	if result.AnnualBilledAmount == 0 && new.AnnualBilledAmount > 0 {
		result.AnnualBilledAmount = new.AnnualBilledAmount
	}
	
	return result
}

// SavePlans saves extracted plans to the database
func (s *PricingV2Service) SavePlans(ctx context.Context, userID string, req model.PricingSaveRequest) (*model.PricingSaveResponse, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return &model.PricingSaveResponse{
			Error: "invalid user ID",
		}, nil
	}

	// Delete existing plans for this user (replace with new extraction)
	if err := s.repo.DeleteByUserID(ctx, uid); err != nil {
		log.Printf("[pricing-v2] failed to delete existing plans: %v", err)
	}

	// Convert and save
	plans := make([]*model.PricingV2Plan, len(req.Plans))
	for i, p := range req.Plans {
		plans[i] = &model.PricingV2Plan{
			UserID:                  uid,
			WebsiteURL:              req.WebsiteURL,
			SourceURL:               req.SourceURL,
			ExtractedAt:             time.Now(),
			PlanName:                p.Name,
			PriceAmount:             p.PriceAmount,
			PriceString:             p.PriceString,
			Currency:                p.Currency,
			PriceFrequency:          p.PriceFrequency,
			BillingPeriod:           p.BillingPeriod,
			MonthlyEquivalentAmount: p.MonthlyEquivalentAmount,
			AnnualBilledAmount:      p.AnnualBilledAmount,
			IncludedUnits:           p.IncludedUnits,
			Features:                p.Features,
			Evidence:                p.Evidence,
		}
	}

	count, err := s.repo.CreateMany(ctx, plans)
	if err != nil {
		return &model.PricingSaveResponse{
			Error: fmt.Sprintf("failed to save plans: %v", err),
		}, nil
	}

	return &model.PricingSaveResponse{
		SavedCount: count,
	}, nil
}

// GetSavedPlans returns saved pricing v2 plans for a user
func (s *PricingV2Service) GetSavedPlans(ctx context.Context, userID string) (*model.SavedPricingV2Response, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID")
	}

	plans, err := s.repo.FindByUserID(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to get plans: %w", err)
	}

	return &model.SavedPricingV2Response{
		Plans: plans,
		Count: len(plans),
	}, nil
}

// Updated LLM Extraction Prompt with strict billing period distinction and evidence requirements
const extractionPrompt = `You are a pricing data extraction specialist. Extract pricing plan information from the provided website content.

STRICT RULES:
1. ONLY extract information that is EXPLICITLY present in the content
2. If a field is not found, use null - NEVER guess or invent data
3. EVIDENCE IS REQUIRED: Include exact text snippets for every extracted value
4. If billing period cannot be determined with evidence, set billing_period: "unknown" and add warning

BILLING PERIOD DISTINCTION (CRITICAL):
- MONTHLY PLAN (billed monthly): Customer pays every month
  - Indicators: "billed monthly", "/mo", "per month", "monthly billing"
  - Evidence must explicitly show monthly billing
  
- YEARLY PLAN (billed annually): Customer pays once per year
  - Indicators: "billed annually", "billed yearly", "/yr", "per year", "annual billing"
  - Evidence must explicitly show annual/yearly billing
  
- MONTHLY EQUIVALENT (only for yearly plans):
  - When yearly plan shows a per-month equivalent price like "$10/mo billed annually"
  - This is a YEARLY plan with monthly_equivalent_amount = 10
  - The actual price they pay is annual_billed_amount = 120/year
  - DO NOT confuse this with an actual monthly plan!

Output ONLY valid JSON in this exact format:
{
  "plans": [
    {
      "name": "Plan Name",
      "price_amount": 19.00,
      "price_string": "$19/mo",
      "currency": "USD",
      "price_frequency": "per_month",
      "billing_period": "monthly",
      "monthly_equivalent_amount": null,
      "annual_billed_amount": null,
      "included_units": [
        {
          "name": "credits",
          "amount": 7500,
          "unit": "per seat per month",
          "raw_text": "7,500 credits/seat/month"
        }
      ],
      "features": ["Feature 1", "Feature 2"],
      "evidence": {
        "name_snippet": "exact text where plan name appears",
        "price_snippet": "exact text showing the price AND billing period",
        "units_snippet": "exact text showing included units",
        "billing_evidence": "exact text proving the billing period (e.g., 'billed monthly' or 'billed annually')"
      }
    }
  ],
  "detected_billing_options": ["monthly", "yearly"],
  "warnings": []
}

EXAMPLES:

Example 1 - Monthly plan:
Text: "Pro Plan $12/mo billed monthly"
Result: billing_period: "monthly", price_amount: 12, evidence.billing_evidence: "billed monthly"

Example 2 - Yearly plan showing monthly equivalent:
Text: "Pro Plan $10/mo billed annually"  
Result: billing_period: "yearly", monthly_equivalent_amount: 10, annual_billed_amount: 120, evidence.billing_evidence: "billed annually"

Example 3 - Yearly plan with direct price:
Text: "Pro Plan $120/year"
Result: billing_period: "yearly", price_amount: 120, price_frequency: "per_year", evidence.billing_evidence: "$120/year"

Example 4 - Cannot determine billing:
Text: "Pro Plan $12/mo" (no billing period indicator)
Result: billing_period: "unknown", add to warnings: "billing_period_unverified_Pro_Plan"

IMPORTANT:
- Create SEPARATE entries for monthly and yearly versions of the same plan
- If page shows "Pay monthly" and "Pay annually" sections, extract plans from BOTH sections
- Currency: $ = USD, € = EUR, £ = GBP
- If features are not visible, return empty array and add "features_not_visible" to warnings
- If pricing requires login/contact sales, add "pricing_gated" to warnings
- Always include billing_evidence in evidence object`

// extractWithLLM uses OpenAI to extract pricing from page content
func (s *PricingV2Service) extractWithLLM(ctx context.Context, content, rawHTML, sourceURL string) ([]model.ExtractedPlan, []string, error) {
	if s.openAIKey == "" {
		return nil, nil, fmt.Errorf("OpenAI API key not configured")
	}

	// Limit content size
	if len(content) > 25000 {
		content = content[:25000] + "\n...[truncated]"
	}

	userPrompt := fmt.Sprintf(`Extract pricing information from this page. Pay special attention to billing period evidence.

Source URL: %s

Page Content:
%s`, sourceURL, content)

	// Call OpenAI
	reqBody := map[string]interface{}{
		"model": "gpt-4o-mini",
		"messages": []map[string]string{
			{"role": "system", "content": extractionPrompt},
			{"role": "user", "content": userPrompt},
		},
		"temperature": 0.1,
		"max_tokens":  4000,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.openAIKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	var apiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, nil, err
	}

	if apiResp.Error.Message != "" {
		return nil, nil, fmt.Errorf("OpenAI error: %s", apiResp.Error.Message)
	}

	if len(apiResp.Choices) == 0 {
		return nil, nil, fmt.Errorf("no response from OpenAI")
	}

	// Parse LLM response
	response := strings.TrimSpace(apiResp.Choices[0].Message.Content)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	var result struct {
		Plans    []model.ExtractedPlan `json:"plans"`
		Warnings []string              `json:"warnings"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		log.Printf("[pricing-v2] failed to parse LLM response: %v, response: %s", err, response)
		return nil, []string{"parse_error"}, fmt.Errorf("failed to parse extraction result")
	}

	return result.Plans, result.Warnings, nil
}

// extractHiddenContent extracts content from hidden elements
func (s *PricingV2Service) extractHiddenContent(htmlContent string) string {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return ""
	}

	var hiddenText strings.Builder
	var extractHidden func(*html.Node)

	extractHidden = func(n *html.Node) {
		if n.Type == html.ElementNode {
			// Skip script, style
			if n.Data == "script" || n.Data == "style" || n.Data == "noscript" {
				return
			}

			// Check for hidden attributes
			isHidden := false
			for _, attr := range n.Attr {
				if attr.Key == "aria-hidden" && attr.Val == "true" {
					isHidden = true
				}
				if attr.Key == "hidden" {
					isHidden = true
				}
				if attr.Key == "style" && (strings.Contains(attr.Val, "display:none") || strings.Contains(attr.Val, "display: none")) {
					isHidden = true
				}
				if attr.Key == "data-state" && (attr.Val == "inactive" || attr.Val == "hidden") {
					isHidden = true
				}
				// Check for tab panels that might be hidden
				if attr.Key == "role" && attr.Val == "tabpanel" {
					// Extract this content as it might be alternate billing
					isHidden = true
				}
			}

			if isHidden {
				// Extract text from this hidden element
				var extractText func(*html.Node)
				extractText = func(child *html.Node) {
					if child.Type == html.TextNode {
						text := strings.TrimSpace(child.Data)
						if text != "" {
							hiddenText.WriteString(text)
							hiddenText.WriteString(" ")
						}
					}
					for c := child.FirstChild; c != nil; c = c.NextSibling {
						extractText(c)
					}
				}
				extractText(n)
				return
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractHidden(c)
		}
	}

	extractHidden(doc)
	return strings.TrimSpace(hiddenText.String())
}

// extractScriptJSON extracts JSON data from script tags
func (s *PricingV2Service) extractScriptJSON(htmlContent string) string {
	var jsonData strings.Builder

	// Pattern for __NEXT_DATA__
	nextDataRe := regexp.MustCompile(`<script[^>]*id="__NEXT_DATA__"[^>]*>([\s\S]*?)</script>`)
	if matches := nextDataRe.FindStringSubmatch(htmlContent); len(matches) > 1 {
		// Parse and extract relevant pricing data
		data := matches[1]
		if len(data) < 50000 { // Limit size
			jsonData.WriteString("NEXT_DATA: ")
			jsonData.WriteString(s.extractPricingFromJSON(data))
			jsonData.WriteString("\n")
		}
	}

	// Pattern for ld+json (structured data)
	ldJsonRe := regexp.MustCompile(`<script[^>]*type="application/ld\+json"[^>]*>([\s\S]*?)</script>`)
	ldMatches := ldJsonRe.FindAllStringSubmatch(htmlContent, -1)
	for _, match := range ldMatches {
		if len(match) > 1 && len(match[1]) < 10000 {
			jsonData.WriteString("LD+JSON: ")
			jsonData.WriteString(match[1])
			jsonData.WriteString("\n")
		}
	}

	// Pattern for __NUXT__
	nuxtRe := regexp.MustCompile(`window\.__NUXT__\s*=\s*(\{[\s\S]*?\});?\s*</script>`)
	if matches := nuxtRe.FindStringSubmatch(htmlContent); len(matches) > 1 {
		data := matches[1]
		if len(data) < 50000 {
			jsonData.WriteString("NUXT_DATA: ")
			jsonData.WriteString(s.extractPricingFromJSON(data))
			jsonData.WriteString("\n")
		}
	}

	return strings.TrimSpace(jsonData.String())
}

// extractPricingFromJSON tries to find pricing-related data in JSON
func (s *PricingV2Service) extractPricingFromJSON(jsonStr string) string {
	// Look for pricing-related keys
	pricingPatterns := []string{
		`"price"`, `"pricing"`, `"plans"`, `"subscription"`,
		`"monthly"`, `"yearly"`, `"annual"`, `"billing"`,
	}

	hasPricing := false
	for _, pattern := range pricingPatterns {
		if strings.Contains(strings.ToLower(jsonStr), strings.ToLower(pattern)) {
			hasPricing = true
			break
		}
	}

	if !hasPricing {
		return ""
	}

	// Return a truncated version
	if len(jsonStr) > 5000 {
		return jsonStr[:5000] + "...[truncated]"
	}
	return jsonStr
}

// detectBillingToggle checks if the page has a billing toggle
func (s *PricingV2Service) detectBillingToggle(visibleText, rawHTML string) bool {
	contentLower := strings.ToLower(visibleText + " " + rawHTML)

	// Check for toggle indicators
	toggleCount := 0
	for _, indicator := range toggleIndicators {
		if strings.Contains(contentLower, indicator) {
			toggleCount++
		}
	}

	// Check for tab/switch elements in HTML
	hasTabList := strings.Contains(rawHTML, `role="tablist"`) ||
		strings.Contains(rawHTML, `role="tab"`) ||
		strings.Contains(rawHTML, "toggle") ||
		strings.Contains(rawHTML, "switch")

	// Need multiple toggle indicators or tablist to confirm
	return toggleCount >= 2 || (toggleCount >= 1 && hasTabList)
}

// Helper functions

func (s *PricingV2Service) normalizeURL(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return ""
	}

	// Add https if no scheme
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "https://" + rawURL
	}

	// Parse and normalize
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	// Ensure trailing slash for root
	if parsed.Path == "" {
		parsed.Path = "/"
	}

	return parsed.String()
}

func (s *PricingV2Service) validateURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL format")
	}

	// Only allow http/https
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("only http/https URLs allowed")
	}

	// Block localhost and private IPs
	host := parsed.Hostname()
	if host == "localhost" || host == "127.0.0.1" || host == "0.0.0.0" {
		return fmt.Errorf("localhost not allowed")
	}

	// Check for private IP ranges
	ip := net.ParseIP(host)
	if ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() {
			return fmt.Errorf("private/internal IPs not allowed")
		}
	}

	return nil
}

func (s *PricingV2Service) urlExists(ctx context.Context, testURL string) bool {
	req, err := http.NewRequestWithContext(ctx, "HEAD", testURL, nil)
	if err != nil {
		return false
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Revalyze/1.0)")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

func (s *PricingV2Service) fetchPageContent(ctx context.Context, pageURL string) (string, string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", pageURL, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// Limit response size
	limitedReader := io.LimitReader(resp.Body, maxResponseSize)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", "", err
	}

	rawHTML := string(body)
	visibleText := s.extractVisibleText(rawHTML)

	return visibleText, rawHTML, nil
}

func (s *PricingV2Service) extractVisibleText(htmlContent string) string {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		// Fallback: strip HTML tags with regex
		re := regexp.MustCompile(`<[^>]*>`)
		return re.ReplaceAllString(htmlContent, " ")
	}

	var textBuilder strings.Builder
	var extractText func(*html.Node)

	extractText = func(n *html.Node) {
		// Skip script, style, and hidden elements
		if n.Type == html.ElementNode {
			switch n.Data {
			case "script", "style", "noscript", "head", "meta", "link":
				return
			}

			// Skip hidden elements for visible text
			for _, attr := range n.Attr {
				if attr.Key == "aria-hidden" && attr.Val == "true" {
					return
				}
				if attr.Key == "hidden" {
					return
				}
			}
		}

		if n.Type == html.TextNode {
			text := strings.TrimSpace(n.Data)
			if text != "" {
				textBuilder.WriteString(text)
				textBuilder.WriteString(" ")
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractText(c)
		}
	}

	extractText(doc)

	// Clean up whitespace
	result := textBuilder.String()
	result = regexp.MustCompile(`\s+`).ReplaceAllString(result, " ")
	return strings.TrimSpace(result)
}

func (s *PricingV2Service) extractLinksFromPage(ctx context.Context, pageURL string) ([]string, error) {
	_, rawHTML, err := s.fetchPageContent(ctx, pageURL)
	if err != nil {
		return nil, err
	}

	doc, err := html.Parse(strings.NewReader(rawHTML))
	if err != nil {
		return nil, err
	}

	var links []string
	var extractLinks func(*html.Node)

	extractLinks = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					links = append(links, attr.Val)
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractLinks(c)
		}
	}

	extractLinks(doc)
	return links, nil
}

func (s *PricingV2Service) resolveURL(base *url.URL, ref string) string {
	refURL, err := url.Parse(ref)
	if err != nil {
		return ""
	}
	resolved := base.ResolveReference(refURL)
	return resolved.String()
}

func (s *PricingV2Service) containsURL(urls []string, target string) bool {
	for _, u := range urls {
		if u == target {
			return true
		}
	}
	return false
}

func (s *PricingV2Service) detectBillingPeriods(plans []model.ExtractedPlan) []string {
	periodSet := make(map[string]bool)
	for _, p := range plans {
		if p.BillingPeriod != "" && p.BillingPeriod != "unknown" {
			periodSet[p.BillingPeriod] = true
		}
	}

	periods := make([]string, 0, len(periodSet))
	for p := range periodSet {
		periods = append(periods, p)
	}
	return periods
}
