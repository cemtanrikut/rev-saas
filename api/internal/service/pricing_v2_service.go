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
	"os"
	"path/filepath"
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

// Monthly signals for price/mode detection
var monthlySignals = []string{
	"/mo", "per month", "monthly", "billed monthly", "/month",
}

// Yearly signals for price/mode detection
var yearlySignals = []string{
	"/yr", "per year", "annually", "billed annually", "yearly", 
	"billed yearly", "/year", "save",
}

// Expand button keywords
var expandKeywords = []string{
	"see all", "show more", "compare", "all features", 
	"view features", "expand", "show all", "more features",
}

// PricingV2Service handles pricing v2 operations
type PricingV2Service struct {
	repo         *mongorepo.PricingV2Repository
	openAIKey    string
	httpClient   *http.Client
	isStaging    bool
	debugDir     string
}

// DebugArtifacts stores debug information for extraction runs
type DebugArtifacts struct {
	RunID              string    `json:"run_id"`
	Timestamp          time.Time `json:"timestamp"`
	URL                string    `json:"url"`
	MonthlyTabLabel    string    `json:"monthly_tab_label"`
	YearlyTabLabel     string    `json:"yearly_tab_label"`
	MonthlyClickAttempted bool   `json:"monthly_click_attempted"`
	YearlyClickAttempted  bool   `json:"yearly_click_attempted"`
	MonthlyStateChanged   bool   `json:"monthly_state_changed"`
	YearlyStateChanged    bool   `json:"yearly_state_changed"`
	FoundMonthlyPrices    bool   `json:"found_monthly_prices"`
	FoundYearlyPrices     bool   `json:"found_yearly_prices"`
	SelectedStateBefore   string `json:"selected_state_before"`
	SelectedStateAfter    string `json:"selected_state_after"`
	ArtifactsPath         string `json:"artifacts_path"`
}

// NewPricingV2Service creates a new PricingV2Service
func NewPricingV2Service(repo *mongorepo.PricingV2Repository, openAIKey string) *PricingV2Service {
	isStaging := os.Getenv("APP_ENV") == "staging" || os.Getenv("APP_ENV") == "local"
	debugDir := "/tmp/pricing-v2-debug"
	
	if isStaging {
		os.MkdirAll(debugDir, 0755)
		log.Printf("[pricing-v2] Debug artifacts enabled, saving to: %s", debugDir)
	}
	
	return &PricingV2Service{
		repo:      repo,
		openAIKey: openAIKey,
		isStaging: isStaging,
		debugDir:  debugDir,
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
	log.Printf("[pricing-v2] DiscoverPricingPage called with: %s", websiteURL)
	
	// Normalize and validate URL
	websiteURL = s.normalizeURL(websiteURL)
	if websiteURL == "" {
		websiteURL = defaultWebsite
	}
	log.Printf("[pricing-v2] normalized URL: %s", websiteURL)

	if err := s.validateURL(websiteURL); err != nil {
		log.Printf("[pricing-v2] URL validation failed: %v", err)
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
	log.Printf("[pricing-v2] trying common pricing paths for %s", baseURL.Host)
	for i, path := range commonPricingPaths {
		testURL := fmt.Sprintf("%s://%s%s", baseURL.Scheme, baseURL.Host, path)
		if s.urlExists(ctx, testURL) {
			candidates = append(candidates, testURL)
			candidateScores[testURL] = 100 - i*10 // Higher score for earlier paths
			log.Printf("[pricing-v2] found candidate: %s (score=%d)", testURL, candidateScores[testURL])
		}
	}

	// Fetch homepage and extract links
	log.Printf("[pricing-v2] extracting links from homepage: %s", websiteURL)
	homepageLinks, err := s.extractLinksFromPage(ctx, websiteURL)
	if err != nil {
		log.Printf("[pricing-v2] failed to extract links from homepage: %v", err)
	} else {
		log.Printf("[pricing-v2] found %d links on homepage", len(homepageLinks))
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
						log.Printf("[pricing-v2] found candidate from link: %s (keyword=%s)", fullURL, keyword)
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

	log.Printf("[pricing-v2] discovery complete: %d candidates found, selected=%v", len(candidates), selected != nil)

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
// NOTE: Temporarily disabled due to timeout issues with toggle clicking
// Users can use "Paste mode" for sites with dynamic monthly/yearly toggles
func (s *PricingV2Service) shouldUseBrowserRender() bool {
	return false // Disabled - use paste mode for toggle sites
}

// extractWithBrowserRender uses chromedp to render page and capture toggle states
func (s *PricingV2Service) extractWithBrowserRender(ctx context.Context, pricingURL string) ([]model.ExtractedPlan, []string, []string, error) {
	// Create debug artifacts for this run
	runID := fmt.Sprintf("%d", time.Now().UnixNano())
	artifacts := &DebugArtifacts{
		RunID:     runID,
		Timestamp: time.Now(),
		URL:       pricingURL,
	}
	artifactsDir := filepath.Join(s.debugDir, runID)
	if s.isStaging {
		os.MkdirAll(artifactsDir, 0755)
		artifacts.ArtifactsPath = artifactsDir
	}

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
	browserCtx, cancelTimeout := context.WithTimeout(browserCtx, 90*time.Second)
	defer cancelTimeout()

	var warnings []string

	// Load the page
	log.Printf("[pricing-v2] loading page in browser: %s", pricingURL)
	
	err := chromedp.Run(browserCtx,
		chromedp.Navigate(pricingURL),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Sleep(3*time.Second),
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load page: %w", err)
	}

	// C) Expand features before snapshot
	s.expandFeatureSections(browserCtx)

	// Get initial state
	var initialHTML, initialText string
	chromedp.Run(browserCtx,
		chromedp.InnerHTML("html", &initialHTML, chromedp.ByQuery),
		chromedp.Text("body", &initialText, chromedp.ByQuery),
	)

	// Detect current billing mode
	currentMode := s.detectCurrentBillingMode(initialText)
	artifacts.SelectedStateBefore = currentMode
	log.Printf("[pricing-v2] detected initial billing mode: %s", currentMode)

	// Find toggle controls
	monthlyTab, yearlyTab := s.findToggleControls(browserCtx)
	artifacts.MonthlyTabLabel = monthlyTab.label
	artifacts.YearlyTabLabel = yearlyTab.label

	// Capture snapshots for both billing modes
	var monthlyHTML, monthlyText string
	var yearlyHTML, yearlyText string
	var monthlySuccess, yearlySuccess bool

	// B) Switch to monthly mode
	if monthlyTab.selector != "" {
		artifacts.MonthlyClickAttempted = true
		result := s.switchBillingMode(browserCtx, "monthly", monthlyTab, currentMode)
		monthlySuccess = result.success
		artifacts.MonthlyStateChanged = result.stateChanged
		
		if monthlySuccess {
			chromedp.Run(browserCtx,
				chromedp.InnerHTML("html", &monthlyHTML, chromedp.ByQuery),
				chromedp.Text("body", &monthlyText, chromedp.ByQuery),
			)
			artifacts.FoundMonthlyPrices = s.hasMonthlyPrices(monthlyText)
			
			// Save debug artifacts
			if s.isStaging {
				s.saveDebugArtifact(artifactsDir, "visible_text_monthly.txt", monthlyText)
				s.saveDebugArtifact(artifactsDir, "html_monthly.html", monthlyHTML)
				s.takeScreenshot(browserCtx, filepath.Join(artifactsDir, "screenshot_monthly.png"))
			}
		} else {
			warnings = append(warnings, "monthly_toggle_failed")
			log.Printf("[pricing-v2] monthly toggle failed: %s", result.reason)
		}
	} else {
		warnings = append(warnings, "monthly_toggle_not_found")
	}

	// B) Switch to yearly mode
	if yearlyTab.selector != "" {
		artifacts.YearlyClickAttempted = true
		result := s.switchBillingMode(browserCtx, "yearly", yearlyTab, currentMode)
		yearlySuccess = result.success
		artifacts.YearlyStateChanged = result.stateChanged
		
		if yearlySuccess {
			chromedp.Run(browserCtx,
				chromedp.InnerHTML("html", &yearlyHTML, chromedp.ByQuery),
				chromedp.Text("body", &yearlyText, chromedp.ByQuery),
			)
			artifacts.FoundYearlyPrices = s.hasYearlyPrices(yearlyText)
			
			// Save debug artifacts
			if s.isStaging {
				s.saveDebugArtifact(artifactsDir, "visible_text_yearly.txt", yearlyText)
				s.saveDebugArtifact(artifactsDir, "html_yearly.html", yearlyHTML)
				s.takeScreenshot(browserCtx, filepath.Join(artifactsDir, "screenshot_yearly.png"))
			}
		} else {
			warnings = append(warnings, "yearly_toggle_failed")
			log.Printf("[pricing-v2] yearly toggle failed: %s", result.reason)
		}
	} else {
		warnings = append(warnings, "yearly_toggle_not_found")
	}

	// Get final state
	var finalText string
	chromedp.Run(browserCtx, chromedp.Text("body", &finalText, chromedp.ByQuery))
	artifacts.SelectedStateAfter = s.detectCurrentBillingMode(finalText)

	// Save debug artifacts summary
	if s.isStaging {
		s.saveDebugArtifacts(artifactsDir, artifacts)
		log.Printf("[pricing-v2] Debug artifacts saved to: %s", artifactsDir)
	}

	// D) Build combined content with explicit snapshot markers
	var combinedContent strings.Builder
	
	if monthlySuccess && monthlyText != "" {
		combinedContent.WriteString("=== SNAPSHOT: MONTHLY BILLING MODE ===\n")
		combinedContent.WriteString("(This snapshot was captured after clicking the monthly billing toggle)\n\n")
		combinedContent.WriteString(monthlyText)
		combinedContent.WriteString("\n\n")
	}
	
	if yearlySuccess && yearlyText != "" {
		combinedContent.WriteString("=== SNAPSHOT: YEARLY/ANNUAL BILLING MODE ===\n")
		combinedContent.WriteString("(This snapshot was captured after clicking the yearly/annual billing toggle)\n\n")
		combinedContent.WriteString(yearlyText)
		combinedContent.WriteString("\n\n")
	}
	
	// Fallback to initial state if no toggle clicked
	if !monthlySuccess && !yearlySuccess {
		combinedContent.WriteString("=== SNAPSHOT: DEFAULT STATE (no toggle clicked) ===\n")
		combinedContent.WriteString(initialText)
		warnings = append(warnings, "no_toggle_clicked")
	}

	// Add script JSON
	combinedHTML := initialHTML
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

	// Check for missing features
	hasFeatures := false
	for _, p := range plans {
		if len(p.Features) > 0 {
			hasFeatures = true
			break
		}
	}
	if !hasFeatures && len(plans) > 0 {
		warnings = append(warnings, "features_not_visible")
	}

	// Deduplicate plans
	plans = s.deduplicatePlans(plans)

	warnings = append(warnings, llmWarnings...)
	periods := s.detectBillingPeriods(plans)

	return plans, periods, warnings, nil
}

// ToggleControl represents a billing toggle control
type ToggleControl struct {
	selector string
	label    string
	element  string // type of element: button, tab, label, etc.
}

// SwitchResult represents the result of switching billing mode
type SwitchResult struct {
	success      bool
	stateChanged bool
	reason       string
}

// findToggleControls finds monthly and yearly toggle controls
func (s *PricingV2Service) findToggleControls(ctx context.Context) (monthly, yearly ToggleControl) {
	var controlsJSON string
	
	err := chromedp.Run(ctx,
		chromedp.Evaluate(`
			(() => {
				const controls = [];
				const monthlyKw = ['monthly', 'month', '/mo', 'per month', 'mo', 'pay monthly'];
				const yearlyKw = ['yearly', 'annual', 'annually', 'year', '/yr', 'per year', 'pay annually', 'billed annually', 'save'];
				
				function scoreText(text, keywords) {
					const normalized = text.toLowerCase().replace(/\s+/g, ' ').trim();
					let score = 0;
					for (const kw of keywords) {
						if (normalized.includes(kw)) {
							score += 10;
							if (normalized === kw) score += 20;
						}
					}
					return score;
				}
				
				function findClickable(el) {
					// Find nearest clickable ancestor
					let current = el;
					while (current && current !== document.body) {
						if (current.tagName === 'BUTTON' || 
						    current.getAttribute('role') === 'tab' ||
						    current.tagName === 'A' ||
						    current.onclick ||
						    current.hasAttribute('data-state') ||
						    current.classList.contains('cursor-pointer')) {
							return current;
						}
						current = current.parentElement;
					}
					return el;
				}
				
				function getSelector(el, index) {
					if (el.id) return '#' + el.id;
					if (el.getAttribute('role') === 'tab') {
						const tabs = document.querySelectorAll('[role="tab"]');
						const idx = Array.from(tabs).indexOf(el);
						if (idx >= 0) return '[role="tab"]:nth-of-type(' + (idx + 1) + ')';
					}
					// Fallback to nth-of-type
					return el.tagName.toLowerCase() + ':nth-of-type(' + (index + 1) + ')';
				}
				
				// Find role="tab" elements (highest priority)
				document.querySelectorAll('[role="tab"]').forEach((el, i) => {
					const text = el.textContent.trim();
					const clickable = findClickable(el);
					controls.push({
						selector: getSelector(clickable, i),
						label: text,
						element: 'role-tab',
						monthlyScore: scoreText(text, monthlyKw),
						yearlyScore: scoreText(text, yearlyKw),
						ariaSelected: el.getAttribute('aria-selected')
					});
				});
				
				// Find buttons with billing keywords
				document.querySelectorAll('button, [role="button"]').forEach((el, i) => {
					const text = el.textContent.trim();
					const mScore = scoreText(text, monthlyKw);
					const yScore = scoreText(text, yearlyKw);
					if (mScore > 0 || yScore > 0) {
						controls.push({
							selector: getSelector(el, i),
							label: text,
							element: 'button',
							monthlyScore: mScore,
							yearlyScore: yScore,
							ariaSelected: el.getAttribute('aria-selected')
						});
					}
				});
				
				// Find labels (for toggle switches)
				document.querySelectorAll('label, span').forEach((el, i) => {
					const text = el.textContent.trim();
					if (text.length > 50) return; // Skip long text
					const mScore = scoreText(text, monthlyKw);
					const yScore = scoreText(text, yearlyKw);
					if (mScore > 0 || yScore > 0) {
						const clickable = findClickable(el);
						controls.push({
							selector: getSelector(clickable, i),
							label: text,
							element: 'label-span',
							monthlyScore: mScore,
							yearlyScore: yScore
						});
					}
				});
				
				return JSON.stringify(controls);
			})()
		`, &controlsJSON),
	)
	
	if err != nil {
		log.Printf("[pricing-v2] failed to find toggle controls: %v", err)
		return
	}

	var controls []struct {
		Selector      string `json:"selector"`
		Label         string `json:"label"`
		Element       string `json:"element"`
		MonthlyScore  int    `json:"monthlyScore"`
		YearlyScore   int    `json:"yearlyScore"`
		AriaSelected  string `json:"ariaSelected"`
	}
	
	if err := json.Unmarshal([]byte(controlsJSON), &controls); err != nil {
		log.Printf("[pricing-v2] failed to parse controls JSON: %v", err)
		return
	}

	// Find best monthly and yearly controls
	var bestMonthly, bestYearly struct {
		control ToggleControl
		score   int
	}

	for _, c := range controls {
		// Bonus for role=tab
		bonus := 0
		if c.Element == "role-tab" {
			bonus = 15
		}
		
		// Only consider as monthly if monthly score > yearly score
		if c.MonthlyScore > c.YearlyScore && c.MonthlyScore+bonus > bestMonthly.score {
			bestMonthly.control = ToggleControl{
				selector: c.Selector,
				label:    c.Label,
				element:  c.Element,
			}
			bestMonthly.score = c.MonthlyScore + bonus
		}
		
		// Only consider as yearly if yearly score > monthly score
		if c.YearlyScore > c.MonthlyScore && c.YearlyScore+bonus > bestYearly.score {
			bestYearly.control = ToggleControl{
				selector: c.Selector,
				label:    c.Label,
				element:  c.Element,
			}
			bestYearly.score = c.YearlyScore + bonus
		}
	}

	log.Printf("[pricing-v2] found monthly control: %s (%s, score=%d)", 
		bestMonthly.control.label, bestMonthly.control.selector, bestMonthly.score)
	log.Printf("[pricing-v2] found yearly control: %s (%s, score=%d)", 
		bestYearly.control.label, bestYearly.control.selector, bestYearly.score)

	return bestMonthly.control, bestYearly.control
}

// switchBillingMode switches to the target billing mode with verification
func (s *PricingV2Service) switchBillingMode(ctx context.Context, target string, control ToggleControl, currentMode string) SwitchResult {
	// If already in target mode, no need to switch
	if currentMode == target {
		log.Printf("[pricing-v2] already in %s mode, skipping switch", target)
		return SwitchResult{success: true, stateChanged: false, reason: "already_in_mode"}
	}

	if control.selector == "" {
		return SwitchResult{success: false, stateChanged: false, reason: "no_selector"}
	}

	// Get text before click
	var textBefore string
	chromedp.Run(ctx, chromedp.Text("body", &textBefore, chromedp.ByQuery))

	// Try clicking with retry
	maxRetries := 2
	for attempt := 0; attempt < maxRetries; attempt++ {
		log.Printf("[pricing-v2] switching to %s mode, attempt %d, selector: %s", target, attempt+1, control.selector)
		
		// Try to click
		err := chromedp.Run(ctx,
			chromedp.Click(control.selector, chromedp.ByQuery),
			chromedp.Sleep(1500*time.Millisecond),
		)
		
		if err != nil {
			log.Printf("[pricing-v2] click failed: %v", err)
			continue
		}

		// Verify state changed
		stateChanged := s.verifyBillingModeChange(ctx, target, textBefore)
		
		if stateChanged {
			return SwitchResult{success: true, stateChanged: true, reason: ""}
		}
		
		log.Printf("[pricing-v2] state did not change after click, attempt %d", attempt+1)
	}

	return SwitchResult{success: false, stateChanged: false, reason: "toggle_failed_after_retries"}
}

// verifyBillingModeChange checks if billing mode actually changed
func (s *PricingV2Service) verifyBillingModeChange(ctx context.Context, target, textBefore string) bool {
	var textAfter string
	chromedp.Run(ctx, chromedp.Text("body", &textAfter, chromedp.ByQuery))

	// Check 1: Text content changed
	if textAfter != textBefore {
		similarity := s.textSimilarity(textBefore, textAfter)
		if similarity < 0.95 {
			log.Printf("[pricing-v2] text changed significantly (similarity: %.2f)", similarity)
			return true
		}
	}

	// Check 2: aria-selected changed
	var ariaChanged bool
	chromedp.Run(ctx,
		chromedp.Evaluate(`
			(() => {
				const tabs = document.querySelectorAll('[role="tab"]');
				for (const tab of tabs) {
					if (tab.getAttribute('aria-selected') === 'true') {
						const text = tab.textContent.toLowerCase();
						const target = '`+target+`';
						if (target === 'monthly' && (text.includes('month') || text.includes('/mo'))) {
							return true;
						}
						if (target === 'yearly' && (text.includes('year') || text.includes('annual'))) {
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

	// Check 3: Price signals match target
	newMode := s.detectCurrentBillingMode(textAfter)
	if newMode == target {
		return true
	}

	return false
}

// detectCurrentBillingMode detects the current billing mode from visible text
func (s *PricingV2Service) detectCurrentBillingMode(text string) string {
	textLower := strings.ToLower(text)
	
	monthlyCount := 0
	yearlyCount := 0
	
	for _, signal := range monthlySignals {
		monthlyCount += strings.Count(textLower, signal)
	}
	for _, signal := range yearlySignals {
		yearlyCount += strings.Count(textLower, signal)
	}
	
	// Check for price patterns
	monthlyPriceRe := regexp.MustCompile(`\$\d+(?:\.\d{2})?\s*/\s*mo`)
	yearlyPriceRe := regexp.MustCompile(`\$\d+(?:\.\d{2})?\s*/\s*(?:yr|year)`)
	
	monthlyCount += len(monthlyPriceRe.FindAllString(textLower, -1)) * 3
	yearlyCount += len(yearlyPriceRe.FindAllString(textLower, -1)) * 3

	if monthlyCount > yearlyCount {
		return "monthly"
	} else if yearlyCount > monthlyCount {
		return "yearly"
	}
	return "unknown"
}

// hasMonthlyPrices checks if text contains monthly price signals
func (s *PricingV2Service) hasMonthlyPrices(text string) bool {
	textLower := strings.ToLower(text)
	for _, signal := range monthlySignals {
		if strings.Contains(textLower, signal) {
			return true
		}
	}
	return false
}

// hasYearlyPrices checks if text contains yearly price signals
func (s *PricingV2Service) hasYearlyPrices(text string) bool {
	textLower := strings.ToLower(text)
	for _, signal := range yearlySignals {
		if strings.Contains(textLower, signal) {
			return true
		}
	}
	return false
}

// expandFeatureSections scrolls and clicks expand buttons to reveal features
func (s *PricingV2Service) expandFeatureSections(ctx context.Context) {
	// Auto-scroll to trigger lazy loading
	chromedp.Run(ctx,
		chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight)`, nil),
		chromedp.Sleep(1*time.Second),
		chromedp.Evaluate(`window.scrollTo(0, 0)`, nil),
		chromedp.Sleep(500*time.Millisecond),
	)

	// Find and click expand buttons (up to 5)
	var expandButtonsJSON string
	chromedp.Run(ctx,
		chromedp.Evaluate(`
			(() => {
				const expandKw = ['see all', 'show more', 'compare', 'all features', 'view features', 'expand', 'show all', 'more features'];
				const buttons = [];
				
				document.querySelectorAll('button, a, [role="button"], span, div').forEach((el, i) => {
					const text = el.textContent.toLowerCase().trim();
					if (text.length > 50) return;
					
					for (const kw of expandKw) {
						if (text.includes(kw)) {
							buttons.push({
								selector: el.id ? '#' + el.id : el.tagName.toLowerCase() + '[data-expand-idx="' + i + '"]',
								text: text,
								index: i
							});
							el.setAttribute('data-expand-idx', i);
							break;
						}
					}
				});
				
				return JSON.stringify(buttons.slice(0, 5));
			})()
		`, &expandButtonsJSON),
	)

	var expandButtons []struct {
		Selector string `json:"selector"`
		Text     string `json:"text"`
		Index    int    `json:"index"`
	}
	
	if err := json.Unmarshal([]byte(expandButtonsJSON), &expandButtons); err == nil {
		for _, btn := range expandButtons {
			log.Printf("[pricing-v2] clicking expand button: %s", btn.Text)
			chromedp.Run(ctx,
				chromedp.Click(btn.Selector, chromedp.ByQuery),
				chromedp.Sleep(500*time.Millisecond),
			)
		}
	}
}

// Debug artifact helpers
func (s *PricingV2Service) saveDebugArtifact(dir, filename, content string) {
	if !s.isStaging {
		return
	}
	path := filepath.Join(dir, filename)
	os.WriteFile(path, []byte(content), 0644)
}

func (s *PricingV2Service) saveDebugArtifacts(dir string, artifacts *DebugArtifacts) {
	if !s.isStaging {
		return
	}
	data, _ := json.MarshalIndent(artifacts, "", "  ")
	os.WriteFile(filepath.Join(dir, "artifacts.json"), data, 0644)
}

func (s *PricingV2Service) takeScreenshot(ctx context.Context, path string) {
	var buf []byte
	chromedp.Run(ctx, chromedp.CaptureScreenshot(&buf))
	if len(buf) > 0 {
		os.WriteFile(path, buf, 0644)
	}
}

// textSimilarity calculates a simple similarity ratio between two texts
func (s *PricingV2Service) textSimilarity(text1, text2 string) float64 {
	if text1 == text2 {
		return 1.0
	}
	
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
	
	return float64(matches) / float64(len(words1)+len(words2)-matches)
}

// deduplicatePlans removes duplicate plans using canonical key
func (s *PricingV2Service) deduplicatePlans(plans []model.ExtractedPlan) []model.ExtractedPlan {
	if len(plans) <= 1 {
		return plans
	}

	deduped := make(map[string]model.ExtractedPlan)
	
	for _, plan := range plans {
		key := s.canonicalPlanKey(plan)
		
		existing, exists := deduped[key]
		if !exists {
			deduped[key] = plan
			continue
		}
		
		merged := s.mergePlans(existing, plan)
		deduped[key] = merged
	}

	result := make([]model.ExtractedPlan, 0, len(deduped))
	for _, plan := range deduped {
		result = append(result, plan)
	}
	
	sort.Slice(result, func(i, j int) bool {
		if result[i].Name != result[j].Name {
			return result[i].Name < result[j].Name
		}
		return result[i].BillingPeriod < result[j].BillingPeriod
	})

	log.Printf("[pricing-v2] deduplicated %d plans to %d unique plans", len(plans), len(result))
	return result
}

func (s *PricingV2Service) canonicalPlanKey(plan model.ExtractedPlan) string {
	name := strings.ToLower(strings.TrimSpace(plan.Name))
	name = regexp.MustCompile(`\s+`).ReplaceAllString(name, " ")
	name = strings.TrimSuffix(name, " plan")
	name = strings.TrimSuffix(name, " tier")
	
	billing := strings.ToLower(plan.BillingPeriod)
	if billing == "" {
		billing = "unknown"
	}
	
	var priceKey string
	if plan.PriceAmount > 0 {
		priceKey = fmt.Sprintf("%.2f", plan.PriceAmount)
	} else if plan.PriceString != "" {
		re := regexp.MustCompile(`[\d,]+\.?\d*`)
		if match := re.FindString(plan.PriceString); match != "" {
			priceKey = strings.ReplaceAll(match, ",", "")
		}
	}
	
	return fmt.Sprintf("%s|%s|%s", name, billing, priceKey)
}

func (s *PricingV2Service) mergePlans(existing, new model.ExtractedPlan) model.ExtractedPlan {
	result := existing
	
	if len(new.Features) > len(result.Features) {
		result.Features = new.Features
	}
	if len(new.IncludedUnits) > len(result.IncludedUnits) {
		result.IncludedUnits = new.IncludedUnits
	}
	if len(new.Evidence.PriceSnippet) > len(result.Evidence.PriceSnippet) {
		result.Evidence = new.Evidence
	}
	if result.MonthlyEquivalentAmount == 0 && new.MonthlyEquivalentAmount > 0 {
		result.MonthlyEquivalentAmount = new.MonthlyEquivalentAmount
	}
	if result.AnnualBilledAmount == 0 && new.AnnualBilledAmount > 0 {
		result.AnnualBilledAmount = new.AnnualBilledAmount
	}
	
	return result
}

// SavePlans saves extracted plans to the database
func (s *PricingV2Service) SavePlans(ctx context.Context, userID string, req model.PricingSaveRequest) (*model.PricingSaveResponse, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return &model.PricingSaveResponse{Error: "invalid user ID"}, nil
	}

	if err := s.repo.DeleteByUserID(ctx, uid); err != nil {
		log.Printf("[pricing-v2] failed to delete existing plans: %v", err)
	}

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
		return &model.PricingSaveResponse{Error: fmt.Sprintf("failed to save plans: %v", err)}, nil
	}

	return &model.PricingSaveResponse{SavedCount: count}, nil
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

	return &model.SavedPricingV2Response{Plans: plans, Count: len(plans)}, nil
}

// DeletePlan deletes a specific pricing plan
func (s *PricingV2Service) DeletePlan(ctx context.Context, userID, planID string) error {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID")
	}

	pid, err := primitive.ObjectIDFromHex(planID)
	if err != nil {
		return fmt.Errorf("invalid plan ID")
	}

	return s.repo.Delete(ctx, pid, uid)
}

// ExtractFromText extracts pricing from pasted text (paste mode fallback)
func (s *PricingV2Service) ExtractFromText(ctx context.Context, req model.PricingExtractFromTextRequest) (*model.PricingExtractFromTextResponse, error) {
	var warnings []string
	
	// Validate input
	if req.MonthlyText == "" && req.YearlyText == "" {
		return &model.PricingExtractFromTextResponse{
			Error: "at least one of monthly_text or yearly_text is required",
		}, nil
	}

	// Add warnings for empty inputs
	if req.MonthlyText == "" {
		warnings = append(warnings, "monthly_text_empty")
	}
	if req.YearlyText == "" {
		warnings = append(warnings, "yearly_text_empty")
	}

	// Build combined content with clear section markers
	var combinedContent strings.Builder
	
	if req.MonthlyText != "" {
		combinedContent.WriteString("=== SNAPSHOT: MONTHLY BILLING MODE ===\n")
		combinedContent.WriteString("(User-pasted text for monthly billing view)\n\n")
		combinedContent.WriteString(req.MonthlyText)
		combinedContent.WriteString("\n\n")
	}
	
	if req.YearlyText != "" {
		combinedContent.WriteString("=== SNAPSHOT: YEARLY/ANNUAL BILLING MODE ===\n")
		combinedContent.WriteString("(User-pasted text for yearly billing view)\n\n")
		combinedContent.WriteString(req.YearlyText)
		combinedContent.WriteString("\n\n")
	}

	// Extract with LLM using the paste mode prompt
	plans, llmWarnings, err := s.extractFromPastedText(ctx, combinedContent.String(), req.WebsiteURL)
	if err != nil {
		return &model.PricingExtractFromTextResponse{
			Error:    fmt.Sprintf("extraction failed: %v", err),
			Warnings: warnings,
		}, nil
	}

	// Deduplicate plans
	plans = s.deduplicatePlans(plans)

	// Detect billing periods
	periods := s.detectBillingPeriods(plans)

	warnings = append(warnings, llmWarnings...)

	return &model.PricingExtractFromTextResponse{
		Plans:           plans,
		DetectedPeriods: periods,
		Warnings:        warnings,
	}, nil
}

// Paste mode specific LLM prompt
const pasteExtractionPrompt = `You are a pricing data extraction specialist. Extract pricing plan information from user-pasted text.

IMPORTANT: The user has manually copied and pasted pricing text from their website. This text is the ONLY source of truth.

STRICT RULES:
1. ONLY extract information that is EXPLICITLY present in the pasted text
2. If a field is not found, use null - NEVER guess or invent data
3. EVIDENCE IS REQUIRED: Include exact text snippets from the pasted text for every extracted value
4. The text is pre-labeled with billing mode (MONTHLY or YEARLY) - use this to set billing_period

SNAPSHOT HANDLING:
- If text is marked "SNAPSHOT: MONTHLY BILLING MODE", all plans from that section have billing_period: "monthly"
- If text is marked "SNAPSHOT: YEARLY/ANNUAL BILLING MODE", all plans from that section have billing_period: "yearly"
- Create SEPARATE plan entries for the same plan in different billing modes

BILLING PERIOD RULES:
- Plans from MONTHLY section → billing_period: "monthly"
- Plans from YEARLY section → billing_period: "yearly"  
- If yearly price shows monthly equivalent (e.g., "$10/mo billed annually"):
  - billing_period: "yearly"
  - monthly_equivalent_amount: 10
  - annual_billed_amount: 120

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
      "included_units": [],
      "features": ["Feature 1", "Feature 2"],
      "evidence": {
        "name_snippet": "exact text from pasted content",
        "price_snippet": "exact text showing the price",
        "units_snippet": "exact text showing included units",
        "billing_evidence": "MONTHLY BILLING MODE section"
      }
    }
  ],
  "detected_billing_options": ["monthly", "yearly"],
  "warnings": []
}

IMPORTANT:
- Extract from BOTH monthly and yearly sections if both are provided
- Currency: $ = USD, € = EUR, £ = GBP
- If features are not in the pasted text, return empty array and add warning
- DO NOT invent or guess any data - only extract what is explicitly written`

// extractFromPastedText uses OpenAI to extract pricing from pasted text
func (s *PricingV2Service) extractFromPastedText(ctx context.Context, content, websiteURL string) ([]model.ExtractedPlan, []string, error) {
	if s.openAIKey == "" {
		return nil, nil, fmt.Errorf("OpenAI API key not configured")
	}

	if len(content) > 50000 {
		content = content[:50000] + "\n...[truncated]"
	}

	sourceInfo := "user-pasted text"
	if websiteURL != "" {
		sourceInfo = fmt.Sprintf("user-pasted text from %s", websiteURL)
	}

	userPrompt := fmt.Sprintf(`Extract pricing information from this %s.

The user has manually copied and pasted this pricing text. Extract ONLY what is explicitly written.

Pasted Content:
%s`, sourceInfo, content)

	reqBody := map[string]interface{}{
		"model": "gpt-4o-mini",
		"messages": []map[string]string{
			{"role": "system", "content": pasteExtractionPrompt},
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
		log.Printf("[pricing-v2] failed to parse paste LLM response: %v, response: %s", err, response)
		return nil, []string{"parse_error"}, fmt.Errorf("failed to parse extraction result")
	}

	return result.Plans, result.Warnings, nil
}

// E) Updated LLM Extraction Prompt with dual snapshot support
const extractionPrompt = `You are a pricing data extraction specialist. Extract pricing plan information from the provided website content.

IMPORTANT: You may receive TWO SNAPSHOTS - one for monthly billing and one for yearly billing. Extract plans from BOTH snapshots separately.

STRICT RULES:
1. ONLY extract information that is EXPLICITLY present in the content
2. If a field is not found, use null - NEVER guess or invent data
3. EVIDENCE IS REQUIRED: Include exact text snippets for every extracted value
4. If billing period cannot be determined with evidence, set billing_period: "unknown" and add warning
5. If features are not visible in the content, return empty features array and add warning "features_not_visible_for_[plan_name]"

SNAPSHOT HANDLING:
- If you see "SNAPSHOT: MONTHLY BILLING MODE", extract plans with billing_period: "monthly"
- If you see "SNAPSHOT: YEARLY/ANNUAL BILLING MODE", extract plans with billing_period: "yearly"
- Create SEPARATE plan entries for the same plan in different billing modes
- The server will deduplicate, so err on the side of including both versions

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
      "included_units": [],
      "features": ["Feature 1", "Feature 2"],
      "evidence": {
        "name_snippet": "exact text where plan name appears",
        "price_snippet": "exact text showing the price AND billing period",
        "units_snippet": "exact text showing included units",
        "billing_evidence": "exact text proving the billing period"
      }
    }
  ],
  "detected_billing_options": ["monthly", "yearly"],
  "warnings": []
}

EXAMPLES:

Example 1 - Same plan in both modes:
From MONTHLY SNAPSHOT: "Pro $12/mo billed monthly" → billing_period: "monthly", price_amount: 12
From YEARLY SNAPSHOT: "Pro $8/mo billed annually" → billing_period: "yearly", monthly_equivalent_amount: 8, annual_billed_amount: 96

Example 2 - Features not visible:
If no features are shown for a plan, return empty array and add to warnings: "features_not_visible_for_Pro"

IMPORTANT:
- Extract plans from BOTH monthly and yearly snapshots if both are present
- Currency: $ = USD, € = EUR, £ = GBP
- Always include billing_evidence in evidence object
- If pricing requires login/contact sales, add "pricing_gated" to warnings`

// extractWithLLM uses OpenAI to extract pricing from page content
func (s *PricingV2Service) extractWithLLM(ctx context.Context, content, rawHTML, sourceURL string) ([]model.ExtractedPlan, []string, error) {
	if s.openAIKey == "" {
		return nil, nil, fmt.Errorf("OpenAI API key not configured")
	}

	// Limit content size for faster LLM processing
	if len(content) > 15000 {
		content = content[:15000] + "\n...[truncated]"
	}

	userPrompt := fmt.Sprintf(`Extract pricing information from this page. You may receive multiple snapshots for different billing modes.

Source URL: %s

Page Content:
%s`, sourceURL, content)

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

// Helper functions (unchanged)
func (s *PricingV2Service) extractHiddenContent(htmlContent string) string {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return ""
	}

	var hiddenText strings.Builder
	var extractHidden func(*html.Node)

	extractHidden = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if n.Data == "script" || n.Data == "style" || n.Data == "noscript" {
				return
			}

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
				if attr.Key == "role" && attr.Val == "tabpanel" {
					isHidden = true
				}
			}

			if isHidden {
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

func (s *PricingV2Service) extractScriptJSON(htmlContent string) string {
	var jsonData strings.Builder

	nextDataRe := regexp.MustCompile(`<script[^>]*id="__NEXT_DATA__"[^>]*>([\s\S]*?)</script>`)
	if matches := nextDataRe.FindStringSubmatch(htmlContent); len(matches) > 1 {
		data := matches[1]
		if len(data) < 50000 {
			jsonData.WriteString("NEXT_DATA: ")
			jsonData.WriteString(s.extractPricingFromJSON(data))
			jsonData.WriteString("\n")
		}
	}

	ldJsonRe := regexp.MustCompile(`<script[^>]*type="application/ld\+json"[^>]*>([\s\S]*?)</script>`)
	ldMatches := ldJsonRe.FindAllStringSubmatch(htmlContent, -1)
	for _, match := range ldMatches {
		if len(match) > 1 && len(match[1]) < 10000 {
			jsonData.WriteString("LD+JSON: ")
			jsonData.WriteString(match[1])
			jsonData.WriteString("\n")
		}
	}

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

func (s *PricingV2Service) extractPricingFromJSON(jsonStr string) string {
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

	if len(jsonStr) > 5000 {
		return jsonStr[:5000] + "...[truncated]"
	}
	return jsonStr
}

func (s *PricingV2Service) detectBillingToggle(visibleText, rawHTML string) bool {
	contentLower := strings.ToLower(visibleText + " " + rawHTML)

	toggleCount := 0
	for _, indicator := range toggleIndicators {
		if strings.Contains(contentLower, indicator) {
			toggleCount++
		}
	}

	hasTabList := strings.Contains(rawHTML, `role="tablist"`) ||
		strings.Contains(rawHTML, `role="tab"`) ||
		strings.Contains(rawHTML, "toggle") ||
		strings.Contains(rawHTML, "switch")

	return toggleCount >= 2 || (toggleCount >= 1 && hasTabList)
}

func (s *PricingV2Service) normalizeURL(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return ""
	}

	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "https://" + rawURL
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

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

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("only http/https URLs allowed")
	}

	host := parsed.Hostname()
	if host == "localhost" || host == "127.0.0.1" || host == "0.0.0.0" {
		return fmt.Errorf("localhost not allowed")
	}

	ip := net.ParseIP(host)
	if ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() {
			return fmt.Errorf("private/internal IPs not allowed")
		}
	}

	return nil
}

func (s *PricingV2Service) urlExists(ctx context.Context, testURL string) bool {
	// Use GET instead of HEAD - many sites block HEAD requests
	req, err := http.NewRequestWithContext(ctx, "GET", testURL, nil)
	if err != nil {
		log.Printf("[pricing-v2] urlExists failed to create request for %s: %v", testURL, err)
		return false
	}
	
	// Use a realistic browser User-Agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		log.Printf("[pricing-v2] urlExists request failed for %s: %v", testURL, err)
		return false
	}
	defer resp.Body.Close()
	
	// Read a small amount to not hang the connection, then close
	io.CopyN(io.Discard, resp.Body, 1024)

	// Accept 200, 301, 302, 303, 307, 308 as "exists"
	exists := resp.StatusCode >= 200 && resp.StatusCode < 400
	log.Printf("[pricing-v2] urlExists %s -> %d (exists=%v)", testURL, resp.StatusCode, exists)
	return exists
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
		re := regexp.MustCompile(`<[^>]*>`)
		return re.ReplaceAllString(htmlContent, " ")
	}

	var textBuilder strings.Builder
	var extractText func(*html.Node)

	extractText = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "script", "style", "noscript", "head", "meta", "link":
				return
			}

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
