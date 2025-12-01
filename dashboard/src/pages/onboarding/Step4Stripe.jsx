const Step4Stripe = ({ data, onChange }) => {
  const handleConnectStripe = () => {
    console.log('Connect Stripe clicked (coming soon)');
    onChange({ willConnectStripe: true });
  };

  const handleSkip = () => {
    onChange({ willConnectStripe: false });
  };

  return (
    <div>
      <div className="mb-8 text-center">
        <h2 className="text-3xl font-bold text-white mb-2">
          Connect Stripe
        </h2>
        <p className="text-slate-400">
          Optional, but recommended for better insights
        </p>
      </div>

      <div className="space-y-6">
        {/* Explanation */}
        <div className="bg-slate-900/50 rounded-xl p-6 border border-slate-700">
          <p className="text-slate-300 leading-relaxed mb-4">
            Revalyze can read your MRR, churn and plan data from Stripe in <span className="font-semibold text-white">read-only mode</span> to build a much more accurate pricing model.
          </p>
          <ul className="space-y-2">
            <li className="flex items-start gap-2 text-slate-400 text-sm">
              <svg className="w-5 h-5 text-blue-400 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
              </svg>
              <span>Analyze your revenue trends and customer behavior</span>
            </li>
            <li className="flex items-start gap-2 text-slate-400 text-sm">
              <svg className="w-5 h-5 text-blue-400 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
              </svg>
              <span>Get more accurate pricing recommendations</span>
            </li>
            <li className="flex items-start gap-2 text-slate-400 text-sm">
              <svg className="w-5 h-5 text-blue-400 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
              </svg>
              <span>We only request read-only access to your data</span>
            </li>
          </ul>
        </div>

        {/* Connect Button */}
        <button
          onClick={handleConnectStripe}
          className="w-full px-6 py-4 bg-gradient-to-r from-blue-500 to-indigo-600 text-white rounded-xl font-semibold hover:from-blue-600 hover:to-indigo-700 hover:scale-105 hover:shadow-xl hover:shadow-blue-500/30 transition-all flex items-center justify-center gap-2"
        >
          <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-8.707l-3-3a1 1 0 00-1.414 1.414L10.586 9H7a1 1 0 100 2h3.586l-1.293 1.293a1 1 0 101.414 1.414l3-3a1 1 0 000-1.414z" clipRule="evenodd" />
          </svg>
          Connect Stripe (coming soon)
        </button>

        {/* Skip Option */}
        <div className="text-center">
          <button
            onClick={handleSkip}
            className="text-slate-400 hover:text-slate-300 transition-colors text-sm font-medium"
          >
            Skip for now
          </button>
        </div>

        {/* Info */}
        <div className="bg-slate-700/30 rounded-xl p-4">
          <p className="text-xs text-slate-400 text-center">
            You can always connect Stripe later from your dashboard settings
          </p>
        </div>
      </div>
    </div>
  );
};

export default Step4Stripe;

