import { useState, useEffect } from 'react';
import { usePlans } from '../../context/PlansContext';
import { businessMetricsApi, simulationApi, downloadBlob } from '../../lib/apiClient';

// Risk level badge component
const RiskBadge = ({ level }) => {
  const styles = {
    low: 'bg-emerald-500/20 text-emerald-400 border-emerald-500/30',
    medium: 'bg-amber-500/20 text-amber-400 border-amber-500/30',
    high: 'bg-red-500/20 text-red-400 border-red-500/30',
  };

  return (
    <span className={`px-2 py-1 text-xs font-medium rounded-full border ${styles[level] || styles.medium}`}>
      {level?.charAt(0).toUpperCase() + level?.slice(1)} Risk
    </span>
  );
};

// Scenario card component
const ScenarioCard = ({ scenario, currency, isPriceIncrease }) => {
  const formatCurrency = (value) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: currency || 'USD',
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(value);
  };

  const scenarioColors = {
    Conservative: 'border-emerald-500/30 bg-emerald-500/5',
    Base: 'border-blue-500/30 bg-blue-500/5',
    Aggressive: 'border-orange-500/30 bg-orange-500/5',
  };

  return (
    <div className={`rounded-xl p-5 border ${scenarioColors[scenario.name] || 'border-slate-700'}`}>
      <div className="flex items-center justify-between mb-4">
        <h4 className="text-lg font-semibold text-white">{scenario.name}</h4>
        <RiskBadge level={scenario.risk_level} />
      </div>

      <div className="space-y-3">
        {/* Customer Impact */}
        <div className="p-3 rounded-lg bg-slate-800/50">
          <p className="text-xs text-slate-500 mb-1">
            {isPriceIncrease ? 'Expected Customer Loss' : 'Expected Customer Gain'}
          </p>
          <p className="text-white font-semibold">
            {isPriceIncrease
              ? `${scenario.customer_loss_min_pct}% – ${scenario.customer_loss_max_pct}%`
              : `${scenario.customer_gain_min_pct}% – ${scenario.customer_gain_max_pct}%`}
          </p>
        </div>

        {/* Customer Count */}
        <div className="p-3 rounded-lg bg-slate-800/50">
          <p className="text-xs text-slate-500 mb-1">Projected Customers</p>
          <p className="text-white font-semibold">
            {scenario.new_customer_count_min.toLocaleString()} – {scenario.new_customer_count_max.toLocaleString()}
          </p>
        </div>

        {/* MRR */}
        <div className="p-3 rounded-lg bg-slate-800/50">
          <p className="text-xs text-slate-500 mb-1">Projected MRR</p>
          <p className="text-white font-semibold">
            {formatCurrency(scenario.new_mrr_min)} – {formatCurrency(scenario.new_mrr_max)}
          </p>
        </div>

        {/* ARR */}
        <div className="p-3 rounded-lg bg-slate-800/50">
          <p className="text-xs text-slate-500 mb-1">Projected ARR (12 mo)</p>
          <p className="text-white font-semibold">
            {formatCurrency(scenario.new_arr_min)} – {formatCurrency(scenario.new_arr_max)}
          </p>
        </div>

        {/* Churn */}
        <div className="p-3 rounded-lg bg-slate-800/50">
          <p className="text-xs text-slate-500 mb-1">Estimated Churn</p>
          <p className="text-white font-semibold">
            {scenario.estimated_churn_min_pct}% – {scenario.estimated_churn_max_pct}%
          </p>
        </div>
      </div>
    </div>
  );
};

// Simulation result component
const SimulationResult = ({ result, onDownloadPdf, isPdfLoading }) => {
  if (!result) return null;

  const formatCurrency = (value, currency) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: currency || 'USD',
      minimumFractionDigits: 0,
      maximumFractionDigits: 2,
    }).format(value);
  };

  const isPriceIncrease = result.price_change_pct >= 0;
  const changeSign = isPriceIncrease ? '+' : '';

  // Derive overall risk (highest among scenarios)
  const riskLevels = { low: 1, medium: 2, high: 3 };
  const overallRisk = result.scenarios?.reduce((highest, sc) => {
    const level = sc.risk_level || 'medium';
    return riskLevels[level] > riskLevels[highest] ? level : highest;
  }, 'low');

  return (
    <div className="space-y-6">
      {/* Header Summary Card */}
      <div className="bg-slate-900/50 backdrop-blur-sm rounded-2xl p-6 border border-slate-800">
        <div className="flex items-start justify-between mb-4">
          <div>
            <h3 className="text-xl font-bold text-white mb-2">{result.plan_name}</h3>
            <div className="flex items-center gap-3 text-lg">
              <span className="text-slate-400">
                {formatCurrency(result.current_price, result.currency)}
              </span>
              <svg className="w-5 h-5 text-slate-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 8l4 4m0 0l-4 4m4-4H3" />
              </svg>
              <span className="text-white font-semibold">
                {formatCurrency(result.new_price, result.currency)}
              </span>
              <span className="text-slate-500">/ month</span>
            </div>
          </div>
          <div className="text-right flex flex-col items-end gap-2">
            <div className={`text-2xl font-bold ${isPriceIncrease ? 'text-emerald-400' : 'text-blue-400'}`}>
              {changeSign}{result.price_change_pct?.toFixed(1)}%
            </div>
            <RiskBadge level={overallRisk} />
          </div>
        </div>

        <div className="grid grid-cols-3 gap-4 pt-4 border-t border-slate-700/50">
          <div>
            <p className="text-xs text-slate-500">Current Customers</p>
            <p className="text-white font-semibold">{result.active_customers_on_plan?.toLocaleString()}</p>
          </div>
          <div>
            <p className="text-xs text-slate-500">Global MRR</p>
            <p className="text-white font-semibold">{formatCurrency(result.global_mrr, result.currency)}</p>
          </div>
          <div>
            <p className="text-xs text-slate-500">Current Churn</p>
            <p className="text-white font-semibold">{result.global_churn_rate}%</p>
          </div>
        </div>

        {/* Download PDF Button */}
        <div className="mt-4 pt-4 border-t border-slate-700/50">
          <button
            onClick={onDownloadPdf}
            disabled={isPdfLoading}
            className="flex items-center gap-2 px-4 py-2 bg-slate-800 hover:bg-slate-700 text-white text-sm font-medium rounded-lg transition-all disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isPdfLoading ? (
              <>
                <svg className="animate-spin h-4 w-4" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                </svg>
                <span>Generating PDF...</span>
              </>
            ) : (
              <>
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
                <span>Download Simulation PDF</span>
              </>
            )}
          </button>
        </div>
      </div>

      {/* Scenario Cards */}
      <div className="grid lg:grid-cols-3 gap-4">
        {result.scenarios?.map((scenario) => (
          <ScenarioCard
            key={scenario.name}
            scenario={scenario}
            currency={result.currency}
            isPriceIncrease={isPriceIncrease}
          />
        ))}
      </div>

      {/* AI Narrative */}
      {result.ai_narrative && (
        <div className="bg-gradient-to-br from-purple-500/10 to-blue-500/10 rounded-2xl p-6 border border-purple-500/20">
          <div className="flex items-center gap-2 mb-4">
            <svg className="w-5 h-5 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
            </svg>
            <h4 className="text-lg font-semibold text-white">AI Pricing Insights</h4>
          </div>
          <div className="text-slate-300 leading-relaxed whitespace-pre-line">
            {result.ai_narrative}
          </div>
        </div>
      )}
    </div>
  );
};

// History item component
const HistoryItem = ({ simulation, onClick, isActive }) => {
  const formatDate = (dateStr) => {
    return new Date(dateStr).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const changeSign = simulation.price_change_pct >= 0 ? '+' : '';

  return (
    <button
      onClick={onClick}
      className={`w-full text-left p-3 rounded-lg transition-all ${
        isActive
          ? 'bg-blue-500/20 border border-blue-500/30'
          : 'bg-slate-800/30 hover:bg-slate-800/50 border border-transparent'
      }`}
    >
      <div className="flex items-center justify-between mb-1">
        <span className="text-white font-medium text-sm">{simulation.plan_name}</span>
        <span className={`text-xs font-medium ${simulation.price_change_pct >= 0 ? 'text-emerald-400' : 'text-blue-400'}`}>
          {changeSign}{simulation.price_change_pct?.toFixed(1)}%
        </span>
      </div>
      <p className="text-xs text-slate-500">{formatDate(simulation.created_at)}</p>
    </button>
  );
};

// Main component
const PricingSimulation = () => {
  const { plans, isLoading: plansLoading } = usePlans();

  // Form state
  const [selectedPlanId, setSelectedPlanId] = useState('');
  const [newPrice, setNewPrice] = useState('');
  const [activeCustomers, setActiveCustomers] = useState('');
  const [formError, setFormError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Business metrics
  const [metrics, setMetrics] = useState(null);
  const [metricsLoading, setMetricsLoading] = useState(true);

  // Results
  const [currentResult, setCurrentResult] = useState(null);
  const [history, setHistory] = useState([]);
  const [historyLoading, setHistoryLoading] = useState(true);

  // PDF download state
  const [isPdfLoading, setIsPdfLoading] = useState(false);
  const [pdfError, setPdfError] = useState('');

  // Get selected plan details
  const selectedPlan = plans?.find((p) => p.id === selectedPlanId);

  // Fetch business metrics
  useEffect(() => {
    const fetchMetrics = async () => {
      try {
        const { data } = await businessMetricsApi.get();
        setMetrics(data);
      } catch (err) {
        console.error('Failed to fetch metrics:', err);
      } finally {
        setMetricsLoading(false);
      }
    };
    fetchMetrics();
  }, []);

  // Fetch simulation history
  useEffect(() => {
    const fetchHistory = async () => {
      try {
        const { data } = await simulationApi.list(5);
        setHistory(data || []);
      } catch (err) {
        console.error('Failed to fetch history:', err);
      } finally {
        setHistoryLoading(false);
      }
    };
    fetchHistory();
  }, []);

  // Handle plan selection change
  const handlePlanChange = (e) => {
    const planId = e.target.value;
    setSelectedPlanId(planId);
    setFormError('');
  };

  // Validate form
  const validate = () => {
    if (!selectedPlanId) {
      setFormError('Please select a plan');
      return false;
    }
    if (!newPrice || parseFloat(newPrice) < 0) {
      setFormError('Please enter a valid new price');
      return false;
    }
    if (!activeCustomers || parseInt(activeCustomers) < 0) {
      setFormError('Please enter a valid number of customers');
      return false;
    }
    return true;
  };

  // Handle form submit
  const handleSubmit = async (e) => {
    e.preventDefault();
    setFormError('');

    if (!validate()) return;

    setIsSubmitting(true);

    try {
      const { data } = await simulationApi.run({
        planId: selectedPlanId,
        currentPrice: selectedPlan.price,
        newPrice: parseFloat(newPrice),
        currency: selectedPlan.currency || 'USD',
        activeCustomersOnPlan: parseInt(activeCustomers),
        globalMrr: metrics?.mrr || 0,
        globalChurnRate: metrics?.monthly_churn_rate || 5,
        pricingGoal: metrics?.pricing_goal || 'revenue',
      });

      setCurrentResult(data);
      // Add to history
      setHistory((prev) => [data, ...prev.slice(0, 4)]);
    } catch (err) {
      setFormError(err.message || 'Failed to run simulation');
    } finally {
      setIsSubmitting(false);
    }
  };

  // Load history item
  const loadHistoryItem = (simulation) => {
    setCurrentResult(simulation);
  };

  // Handle PDF download
  const handleDownloadPdf = async () => {
    if (!currentResult?.id) return;

    setIsPdfLoading(true);
    setPdfError('');

    try {
      const { ok, blob } = await simulationApi.exportPdf(currentResult.id);
      if (ok && blob) {
        // Generate filename
        const planName = currentResult.plan_name?.replace(/\s+/g, '-').toLowerCase() || 'simulation';
        const date = new Date(currentResult.created_at).toISOString().split('T')[0].replace(/-/g, '');
        const filename = `pricing-simulation-${planName}-${date}.pdf`;
        
        downloadBlob(blob, filename);
      }
    } catch (err) {
      console.error('Failed to download PDF:', err);
      setPdfError(err.message || 'Failed to generate PDF');
    } finally {
      setIsPdfLoading(false);
    }
  };

  const formatPrice = (price, currency) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: currency || 'USD',
      minimumFractionDigits: 0,
      maximumFractionDigits: 2,
    }).format(price);
  };

  return (
    <div className="max-w-7xl mx-auto space-y-8">
      {/* Page Header */}
      <div>
        <h1 className="text-3xl font-bold text-white mb-2">Pricing Simulation</h1>
        <p className="text-slate-400">
          Test new price points before rolling them out. See projected impact on customers, MRR, and churn.
        </p>
      </div>

      <div className="grid lg:grid-cols-5 gap-8">
        {/* Left Column - Form */}
        <div className="lg:col-span-2 space-y-6">
          {/* Simulation Form */}
          <div className="bg-slate-900/50 backdrop-blur-sm rounded-2xl p-6 border border-slate-800">
            <h3 className="text-lg font-semibold text-white mb-4">New Simulation</h3>

            <form onSubmit={handleSubmit} className="space-y-5">
              {/* Plan Selector */}
              <div>
                <label className="block text-sm font-semibold text-slate-300 mb-2">
                  Select Plan
                </label>
                <select
                  value={selectedPlanId}
                  onChange={handlePlanChange}
                  disabled={plansLoading || isSubmitting}
                  className="w-full px-4 py-3 rounded-xl bg-slate-900/50 border border-slate-700 text-white focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all disabled:opacity-50"
                >
                  <option value="">Choose a plan...</option>
                  {plans?.map((plan) => (
                    <option key={plan.id} value={plan.id}>
                      {plan.name} - {formatPrice(plan.price, plan.currency)}
                    </option>
                  ))}
                </select>

                {selectedPlan && (
                  <div className="mt-2 p-3 rounded-lg bg-slate-800/50 text-sm">
                    <div className="flex items-center justify-between">
                      <span className="text-slate-400">Current Price</span>
                      <span className="text-white font-medium">
                        {formatPrice(selectedPlan.price, selectedPlan.currency)} / {selectedPlan.interval || 'month'}
                      </span>
                    </div>
                  </div>
                )}
              </div>

              {/* New Price */}
              <div>
                <label className="block text-sm font-semibold text-slate-300 mb-2">
                  New Price {selectedPlan && <span className="text-slate-500 font-normal">({selectedPlan.currency || 'USD'})</span>}
                </label>
                <input
                  type="number"
                  step="0.01"
                  min="0"
                  value={newPrice}
                  onChange={(e) => {
                    setNewPrice(e.target.value);
                    setFormError('');
                  }}
                  disabled={isSubmitting}
                  className="w-full px-4 py-3 rounded-xl bg-slate-900/50 border border-slate-700 text-white placeholder-slate-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all disabled:opacity-50"
                  placeholder="Enter new price"
                />
              </div>

              {/* Active Customers */}
              <div>
                <label className="block text-sm font-semibold text-slate-300 mb-2">
                  Active Customers on This Plan
                </label>
                <input
                  type="number"
                  min="0"
                  value={activeCustomers}
                  onChange={(e) => {
                    setActiveCustomers(e.target.value);
                    setFormError('');
                  }}
                  disabled={isSubmitting}
                  className="w-full px-4 py-3 rounded-xl bg-slate-900/50 border border-slate-700 text-white placeholder-slate-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all disabled:opacity-50"
                  placeholder="e.g., 500"
                />
              </div>

              {/* Business Metrics Summary */}
              {metricsLoading ? (
                <div className="p-4 rounded-xl bg-slate-800/30 border border-slate-700/50">
                  <div className="flex items-center gap-2 text-slate-500">
                    <div className="w-4 h-4 border-2 border-slate-500 border-t-transparent rounded-full animate-spin"></div>
                    Loading metrics...
                  </div>
                </div>
              ) : metrics ? (
                <div className="p-4 rounded-xl bg-slate-800/30 border border-slate-700/50">
                  <p className="text-xs text-slate-500 mb-3">Business Metrics (from settings)</p>
                  <div className="grid grid-cols-3 gap-3">
                    <div>
                      <p className="text-xs text-slate-500">MRR</p>
                      <p className="text-white font-medium">
                        {formatPrice(metrics.mrr || 0, metrics.currency)}
                      </p>
                    </div>
                    <div>
                      <p className="text-xs text-slate-500">Churn</p>
                      <p className="text-white font-medium">{metrics.monthly_churn_rate || 0}%</p>
                    </div>
                    <div>
                      <p className="text-xs text-slate-500">Goal</p>
                      <p className="text-white font-medium capitalize">{metrics.pricing_goal || 'Revenue'}</p>
                    </div>
                  </div>
                </div>
              ) : (
                <div className="p-4 rounded-xl bg-amber-500/10 border border-amber-500/30">
                  <p className="text-amber-400 text-sm">
                    No business metrics set. Go to Settings to configure your metrics for better simulations.
                  </p>
                </div>
              )}

              {/* Error */}
              {formError && (
                <div className="p-3 rounded-lg bg-red-500/10 border border-red-500/30">
                  <p className="text-red-400 text-sm">{formError}</p>
                </div>
              )}

              {/* Submit Button */}
              <button
                type="submit"
                disabled={isSubmitting || !selectedPlanId}
                className="w-full px-6 py-3 bg-gradient-to-r from-blue-500 to-indigo-600 text-white rounded-xl font-semibold hover:from-blue-600 hover:to-indigo-700 hover:scale-[1.02] transition-all shadow-lg shadow-blue-500/20 disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:scale-100"
              >
                {isSubmitting ? (
                  <span className="flex items-center justify-center gap-2">
                    <svg className="animate-spin h-5 w-5" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                    </svg>
                    Running Simulation...
                  </span>
                ) : (
                  'Run Simulation'
                )}
              </button>
            </form>
          </div>

          {/* History */}
          {!historyLoading && history.length > 0 && (
            <div className="bg-slate-900/50 backdrop-blur-sm rounded-2xl p-6 border border-slate-800">
              <h3 className="text-sm font-semibold text-slate-400 mb-3">Recent Simulations</h3>
              <div className="space-y-2">
                {history.map((sim) => (
                  <HistoryItem
                    key={sim.id}
                    simulation={sim}
                    onClick={() => loadHistoryItem(sim)}
                    isActive={currentResult?.id === sim.id}
                  />
                ))}
              </div>
            </div>
          )}
        </div>

        {/* Right Column - Results */}
        <div className="lg:col-span-3">
          {pdfError && (
            <div className="mb-4 p-3 rounded-lg bg-red-500/10 border border-red-500/30">
              <p className="text-red-400 text-sm">{pdfError}</p>
            </div>
          )}
          {currentResult ? (
            <SimulationResult 
              result={currentResult} 
              onDownloadPdf={handleDownloadPdf}
              isPdfLoading={isPdfLoading}
            />
          ) : (
            <div className="bg-slate-900/50 backdrop-blur-sm rounded-2xl p-12 border border-slate-800 text-center">
              <div className="w-16 h-16 bg-slate-800 rounded-2xl flex items-center justify-center mx-auto mb-4">
                <svg className="w-8 h-8 text-slate-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                </svg>
              </div>
              <h3 className="text-xl font-semibold text-white mb-2">No Simulations Yet</h3>
              <p className="text-slate-400 max-w-sm mx-auto">
                Select a plan and enter a new price point to see how it might impact your customers, revenue, and churn.
              </p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default PricingSimulation;

