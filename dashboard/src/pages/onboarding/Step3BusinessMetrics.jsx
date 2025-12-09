const PRICING_GOALS = [
  { value: 'revenue', label: 'Maximize Revenue' },
  { value: 'retention', label: 'Improve Retention' },
  { value: 'conversion', label: 'Increase Conversion' },
  { value: 'differentiation', label: 'Better Differentiation' },
];

const CURRENCIES = [
  { value: 'USD', label: 'USD ($)' },
  { value: 'EUR', label: 'EUR (€)' },
  { value: 'GBP', label: 'GBP (£)' },
  { value: 'TRY', label: 'TRY (₺)' },
];

const Step3BusinessMetrics = ({ data, onChange }) => {
  const handleChange = (e) => {
    const { name, value, type } = e.target;
    
    // Handle numeric inputs
    if (type === 'number') {
      const numValue = value === '' ? '' : parseFloat(value);
      onChange({ [name]: numValue });
    } else {
      onChange({ [name]: value });
    }
  };

  return (
    <div>
      <div className="mb-8 text-center">
        <h2 className="text-3xl font-bold text-white mb-2">Business Metrics</h2>
        <p className="text-slate-400">
          Share your key metrics to get personalized pricing recommendations.
        </p>
      </div>

      <div className="space-y-6">
        {/* Currency and MRR */}
        <div className="grid md:grid-cols-2 gap-6">
          {/* Currency */}
          <div>
            <label htmlFor="currency" className="block text-sm font-semibold text-slate-300 mb-2">
              Currency
            </label>
            <select
              id="currency"
              name="currency"
              value={data.currency || 'USD'}
              onChange={handleChange}
              className="w-full px-4 py-3 pr-10 rounded-xl bg-slate-900/50 border border-slate-700 text-white focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all appearance-none bg-[url('data:image/svg+xml;charset=UTF-8,%3csvg xmlns=%27http://www.w3.org/2000/svg%27 viewBox=%270 0 24 24%27 fill=%27none%27 stroke=%27rgb(148 163 184)%27 stroke-width=%272%27 stroke-linecap=%27round%27 stroke-linejoin=%27round%27%3e%3cpolyline points=%276 9 12 15 18 9%27%3e%3c/polyline%3e%3c/svg%3e')] bg-[length:1.25rem_1.25rem] bg-[position:right_0.75rem_center] bg-no-repeat"
            >
              {CURRENCIES.map((c) => (
                <option key={c.value} value={c.value}>
                  {c.label}
                </option>
              ))}
            </select>
          </div>

          {/* MRR */}
          <div>
            <label htmlFor="mrr" className="block text-sm font-semibold text-slate-300 mb-2">
              Monthly Recurring Revenue (MRR) <span className="text-red-400">*</span>
            </label>
            <div className="relative">
              <span className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-400">
                {data.currency === 'EUR' ? '€' : data.currency === 'GBP' ? '£' : data.currency === 'TRY' ? '₺' : '$'}
              </span>
              <input
                id="mrr"
                name="mrr"
                type="number"
                min="0"
                step="1"
                value={data.mrr ?? ''}
                onChange={handleChange}
                className="w-full px-4 py-3 pl-8 rounded-xl bg-slate-900/50 border border-slate-700 text-white placeholder-slate-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all"
                placeholder="10000"
                required
              />
            </div>
          </div>
        </div>

        {/* Monthly Churn */}
        <div>
          <label htmlFor="monthlyChurnRate" className="block text-sm font-semibold text-slate-300 mb-2">
            Monthly Churn Rate (%) <span className="text-red-400">*</span>
          </label>
          <div className="relative">
            <input
              id="monthlyChurnRate"
              name="monthlyChurnRate"
              type="number"
              min="0"
              max="100"
              step="0.1"
              value={data.monthlyChurnRate ?? ''}
              onChange={handleChange}
              className="w-full px-4 py-3 pr-10 rounded-xl bg-slate-900/50 border border-slate-700 text-white placeholder-slate-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all"
              placeholder="5"
              required
            />
            <span className="absolute right-4 top-1/2 -translate-y-1/2 text-slate-400">%</span>
          </div>
          <p className="mt-1 text-xs text-slate-500">
            The percentage of customers who cancel each month.
          </p>
        </div>

        {/* Pricing Goal */}
        <div>
          <label className="block text-sm font-semibold text-slate-300 mb-3">
            Primary Pricing Goal <span className="text-red-400">*</span>
          </label>
          <div className="grid md:grid-cols-2 gap-3">
            {PRICING_GOALS.map((goal) => (
              <label
                key={goal.value}
                className={`flex items-center p-4 rounded-xl border cursor-pointer transition-all ${
                  data.pricingGoal === goal.value
                    ? 'bg-blue-500/10 border-blue-500/50 shadow-sm'
                    : 'bg-slate-900/30 border-slate-700 hover:border-slate-600'
                }`}
              >
                <input
                  type="radio"
                  name="pricingGoal"
                  value={goal.value}
                  checked={data.pricingGoal === goal.value}
                  onChange={handleChange}
                  className="w-5 h-5 text-blue-500 border-slate-600 focus:ring-blue-500 focus:ring-2"
                />
                <span className="ml-3 text-slate-300">{goal.label}</span>
              </label>
            ))}
          </div>
        </div>

        {/* Target ARR Growth (Optional) */}
        <div>
          <label htmlFor="targetArrGrowth" className="block text-sm font-semibold text-slate-300 mb-2">
            Target ARR Growth (%) <span className="text-slate-500 font-normal">(optional)</span>
          </label>
          <div className="relative">
            <input
              id="targetArrGrowth"
              name="targetArrGrowth"
              type="number"
              min="0"
              max="1000"
              step="1"
              value={data.targetArrGrowth ?? ''}
              onChange={handleChange}
              className="w-full px-4 py-3 pr-10 rounded-xl bg-slate-900/50 border border-slate-700 text-white placeholder-slate-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all"
              placeholder="50"
            />
            <span className="absolute right-4 top-1/2 -translate-y-1/2 text-slate-400">%</span>
          </div>
          <p className="mt-1 text-xs text-slate-500">
            Your target annual recurring revenue growth. Leave empty if you don't have a specific target.
          </p>
        </div>

        {/* Info box */}
        <div className="bg-blue-500/10 border border-blue-500/30 rounded-xl p-4">
          <div className="flex items-start gap-3">
            <svg
              className="w-5 h-5 text-blue-400 flex-shrink-0 mt-0.5"
              fill="currentColor"
              viewBox="0 0 20 20"
            >
              <path
                fillRule="evenodd"
                d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z"
                clipRule="evenodd"
              />
            </svg>
            <div>
              <p className="text-sm text-slate-300">
                <span className="font-semibold">Why we need this:</span> Your business metrics help us generate personalized pricing recommendations that align with your goals.
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Step3BusinessMetrics;


