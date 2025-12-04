import { useNavigate } from 'react-router-dom';
import { useAnalysis } from '../../context/AnalysisContext';

const Analyses = () => {
  const navigate = useNavigate();
  const { analyses, selectedAnalysis, selectAnalysis } = useAnalysis();

  const formatPrice = (price, interval) => {
    const formattedPrice = new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 0,
      maximumFractionDigits: 2
    }).format(price);

    return `${formattedPrice} / ${interval === 'monthly' ? 'month' : 'year'}`;
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
    return percent > 0 ? `+${percent.toFixed(1)}%` : `${percent.toFixed(1)}%`;
  };

  const formatChangeAbsolute = (amount) => {
    const formatted = new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 0,
      maximumFractionDigits: 2
    }).format(Math.abs(amount));

    return amount > 0 ? `+${formatted}` : `-${formatted}`;
  };

  const getAverageChangePercent = (analysis) => {
    if (!analysis.suggestions || !analysis.suggestions.length) return null;
    const total = analysis.suggestions.reduce((sum, s) => sum + s.changePercent, 0);
    return total / analysis.suggestions.length;
  };

  const generateKeyInsights = (analysis) => {
    const insights = [];
    const { averageChangePercent, maxChangePercent } = analysis.stats;

    // Insight 1: Overall pricing position
    if (averageChangePercent < 5) {
      insights.push("Your pricing is broadly aligned with typical SaaS ranges. Only minor adjustments are recommended.");
    } else if (averageChangePercent < 15) {
      insights.push("Your pricing seems somewhat underpriced. A moderate uplift could unlock additional revenue without significant churn risk.");
    } else {
      insights.push("Your plans are likely underpriced. A more assertive price realignment may be justified to capture the value you deliver.");
    }

    // Insight 2: Single plan with high uplift
    if (maxChangePercent > 20) {
      const highestPlan = analysis.suggestions.find(s => s.changePercent === maxChangePercent);
      insights.push(`Your "${highestPlan.planName}" plan appears significantly underpriced and may warrant a more aggressive increase.`);
    }

    // Insight 3: Competitive context
    if (analysis.stats.numCompetitors === 0) {
      insights.push("Adding competitor pricing data would strengthen these recommendations and provide better benchmarking context.");
    } else if (analysis.stats.numCompetitors < 3) {
      insights.push("Consider adding more competitors to improve the accuracy of our pricing benchmarks and market positioning analysis.");
    } else {
      insights.push("Your competitive set provides solid benchmarking context. These recommendations reflect common patterns in your market segment.");
    }

    return insights;
  };

  // Empty state
  if (analyses.length === 0) {
    return (
      <div className="max-w-4xl mx-auto">
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
    return null; // Shouldn't happen if analyses.length > 0
  }

  const keyInsights = generateKeyInsights(selectedAnalysis);

  return (
    <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
      {/* Main Content - Analysis Details */}
      <div className="lg:col-span-2 space-y-6">
        {/* Page Header */}
        <div>
          <h1 className="text-3xl font-bold text-white mb-3">
            {selectedAnalysis.id === analyses[0]?.id ? 'Latest Pricing Analysis' : 'Selected Pricing Analysis'}
          </h1>
          <p className="text-slate-400">
            Run on {formatDate(selectedAnalysis.createdAt)}. Used {selectedAnalysis.stats.numPlans} plan
            {selectedAnalysis.stats.numPlans > 1 ? 's' : ''} and {selectedAnalysis.stats.numCompetitors} competitor
            {selectedAnalysis.stats.numCompetitors !== 1 ? 's' : ''} as input.
          </p>
        </div>

        {/* Summary Section */}
        <div className="bg-slate-900/50 backdrop-blur-sm rounded-2xl p-8 border border-slate-800">
          <div className="flex items-start gap-4 mb-6">
            <div className="w-10 h-10 bg-blue-500/10 rounded-xl flex items-center justify-center flex-shrink-0">
              <svg className="w-5 h-5 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
            </div>
            <div>
              <h2 className="text-xl font-semibold text-white mb-3">
                Summary
              </h2>
              <p className="text-slate-300 leading-relaxed">
                {selectedAnalysis.summary}
              </p>
            </div>
          </div>
        </div>

        {/* Key Insights */}
        <div className="bg-slate-900/50 backdrop-blur-sm rounded-2xl p-8 border border-slate-800">
          <div className="flex items-start gap-4">
            <div className="w-10 h-10 bg-indigo-500/10 rounded-xl flex items-center justify-center flex-shrink-0">
              <svg className="w-5 h-5 text-indigo-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
              </svg>
            </div>
            <div className="flex-1">
              <h2 className="text-xl font-semibold text-white mb-4">
                Key Insights
              </h2>
              <ul className="space-y-3">
                {keyInsights.map((insight, index) => (
                  <li key={index} className="flex items-start gap-3">
                    <span className="w-1.5 h-1.5 bg-indigo-400 rounded-full mt-2 flex-shrink-0"></span>
                    <span className="text-slate-300 leading-relaxed">{insight}</span>
                  </li>
                ))}
              </ul>
            </div>
          </div>
        </div>

        {/* Pricing Recommendations Table */}
        <div className="bg-slate-900/50 backdrop-blur-sm rounded-2xl border border-slate-800 overflow-hidden">
          <div className="p-6 border-b border-slate-800">
            <h2 className="text-xl font-semibold text-white">
              Suggested Price Changes
            </h2>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-slate-800/50">
                <tr>
                  <th className="text-left px-6 py-4 text-sm font-semibold text-slate-300">
                    Plan
                  </th>
                  <th className="text-right px-6 py-4 text-sm font-semibold text-slate-300">
                    Current Price
                  </th>
                  <th className="text-right px-6 py-4 text-sm font-semibold text-slate-300">
                    Suggested Price
                  </th>
                  <th className="text-right px-6 py-4 text-sm font-semibold text-slate-300">
                    Change
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-800">
                {selectedAnalysis.suggestions.map((suggestion) => (
                  <tr key={suggestion.planId} className="hover:bg-slate-800/30 transition-colors">
                    <td className="px-6 py-5">
                      <div className="font-semibold text-white">
                        {suggestion.planName}
                      </div>
                      <div className="text-sm text-slate-400">
                        Billed {suggestion.planInterval}
                      </div>
                    </td>
                    <td className="px-6 py-5 text-right">
                      <div className="text-slate-300 font-medium">
                        {formatPrice(suggestion.currentPrice, suggestion.planInterval)}
                      </div>
                    </td>
                    <td className="px-6 py-5 text-right">
                      <div className="text-blue-400 font-semibold text-lg">
                        {formatPrice(suggestion.suggestedPrice, suggestion.planInterval)}
                      </div>
                    </td>
                    <td className="px-6 py-5 text-right">
                      <div className="flex flex-col items-end gap-1">
                        <span className="text-emerald-400 font-semibold">
                          {formatChangeAbsolute(suggestion.changeAbsolute)}
                        </span>
                        <span className="text-sm text-emerald-400/70">
                          {formatChangePercent(suggestion.changePercent)}
                        </span>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* Disclaimer */}
          <div className="px-6 py-4 bg-slate-800/30 border-t border-slate-800">
            <p className="text-sm text-slate-500 leading-relaxed">
              These recommendations are based on heuristic logic for now. In the next versions, we will incorporate your revenue metrics and a more advanced AI model to estimate MRR impact and churn risk with greater precision.
            </p>
          </div>
        </div>

        {/* Action Buttons */}
        <div className="flex gap-4">
          <button
            onClick={() => navigate('/app/overview')}
            className="px-6 py-3 bg-slate-800 text-white rounded-xl font-semibold hover:bg-slate-700 transition-all"
          >
            Back to Overview
          </button>
          <button
            onClick={() => navigate('/app/plans')}
            className="px-6 py-3 bg-slate-800 text-white rounded-xl font-semibold hover:bg-slate-700 transition-all"
          >
            Adjust My Plans
          </button>
          <button
            onClick={() => navigate('/app/competitors')}
            className="px-6 py-3 bg-slate-800 text-white rounded-xl font-semibold hover:bg-slate-700 transition-all"
          >
            Update Competitors
          </button>
        </div>
      </div>

      {/* Sidebar - Analysis History */}
      <aside className="lg:col-span-1">
        <div className="bg-slate-900/50 backdrop-blur-sm rounded-2xl p-6 border border-slate-800 sticky top-6">
          <h3 className="text-lg font-semibold text-white mb-4">
            Analysis History
          </h3>
          
          {analyses.length === 1 ? (
            <p className="text-sm text-slate-400 mb-4">
              You have run 1 analysis so far.
            </p>
          ) : (
            <p className="text-sm text-slate-400 mb-4">
              You've run {analyses.length} analyses so far.
            </p>
          )}

          <div className="space-y-2">
            {analyses.map((analysis) => {
              const isSelected = analysis.id === selectedAnalysis.id;
              const avgChange = getAverageChangePercent(analysis);

              return (
                <button
                  key={analysis.id}
                  onClick={() => selectAnalysis(analysis.id)}
                  className={`w-full text-left p-4 rounded-xl border transition-all ${
                    isSelected
                      ? 'border-blue-500 bg-slate-800/50 shadow-lg shadow-blue-500/10'
                      : 'border-slate-800 hover:bg-slate-800/30 hover:border-slate-700'
                  }`}
                >
                  <div className="flex items-start justify-between mb-2">
                    <div className="text-sm font-medium text-white">
                      {formatDateShort(analysis.createdAt)}
                    </div>
                    {isSelected && (
                      <svg className="w-4 h-4 text-blue-400" fill="currentColor" viewBox="0 0 20 20">
                        <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                      </svg>
                    )}
                  </div>
                  
                  <div className="space-y-1">
                    <div className="text-xs text-slate-400">
                      {analysis.stats.numPlans} plan{analysis.stats.numPlans > 1 ? 's' : ''} • {analysis.stats.numCompetitors} competitor{analysis.stats.numCompetitors !== 1 ? 's' : ''}
                    </div>
                    {avgChange !== null && (
                      <div className="text-xs font-medium text-emerald-400">
                        Avg. uplift: {formatChangePercent(avgChange)}
                      </div>
                    )}
                  </div>
                </button>
              );
            })}
          </div>
        </div>
      </aside>
    </div>
  );
};

export default Analyses;
