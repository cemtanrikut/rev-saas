import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAnalysis } from '../../context/AnalysisContext';
import { useBusinessMetrics } from '../../context/BusinessMetricsContext';
import { analysisApi, downloadBlob } from '../../lib/apiClient';

const Analyses = () => {
  const navigate = useNavigate();
  const { analyses, selectedAnalysis, selectAnalysis, isLoading, isRunning, runAnalysis, limitError, clearLimitError } = useAnalysis();
  const { metrics, hasMetrics, isLoading: metricsLoading } = useBusinessMetrics();
  
  // PDF export state
  const [exportingId, setExportingId] = useState(null);
  const [exportError, setExportError] = useState(null);

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

  const formatPrice = (price) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 0,
      maximumFractionDigits: 2
    }).format(price);
  };

  const formatCurrency = (value, currency = 'USD') => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency,
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(value);
  };

  const formatDate = (isoString) => {
    const date = new Date(isoString);
    return new Intl.DateTimeFormat('en-US', {
      month: 'long',
      day: 'numeric',
      year: 'numeric',
      hour: 'numeric',
      minute: '2-digit'
    }).format(date);
  };

  const formatDateShort = (isoString) => {
    const date = new Date(isoString);
    return new Intl.DateTimeFormat('en-US', {
      month: 'short',
      day: 'numeric',
      hour: 'numeric',
      minute: '2-digit'
    }).format(date);
  };

  const formatChangePercent = (percent) => {
    if (percent === 0) return '0%';
    return percent > 0 ? `+${percent.toFixed(1)}%` : `${percent.toFixed(1)}%`;
  };

  const formatChangeAbsolute = (amount) => {
    if (amount === 0) return '$0';
    const formatted = new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 0,
      maximumFractionDigits: 2
    }).format(Math.abs(amount));

    return amount > 0 ? `+${formatted}` : `-${formatted}`;
  };

  const getPositionBadge = (position) => {
    const badges = {
      'lowest': { label: 'Lowest', color: 'bg-red-500/20 text-red-400' },
      'below_median': { label: 'Below Median', color: 'bg-amber-500/20 text-amber-400' },
      'around_median': { label: 'Around Median', color: 'bg-emerald-500/20 text-emerald-400' },
      'above_median': { label: 'Above Median', color: 'bg-blue-500/20 text-blue-400' },
      'highest': { label: 'Highest', color: 'bg-purple-500/20 text-purple-400' },
      'unknown': { label: 'Unknown', color: 'bg-slate-500/20 text-slate-400' }
    };
    return badges[position] || badges['unknown'];
  };

  const getActionBadge = (action) => {
    const badges = {
      'raise_price': { label: 'Raise Price', color: 'text-emerald-400' },
      'raise_price_conservative': { label: 'Raise (Conservative)', color: 'text-emerald-400/70' },
      'raise_price_aggressive': { label: 'Raise (Aggressive)', color: 'text-emerald-500' },
      'lower_price': { label: 'Lower Price', color: 'text-amber-400' },
      'keep': { label: 'Keep Current', color: 'text-blue-400' },
      'keep_for_growth': { label: 'Keep for Growth', color: 'text-blue-400/70' },
      'consider_increase': { label: 'Consider Increase', color: 'text-teal-400' },
      'add_competitors': { label: 'Add Data', color: 'text-slate-400' },
      'restructure': { label: 'Restructure', color: 'text-purple-400' }
    };
    return badges[action] || { label: action, color: 'text-slate-400' };
  };

  // Business Snapshot Component
  const BusinessSnapshotCard = () => {
    if (metricsLoading) {
      return (
        <div className="bg-slate-900/50 backdrop-blur-sm rounded-2xl p-6 border border-slate-800">
          <div className="flex items-center gap-3">
            <svg className="animate-spin h-5 w-5 text-blue-500" viewBox="0 0 24 24">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
            </svg>
            <span className="text-slate-400 text-sm">Loading business metrics...</span>
          </div>
        </div>
      );
    }

    if (!hasMetrics || !metrics) {
      return (
        <div className="bg-slate-900/50 backdrop-blur-sm rounded-2xl p-5 border border-slate-800">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-9 h-9 bg-indigo-500/10 rounded-lg flex items-center justify-center flex-shrink-0">
                <svg className="w-5 h-5 text-indigo-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                </svg>
              </div>
              <div>
                <p className="text-sm text-slate-300">
                  Add your business metrics to improve analysis accuracy
                </p>
              </div>
            </div>
            <button
              onClick={() => navigate('/app/settings')}
              className="px-4 py-2 text-sm font-medium text-indigo-400 hover:text-indigo-300 hover:bg-indigo-500/10 rounded-lg transition-all"
            >
              Add Metrics
            </button>
          </div>
        </div>
      );
    }

    return (
      <div className="bg-slate-900/50 backdrop-blur-sm rounded-2xl p-5 border border-slate-800">
        <div className="flex items-center gap-3 mb-4">
          <div className="w-8 h-8 bg-emerald-500/10 rounded-lg flex items-center justify-center flex-shrink-0">
            <svg className="w-4 h-4 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
            </svg>
          </div>
          <h3 className="text-sm font-semibold text-slate-300">Business Snapshot</h3>
          <button
            onClick={() => navigate('/app/settings')}
            className="ml-auto text-xs text-slate-500 hover:text-slate-400 transition-colors"
          >
            Edit
          </button>
        </div>
        <div className="grid grid-cols-4 gap-4">
          <div>
            <div className="text-xs text-slate-500 mb-1">MRR</div>
            <div className="text-lg font-bold text-white">
              {formatCurrency(metrics.mrr, metrics.currency)}
            </div>
          </div>
          <div>
            <div className="text-xs text-slate-500 mb-1">Customers</div>
            <div className="text-lg font-bold text-white">
              {metrics.customers.toLocaleString()}
            </div>
          </div>
          <div>
            <div className="text-xs text-slate-500 mb-1">Churn Rate</div>
            <div className={`text-lg font-bold ${
              metrics.monthly_churn_rate > 7 ? 'text-red-400' : 
              metrics.monthly_churn_rate < 3 ? 'text-emerald-400' : 'text-amber-400'
            }`}>
              {metrics.monthly_churn_rate.toFixed(1)}%
            </div>
          </div>
          <div>
            <div className="text-xs text-slate-500 mb-1">Currency</div>
            <div className="text-lg font-bold text-white">
              {metrics.currency}
            </div>
          </div>
        </div>
      </div>
    );
  };

  const generateKeyInsights = (analysis) => {
    if (!analysis || !analysis.recommendations) return [];
    
    const insights = [];
    const { stats, recommendations } = analysis;

    // Insight based on recommendations
    const raiseCount = recommendations.filter(r => r.suggestedAction === 'raise_price').length;
    const lowerCount = recommendations.filter(r => r.suggestedAction === 'lower_price').length;
    const keepCount = recommendations.filter(r => r.suggestedAction === 'keep').length;

    if (raiseCount > 0 && lowerCount === 0) {
      insights.push(`All ${raiseCount} plan(s) could benefit from a price increase based on market positioning.`);
    } else if (lowerCount > 0 && raiseCount === 0) {
      insights.push(`${lowerCount} plan(s) may be overpriced relative to the market median.`);
    } else if (keepCount === recommendations.length) {
      insights.push("Your pricing is well-aligned with market benchmarks. Maintain current pricing strategy.");
    } else {
      insights.push(`Mixed positioning: ${raiseCount} underpriced, ${keepCount} well-positioned, ${lowerCount} potentially overpriced.`);
    }

    // Insight about competitive context
    if (stats.numCompetitors === 0) {
      insights.push("Add competitor pricing data to improve the accuracy of these recommendations.");
    } else if (stats.numCompetitors < 3) {
      insights.push("Consider adding more competitors for better market benchmarking accuracy.");
    } else {
      insights.push(`Analysis based on ${stats.numCompetitors} competitors provides solid market context.`);
    }

    // Insight about potential revenue impact
    if (raiseCount > 0) {
      insights.push("Implementing suggested price increases could unlock additional revenue without significant churn risk.");
    }

    return insights;
  };

  const handleRunNewAnalysis = async () => {
    const result = await runAnalysis();
    if (!result.success) {
      console.error('Failed to run analysis:', result.error);
    }
  };

  // Handle PDF export
  const handleExportPdf = async (analysisId, analysisDate) => {
    setExportingId(analysisId);
    setExportError(null);
    
    try {
      const { ok, blob } = await analysisApi.exportPdf(analysisId);
      
      if (ok && blob) {
        // Generate filename with date
        const dateStr = new Date(analysisDate).toISOString().split('T')[0];
        const filename = `pricing-report-${dateStr}.pdf`;
        downloadBlob(blob, filename);
      }
    } catch (error) {
      console.error('Failed to export PDF:', error);
      setExportError(error.message || 'Failed to export PDF');
      // Clear error after 5 seconds
      setTimeout(() => setExportError(null), 5000);
    } finally {
      setExportingId(null);
    }
  };

  // Loading state
  if (isLoading) {
    return (
      <div className="max-w-4xl mx-auto">
        <div className="text-center py-20">
          <div className="w-12 h-12 border-4 border-blue-500 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
          <p className="text-slate-400">Loading analysis history...</p>
        </div>
      </div>
    );
  }

  // Empty state
  if (analyses.length === 0) {
    return (
      <div className="max-w-4xl mx-auto space-y-6">
        {/* Business Snapshot - visible even in empty state */}
        <BusinessSnapshotCard />

        <div className="text-center py-20">
          <div className="w-20 h-20 bg-slate-800 rounded-2xl flex items-center justify-center mx-auto mb-6">
            <svg className="w-10 h-10 text-slate-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
            </svg>
          </div>
          <h2 className="text-2xl font-bold text-white mb-3">
            No analyses yet
          </h2>
          <p className="text-slate-400 mb-8 max-w-md mx-auto leading-relaxed">
            Run your first pricing analysis from the Overview page once you have defined your plans and added at least one competitor.
          </p>
          <button
            onClick={() => navigate('/app/overview')}
            className="px-6 py-3 bg-gradient-to-r from-blue-500 to-indigo-600 text-white rounded-xl font-semibold hover:from-blue-600 hover:to-indigo-700 hover:scale-105 transition-all shadow-lg shadow-blue-500/20"
          >
            Go to Overview
          </button>
        </div>
      </div>
    );
  }

  if (!selectedAnalysis) {
    return null;
  }

  const keyInsights = generateKeyInsights(selectedAnalysis);
  const recommendations = selectedAnalysis.recommendations || [];

  // Format report date
  const reportDate = new Intl.DateTimeFormat('en-US', {
    month: 'long',
    day: '2-digit',
    year: 'numeric'
  }).format(new Date(selectedAnalysis.createdAt));

  return (
    <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
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

      {/* Export Error Toast */}
      {exportError && (
        <div className="fixed bottom-6 right-6 z-50 animate-fade-in-up">
          <div className="flex items-center gap-3 px-4 py-3 bg-red-500/10 border border-red-500/30 rounded-xl shadow-lg backdrop-blur-sm">
            <svg className="w-5 h-5 text-red-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <span className="text-sm text-red-300">{exportError}</span>
            <button 
              onClick={() => setExportError(null)}
              className="text-red-400 hover:text-red-300 ml-2"
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>
      )}
      
      {/* Main Content - Pricing Strategy Report */}
      <div className="lg:col-span-2 space-y-6">
        
        {/* ═══════════════════════════════════════════════════════════════
            REPORT HEADER
        ═══════════════════════════════════════════════════════════════ */}
        <div className="bg-gradient-to-br from-slate-900 via-slate-900 to-slate-800 rounded-2xl p-8 border border-slate-800 relative overflow-hidden">
          {/* Subtle background pattern */}
          <div className="absolute inset-0 opacity-5">
            <div className="absolute top-0 right-0 w-64 h-64 bg-blue-500 rounded-full blur-3xl"></div>
            <div className="absolute bottom-0 left-0 w-48 h-48 bg-indigo-500 rounded-full blur-3xl"></div>
          </div>
          
          <div className="relative">
            <div className="flex items-start justify-between mb-6">
              <div>
                <div className="flex items-center gap-3 mb-3">
                  <div className="w-10 h-10 bg-blue-500/20 rounded-xl flex items-center justify-center">
                    <svg className="w-5 h-5 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                    </svg>
                  </div>
                  <span className="text-xs font-medium text-blue-400 bg-blue-500/10 px-3 py-1 rounded-full uppercase tracking-wider">
                    Pricing Report
                  </span>
                  {/* AI-powered badge */}
                  {(selectedAnalysis.ai_summary || (selectedAnalysis.ai_scenarios && selectedAnalysis.ai_scenarios.length > 0)) && (
                    <span className="flex items-center gap-1.5 text-xs font-medium text-purple-400 bg-gradient-to-r from-purple-500/20 to-blue-500/20 px-3 py-1 rounded-full border border-purple-500/30">
                      <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                      </svg>
                      AI-powered insight
                    </span>
                  )}
                </div>
                <h1 className="text-3xl font-bold text-white mb-2">
                  Pricing Strategy Report
                </h1>
                <div className="flex items-center gap-4 text-sm text-slate-400">
                  <span className="flex items-center gap-1.5">
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                    </svg>
                    {reportDate}
                  </span>
                  <span className="w-1 h-1 bg-slate-600 rounded-full"></span>
                  <span>{selectedAnalysis.numPlans} plan{selectedAnalysis.numPlans !== 1 ? 's' : ''}</span>
                  <span className="w-1 h-1 bg-slate-600 rounded-full"></span>
                  <span>{selectedAnalysis.numCompetitors} competitor{selectedAnalysis.numCompetitors !== 1 ? 's' : ''}</span>
                </div>
              </div>
              <div className="flex items-center gap-2">
                {/* Export PDF Button */}
                <button
                  onClick={() => handleExportPdf(selectedAnalysis.id, selectedAnalysis.createdAt)}
                  disabled={exportingId === selectedAnalysis.id}
                  className="flex items-center gap-2 px-4 py-2.5 bg-slate-800 text-slate-300 rounded-xl font-medium hover:bg-slate-700 hover:text-white transition-all disabled:opacity-50 disabled:cursor-not-allowed border border-slate-700"
                >
                  {exportingId === selectedAnalysis.id ? (
                    <>
                      <svg className="animate-spin h-4 w-4" viewBox="0 0 24 24">
                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                      </svg>
                      Exporting...
                    </>
                  ) : (
                    <>
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                      </svg>
                      Export PDF
                    </>
                  )}
                </button>
                
                {/* Run New Analysis Button */}
                <button
                  onClick={handleRunNewAnalysis}
                  disabled={isRunning}
                  className="flex items-center gap-2 px-4 py-2.5 bg-blue-500/10 text-blue-400 rounded-xl font-medium hover:bg-blue-500/20 transition-all disabled:opacity-50 disabled:cursor-not-allowed border border-blue-500/20"
                >
                  {isRunning ? (
                    <>
                      <svg className="animate-spin h-4 w-4" viewBox="0 0 24 24">
                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                      </svg>
                      Running...
                    </>
                  ) : (
                    <>
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                      </svg>
                      Run New
                    </>
                  )}
                </button>
              </div>
            </div>
          </div>
        </div>

        {/* ═══════════════════════════════════════════════════════════════
            BUSINESS SNAPSHOT + EXECUTIVE SUMMARY
        ═══════════════════════════════════════════════════════════════ */}
        <div className="grid grid-cols-1 lg:grid-cols-5 gap-6">
          {/* Business Snapshot - Compact */}
          <div className="lg:col-span-2">
            <div className="bg-slate-900/50 backdrop-blur-sm rounded-2xl p-5 border border-slate-800 h-full">
              <div className="flex items-center gap-2 mb-4">
                <svg className="w-4 h-4 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                </svg>
                <h3 className="text-sm font-semibold text-slate-300">Business Snapshot</h3>
              </div>
              
              {metricsLoading ? (
                <div className="flex items-center justify-center py-4">
                  <svg className="animate-spin h-5 w-5 text-blue-500" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                  </svg>
                </div>
              ) : hasMetrics && metrics ? (
                <div className="space-y-3">
                  <div className="flex items-center justify-between py-2 border-b border-slate-800/50">
                    <span className="text-xs text-slate-500">MRR</span>
                    <span className="text-sm font-semibold text-white">{formatCurrency(metrics.mrr, metrics.currency)}</span>
                  </div>
                  <div className="flex items-center justify-between py-2 border-b border-slate-800/50">
                    <span className="text-xs text-slate-500">Customers</span>
                    <span className="text-sm font-semibold text-white">{metrics.customers.toLocaleString()}</span>
                  </div>
                  <div className="flex items-center justify-between py-2 border-b border-slate-800/50">
                    <span className="text-xs text-slate-500">Churn Rate</span>
                    <span className={`text-sm font-semibold ${
                      metrics.monthly_churn_rate > 7 ? 'text-red-400' : 
                      metrics.monthly_churn_rate < 3 ? 'text-emerald-400' : 'text-amber-400'
                    }`}>{metrics.monthly_churn_rate.toFixed(1)}%</span>
                  </div>
                  <div className="flex items-center justify-between py-2">
                    <span className="text-xs text-slate-500">Currency</span>
                    <span className="text-sm font-semibold text-white">{metrics.currency}</span>
                  </div>
                </div>
              ) : (
                <div className="text-center py-4">
                  <p className="text-xs text-slate-500 mb-3">No metrics configured</p>
                  <button
                    onClick={() => navigate('/app/settings')}
                    className="text-xs text-blue-400 hover:text-blue-300"
                  >
                    Add metrics →
                  </button>
                </div>
              )}
            </div>
          </div>

          {/* Executive Summary */}
          <div className="lg:col-span-3">
            <div className="bg-slate-900/50 backdrop-blur-sm rounded-2xl p-6 border border-slate-800 h-full">
              <div className="flex items-center gap-2 mb-4">
                <svg className="w-4 h-4 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
                <h3 className="text-sm font-semibold text-slate-300">Executive Summary</h3>
              </div>
              <p className="text-slate-300 text-sm leading-relaxed whitespace-pre-line">
                {selectedAnalysis.summary}
              </p>
              <p className="text-xs text-slate-500 mt-4 pt-3 border-t border-slate-800/50 flex items-center gap-1.5">
                <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                Based on your current business metrics and market positioning.
              </p>
            </div>
          </div>
        </div>

        {/* ═══════════════════════════════════════════════════════════════
            KEY INSIGHTS
        ═══════════════════════════════════════════════════════════════ */}
        {keyInsights.length > 0 && (
          <div className="bg-slate-900/50 backdrop-blur-sm rounded-2xl p-6 border border-slate-800">
            <div className="flex items-center gap-2 mb-4">
              <svg className="w-4 h-4 text-indigo-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
              </svg>
              <h3 className="text-sm font-semibold text-slate-300">Key Insights</h3>
            </div>
            <div className="grid grid-cols-1 gap-3">
              {keyInsights.map((insight, index) => (
                <div key={index} className="flex items-start gap-3 p-3 bg-slate-800/30 rounded-xl">
                  <span className="w-5 h-5 bg-indigo-500/20 rounded-full flex items-center justify-center flex-shrink-0 mt-0.5">
                    <span className="text-xs font-semibold text-indigo-400">{index + 1}</span>
                  </span>
                  <span className="text-sm text-slate-300 leading-relaxed">{insight}</span>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* ═══════════════════════════════════════════════════════════════
            AI PRICING INSIGHTS (only shown for paid plans with AI data)
        ═══════════════════════════════════════════════════════════════ */}
        {(selectedAnalysis.ai_summary || (selectedAnalysis.ai_scenarios && selectedAnalysis.ai_scenarios.length > 0)) && (
          <div className="bg-gradient-to-br from-purple-500/5 via-slate-900/50 to-blue-500/5 backdrop-blur-sm rounded-2xl border border-purple-500/20 overflow-hidden">
            {/* AI Badge Header */}
            <div className="px-6 py-4 bg-gradient-to-r from-purple-500/10 to-blue-500/10 border-b border-purple-500/20">
              <div className="flex items-center gap-3">
                <div className="w-8 h-8 bg-gradient-to-br from-purple-500 to-blue-500 rounded-lg flex items-center justify-center">
                  <svg className="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                  </svg>
                </div>
                <div>
                  <div className="flex items-center gap-2">
                    <h3 className="text-base font-semibold text-white">AI Pricing Insights</h3>
                    <span className="text-xs font-medium text-purple-400 bg-purple-500/20 px-2 py-0.5 rounded-full">
                      AI-powered
                    </span>
                  </div>
                  <p className="text-xs text-slate-400 mt-0.5">
                    Advanced analysis powered by machine learning
                  </p>
                </div>
              </div>
            </div>

            <div className="p-6 space-y-5">
              {/* AI Summary */}
              {selectedAnalysis.ai_summary && (
                <div>
                  <div className="flex items-center gap-2 mb-3">
                    <svg className="w-4 h-4 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                    </svg>
                    <h4 className="text-sm font-semibold text-slate-300">AI Summary</h4>
                  </div>
                  <div className="p-4 bg-slate-800/40 rounded-xl border border-slate-700/50">
                    <p className="text-sm text-slate-300 leading-relaxed whitespace-pre-line">
                      {selectedAnalysis.ai_summary}
                    </p>
                  </div>
                </div>
              )}

              {/* AI Scenarios */}
              {selectedAnalysis.ai_scenarios && selectedAnalysis.ai_scenarios.length > 0 && (
                <div>
                  <div className="flex items-center gap-2 mb-3">
                    <svg className="w-4 h-4 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
                    </svg>
                    <h4 className="text-sm font-semibold text-slate-300">Pricing Scenarios</h4>
                    <span className="text-xs text-slate-500 bg-slate-800/60 px-2 py-0.5 rounded-full">
                      {selectedAnalysis.ai_scenarios.length} scenario{selectedAnalysis.ai_scenarios.length !== 1 ? 's' : ''}
                    </span>
                  </div>
                  <div className="space-y-2">
                    {selectedAnalysis.ai_scenarios.map((scenario, index) => (
                      <div 
                        key={index} 
                        className="flex items-start gap-3 p-3 bg-slate-800/30 rounded-xl border border-slate-700/30 hover:border-slate-600/50 transition-colors"
                      >
                        <span className="w-6 h-6 bg-gradient-to-br from-purple-500/30 to-blue-500/30 rounded-full flex items-center justify-center flex-shrink-0 mt-0.5">
                          <span className="text-xs font-bold text-white">{index + 1}</span>
                        </span>
                        <p className="text-sm text-slate-300 leading-relaxed">{scenario}</p>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* AI Disclaimer */}
              <div className="flex items-start gap-2 pt-3 border-t border-slate-800/50">
                <svg className="w-4 h-4 text-slate-500 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <p className="text-xs text-slate-500">
                  AI insights are generated based on your data and market context. Results should be reviewed alongside your business knowledge.
                </p>
              </div>
            </div>
          </div>
        )}

        {/* ═══════════════════════════════════════════════════════════════
            PRICING RECOMMENDATIONS
        ═══════════════════════════════════════════════════════════════ */}
        {recommendations.length > 0 && (
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <svg className="w-5 h-5 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <h2 className="text-lg font-semibold text-white">Pricing Recommendations</h2>
              </div>
              <span className="text-xs text-slate-500 bg-slate-800 px-3 py-1 rounded-full">
                {recommendations.length} plan{recommendations.length !== 1 ? 's' : ''}
              </span>
            </div>

            {/* Recommendation Cards */}
            <div className="space-y-4">
              {recommendations.map((rec, index) => {
                const positionBadge = getPositionBadge(rec.position);
                const actionBadge = getActionBadge(rec.suggestedAction);
                const priceChange = rec.suggestedPrice - rec.currentPrice;
                const changePercent = rec.currentPrice > 0 
                  ? ((priceChange / rec.currentPrice) * 100) 
                  : 0;
                
                return (
                  <div 
                    key={rec.planId || index} 
                    className="bg-slate-900/50 backdrop-blur-sm rounded-2xl border border-slate-800 overflow-hidden hover:border-slate-700 transition-all hover:shadow-lg hover:shadow-slate-900/50"
                  >
                    {/* Card Header with Plan Name and Action Badge */}
                    <div className="p-5 pb-4 border-b border-slate-800/50">
                      <div className="flex items-start justify-between">
                        <div className="flex items-center gap-3">
                          <div className="w-10 h-10 bg-slate-800 rounded-xl flex items-center justify-center">
                            <span className="text-lg font-bold text-slate-300">
                              {rec.planName.charAt(0).toUpperCase()}
                            </span>
                          </div>
                          <div>
                            <h3 className="text-base font-semibold text-white">
                              {rec.planName}
                            </h3>
                            <span className={`inline-flex px-2 py-0.5 rounded text-xs font-medium ${positionBadge.color}`}>
                              {positionBadge.label}
                            </span>
                          </div>
                        </div>
                        <div className={`px-3 py-1.5 rounded-lg text-sm font-semibold ${
                          actionBadge.color.includes('emerald') ? 'bg-emerald-500/10 border border-emerald-500/20' :
                          actionBadge.color.includes('amber') ? 'bg-amber-500/10 border border-amber-500/20' :
                          actionBadge.color.includes('blue') ? 'bg-blue-500/10 border border-blue-500/20' :
                          actionBadge.color.includes('teal') ? 'bg-teal-500/10 border border-teal-500/20' :
                          'bg-slate-500/10 border border-slate-500/20'
                        } ${actionBadge.color}`}>
                          {actionBadge.label}
                        </div>
                      </div>
                    </div>

                    {/* Price Comparison Row */}
                    <div className="p-5 bg-slate-800/20">
                      <div className="flex items-center justify-between">
                        {/* Current Price */}
                        <div className="text-center flex-1">
                          <div className="text-xs text-slate-500 mb-1 uppercase tracking-wider">Current</div>
                          <div className="text-2xl font-bold text-slate-300">
                            {formatPrice(rec.currentPrice)}
                          </div>
                        </div>

                        {/* Arrow / Change Indicator */}
                        <div className="flex-shrink-0 px-6">
                          <div className={`flex items-center gap-2 px-4 py-2 rounded-full ${
                            priceChange > 0 ? 'bg-emerald-500/10 text-emerald-400' :
                            priceChange < 0 ? 'bg-amber-500/10 text-amber-400' :
                            'bg-slate-500/10 text-slate-400'
                          }`}>
                            {priceChange !== 0 && (
                              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M14 5l7 7m0 0l-7 7m7-7H3" />
                              </svg>
                            )}
                            <span className="text-sm font-semibold">
                              {changePercent !== 0 ? `${changePercent > 0 ? '+' : ''}${changePercent.toFixed(0)}%` : 'No change'}
                            </span>
                          </div>
                        </div>

                        {/* Suggested Price */}
                        <div className="text-center flex-1">
                          <div className="text-xs text-slate-500 mb-1 uppercase tracking-wider">Suggested</div>
                          <div className={`text-2xl font-bold ${
                            priceChange > 0 ? 'text-emerald-400' :
                            priceChange < 0 ? 'text-amber-400' :
                            'text-blue-400'
                          }`}>
                            {formatPrice(rec.suggestedPrice)}
                          </div>
                        </div>
                      </div>
                    </div>

                    {/* Rationale */}
                    <div className="px-5 py-4 border-t border-slate-800/50">
                      <p className="text-sm text-slate-400 leading-relaxed">
                        {rec.rationale}
                      </p>
                    </div>
                  </div>
                );
              })}
            </div>

            {/* Report Footer / Disclaimer */}
            <div className="flex items-start gap-3 p-4 bg-slate-800/20 rounded-xl border border-slate-800/50">
              <svg className="w-4 h-4 text-slate-500 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
              </svg>
              <p className="text-xs text-slate-500 leading-relaxed">
                <strong className="text-slate-400">Disclaimer:</strong> These recommendations are generated using rule-based analysis comparing your pricing to competitor benchmarks. 
                Results are adjusted based on your business metrics including MRR, customer count, and churn rate.
              </p>
            </div>
          </div>
        )}

        {/* ═══════════════════════════════════════════════════════════════
            ACTION BUTTONS
        ═══════════════════════════════════════════════════════════════ */}
        <div className="flex items-center gap-3 pt-2">
          <button
            onClick={() => navigate('/app/overview')}
            className="flex items-center gap-2 px-5 py-2.5 bg-slate-800 text-slate-300 rounded-xl font-medium hover:bg-slate-700 hover:text-white transition-all"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 19l-7-7m0 0l7-7m-7 7h18" />
            </svg>
            Overview
          </button>
          <button
            onClick={() => navigate('/app/plans')}
            className="flex items-center gap-2 px-5 py-2.5 bg-slate-800 text-slate-300 rounded-xl font-medium hover:bg-slate-700 hover:text-white transition-all"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
            </svg>
            Edit Plans
          </button>
          <button
            onClick={() => navigate('/app/competitors')}
            className="flex items-center gap-2 px-5 py-2.5 bg-slate-800 text-slate-300 rounded-xl font-medium hover:bg-slate-700 hover:text-white transition-all"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
            Competitors
          </button>
        </div>
      </div>

      {/* Sidebar - Analysis History */}
      <aside className="lg:col-span-1">
        <div className="bg-slate-900/50 backdrop-blur-sm rounded-2xl p-6 border border-slate-800 sticky top-6">
          {/* Header */}
          <div className="flex items-center justify-between mb-5">
            <div className="flex items-center gap-2">
              <div className="w-8 h-8 bg-indigo-500/10 rounded-lg flex items-center justify-center">
                <svg className="w-4 h-4 text-indigo-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
              <h3 className="text-base font-semibold text-white">
                Analysis History
              </h3>
            </div>
            <span className="text-xs text-slate-400 bg-slate-800/80 px-2.5 py-1 rounded-full font-medium">
              {analyses.length} report{analyses.length !== 1 ? 's' : ''}
            </span>
          </div>

          {/* Analysis Cards */}
          <div className="space-y-3">
            {analyses.map((analysis, index) => {
              const isSelected = analysis.id === selectedAnalysis.id;
              const isLatest = index === 0;
              const raiseCount = (analysis.recommendations || []).filter(r => 
                r.suggestedAction === 'raise_price' || 
                r.suggestedAction === 'raise_price_conservative' || 
                r.suggestedAction === 'raise_price_aggressive'
              ).length;
              const lowerCount = (analysis.recommendations || []).filter(r => r.suggestedAction === 'lower_price').length;
              const keepCount = (analysis.recommendations || []).filter(r => 
                r.suggestedAction === 'keep' || r.suggestedAction === 'keep_for_growth'
              ).length;
              
              // Truncate summary - get first sentence only
              const summaryPreview = analysis.summary 
                ? analysis.summary.split('\n')[0].split('. ')[0] + '.'
                : 'No summary available';

              // Format date nicely
              const formattedDate = new Intl.DateTimeFormat('en-US', {
                month: 'short',
                day: '2-digit',
                year: 'numeric'
              }).format(new Date(analysis.createdAt));

              return (
                <div
                  key={analysis.id}
                  className={`group rounded-xl border transition-all duration-200 ${
                    isSelected
                      ? 'border-blue-500/50 bg-blue-500/5 shadow-lg shadow-blue-500/5'
                      : 'border-slate-800 bg-slate-900/30 hover:border-slate-700 hover:bg-slate-800/30 hover:shadow-md hover:shadow-slate-900/50'
                  }`}
                >
                  {/* Card Content */}
                  <div className="p-4">
                    {/* Date Header */}
                    <div className="flex items-center justify-between mb-3">
                      <div className="flex items-center gap-2">
                        <svg className={`w-4 h-4 ${isSelected ? 'text-blue-400' : 'text-slate-500'}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                        </svg>
                        <span className="text-sm font-medium text-white">
                          {formattedDate}
                        </span>
                      </div>
                      <div className="flex items-center gap-1.5">
                        {/* AI Badge */}
                        {(analysis.ai_summary || (analysis.ai_scenarios && analysis.ai_scenarios.length > 0)) && (
                          <span className="flex items-center gap-1 text-xs font-medium text-purple-400 bg-purple-500/10 px-1.5 py-0.5 rounded">
                            <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                            </svg>
                            AI
                          </span>
                        )}
                        {isLatest && (
                          <span className="text-xs font-medium text-indigo-400 bg-indigo-500/10 px-2 py-0.5 rounded-full">
                            Latest
                          </span>
                        )}
                        {isSelected && !isLatest && (
                          <span className="text-xs font-medium text-blue-400 bg-blue-500/10 px-2 py-0.5 rounded-full">
                            Active
                          </span>
                        )}
                      </div>
                    </div>

                    {/* Stats Row */}
                    <div className="flex items-center gap-4 mb-3 py-2 px-3 bg-slate-800/30 rounded-lg">
                      <div className="flex items-center gap-1.5">
                        <svg className="w-3.5 h-3.5 text-slate-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                        </svg>
                        <span className="text-xs text-slate-400">
                          <span className="font-semibold text-slate-300">{analysis.numPlans}</span> plan{analysis.numPlans !== 1 ? 's' : ''}
                        </span>
                      </div>
                      <div className="w-px h-3 bg-slate-700"></div>
                      <div className="flex items-center gap-1.5">
                        <svg className="w-3.5 h-3.5 text-slate-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
                        </svg>
                        <span className="text-xs text-slate-400">
                          <span className="font-semibold text-slate-300">{analysis.numCompetitors}</span> competitor{analysis.numCompetitors !== 1 ? 's' : ''}
                        </span>
                      </div>
                    </div>

                    {/* Summary Preview */}
                    <p className="text-xs text-slate-500 leading-relaxed mb-3 line-clamp-1">
                      {summaryPreview}
                    </p>

                    {/* Insights Row */}
                    <div className="flex items-center gap-1.5 flex-wrap">
                      {raiseCount > 0 && (
                        <span className="inline-flex items-center gap-1 text-xs font-medium text-emerald-400 bg-emerald-500/10 px-2 py-1 rounded-md">
                          <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 10l7-7m0 0l7 7m-7-7v18" />
                          </svg>
                          {raiseCount}
                        </span>
                      )}
                      {lowerCount > 0 && (
                        <span className="inline-flex items-center gap-1 text-xs font-medium text-amber-400 bg-amber-500/10 px-2 py-1 rounded-md">
                          <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 14l-7 7m0 0l-7-7m7 7V3" />
                          </svg>
                          {lowerCount}
                        </span>
                      )}
                      {keepCount > 0 && (
                        <span className="inline-flex items-center gap-1 text-xs font-medium text-blue-400 bg-blue-500/10 px-2 py-1 rounded-md">
                          <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 12h14" />
                          </svg>
                          {keepCount}
                        </span>
                      )}
                    </div>
                  </div>

                  {/* Action Buttons */}
                  <div className="px-4 pb-4 space-y-2">
                    <button
                      onClick={() => selectAnalysis(analysis.id)}
                      disabled={isSelected}
                      className={`w-full flex items-center justify-center gap-2 text-xs font-semibold py-2.5 rounded-lg transition-all ${
                        isSelected
                          ? 'text-slate-500 bg-slate-800/50 cursor-default'
                          : 'text-white bg-slate-800 hover:bg-slate-700 group-hover:bg-slate-700'
                      }`}
                    >
                      {isSelected ? (
                        <>
                          <svg className="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 20 20">
                            <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                          </svg>
                          Currently Viewing
                        </>
                      ) : (
                        <>
                          <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                          </svg>
                          View Report
                        </>
                      )}
                    </button>
                    
                    {/* Export PDF Button */}
                    <button
                      onClick={() => handleExportPdf(analysis.id, analysis.createdAt)}
                      disabled={exportingId === analysis.id}
                      className="w-full flex items-center justify-center gap-2 text-xs font-medium py-2 rounded-lg transition-all text-slate-400 bg-slate-800/30 hover:bg-slate-800 hover:text-slate-300 border border-slate-700/50 disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      {exportingId === analysis.id ? (
                        <>
                          <svg className="animate-spin w-3.5 h-3.5" viewBox="0 0 24 24">
                            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                          </svg>
                          Exporting...
                        </>
                      ) : (
                        <>
                          <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                          </svg>
                          Export PDF
                        </>
                      )}
                    </button>
                  </div>
                </div>
              );
            })}
          </div>

          {/* Run New Analysis CTA */}
          <div className="mt-4 pt-4 border-t border-slate-800">
            <button
              onClick={handleRunNewAnalysis}
              disabled={isRunning}
              className="w-full text-center text-sm font-medium py-2.5 rounded-xl bg-gradient-to-r from-blue-500/10 to-indigo-500/10 text-blue-400 hover:from-blue-500/20 hover:to-indigo-500/20 transition-all border border-blue-500/20 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isRunning ? (
                <span className="flex items-center justify-center gap-2">
                  <svg className="animate-spin h-4 w-4" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                  </svg>
                  Running...
                </span>
              ) : (
                '+ Run New Analysis'
              )}
            </button>
          </div>
        </div>
      </aside>
    </div>
  );
};

export default Analyses;
