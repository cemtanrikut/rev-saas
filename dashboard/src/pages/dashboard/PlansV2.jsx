import { useState, useEffect } from 'react';
import { pricingV2Api } from '../../lib/apiClient';

const PlansV2 = () => {
  // Website URL state
  const [websiteUrl, setWebsiteUrl] = useState('');
  const [pricingUrl, setPricingUrl] = useState('');
  const [manualPricingUrl, setManualPricingUrl] = useState('');
  
  // Discovery state
  const [candidates, setCandidates] = useState([]);
  const [isDiscovering, setIsDiscovering] = useState(false);
  const [discoverError, setDiscoverError] = useState('');
  
  // Extraction state
  const [extractedPlans, setExtractedPlans] = useState([]);
  const [detectedPeriods, setDetectedPeriods] = useState([]);
  const [warnings, setWarnings] = useState([]);
  const [isExtracting, setIsExtracting] = useState(false);
  const [extractError, setExtractError] = useState('');
  
  // Selection state
  const [selectedPlans, setSelectedPlans] = useState(new Set());
  
  // Save state
  const [isSaving, setIsSaving] = useState(false);
  const [saveSuccess, setSaveSuccess] = useState(false);
  
  // Saved plans state
  const [savedPlans, setSavedPlans] = useState([]);
  const [isLoadingSaved, setIsLoadingSaved] = useState(true);
  
  // View state
  const [showManualInput, setShowManualInput] = useState(false);
  const [activeFilter, setActiveFilter] = useState('all');

  // Load saved plans on mount
  useEffect(() => {
    loadSavedPlans();
  }, []);

  const loadSavedPlans = async () => {
    try {
      setIsLoadingSaved(true);
      const response = await pricingV2Api.list();
      setSavedPlans(response.plans || []);
    } catch (err) {
      console.error('Failed to load saved plans:', err);
    } finally {
      setIsLoadingSaved(false);
    }
  };

  const handleDiscover = async () => {
    if (!websiteUrl.trim()) {
      setDiscoverError('Please enter a website URL');
      return;
    }

    setIsDiscovering(true);
    setDiscoverError('');
    setCandidates([]);
    setPricingUrl('');
    setExtractedPlans([]);

    try {
      const response = await pricingV2Api.discover({ website_url: websiteUrl });
      
      if (response.error) {
        setDiscoverError(response.error);
        setShowManualInput(true);
        return;
      }

      setCandidates(response.pricing_candidates || []);
      
      if (response.selected_pricing_url) {
        setPricingUrl(response.selected_pricing_url);
        // Auto-extract
        await handleExtract(response.selected_pricing_url);
      } else {
        setShowManualInput(true);
        setDiscoverError('Could not auto-detect pricing page. Please enter the URL manually.');
      }
    } catch (err) {
      setDiscoverError(err.message || 'Discovery failed');
      setShowManualInput(true);
    } finally {
      setIsDiscovering(false);
    }
  };

  const handleExtract = async (url) => {
    const targetUrl = url || manualPricingUrl || pricingUrl;
    
    if (!targetUrl) {
      setExtractError('Please enter a pricing page URL');
      return;
    }

    setIsExtracting(true);
    setExtractError('');
    setExtractedPlans([]);
    setWarnings([]);
    setSelectedPlans(new Set());

    try {
      const response = await pricingV2Api.extract({ pricing_url: targetUrl });
      
      if (response.error) {
        setExtractError(response.error);
        return;
      }

      setPricingUrl(targetUrl);
      setExtractedPlans(response.plans || []);
      setDetectedPeriods(response.detected_periods || []);
      setWarnings(response.warnings || []);
      
      // Auto-select all plans
      const allIds = new Set(response.plans?.map((_, i) => i) || []);
      setSelectedPlans(allIds);
    } catch (err) {
      setExtractError(err.message || 'Extraction failed');
    } finally {
      setIsExtracting(false);
    }
  };

  const handleSave = async () => {
    const plansToSave = extractedPlans.filter((_, i) => selectedPlans.has(i));
    
    if (plansToSave.length === 0) {
      return;
    }

    setIsSaving(true);
    setSaveSuccess(false);

    try {
      const response = await pricingV2Api.save({
        plans: plansToSave,
        source_url: pricingUrl,
        website_url: websiteUrl
      });

      if (response.error) {
        setExtractError(response.error);
        return;
      }

      setSaveSuccess(true);
      setExtractedPlans([]);
      setSelectedPlans(new Set());
      
      // Reload saved plans
      await loadSavedPlans();
      
      // Clear success after 3 seconds
      setTimeout(() => setSaveSuccess(false), 3000);
    } catch (err) {
      setExtractError(err.message || 'Save failed');
    } finally {
      setIsSaving(false);
    }
  };

  const togglePlanSelection = (index) => {
    const newSelected = new Set(selectedPlans);
    if (newSelected.has(index)) {
      newSelected.delete(index);
    } else {
      newSelected.add(index);
    }
    setSelectedPlans(newSelected);
  };

  const selectAllPlans = () => {
    setSelectedPlans(new Set(extractedPlans.map((_, i) => i)));
  };

  const deselectAllPlans = () => {
    setSelectedPlans(new Set());
  };

  const formatPrice = (amount, currency) => {
    if (!amount) return 'N/A';
    const formatter = new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: currency || 'USD',
      minimumFractionDigits: 0,
      maximumFractionDigits: 2
    });
    return formatter.format(amount);
  };

  const filteredPlans = extractedPlans.filter(plan => {
    if (activeFilter === 'all') return true;
    return plan.billing_period === activeFilter;
  });

  const filteredSavedPlans = savedPlans.filter(plan => {
    if (activeFilter === 'all') return true;
    return plan.billing_period === activeFilter;
  });

  return (
    <div className="max-w-6xl mx-auto space-y-8">
      {/* Page Header */}
      <div>
        <h1 className="text-3xl font-bold text-white mb-2">
          My Pricing V2
        </h1>
        <p className="text-slate-400">
          Auto-import your pricing plans from your website. This experimental feature extracts pricing data using AI.
        </p>
      </div>

      {/* Auto-Import Section */}
      <div className="bg-slate-900/50 backdrop-blur-sm rounded-2xl p-6 border border-slate-800">
        <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
          <svg className="w-5 h-5 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
          </svg>
          Auto-detect & Import Pricing
        </h3>

        {/* Website URL Input */}
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-semibold text-slate-300 mb-2">
              Your Website URL
            </label>
            <div className="flex gap-3">
              <input
                type="url"
                value={websiteUrl}
                onChange={(e) => setWebsiteUrl(e.target.value)}
                placeholder="https://www.example.com"
                className="flex-1 px-4 py-3 rounded-xl bg-slate-900/50 border border-slate-700 text-white placeholder-slate-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all"
              />
              <button
                onClick={handleDiscover}
                disabled={isDiscovering || isExtracting}
                className="px-6 py-3 bg-gradient-to-r from-blue-500 to-indigo-600 text-white rounded-xl font-semibold hover:from-blue-600 hover:to-indigo-700 hover:scale-105 transition-all shadow-lg shadow-blue-500/20 disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:scale-100 whitespace-nowrap"
              >
                {isDiscovering ? (
                  <span className="flex items-center gap-2">
                    <svg className="animate-spin h-5 w-5" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                    </svg>
                    Discovering...
                  </span>
                ) : (
                  'Auto-detect & Import'
                )}
              </button>
            </div>
          </div>

          {/* Error Display */}
          {discoverError && (
            <div className="bg-red-500/10 border border-red-500/30 rounded-xl p-4">
              <p className="text-red-400 text-sm">{discoverError}</p>
            </div>
          )}

          {/* Candidates Display */}
          {candidates.length > 0 && (
            <div className="bg-slate-800/30 rounded-xl p-4">
              <p className="text-sm text-slate-400 mb-2">Found pricing page candidates:</p>
              <div className="flex flex-wrap gap-2">
                {candidates.map((url, i) => (
                  <button
                    key={i}
                    onClick={() => handleExtract(url)}
                    className={`px-3 py-1.5 rounded-lg text-sm transition-all ${
                      pricingUrl === url
                        ? 'bg-blue-500 text-white'
                        : 'bg-slate-700 text-slate-300 hover:bg-slate-600'
                    }`}
                  >
                    {new URL(url).pathname || '/'}
                  </button>
                ))}
              </div>
            </div>
          )}

          {/* Manual URL Input */}
          {showManualInput && (
            <div className="pt-4 border-t border-slate-700">
              <label className="block text-sm font-semibold text-slate-300 mb-2">
                Or enter pricing page URL manually:
              </label>
              <div className="flex gap-3">
                <input
                  type="url"
                  value={manualPricingUrl}
                  onChange={(e) => setManualPricingUrl(e.target.value)}
                  placeholder="https://www.example.com/pricing"
                  className="flex-1 px-4 py-3 rounded-xl bg-slate-900/50 border border-slate-700 text-white placeholder-slate-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all"
                />
                <button
                  onClick={() => handleExtract()}
                  disabled={isExtracting || !manualPricingUrl}
                  className="px-6 py-3 bg-slate-700 text-white rounded-xl font-semibold hover:bg-slate-600 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {isExtracting ? 'Extracting...' : 'Extract'}
                </button>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Extraction Results */}
      {(isExtracting || extractedPlans.length > 0 || extractError) && (
        <div className="bg-slate-900/50 backdrop-blur-sm rounded-2xl p-6 border border-slate-800">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold text-white flex items-center gap-2">
              <svg className="w-5 h-5 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
              </svg>
              Extracted Plans
              {pricingUrl && (
                <span className="text-sm font-normal text-slate-500">
                  from {new URL(pricingUrl).hostname}
                </span>
              )}
            </h3>

            {/* Billing Period Filter */}
            {detectedPeriods.length > 1 && (
              <div className="flex gap-2">
                <button
                  onClick={() => setActiveFilter('all')}
                  className={`px-3 py-1.5 rounded-lg text-sm transition-all ${
                    activeFilter === 'all' ? 'bg-blue-500 text-white' : 'bg-slate-700 text-slate-300 hover:bg-slate-600'
                  }`}
                >
                  All
                </button>
                {detectedPeriods.map(period => (
                  <button
                    key={period}
                    onClick={() => setActiveFilter(period)}
                    className={`px-3 py-1.5 rounded-lg text-sm capitalize transition-all ${
                      activeFilter === period ? 'bg-blue-500 text-white' : 'bg-slate-700 text-slate-300 hover:bg-slate-600'
                    }`}
                  >
                    {period}
                  </button>
                ))}
              </div>
            )}
          </div>

          {/* Render Status Badge */}
          {extractedPlans.length > 0 && (
            <div className="flex items-center gap-2 mb-4">
              {detectedPeriods.length > 1 && (
                <span className="px-2 py-1 bg-emerald-500/10 text-emerald-400 text-xs rounded-lg flex items-center gap-1">
                  <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                  </svg>
                  Both monthly & yearly detected
                </span>
              )}
            </div>
          )}

          {/* Warnings */}
          {warnings.length > 0 && (
            <div className="bg-amber-500/10 border border-amber-500/30 rounded-xl p-3 mb-4">
              <p className="text-amber-400 text-sm flex items-center gap-2">
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                </svg>
                {warnings.join(', ')}
              </p>
            </div>
          )}

          {/* Error */}
          {extractError && (
            <div className="bg-red-500/10 border border-red-500/30 rounded-xl p-4 mb-4">
              <p className="text-red-400 text-sm">{extractError}</p>
            </div>
          )}

          {/* Loading */}
          {isExtracting && (
            <div className="py-12 text-center">
              <div className="w-12 h-12 border-4 border-blue-500 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
              <p className="text-slate-400">Extracting pricing information with AI...</p>
            </div>
          )}

          {/* Plans Grid */}
          {!isExtracting && filteredPlans.length > 0 && (
            <>
              {/* Selection Controls */}
              <div className="flex items-center justify-between mb-4 pb-4 border-b border-slate-700">
                <div className="flex items-center gap-4">
                  <button
                    onClick={selectAllPlans}
                    className="text-sm text-blue-400 hover:text-blue-300"
                  >
                    Select All
                  </button>
                  <button
                    onClick={deselectAllPlans}
                    className="text-sm text-slate-400 hover:text-slate-300"
                  >
                    Deselect All
                  </button>
                  <span className="text-sm text-slate-500">
                    {selectedPlans.size} of {extractedPlans.length} selected
                  </span>
                </div>
              </div>

              <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-4">
                {filteredPlans.map((plan, index) => (
                  <div
                    key={index}
                    onClick={() => togglePlanSelection(index)}
                    className={`p-5 rounded-xl border cursor-pointer transition-all ${
                      selectedPlans.has(index)
                        ? 'bg-blue-500/10 border-blue-500/50'
                        : 'bg-slate-800/30 border-slate-700 hover:border-slate-600'
                    }`}
                  >
                    {/* Checkbox */}
                    <div className="flex items-start justify-between mb-3">
                      <div className="flex items-center gap-3">
                        <div className={`w-5 h-5 rounded border-2 flex items-center justify-center transition-all ${
                          selectedPlans.has(index)
                            ? 'bg-blue-500 border-blue-500'
                            : 'border-slate-600'
                        }`}>
                          {selectedPlans.has(index) && (
                            <svg className="w-3 h-3 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" />
                            </svg>
                          )}
                        </div>
                        <h4 className="text-lg font-bold text-white">{plan.name}</h4>
                      </div>
                      <span className={`px-2 py-0.5 rounded text-xs font-medium ${
                        plan.billing_period === 'monthly' ? 'bg-blue-500/20 text-blue-300' :
                        plan.billing_period === 'yearly' ? 'bg-emerald-500/20 text-emerald-300' :
                        'bg-slate-500/20 text-slate-300'
                      }`}>
                        {plan.billing_period || 'unknown'}
                      </span>
                    </div>

                    {/* Price */}
                    <div className="mb-3">
                      <div className="text-2xl font-bold text-blue-400">
                        {plan.price_string || formatPrice(plan.price_amount, plan.currency)}
                      </div>
                      {/* Monthly equivalent for yearly plans */}
                      {plan.billing_period === 'yearly' && plan.monthly_equivalent_amount > 0 && (
                        <div className="text-sm text-slate-400 mt-1">
                          {formatPrice(plan.monthly_equivalent_amount, plan.currency)}/mo equivalent
                        </div>
                      )}
                      {/* Annual total for yearly plans */}
                      {plan.billing_period === 'yearly' && plan.annual_billed_amount > 0 && (
                        <div className="text-xs text-slate-500">
                          {formatPrice(plan.annual_billed_amount, plan.currency)}/year billed
                        </div>
                      )}
                    </div>

                    {/* Included Units */}
                    {plan.included_units?.length > 0 && (
                      <div className="mb-3 space-y-1">
                        {plan.included_units.map((unit, i) => (
                          <div key={i} className="text-sm text-slate-400 flex items-center gap-2">
                            <svg className="w-4 h-4 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                            </svg>
                            {unit.raw_text || `${unit.amount} ${unit.name}`}
                          </div>
                        ))}
                      </div>
                    )}

                    {/* Features */}
                    {plan.features?.length > 0 && (
                      <div className="pt-3 border-t border-slate-700">
                        <p className="text-xs text-slate-500 mb-2">Top Features:</p>
                        <ul className="space-y-1">
                          {plan.features.slice(0, 3).map((feature, i) => (
                            <li key={i} className="text-sm text-slate-400 flex items-center gap-2">
                              <span className="w-1 h-1 bg-slate-500 rounded-full"></span>
                              {feature}
                            </li>
                          ))}
                          {plan.features.length > 3 && (
                            <li className="text-sm text-slate-500">
                              +{plan.features.length - 3} more
                            </li>
                          )}
                        </ul>
                      </div>
                    )}

                    {/* Billing Evidence */}
                    {plan.evidence?.billing_evidence && (
                      <div className="mt-3 pt-3 border-t border-slate-700/50">
                        <p className="text-xs text-slate-500 flex items-center gap-1">
                          <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                          </svg>
                          Evidence: "{plan.evidence.billing_evidence}"
                        </p>
                      </div>
                    )}
                  </div>
                ))}
              </div>

              {/* Save Button */}
              <div className="mt-6 flex items-center justify-between pt-6 border-t border-slate-700">
                <p className="text-sm text-slate-400">
                  Selected plans will be saved to your My Pricing V2
                </p>
                <button
                  onClick={handleSave}
                  disabled={selectedPlans.size === 0 || isSaving}
                  className="px-6 py-3 bg-gradient-to-r from-emerald-500 to-teal-600 text-white rounded-xl font-semibold hover:from-emerald-600 hover:to-teal-700 hover:scale-105 transition-all shadow-lg shadow-emerald-500/20 disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:scale-100"
                >
                  {isSaving ? (
                    <span className="flex items-center gap-2">
                      <svg className="animate-spin h-5 w-5" viewBox="0 0 24 24">
                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                      </svg>
                      Saving...
                    </span>
                  ) : (
                    `Save ${selectedPlans.size} Plan${selectedPlans.size !== 1 ? 's' : ''} to My Pricing V2`
                  )}
                </button>
              </div>
            </>
          )}

          {/* Empty State */}
          {!isExtracting && filteredPlans.length === 0 && !extractError && (
            <div className="py-12 text-center">
              <p className="text-slate-400">No plans found. Try a different pricing page URL.</p>
            </div>
          )}
        </div>
      )}

      {/* Success Message */}
      {saveSuccess && (
        <div className="bg-emerald-500/10 border border-emerald-500/30 rounded-xl p-4 flex items-center gap-3">
          <svg className="w-5 h-5 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
          </svg>
          <p className="text-emerald-400">Plans saved successfully!</p>
        </div>
      )}

      {/* Saved Plans Section */}
      <div className="bg-slate-900/50 backdrop-blur-sm rounded-2xl border border-slate-800 overflow-hidden">
        <div className="p-6 border-b border-slate-800">
          <h3 className="text-lg font-semibold text-white">
            Saved Plans ({savedPlans.length})
          </h3>
        </div>

        {isLoadingSaved ? (
          <div className="p-12 text-center">
            <div className="w-12 h-12 border-4 border-blue-500 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
            <p className="text-slate-400">Loading saved plans...</p>
          </div>
        ) : savedPlans.length === 0 ? (
          <div className="p-12 text-center">
            <div className="w-16 h-16 bg-slate-800 rounded-2xl flex items-center justify-center mx-auto mb-4">
              <svg className="w-8 h-8 text-slate-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 14l6-6m-5.5.5h.01m4.99 5h.01M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16l3.5-2 3.5 2 3.5-2 3.5 2zM10 8.5a.5.5 0 11-1 0 .5.5 0 011 0zm5 5a.5.5 0 11-1 0 .5.5 0 011 0z" />
              </svg>
            </div>
            <p className="text-slate-400 text-lg">No plans imported yet.</p>
            <p className="text-slate-500 text-sm mt-2">
              Use the auto-import feature above to extract pricing from your website.
            </p>
          </div>
        ) : (
          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-4 p-6">
            {filteredSavedPlans.map((plan) => (
              <div
                key={plan.id}
                className="bg-slate-800/30 rounded-xl p-5 border border-slate-700"
              >
                <div className="flex items-start justify-between mb-3">
                  <h4 className="text-lg font-bold text-white">{plan.plan_name}</h4>
                  <span className={`px-2 py-0.5 rounded text-xs font-medium ${
                    plan.billing_period === 'monthly' ? 'bg-blue-500/20 text-blue-300' :
                    plan.billing_period === 'yearly' ? 'bg-emerald-500/20 text-emerald-300' :
                    'bg-slate-500/20 text-slate-300'
                  }`}>
                    {plan.billing_period || 'unknown'}
                  </span>
                </div>

                <div className="mb-2">
                  <div className="text-2xl font-bold text-blue-400">
                    {plan.price_string || formatPrice(plan.price_amount, plan.currency)}
                  </div>
                  {/* Monthly equivalent for yearly plans */}
                  {plan.billing_period === 'yearly' && plan.monthly_equivalent_amount > 0 && (
                    <div className="text-sm text-slate-400">
                      {formatPrice(plan.monthly_equivalent_amount, plan.currency)}/mo equivalent
                    </div>
                  )}
                </div>

                {plan.source_url && (
                  <p className="text-xs text-slate-500 truncate" title={plan.source_url}>
                    Source: {new URL(plan.source_url).hostname}
                  </p>
                )}

                {plan.features?.length > 0 && (
                  <div className="mt-3 pt-3 border-t border-slate-700">
                    <ul className="space-y-1">
                      {plan.features.slice(0, 3).map((feature, i) => (
                        <li key={i} className="text-sm text-slate-400 flex items-center gap-2">
                          <span className="w-1 h-1 bg-slate-500 rounded-full"></span>
                          {feature}
                        </li>
                      ))}
                    </ul>
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default PlansV2;
