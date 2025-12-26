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
  const [activeTab, setActiveTab] = useState('import'); // 'import' or 'saved'
  
  // Paste mode state
  const [showPasteMode, setShowPasteMode] = useState(false);
  const [monthlyText, setMonthlyText] = useState('');
  const [yearlyText, setYearlyText] = useState('');
  const [isPasteExtracting, setIsPasteExtracting] = useState(false);

  // Load saved plans on mount
  useEffect(() => {
    loadSavedPlans();
  }, []);

  const loadSavedPlans = async () => {
    try {
      setIsLoadingSaved(true);
      const response = await pricingV2Api.list();
      const data = response.data || response;
      setSavedPlans(data.plans || []);
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
      const data = response.data || response;
      
      if (data.error) {
        setDiscoverError(data.error);
        setShowManualInput(true);
        return;
      }

      setCandidates(data.pricing_candidates || []);
      
      if (data.selected_pricing_url) {
        setPricingUrl(data.selected_pricing_url);
        await handleExtract(data.selected_pricing_url);
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
      const data = response.data || response;
      
      if (data.error) {
        setExtractError(data.error);
        return;
      }

      setPricingUrl(targetUrl);
      setExtractedPlans(data.plans || []);
      setDetectedPeriods(data.detected_periods || []);
      setWarnings(data.warnings || []);
      
      const allIds = new Set(data.plans?.map((_, i) => i) || []);
      setSelectedPlans(allIds);
    } catch (err) {
      setExtractError(err.message || 'Extraction failed');
    } finally {
      setIsExtracting(false);
    }
  };

  const handleExtractFromText = async () => {
    if (!monthlyText.trim() && !yearlyText.trim()) {
      setExtractError('Please paste at least one pricing text (monthly or yearly view)');
      return;
    }

    setIsPasteExtracting(true);
    setExtractError('');
    setExtractedPlans([]);
    setWarnings([]);
    setSelectedPlans(new Set());

    try {
      const response = await pricingV2Api.extractFromText({
        monthly_text: monthlyText.trim(),
        yearly_text: yearlyText.trim(),
        website_url: websiteUrl.trim() || undefined
      });
      const data = response.data || response;
      
      if (data.error) {
        setExtractError(data.error);
        return;
      }

      setPricingUrl('pasted-text');
      setExtractedPlans(data.plans || []);
      setDetectedPeriods(data.detected_periods || []);
      setWarnings(data.warnings || []);
      
      const allIds = new Set(data.plans?.map((_, i) => i) || []);
      setSelectedPlans(allIds);
    } catch (err) {
      setExtractError(err.message || 'Extraction from pasted text failed');
    } finally {
      setIsPasteExtracting(false);
    }
  };

  const handleSave = async () => {
    const plansToSave = extractedPlans.filter((_, i) => selectedPlans.has(i));
    
    if (plansToSave.length === 0) return;

    setIsSaving(true);
    setSaveSuccess(false);

    try {
      const response = await pricingV2Api.save({
        plans: plansToSave,
        source_url: pricingUrl,
        website_url: websiteUrl
      });
      const data = response.data || response;

      if (data.error) {
        setExtractError(data.error);
        return;
      }

      setSaveSuccess(true);
      setExtractedPlans([]);
      setSelectedPlans(new Set());
      await loadSavedPlans();
      setActiveTab('saved');
      setTimeout(() => setSaveSuccess(false), 3000);
    } catch (err) {
      setExtractError(err.message || 'Save failed');
    } finally {
      setIsSaving(false);
    }
  };

  const handleDeletePlan = async (planId) => {
    if (!window.confirm('Are you sure you want to delete this plan?')) return;
    
    try {
      const response = await pricingV2Api.delete(planId);
      const data = response.data || response;
      
      if (data.error) {
        console.error('Delete failed:', data.error);
        return;
      }
      
      // Remove from local state
      setSavedPlans(prev => prev.filter(p => p.id !== planId));
    } catch (err) {
      console.error('Delete failed:', err);
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

  const currentStep = extractedPlans.length > 0 ? 3 : (isDiscovering || isExtracting || pricingUrl) ? 2 : 1;

  return (
    <div className="min-h-screen">
      {/* Hero Header */}
      <div className="relative mb-10">
        <div className="absolute inset-0 bg-gradient-to-r from-violet-600/20 via-fuchsia-600/10 to-transparent rounded-3xl blur-3xl" />
        <div className="relative">
          <div className="flex items-center gap-4 mb-3">
            <div className="w-14 h-14 rounded-2xl bg-gradient-to-br from-violet-500 to-fuchsia-600 flex items-center justify-center shadow-lg shadow-violet-500/30">
              <svg className="w-7 h-7 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 14l6-6m-5.5.5h.01m4.99 5h.01M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16l3.5-2 3.5 2 3.5-2 3.5 2zM10 8.5a.5.5 0 11-1 0 .5.5 0 011 0zm5 5a.5.5 0 11-1 0 .5.5 0 011 0z" />
              </svg>
            </div>
            <div>
              <h1 className="text-4xl font-bold bg-gradient-to-r from-white via-violet-200 to-fuchsia-200 bg-clip-text text-transparent">
                My Pricing
              </h1>
              <p className="text-slate-400 text-lg">
                Import your pricing plans automatically with AI
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* Tab Navigation */}
      <div className="flex gap-2 mb-8">
        <button
          onClick={() => setActiveTab('import')}
          className={`px-6 py-3 rounded-xl font-medium transition-all ${
            activeTab === 'import'
              ? 'bg-gradient-to-r from-violet-500 to-fuchsia-600 text-white shadow-lg shadow-violet-500/25'
              : 'bg-slate-800/50 text-slate-400 hover:text-white hover:bg-slate-800'
          }`}
        >
          <span className="flex items-center gap-2">
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
            </svg>
            Import Plans
          </span>
        </button>
        <button
          onClick={() => setActiveTab('saved')}
          className={`px-6 py-3 rounded-xl font-medium transition-all ${
            activeTab === 'saved'
              ? 'bg-gradient-to-r from-violet-500 to-fuchsia-600 text-white shadow-lg shadow-violet-500/25'
              : 'bg-slate-800/50 text-slate-400 hover:text-white hover:bg-slate-800'
          }`}
        >
          <span className="flex items-center gap-2">
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
            </svg>
            Saved Plans
            {savedPlans.length > 0 && (
              <span className="bg-white/20 text-white text-xs px-2 py-0.5 rounded-full">
                {savedPlans.length}
              </span>
            )}
          </span>
        </button>
      </div>

      {activeTab === 'import' && (
        <div className="space-y-6">
          {/* Progress Steps */}
          <div className="flex items-center justify-center gap-4 mb-8">
            {[
              { num: 1, label: 'Enter URL' },
              { num: 2, label: 'Extract Plans' },
              { num: 3, label: 'Review & Save' }
            ].map((step, idx) => (
              <div key={step.num} className="flex items-center gap-4">
                <div className="flex items-center gap-3">
                  <div className={`w-10 h-10 rounded-full flex items-center justify-center font-bold transition-all ${
                    currentStep >= step.num
                      ? 'bg-gradient-to-r from-violet-500 to-fuchsia-600 text-white shadow-lg shadow-violet-500/30'
                      : 'bg-slate-800 text-slate-500'
                  }`}>
                    {currentStep > step.num ? (
                      <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" />
                      </svg>
                    ) : step.num}
                  </div>
                  <span className={`font-medium hidden sm:block ${currentStep >= step.num ? 'text-white' : 'text-slate-500'}`}>
                    {step.label}
                  </span>
                </div>
                {idx < 2 && (
                  <div className={`w-12 h-0.5 ${currentStep > step.num ? 'bg-violet-500' : 'bg-slate-700'}`} />
                )}
              </div>
            ))}
          </div>

          {/* URL Input Card */}
          <div className="relative group">
            <div className="absolute -inset-1 bg-gradient-to-r from-violet-600 to-fuchsia-600 rounded-3xl blur opacity-25 group-hover:opacity-40 transition-opacity" />
            <div className="relative bg-slate-900/80 backdrop-blur-xl rounded-2xl p-8 border border-slate-700/50">
              <div className="flex items-start gap-4 mb-6">
                <div className="w-12 h-12 rounded-xl bg-violet-500/10 flex items-center justify-center flex-shrink-0">
                  <svg className="w-6 h-6 text-violet-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
                  </svg>
                </div>
                <div>
                  <h3 className="text-xl font-bold text-white mb-1">Enter Your Website</h3>
                  <p className="text-slate-400">We'll automatically find and extract your pricing plans</p>
                </div>
              </div>

              <div className="flex gap-3">
                <div className="flex-1 relative">
                  <div className="absolute inset-y-0 left-4 flex items-center pointer-events-none">
                    <svg className="w-5 h-5 text-slate-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
                    </svg>
                  </div>
                  <input
                    type="url"
                    value={websiteUrl}
                    onChange={(e) => setWebsiteUrl(e.target.value)}
                    onKeyDown={(e) => e.key === 'Enter' && handleDiscover()}
                    placeholder="https://www.yourcompany.com"
                    className="w-full pl-12 pr-4 py-4 rounded-xl bg-slate-800/50 border-2 border-slate-700 text-white text-lg placeholder-slate-500 focus:border-violet-500 focus:ring-4 focus:ring-violet-500/20 outline-none transition-all"
                  />
                </div>
                <button
                  onClick={handleDiscover}
                  disabled={isDiscovering || isExtracting || !websiteUrl.trim()}
                  className="px-8 py-4 bg-gradient-to-r from-violet-500 to-fuchsia-600 text-white rounded-xl font-bold text-lg hover:from-violet-600 hover:to-fuchsia-700 hover:scale-[1.02] hover:shadow-xl hover:shadow-violet-500/25 transition-all disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:scale-100 whitespace-nowrap"
                >
                  {isDiscovering ? (
                    <span className="flex items-center gap-3">
                      <svg className="animate-spin h-5 w-5" viewBox="0 0 24 24">
                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                      </svg>
                      Finding...
                    </span>
                  ) : isExtracting ? (
                    <span className="flex items-center gap-3">
                      <svg className="animate-spin h-5 w-5" viewBox="0 0 24 24">
                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                      </svg>
                      Extracting...
                    </span>
                  ) : (
                    <span className="flex items-center gap-2">
                      <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                      </svg>
                      Import
                    </span>
                  )}
                </button>
              </div>

              {/* Error */}
              {discoverError && (
                <div className="mt-4 bg-red-500/10 border border-red-500/30 rounded-xl p-4 flex items-start gap-3">
                  <svg className="w-5 h-5 text-red-400 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  <p className="text-red-300">{discoverError}</p>
                </div>
              )}

              {/* Candidates */}
              {candidates.length > 0 && (
                <div className="mt-4 bg-slate-800/50 rounded-xl p-4">
                  <p className="text-sm text-slate-400 mb-3 flex items-center gap-2">
                    <svg className="w-4 h-4 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                    </svg>
                    Found {candidates.length} pricing page{candidates.length > 1 ? 's' : ''}:
                  </p>
                  <div className="flex flex-wrap gap-2">
                    {candidates.map((url, i) => (
                      <button
                        key={i}
                        onClick={() => handleExtract(url)}
                        className={`px-4 py-2 rounded-lg text-sm font-medium transition-all ${
                          pricingUrl === url
                            ? 'bg-violet-500 text-white shadow-lg shadow-violet-500/30'
                            : 'bg-slate-700 text-slate-300 hover:bg-slate-600 hover:text-white'
                        }`}
                      >
                        {new URL(url).pathname || '/'}
                      </button>
                    ))}
                  </div>
                </div>
              )}

              {/* Manual Input */}
              {showManualInput && (
                <div className="mt-6 pt-6 border-t border-slate-700/50">
                  <label className="block text-sm font-medium text-slate-300 mb-3">
                    Or enter pricing page URL manually:
                  </label>
                  <div className="flex gap-3">
                    <input
                      type="url"
                      value={manualPricingUrl}
                      onChange={(e) => setManualPricingUrl(e.target.value)}
                      placeholder="https://www.yourcompany.com/pricing"
                      className="flex-1 px-4 py-3 rounded-xl bg-slate-800/50 border border-slate-700 text-white placeholder-slate-500 focus:border-violet-500 focus:ring-2 focus:ring-violet-500/20 outline-none transition-all"
                    />
                    <button
                      onClick={() => handleExtract()}
                      disabled={isExtracting || !manualPricingUrl}
                      className="px-6 py-3 bg-slate-700 text-white rounded-xl font-medium hover:bg-slate-600 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      Extract
                    </button>
                  </div>
                </div>
              )}
            </div>
          </div>

          {/* Paste Mode (Collapsible) */}
          <div className="bg-slate-900/60 backdrop-blur-sm rounded-2xl border border-slate-800 overflow-hidden">
            <button
              onClick={() => setShowPasteMode(!showPasteMode)}
              className="w-full px-6 py-4 flex items-center justify-between text-left hover:bg-slate-800/30 transition-colors"
            >
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 bg-amber-500/10 rounded-xl flex items-center justify-center">
                  <svg className="w-5 h-5 text-amber-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                  </svg>
                </div>
                <div>
                  <h3 className="text-white font-medium">Manual Paste Mode</h3>
                  <p className="text-sm text-slate-500">For sites with dynamic monthly/yearly toggles</p>
                </div>
              </div>
              <svg className={`w-5 h-5 text-slate-400 transition-transform ${showPasteMode ? 'rotate-180' : ''}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
              </svg>
            </button>

            {showPasteMode && (
              <div className="px-6 pb-6 space-y-4">
                <div className="grid md:grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-slate-300 mb-2 flex items-center gap-2">
                      <span className="w-2 h-2 bg-blue-400 rounded-full" />
                      Monthly View
                    </label>
                    <textarea
                      value={monthlyText}
                      onChange={(e) => setMonthlyText(e.target.value)}
                      placeholder="Paste pricing text when Monthly is selected..."
                      className="w-full h-40 px-4 py-3 rounded-xl bg-slate-800/50 border border-slate-700 text-white placeholder-slate-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all resize-none text-sm"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-slate-300 mb-2 flex items-center gap-2">
                      <span className="w-2 h-2 bg-emerald-400 rounded-full" />
                      Yearly View
                    </label>
                    <textarea
                      value={yearlyText}
                      onChange={(e) => setYearlyText(e.target.value)}
                      placeholder="Paste pricing text when Yearly is selected..."
                      className="w-full h-40 px-4 py-3 rounded-xl bg-slate-800/50 border border-slate-700 text-white placeholder-slate-500 focus:border-emerald-500 focus:ring-2 focus:ring-emerald-500/20 outline-none transition-all resize-none text-sm"
                    />
                  </div>
                </div>
                <button
                  onClick={handleExtractFromText}
                  disabled={isPasteExtracting || (!monthlyText.trim() && !yearlyText.trim())}
                  className="w-full px-6 py-3 bg-gradient-to-r from-amber-500 to-orange-600 text-white rounded-xl font-medium hover:from-amber-600 hover:to-orange-700 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {isPasteExtracting ? 'Extracting...' : 'Extract from Pasted Text'}
                </button>
              </div>
            )}
          </div>

          {/* Extraction Results */}
          {(isExtracting || extractedPlans.length > 0 || extractError) && (
            <div className="bg-slate-900/60 backdrop-blur-sm rounded-2xl border border-slate-800 overflow-hidden">
              <div className="p-6 border-b border-slate-800">
                <div className="flex items-center justify-between">
                  <h3 className="text-xl font-bold text-white flex items-center gap-3">
                    <div className="w-10 h-10 rounded-xl bg-emerald-500/10 flex items-center justify-center">
                      <svg className="w-5 h-5 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                      </svg>
                    </div>
                    Extracted Plans
                  </h3>
                  {detectedPeriods.length > 0 && (
                    <div className="flex gap-2">
                      {detectedPeriods.map(period => (
                        <span key={period} className={`px-3 py-1 rounded-lg text-xs font-medium ${
                          period === 'monthly' ? 'bg-blue-500/20 text-blue-300' : 'bg-emerald-500/20 text-emerald-300'
                        }`}>
                          {period}
                        </span>
                      ))}
                    </div>
                  )}
                </div>
              </div>

              {/* Warnings */}
              {warnings.length > 0 && (
                <div className="mx-6 mt-4 bg-amber-500/10 border border-amber-500/30 rounded-xl p-3">
                  <p className="text-amber-300 text-sm flex items-center gap-2">
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                    </svg>
                    {warnings.join(', ')}
                  </p>
                </div>
              )}

              {/* Error */}
              {extractError && (
                <div className="mx-6 mt-4 bg-red-500/10 border border-red-500/30 rounded-xl p-4">
                  <p className="text-red-300">{extractError}</p>
                </div>
              )}

              {/* Loading */}
              {isExtracting && (
                <div className="p-16 text-center">
                  <div className="relative w-20 h-20 mx-auto mb-6">
                    <div className="absolute inset-0 rounded-full border-4 border-violet-500/20" />
                    <div className="absolute inset-0 rounded-full border-4 border-violet-500 border-t-transparent animate-spin" />
                    <div className="absolute inset-2 rounded-full bg-gradient-to-br from-violet-500/20 to-fuchsia-500/20 flex items-center justify-center">
                      <svg className="w-8 h-8 text-violet-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
                      </svg>
                    </div>
                  </div>
                  <p className="text-lg text-white font-medium mb-2">Extracting pricing data...</p>
                  <p className="text-slate-400">AI is analyzing your pricing page</p>
                </div>
              )}

              {/* Plans Grid */}
              {!isExtracting && extractedPlans.length > 0 && (
                <div className="p-6">
                  <div className="flex items-center justify-between mb-4">
                    <div className="flex items-center gap-4">
                      <button onClick={() => setSelectedPlans(new Set(extractedPlans.map((_, i) => i)))} className="text-sm text-violet-400 hover:text-violet-300">
                        Select All
                      </button>
                      <button onClick={() => setSelectedPlans(new Set())} className="text-sm text-slate-400 hover:text-slate-300">
                        Deselect All
                      </button>
                    </div>
                    <span className="text-sm text-slate-500">{selectedPlans.size} selected</span>
                  </div>

                  <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-4">
                    {extractedPlans.map((plan, index) => (
                      <div
                        key={index}
                        onClick={() => togglePlanSelection(index)}
                        className={`relative p-5 rounded-2xl border-2 cursor-pointer transition-all hover:scale-[1.02] ${
                          selectedPlans.has(index)
                            ? 'bg-gradient-to-br from-violet-500/10 to-fuchsia-500/10 border-violet-500/50 shadow-lg shadow-violet-500/10'
                            : 'bg-slate-800/30 border-slate-700/50 hover:border-slate-600'
                        }`}
                      >
                        {/* Selection indicator */}
                        <div className={`absolute top-4 right-4 w-6 h-6 rounded-full border-2 flex items-center justify-center transition-all ${
                          selectedPlans.has(index) ? 'bg-violet-500 border-violet-500' : 'border-slate-600'
                        }`}>
                          {selectedPlans.has(index) && (
                            <svg className="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" />
                            </svg>
                          )}
                        </div>

                        <div className="mb-3">
                          <span className={`inline-block px-2 py-1 rounded-lg text-xs font-medium mb-2 ${
                            plan.billing_period === 'monthly' ? 'bg-blue-500/20 text-blue-300' :
                            plan.billing_period === 'yearly' ? 'bg-emerald-500/20 text-emerald-300' :
                            'bg-slate-500/20 text-slate-300'
                          }`}>
                            {plan.billing_period || 'unknown'}
                          </span>
                          <h4 className="text-xl font-bold text-white">{plan.name}</h4>
                        </div>

                        <div className="mb-4">
                          <div className="text-3xl font-bold bg-gradient-to-r from-violet-400 to-fuchsia-400 bg-clip-text text-transparent">
                            {plan.price_string || formatPrice(plan.price_amount, plan.currency)}
                          </div>
                          {plan.billing_period === 'yearly' && plan.monthly_equivalent_amount > 0 && (
                            <div className="text-sm text-slate-400 mt-1">
                              {formatPrice(plan.monthly_equivalent_amount, plan.currency)}/mo
                            </div>
                          )}
                        </div>

                        {plan.features?.length > 0 && (
                          <ul className="space-y-2">
                            {plan.features.slice(0, 3).map((feature, i) => (
                              <li key={i} className="text-sm text-slate-400 flex items-center gap-2">
                                <svg className="w-4 h-4 text-emerald-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                                </svg>
                                {feature}
                              </li>
                            ))}
                            {plan.features.length > 3 && (
                              <li className="text-sm text-slate-500">+{plan.features.length - 3} more</li>
                            )}
                          </ul>
                        )}
                      </div>
                    ))}
                  </div>

                  {/* Save Button */}
                  <div className="mt-8 flex items-center justify-between">
                    <p className="text-slate-400">
                      {selectedPlans.size} plan{selectedPlans.size !== 1 ? 's' : ''} selected
                    </p>
                    <button
                      onClick={handleSave}
                      disabled={selectedPlans.size === 0 || isSaving}
                      className="px-8 py-4 bg-gradient-to-r from-emerald-500 to-teal-600 text-white rounded-xl font-bold text-lg hover:from-emerald-600 hover:to-teal-700 hover:scale-[1.02] hover:shadow-xl hover:shadow-emerald-500/25 transition-all disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:scale-100"
                    >
                      {isSaving ? (
                        <span className="flex items-center gap-3">
                          <svg className="animate-spin h-5 w-5" viewBox="0 0 24 24">
                            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                          </svg>
                          Saving...
                        </span>
                      ) : (
                        <span className="flex items-center gap-2">
                          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                          </svg>
                          Save {selectedPlans.size} Plan{selectedPlans.size !== 1 ? 's' : ''}
                        </span>
                      )}
                    </button>
                  </div>
                </div>
              )}

              {/* Empty State */}
              {!isExtracting && extractedPlans.length === 0 && !extractError && (
                <div className="p-12 text-center">
                  <p className="text-slate-400">No plans found. Try a different URL.</p>
                </div>
              )}
            </div>
          )}
        </div>
      )}

      {activeTab === 'saved' && (
        <div>
          {/* Success Toast */}
          {saveSuccess && (
            <div className="mb-6 bg-emerald-500/10 border border-emerald-500/30 rounded-xl p-4 flex items-center gap-3 animate-pulse">
              <svg className="w-6 h-6 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
              </svg>
              <p className="text-emerald-300 font-medium">Plans saved successfully!</p>
            </div>
          )}

          {isLoadingSaved ? (
            <div className="bg-slate-900/60 rounded-2xl border border-slate-800 p-16 text-center">
              <div className="w-12 h-12 border-4 border-violet-500 border-t-transparent rounded-full animate-spin mx-auto mb-4" />
              <p className="text-slate-400">Loading your plans...</p>
            </div>
          ) : savedPlans.length === 0 ? (
            <div className="bg-slate-900/60 rounded-2xl border border-slate-800 p-16 text-center">
              <div className="w-24 h-24 bg-gradient-to-br from-violet-500/10 to-fuchsia-500/10 rounded-3xl flex items-center justify-center mx-auto mb-6">
                <svg className="w-12 h-12 text-violet-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 14l6-6m-5.5.5h.01m4.99 5h.01M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16l3.5-2 3.5 2 3.5-2 3.5 2zM10 8.5a.5.5 0 11-1 0 .5.5 0 011 0zm5 5a.5.5 0 11-1 0 .5.5 0 011 0z" />
                </svg>
              </div>
              <h3 className="text-2xl font-bold text-white mb-3">No plans yet</h3>
              <p className="text-slate-400 mb-6 max-w-md mx-auto">
                Import your pricing plans from your website to get started with analysis and simulations.
              </p>
              <button
                onClick={() => setActiveTab('import')}
                className="px-6 py-3 bg-gradient-to-r from-violet-500 to-fuchsia-600 text-white rounded-xl font-medium hover:from-violet-600 hover:to-fuchsia-700 transition-all"
              >
                Import Plans
              </button>
            </div>
          ) : (
            <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-5">
              {savedPlans.map((plan) => (
                <div
                  key={plan.id}
                  className="group relative bg-slate-900/60 backdrop-blur-sm rounded-2xl p-6 border border-slate-800 hover:border-violet-500/30 transition-all hover:shadow-lg hover:shadow-violet-500/5"
                >
                  <div className="absolute inset-0 bg-gradient-to-br from-violet-500/5 to-fuchsia-500/5 rounded-2xl opacity-0 group-hover:opacity-100 transition-opacity" />
                  
                  <div className="relative">
                    {/* Delete Button */}
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        handleDeletePlan(plan.id);
                      }}
                      className="absolute top-0 right-0 p-2 rounded-lg text-slate-500 hover:text-red-400 hover:bg-red-500/10 opacity-0 group-hover:opacity-100 transition-all z-10"
                      title="Delete plan"
                    >
                      <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                      </svg>
                    </button>

                    <div className="flex items-start justify-between mb-4">
                      <div>
                        <span className={`inline-block px-2 py-1 rounded-lg text-xs font-medium mb-2 ${
                          plan.billing_period === 'monthly' ? 'bg-blue-500/20 text-blue-300' :
                          plan.billing_period === 'yearly' ? 'bg-emerald-500/20 text-emerald-300' :
                          'bg-slate-500/20 text-slate-300'
                        }`}>
                          {plan.billing_period || 'unknown'}
                        </span>
                        <h4 className="text-xl font-bold text-white">{plan.plan_name}</h4>
                      </div>
                    </div>

                    <div className="mb-4">
                      <div className="text-3xl font-bold bg-gradient-to-r from-violet-400 to-fuchsia-400 bg-clip-text text-transparent">
                        {plan.price_string || formatPrice(plan.price_amount, plan.currency)}
                      </div>
                      {plan.billing_period === 'yearly' && plan.monthly_equivalent_amount > 0 && (
                        <div className="text-sm text-slate-400 mt-1">
                          {formatPrice(plan.monthly_equivalent_amount, plan.currency)}/mo equivalent
                        </div>
                      )}
                    </div>

                    {plan.source_url && plan.source_url !== 'pasted-text' && (
                      <p className="text-xs text-slate-500 mb-4 truncate">
                        Source: {new URL(plan.source_url).hostname}
                      </p>
                    )}

                    {plan.features?.length > 0 && (
                      <ul className="space-y-2 pt-4 border-t border-slate-700/50">
                        {plan.features.slice(0, 3).map((feature, i) => (
                          <li key={i} className="text-sm text-slate-400 flex items-center gap-2">
                            <svg className="w-4 h-4 text-emerald-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                            </svg>
                            {feature}
                          </li>
                        ))}
                        {plan.features.length > 3 && (
                          <li className="text-sm text-slate-500">+{plan.features.length - 3} more features</li>
                        )}
                      </ul>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default PlansV2;
