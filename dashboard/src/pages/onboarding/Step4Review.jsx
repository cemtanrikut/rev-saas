const PRICING_GOAL_LABELS = {
  revenue: 'Maximize Revenue',
  retention: 'Improve Retention',
  conversion: 'Increase Conversion',
  differentiation: 'Better Differentiation',
};

const Step4Review = ({ plans, competitors, businessMetrics, onGenerate, isLoading }) => {
  const formatPrice = (price, currency) => {
    const symbols = { USD: '$', EUR: '€', GBP: '£', TRY: '₺' };
    return `${symbols[currency] || currency}${price}`;
  };

  const formatCurrency = (value, currency) => {
    const symbols = { USD: '$', EUR: '€', GBP: '£', TRY: '₺' };
    return `${symbols[currency] || currency}${value?.toLocaleString() || 0}`;
  };

  // Count total competitor plans
  const totalCompetitorPlans = competitors.reduce((acc, c) => acc + (c.plans?.length || 0), 0);

  return (
    <div>
      <div className="mb-8 text-center">
        <h2 className="text-3xl font-bold text-white mb-2">Review & Confirm</h2>
        <p className="text-slate-400">
          Review your information before generating your first pricing analysis.
        </p>
      </div>

      <div className="space-y-6">
        {/* Pricing Plans */}
        <div className="bg-slate-900/50 rounded-xl p-6 border border-slate-700">
          <div className="flex items-center gap-3 mb-4">
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
            <h3 className="text-lg font-semibold text-white">Your Pricing Plans</h3>
            <span className="ml-auto text-sm text-slate-400">{plans.length} plan(s)</span>
          </div>
          
          {plans.length > 0 ? (
            <div className="space-y-2">
              {plans.map((plan) => (
                <div
                  key={plan.id}
                  className="flex items-center justify-between p-3 rounded-lg bg-slate-800/50"
                >
                  <span className="text-white font-medium">{plan.name}</span>
                  <span className="text-slate-400">
                    {formatPrice(plan.price, plan.currency)} / {plan.billingCycle}
                  </span>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-slate-500 italic">No plans added</p>
          )}
        </div>

        {/* Competitors */}
        <div className="bg-slate-900/50 rounded-xl p-6 border border-slate-700">
          <div className="flex items-center gap-3 mb-4">
            <div className="w-10 h-10 rounded-lg bg-purple-500/20 flex items-center justify-center">
              <svg
                className="w-5 h-5 text-purple-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4"
                />
              </svg>
            </div>
            <h3 className="text-lg font-semibold text-white">Competitors</h3>
            <span className="ml-auto text-sm text-slate-400">
              {competitors.length > 0 
                ? `${competitors.length} competitor(s), ${totalCompetitorPlans} plan(s)` 
                : 'None added'}
            </span>
          </div>
          
          {competitors.length > 0 ? (
            <div className="space-y-3">
              {competitors.map((comp) => (
                <div
                  key={comp.id}
                  className="p-3 rounded-lg bg-slate-800/50"
                >
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-white font-medium">{comp.name}</span>
                    <span className="text-slate-500 text-sm">{comp.plans?.length || 0} plan(s)</span>
                  </div>
                  {comp.plans && comp.plans.length > 0 && (
                    <div className="space-y-1 pl-3 border-l-2 border-slate-700">
                      {comp.plans.map((plan, idx) => (
                        <div key={idx} className="flex items-center justify-between text-sm">
                          <span className="text-slate-400">{plan.name}</span>
                          <span className="text-slate-500">
                            {formatPrice(plan.price, plan.currency)} / {plan.billingCycle}
                          </span>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              ))}
            </div>
          ) : (
            <p className="text-slate-500 italic">No competitors added — you can add them later from the dashboard.</p>
          )}
        </div>

        {/* Business Metrics */}
        <div className="bg-slate-900/50 rounded-xl p-6 border border-slate-700">
          <div className="flex items-center gap-3 mb-4">
            <div className="w-10 h-10 rounded-lg bg-green-500/20 flex items-center justify-center">
              <svg
                className="w-5 h-5 text-green-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"
                />
              </svg>
            </div>
            <h3 className="text-lg font-semibold text-white">Business Metrics</h3>
          </div>
          
          <div className="grid md:grid-cols-2 gap-4">
            <div className="p-3 rounded-lg bg-slate-800/50">
              <p className="text-slate-400 text-sm mb-1">MRR</p>
              <p className="text-white font-semibold text-lg">
                {formatCurrency(businessMetrics.mrr, businessMetrics.currency)}
              </p>
            </div>
            <div className="p-3 rounded-lg bg-slate-800/50">
              <p className="text-slate-400 text-sm mb-1">Monthly Churn</p>
              <p className="text-white font-semibold text-lg">
                {businessMetrics.monthlyChurnRate ?? 0}%
              </p>
            </div>
            <div className="p-3 rounded-lg bg-slate-800/50">
              <p className="text-slate-400 text-sm mb-1">Pricing Goal</p>
              <p className="text-white font-semibold">
                {PRICING_GOAL_LABELS[businessMetrics.pricingGoal] || 'Not set'}
              </p>
            </div>
            {(businessMetrics.targetArrGrowth !== '' && businessMetrics.targetArrGrowth !== undefined && businessMetrics.targetArrGrowth !== null) && (
              <div className="p-3 rounded-lg bg-slate-800/50">
                <p className="text-slate-400 text-sm mb-1">Target ARR Growth</p>
                <p className="text-white font-semibold text-lg">
                  {businessMetrics.targetArrGrowth}%
                </p>
              </div>
            )}
          </div>
        </div>

        {/* Generate Button */}
        <button
          onClick={onGenerate}
          disabled={isLoading}
          className="w-full px-8 py-4 bg-gradient-to-r from-blue-500 to-indigo-600 text-white rounded-xl font-semibold hover:from-blue-600 hover:to-indigo-700 hover:scale-[1.02] hover:shadow-xl hover:shadow-blue-500/30 transition-all flex items-center justify-center gap-3 disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:scale-100"
        >
          {isLoading ? (
            <>
              <svg className="animate-spin w-5 h-5" fill="none" viewBox="0 0 24 24">
                <circle
                  className="opacity-25"
                  cx="12"
                  cy="12"
                  r="10"
                  stroke="currentColor"
                  strokeWidth="4"
                />
                <path
                  className="opacity-75"
                  fill="currentColor"
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                />
              </svg>
              Generating Analysis...
            </>
          ) : (
            <>
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M13 10V3L4 14h7v7l9-11h-7z"
                />
              </svg>
              Generate My First Pricing Analysis
            </>
          )}
        </button>

        {/* Info */}
        <p className="text-center text-xs text-slate-500">
          Your data is securely stored and only used to generate personalized pricing recommendations.
        </p>
      </div>
    </div>
  );
};

export default Step4Review;
