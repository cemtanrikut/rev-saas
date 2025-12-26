import { useState } from 'react';
import { pricingV2Api } from '../../lib/apiClient';

const Step1Plans = ({ data, onChange }) => {
  const [websiteUrl, setWebsiteUrl] = useState('');
  const [pricingUrl, setPricingUrl] = useState('');
  const [manualPricingUrl, setManualPricingUrl] = useState('');
  const [isDiscovering, setIsDiscovering] = useState(false);
  const [isExtracting, setIsExtracting] = useState(false);
  const [candidates, setCandidates] = useState([]);
  const [extractedPlans, setExtractedPlans] = useState([]);
  const [error, setError] = useState('');
  const [showManualInput, setShowManualInput] = useState(false);

  const handleDiscover = async () => {
    if (!websiteUrl.trim()) {
      setError('Please enter your website URL');
      return;
    }

    setIsDiscovering(true);
    setError('');
    setCandidates([]);
    setExtractedPlans([]);

    try {
      const response = await pricingV2Api.discover({ website_url: websiteUrl });
      const result = response.data || response;
      
      if (result.error) {
        setError(result.error);
        setShowManualInput(true);
        return;
      }

      setCandidates(result.pricing_candidates || []);
      
      if (result.selected_pricing_url) {
        setPricingUrl(result.selected_pricing_url);
        await handleExtract(result.selected_pricing_url);
      } else {
        setShowManualInput(true);
        setError('Could not auto-detect pricing page. Please enter the URL manually.');
      }
    } catch (err) {
      setError(err.message || 'Discovery failed');
      setShowManualInput(true);
    } finally {
      setIsDiscovering(false);
    }
  };

  const handleExtract = async (url) => {
    const targetUrl = url || manualPricingUrl || pricingUrl;
    
    if (!targetUrl) {
      setError('Please enter a pricing page URL');
      return;
    }

    setIsExtracting(true);
    setError('');

    try {
      const response = await pricingV2Api.extract({ pricing_url: targetUrl });
      const result = response.data || response;
      
      if (result.error) {
        setError(result.error);
        return;
      }

      setPricingUrl(targetUrl);
      setExtractedPlans(result.plans || []);
      
      // Auto-add all extracted plans to onboarding data
      const plansForOnboarding = (result.plans || []).map((plan, idx) => ({
        id: Date.now() + idx,
        name: plan.name,
        price: plan.price_amount || 0,
        priceString: plan.price_string,
        currency: plan.currency || 'USD',
        billingCycle: plan.billing_period || 'monthly',
        features: plan.features || [],
        includedUnits: plan.included_units || [],
      }));
      
      onChange(plansForOnboarding);
    } catch (err) {
      setError(err.message || 'Extraction failed');
    } finally {
      setIsExtracting(false);
    }
  };

  const handleRemovePlan = (id) => {
    onChange(data.filter((p) => p.id !== id));
  };

  const formatPrice = (plan) => {
    if (plan.priceString) return plan.priceString;
    const symbols = { USD: '$', EUR: '€', GBP: '£', TRY: '₺' };
    return `${symbols[plan.currency] || plan.currency}${plan.price}`;
  };

  return (
    <div>
      <div className="mb-8 text-center">
        <h2 className="text-3xl font-bold text-white mb-2">Your Pricing Plans</h2>
        <p className="text-slate-400">
          Auto-import your pricing plans from your website.
        </p>
      </div>

      {/* Auto-detect Section */}
      <div className="bg-slate-900/30 rounded-xl p-6 border border-slate-700/50 mb-6">
        <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
          <svg className="w-5 h-5 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          Auto-detect Pricing
        </h3>

        <div className="space-y-4">
          <div>
            <label htmlFor="websiteUrl" className="block text-sm font-semibold text-slate-300 mb-2">
              Your Website URL
            </label>
            <div className="flex gap-3">
              <input
                id="websiteUrl"
                type="url"
                value={websiteUrl}
                onChange={(e) => setWebsiteUrl(e.target.value)}
                className="flex-1 px-4 py-3 rounded-xl bg-slate-900/50 border border-slate-700 text-white placeholder-slate-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all"
                placeholder="https://www.yourcompany.com"
              />
              <button
                onClick={handleDiscover}
                disabled={isDiscovering || isExtracting}
                className="px-6 py-3 bg-gradient-to-r from-blue-500 to-indigo-600 text-white rounded-xl font-semibold hover:from-blue-600 hover:to-indigo-700 transition-all disabled:opacity-50 disabled:cursor-not-allowed whitespace-nowrap"
              >
                {isDiscovering ? (
                  <span className="flex items-center gap-2">
                    <svg className="animate-spin h-5 w-5" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                    </svg>
                    Detecting...
                  </span>
                ) : isExtracting ? (
                  <span className="flex items-center gap-2">
                    <svg className="animate-spin h-5 w-5" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                    </svg>
                    Extracting...
                  </span>
                ) : (
                  'Auto-detect & Import'
                )}
              </button>
            </div>
          </div>

          {/* Error Display */}
          {error && (
            <div className="bg-red-500/10 border border-red-500/30 rounded-xl p-4">
              <p className="text-red-400 text-sm">{error}</p>
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
                  placeholder="https://www.yourcompany.com/pricing"
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

      {/* Extracted/Imported Plans List */}
      {data.length > 0 && (
        <div className="space-y-3">
          <h3 className="text-lg font-semibold text-white flex items-center gap-2">
            <svg className="w-5 h-5 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
            </svg>
            Imported Plans ({data.length})
          </h3>
          
          {data.map((plan) => (
            <div
              key={plan.id}
              className="flex items-center justify-between p-4 rounded-xl bg-slate-900/50 border border-slate-700"
            >
              <div className="flex items-center gap-4">
                <div className="w-10 h-10 rounded-lg bg-blue-500/20 flex items-center justify-center">
                  <svg
                    className="w-5 h-5 text-blue-400"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                    />
                  </svg>
                </div>
                <div>
                  <p className="text-white font-semibold">{plan.name}</p>
                  <p className="text-slate-400 text-sm">
                    {formatPrice(plan)} / {plan.billingCycle}
                  </p>
                  {plan.features?.length > 0 && (
                    <p className="text-slate-500 text-xs mt-1">
                      {plan.features.slice(0, 2).join(', ')}
                      {plan.features.length > 2 && ` +${plan.features.length - 2} more`}
                    </p>
                  )}
                </div>
              </div>
              <button
                onClick={() => handleRemovePlan(plan.id)}
                className="p-2 text-slate-400 hover:text-red-400 hover:bg-red-500/10 rounded-lg transition-all"
              >
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                  />
                </svg>
              </button>
            </div>
          ))}
        </div>
      )}

      {/* Requirement Info */}
      {data.length === 0 && (
        <div className="mt-4 bg-amber-500/10 border border-amber-500/30 rounded-xl p-4">
          <div className="flex items-start gap-3">
            <svg
              className="w-5 h-5 text-amber-400 flex-shrink-0 mt-0.5"
              fill="currentColor"
              viewBox="0 0 20 20"
            >
              <path
                fillRule="evenodd"
                d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z"
                clipRule="evenodd"
              />
            </svg>
            <p className="text-sm text-amber-300">
              Enter your website URL above to auto-import your pricing plans, or at least one plan is required to continue.
            </p>
          </div>
        </div>
      )}
    </div>
  );
};

export default Step1Plans;
