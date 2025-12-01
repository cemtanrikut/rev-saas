import { Link } from 'react-router-dom';

const Overview = () => {
  const checklistItems = [
    { label: 'Created account', completed: true },
    { label: 'Completed onboarding', completed: true },
    { label: 'Connected Stripe', completed: false },
    { label: 'Added competitors', completed: false },
    { label: 'Ran first pricing analysis', completed: false }
  ];

  return (
    <div className="max-w-7xl mx-auto space-y-8">
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
                    className={`w-5 h-5 rounded-full flex items-center justify-center flex-shrink-0 ${
                      item.completed
                        ? 'bg-blue-500'
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
                    className={item.completed ? 'text-slate-300' : 'text-slate-400'}
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
            <div className="w-10 h-10 bg-blue-500 rounded-lg flex items-center justify-center flex-shrink-0">
              <svg className="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
            </div>
            <div>
              <h3 className="text-lg font-semibold text-white mb-1">
                Next best action
              </h3>
              <p className="text-sm text-slate-400">
                Add your competitors to get started with pricing analysis
              </p>
            </div>
          </div>

          <Link
            to="/app/competitors"
            className="block w-full px-4 py-3 bg-gradient-to-r from-blue-500 to-indigo-600 text-white rounded-xl font-semibold hover:from-blue-600 hover:to-indigo-700 transition-all hover:scale-105 text-center shadow-lg shadow-blue-500/20"
          >
            Go to competitors
          </Link>
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

      {/* Quick Stats */}
      <div>
        <h2 className="text-2xl font-bold text-white mb-4">
          Quick Stats
        </h2>
        <div className="grid md:grid-cols-4 gap-4">
          <div className="bg-slate-900/50 backdrop-blur-sm rounded-xl p-6 border border-slate-800">
            <div className="text-sm text-slate-400 mb-1">Competitors Added</div>
            <div className="text-3xl font-bold text-white">0</div>
          </div>
          <div className="bg-slate-900/50 backdrop-blur-sm rounded-xl p-6 border border-slate-800">
            <div className="text-sm text-slate-400 mb-1">Analyses Run</div>
            <div className="text-3xl font-bold text-white">0</div>
          </div>
          <div className="bg-slate-900/50 backdrop-blur-sm rounded-xl p-6 border border-slate-800">
            <div className="text-sm text-slate-400 mb-1">Reports Generated</div>
            <div className="text-3xl font-bold text-white">0</div>
          </div>
          <div className="bg-slate-900/50 backdrop-blur-sm rounded-xl p-6 border border-slate-800">
            <div className="text-sm text-slate-400 mb-1">Data Sources</div>
            <div className="text-3xl font-bold text-white">0</div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Overview;

