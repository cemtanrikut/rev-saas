import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useCompetitors } from '../../context/CompetitorsContext';
import { usePlans } from '../../context/PlansContext';
import { useAnalysis } from '../../context/AnalysisContext';
import { useBusinessMetrics } from '../../context/BusinessMetricsContext';

const Overview = () => {
  const navigate = useNavigate();
  const { competitors } = useCompetitors();
  const { plans } = usePlans();
  const { analyses, runAnalysis, isRunning, error: analysisError, limitError, clearLimitError } = useAnalysis();
  const { metrics, hasMetrics, isLoading: metricsLoading } = useBusinessMetrics();
  const [runError, setRunError] = useState(null);

  // Get user-friendly error title
  const getLimitErrorTitle = (errorCode) => {
    const titles = {
      'LIMIT_ANALYSES': 'Analysis Limit Reached',
      'LIMIT_COMPETITORS': 'Competitor Limit Reached',
      'LIMIT_PLANS': 'Plan Limit Reached',
      'LIMIT_TRIAL_EXPIRED': 'Trial Expired',
    };
    return titles[errorCode] || 'Limit Reached';
  };

  // Get upgrade message
  const getLimitErrorMessage = (limitError) => {
    if (limitError.errorCode === 'LIMIT_TRIAL_EXPIRED') {
      return 'Your free trial has expired. Upgrade to continue using Revalyze.';
    }
    if (limitError.errorCode === 'LIMIT_ANALYSES') {
      if (limitError.plan === 'free') {
        return `You've used all ${limitError.limit} free analyses. Upgrade to get more analyses each month.`;
      }
      return `You've reached your monthly limit of ${limitError.limit} analyses. Upgrade for more.`;
    }
    return limitError.reason || 'You have reached a limit on your current plan.';
  };

  // Currency formatter
  const formatCurrency = (value, currency = 'USD') => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency,
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(value);
  };

  const checklistItems = [
    { label: 'Created account', completed: true },
    { label: 'Completed onboarding', completed: true },
    { label: 'Defined current plans', completed: plans.length > 0 },
    { label: 'Added competitors', completed: competitors.length > 0 },
    { label: 'Ran first pricing analysis', completed: analyses.length > 0 }
  ];

  const hasPlans = plans.length > 0;
  const hasCompetitors = competitors.length > 0;
  const hasAnalysis = analyses.length > 0;

  const handleRunAnalysis = async () => {
    if (!hasPlans || !hasCompetitors) return;
    
    setRunError(null);
    const result = await runAnalysis();
    
    if (result.success) {
      navigate('/app/analyses');
    } else {
      setRunError(result.error || 'Failed to run analysis');
    }
  };

  return (
    <div className="max-w-7xl mx-auto space-y-8">
      {/* Limit Error Modal */}
      {limitError && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
          <div className="bg-slate-900 rounded-2xl border border-slate-700 shadow-2xl max-w-md w-full mx-4 overflow-hidden">
            {/* Modal Header */}
            <div className="p-6 pb-4 bg-gradient-to-br from-amber-500/10 to-orange-500/10 border-b border-slate-800">
              <div className="flex items-center gap-4">
                <div className="w-12 h-12 bg-amber-500/20 rounded-xl flex items-center justify-center">
                  <svg className="w-6 h-6 text-amber-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                  </svg>
                </div>
                <div>
                  <h3 className="text-lg font-bold text-white">
                    {getLimitErrorTitle(limitError.errorCode)}
                  </h3>
                  <p className="text-sm text-slate-400">
                    Current plan: <span className="text-amber-400 font-medium capitalize">{limitError.plan || 'Free'}</span>
                  </p>
                </div>
              </div>
            </div>

            {/* Modal Body */}
            <div className="p-6 space-y-4">
              <p className="text-slate-300 leading-relaxed">
                {getLimitErrorMessage(limitError)}
              </p>

              {/* Usage Stats */}
              {limitError.limit && limitError.current !== undefined && (
                <div className="bg-slate-800/50 rounded-xl p-4 border border-slate-700/50">
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-sm text-slate-400">Usage</span>
                    <span className="text-sm font-medium text-white">
                      {limitError.current} / {limitError.limit}
                    </span>
                  </div>
                  <div className="w-full h-2 bg-slate-700 rounded-full overflow-hidden">
                    <div 
                      className="h-full bg-gradient-to-r from-amber-500 to-red-500 rounded-full transition-all"
                      style={{ width: `${Math.min((limitError.current / limitError.limit) * 100, 100)}%` }}
                    />
                  </div>
                </div>
              )}

              {/* Benefits */}
              <div className="space-y-2">
                <p className="text-xs font-semibold text-slate-400 uppercase tracking-wider">
                  Upgrade to unlock:
                </p>
                <ul className="space-y-2">
                  <li className="flex items-center gap-2 text-sm text-slate-300">
                    <svg className="w-4 h-4 text-emerald-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                    </svg>
                    More analyses per month
                  </li>
                  <li className="flex items-center gap-2 text-sm text-slate-300">
                    <svg className="w-4 h-4 text-emerald-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                    </svg>
                    AI-powered pricing insights
                  </li>
                  <li className="flex items-center gap-2 text-sm text-slate-300">
                    <svg className="w-4 h-4 text-emerald-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                    </svg>
                    Track more competitors
                  </li>
                </ul>
              </div>
            </div>

            {/* Modal Footer */}
            <div className="p-6 pt-4 bg-slate-800/30 border-t border-slate-800 flex items-center gap-3">
              <button
                onClick={clearLimitError}
                className="flex-1 px-4 py-2.5 bg-slate-800 text-slate-300 rounded-xl font-medium hover:bg-slate-700 transition-all"
              >
                Maybe Later
              </button>
              <button
                onClick={() => {
                  clearLimitError();
                  navigate('/app/billing');
                }}
                className="flex-1 px-4 py-2.5 bg-gradient-to-r from-blue-500 to-indigo-600 text-white rounded-xl font-semibold hover:from-blue-600 hover:to-indigo-700 transition-all shadow-lg shadow-blue-500/20"
              >
                Upgrade Now
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Welcome Section */}
      <div className="grid md:grid-cols-3 gap-6">
        {/* Left: Welcome + Checklist */}
        <div className="md:col-span-2 space-y-6">
          <div className="bg-slate-900/50 backdrop-blur-sm rounded-2xl p-8 border border-slate-800">
            <h1 className="text-3xl font-bold text-white mb-2">
              Welcome to Revalyze
            </h1>
            <p className="text-slate-400 text-lg mb-8">
              Your AI Pricing Brain for SaaS
            </p>

            {/* Checklist */}
            <div className="space-y-3">
              <h3 className="text-sm font-semibold text-slate-300 mb-4">
                Getting Started
              </h3>
              {checklistItems.map((item, index) => (
                <div
                  key={index}
                  className="flex items-center gap-3 p-3 rounded-xl bg-slate-800/50 border border-slate-700"
                >
                  <div
                    className={`w-5 h-5 rounded-full flex items-center justify-center flex-shrink-0 transition-all ${
                      item.completed
                        ? 'bg-blue-500 shadow-lg shadow-blue-500/50'
                        : 'bg-slate-700 border border-slate-600'
                    }`}
                  >
                    {item.completed && (
                      <svg className="w-3 h-3 text-white" fill="currentColor" viewBox="0 0 20 20">
                        <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                      </svg>
                    )}
                  </div>
                  <span
                    className={`${item.completed ? 'text-slate-300' : 'text-slate-400'} transition-colors`}
                  >
                    {item.label}
                  </span>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Right: Next Action */}
        <div className="bg-gradient-to-br from-blue-500/10 to-indigo-500/10 backdrop-blur-sm rounded-2xl p-6 border border-blue-500/20">
          <div className="flex items-start gap-3 mb-6">
            <div className="w-10 h-10 bg-blue-500 rounded-lg flex items-center justify-center flex-shrink-0 shadow-lg shadow-blue-500/50">
              <svg className="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
            </div>
            <div>
              <h3 className="text-lg font-semibold text-white mb-1">
                Next best action
              </h3>
              <p className="text-sm text-slate-400">
                {!hasPlans 
                  ? "Define your current pricing plans to start the analysis."
                  : !hasCompetitors
                  ? "Add competitors so Revalyze can benchmark your pricing."
                  : !hasAnalysis
                  ? "You're all set! Run your first pricing analysis."
                  : "Review your latest pricing analysis or run a new one."
                }
              </p>
            </div>
          </div>

          {!hasPlans ? (
            <button
              onClick={() => navigate('/app/plans')}
              className="block w-full px-4 py-3 bg-gradient-to-r from-blue-500 to-indigo-600 text-white rounded-xl font-semibold hover:from-blue-600 hover:to-indigo-700 transition-all hover:scale-105 text-center shadow-lg shadow-blue-500/20"
            >
              Go to My Pricing
            </button>
          ) : !hasCompetitors ? (
            <button
              onClick={() => navigate('/app/competitors')}
              className="block w-full px-4 py-3 bg-gradient-to-r from-blue-500 to-indigo-600 text-white rounded-xl font-semibold hover:from-blue-600 hover:to-indigo-700 transition-all hover:scale-105 text-center shadow-lg shadow-blue-500/20"
            >
              Go to Competitors
            </button>
          ) : !hasAnalysis ? (
            <button
              onClick={handleRunAnalysis}
              disabled={isRunning}
              className="block w-full px-4 py-3 bg-gradient-to-r from-blue-500 to-indigo-600 text-white rounded-xl font-semibold hover:from-blue-600 hover:to-indigo-700 transition-all hover:scale-105 text-center shadow-lg shadow-blue-500/20 disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:scale-100"
            >
              {isRunning ? (
                <span className="flex items-center justify-center gap-2">
                  <svg className="animate-spin h-5 w-5" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                  </svg>
                  Running Analysis...
                </span>
              ) : (
                'Run Analysis'
              )}
            </button>
          ) : (
            <button
              onClick={() => navigate('/app/analyses')}
              className="block w-full px-4 py-3 bg-gradient-to-r from-blue-500 to-indigo-600 text-white rounded-xl font-semibold hover:from-blue-600 hover:to-indigo-700 transition-all hover:scale-105 text-center shadow-lg shadow-blue-500/20"
            >
              View Analysis
            </button>
          )}
        </div>
      </div>

      {/* Pricing Insights Section */}
      <div>
        <h2 className="text-2xl font-bold text-white mb-4">
          Pricing Insights
        </h2>
        <div className="grid md:grid-cols-2 gap-6">
          {/* Insight Card 1 */}
          <div className="bg-slate-900/50 backdrop-blur-sm rounded-2xl p-6 border border-slate-800 hover:border-slate-700 transition-colors">
            <div className="flex items-start gap-4 mb-4">
              <div className="w-12 h-12 bg-amber-500/10 rounded-xl flex items-center justify-center flex-shrink-0">
                <svg className="w-6 h-6 text-amber-500" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                </svg>
              </div>
              <div>
                <h3 className="text-lg font-semibold text-white mb-2">
                  Potential underpricing detected
                </h3>
                <p className="text-slate-400 leading-relaxed">
                  Your current plans are likely underpriced compared to the market. Based on similar SaaS products, you may be leaving revenue on the table.
                </p>
              </div>
            </div>
            <div className="pt-4 border-t border-slate-800">
              <p className="text-sm text-slate-500">
                Connect Stripe to get personalized insights
              </p>
            </div>
          </div>

          {/* Insight Card 2 */}
          <div className="bg-slate-900/50 backdrop-blur-sm rounded-2xl p-6 border border-slate-800 hover:border-slate-700 transition-colors">
            <div className="flex items-start gap-4 mb-4">
              <div className="w-12 h-12 bg-blue-500/10 rounded-xl flex items-center justify-center flex-shrink-0">
                <svg className="w-6 h-6 text-blue-500" fill="currentColor" viewBox="0 0 20 20">
                  <path d="M2 11a1 1 0 011-1h2a1 1 0 011 1v5a1 1 0 01-1 1H3a1 1 0 01-1-1v-5zM8 7a1 1 0 011-1h2a1 1 0 011 1v9a1 1 0 01-1 1H9a1 1 0 01-1-1V7zM14 4a1 1 0 011-1h2a1 1 0 011 1v12a1 1 0 01-1 1h-2a1 1 0 01-1-1V4z" />
                </svg>
              </div>
              <div>
                <h3 className="text-lg font-semibold text-white mb-2">
                  Price increase opportunity
                </h3>
                <p className="text-slate-400 leading-relaxed">
                  Increasing your Pro plan by 10–15% could increase MRR without a major churn risk. Industry benchmarks show similar products at higher price points.
                </p>
              </div>
            </div>
            <div className="pt-4 border-t border-slate-800">
              <p className="text-sm text-slate-500">
                Add competitors for detailed comparison
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* Business Snapshot */}
      <div>
        <h2 className="text-2xl font-bold text-white mb-4">
          Business Snapshot
        </h2>
        {metricsLoading ? (
          <div className="bg-slate-900/50 backdrop-blur-sm rounded-2xl p-6 border border-slate-800">
            <div className="flex items-center justify-center py-8">
              <svg className="animate-spin h-8 w-8 text-blue-500" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
              </svg>
            </div>
          </div>
        ) : hasMetrics && metrics ? (
          <div className="grid md:grid-cols-3 gap-4">
            {/* MRR */}
            <div className="bg-slate-900/50 backdrop-blur-sm rounded-xl p-6 border border-slate-800 hover:border-emerald-500/30 transition-colors">
              <div className="flex items-center gap-3 mb-3">
                <div className="w-10 h-10 bg-emerald-500/10 rounded-lg flex items-center justify-center">
                  <svg className="w-5 h-5 text-emerald-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                </div>
                <div className="text-sm text-slate-400">Monthly Recurring Revenue</div>
              </div>
              <div className="text-3xl font-bold text-white">
                {formatCurrency(metrics.mrr, metrics.currency)}
              </div>
              <div className="text-xs text-slate-500 mt-1">
                {metrics.currency}
              </div>
            </div>

            {/* Customers */}
            <div className="bg-slate-900/50 backdrop-blur-sm rounded-xl p-6 border border-slate-800 hover:border-blue-500/30 transition-colors">
              <div className="flex items-center gap-3 mb-3">
                <div className="w-10 h-10 bg-blue-500/10 rounded-lg flex items-center justify-center">
                  <svg className="w-5 h-5 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                  </svg>
                </div>
                <div className="text-sm text-slate-400">Customers</div>
              </div>
              <div className="text-3xl font-bold text-white">
                {metrics.customers.toLocaleString()}
              </div>
              <div className="text-xs text-slate-500 mt-1">
                Active subscribers
              </div>
            </div>

            {/* Churn Rate */}
            <div className="bg-slate-900/50 backdrop-blur-sm rounded-xl p-6 border border-slate-800 hover:border-amber-500/30 transition-colors">
              <div className="flex items-center gap-3 mb-3">
                <div className="w-10 h-10 bg-amber-500/10 rounded-lg flex items-center justify-center">
                  <svg className="w-5 h-5 text-amber-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 17h8m0 0V9m0 8l-8-8-4 4-6-6" />
                  </svg>
                </div>
                <div className="text-sm text-slate-400">Monthly Churn Rate</div>
              </div>
              <div className="text-3xl font-bold text-white">
                {metrics.monthly_churn_rate.toFixed(1)}%
              </div>
              <div className="text-xs text-slate-500 mt-1">
                {metrics.monthly_churn_rate < 3 ? 'Healthy' : metrics.monthly_churn_rate < 5 ? 'Moderate' : 'High'}
              </div>
            </div>
          </div>
        ) : (
          // CTA when metrics not set
          <div className="bg-gradient-to-br from-indigo-500/10 to-purple-500/10 backdrop-blur-sm rounded-2xl p-6 border border-indigo-500/20">
            <div className="flex items-start gap-4">
              <div className="w-12 h-12 bg-indigo-500/20 rounded-xl flex items-center justify-center flex-shrink-0">
                <svg className="w-6 h-6 text-indigo-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                </svg>
              </div>
              <div className="flex-1">
                <h3 className="text-lg font-semibold text-white mb-2">
                  Add your business metrics
                </h3>
                <p className="text-slate-400 text-sm leading-relaxed mb-4">
                  Add your business metrics to get more context in your pricing analysis. Track your MRR, customer count, and churn rate to receive more accurate recommendations.
                </p>
                <button
                  onClick={() => navigate('/app/settings')}
                  className="px-5 py-2.5 bg-gradient-to-r from-indigo-500 to-purple-600 text-white rounded-xl font-medium text-sm hover:from-indigo-600 hover:to-purple-700 transition-all hover:scale-105 shadow-lg shadow-indigo-500/20"
                >
                  Add Business Metrics
                </button>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Quick Stats */}
      <div>
        <h2 className="text-2xl font-bold text-white mb-4">
          Quick Stats
        </h2>
        <div className="grid md:grid-cols-4 gap-4">
          <div className="bg-slate-900/50 backdrop-blur-sm rounded-xl p-6 border border-slate-800">
            <div className="text-sm text-slate-400 mb-1">Plans Defined</div>
            <div className="text-3xl font-bold text-white">{plans.length}</div>
          </div>
          <div className="bg-slate-900/50 backdrop-blur-sm rounded-xl p-6 border border-slate-800">
            <div className="text-sm text-slate-400 mb-1">Competitors Added</div>
            <div className="text-3xl font-bold text-white">{competitors.length}</div>
          </div>
          <div className="bg-slate-900/50 backdrop-blur-sm rounded-xl p-6 border border-slate-800">
            <div className="text-sm text-slate-400 mb-1">Analyses Run</div>
            <div className="text-3xl font-bold text-white">{analyses.length}</div>
          </div>
          <div className="bg-slate-900/50 backdrop-blur-sm rounded-xl p-6 border border-slate-800">
            <div className="text-sm text-slate-400 mb-1">Reports Generated</div>
            <div className="text-3xl font-bold text-white">0</div>
          </div>
        </div>
        
        {/* Analysis Progress Insight */}
        {analyses.length > 0 && (
          <div className="mt-4 p-4 bg-blue-500/5 border border-blue-500/20 rounded-xl">
            <p className="text-sm text-slate-300">
              {analyses.length === 1
                ? "You've run 1 analysis so far. "
                : `You've run ${analyses.length} analyses so far. `}
              {analyses.length >= 3
                ? "Great progress! Track your pricing evolution over time in the Analyses page."
                : "Keep running analyses to track your pricing evolution over time."}
            </p>
          </div>
        )}
      </div>
    </div>
  );
};

export default Overview;
