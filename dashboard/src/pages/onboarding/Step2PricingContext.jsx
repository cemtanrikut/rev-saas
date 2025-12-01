const Step2PricingContext = ({ data, onChange }) => {
  const handleChange = (e) => {
    const { name, value } = e.target;
    onChange({ [name]: value });
  };

  return (
    <div>
      <div className="mb-8 text-center">
        <h2 className="text-3xl font-bold text-white mb-2">
          Pricing Context
        </h2>
        <p className="text-slate-400">
          Help us understand your current pricing situation
        </p>
      </div>

      <div className="space-y-6">
        {/* Plan Count */}
        <div>
          <label htmlFor="planCount" className="block text-sm font-semibold text-slate-300 mb-2">
            How many pricing plans do you currently have?
          </label>
          <select
            id="planCount"
            name="planCount"
            value={data.planCount}
            onChange={handleChange}
            className="w-full px-4 py-3 pr-10 rounded-xl bg-slate-900/50 border border-slate-700 text-white focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all appearance-none bg-[url('data:image/svg+xml;charset=UTF-8,%3csvg xmlns=%27http://www.w3.org/2000/svg%27 viewBox=%270 0 24 24%27 fill=%27none%27 stroke=%27rgb(148 163 184)%27 stroke-width=%272%27 stroke-linecap=%27round%27 stroke-linejoin=%27round%27%3e%3cpolyline points=%276 9 12 15 18 9%27%3e%3c/polyline%3e%3c/svg%3e')] bg-[length:1.25rem_1.25rem] bg-[position:right_0.75rem_center] bg-no-repeat"
            required
          >
            <option value="">Select plan count</option>
            <option value="1">1</option>
            <option value="2-3">2–3</option>
            <option value="4-6">4–6</option>
            <option value="7+">7+</option>
          </select>
        </div>

        {/* Main Pricing Concern */}
        <div>
          <label className="block text-sm font-semibold text-slate-300 mb-3">
            What's your main pricing concern?
          </label>
          <div className="space-y-3">
            {[
              { value: 'underpriced', label: 'We think we are underpriced' },
              { value: 'afraid-increase', label: "We're afraid to increase prices" },
              { value: 'messy-plans', label: 'Our plans feel messy / unclear' },
              { value: 'competitor-knowledge', label: "We don't know what competitors charge" },
              { value: 'launch-new', label: 'We want to launch new plans' }
            ].map((option) => (
              <label
                key={option.value}
                className={`flex items-center p-4 rounded-xl border cursor-pointer transition-all ${
                  data.mainPricingConcern === option.value
                    ? 'bg-blue-500/10 border-blue-500/50 shadow-sm'
                    : 'bg-slate-900/30 border-slate-700 hover:border-slate-600'
                }`}
              >
                <input
                  type="radio"
                  name="mainPricingConcern"
                  value={option.value}
                  checked={data.mainPricingConcern === option.value}
                  onChange={handleChange}
                  className="w-5 h-5 text-blue-500 border-slate-600 focus:ring-blue-500 focus:ring-2"
                />
                <span className="ml-3 text-slate-300">{option.label}</span>
              </label>
            ))}
          </div>
        </div>

        {/* Optional Notes */}
        <div>
          <label htmlFor="pricingNotes" className="block text-sm font-semibold text-slate-300 mb-2">
            Additional notes <span className="text-slate-500 font-normal">(optional)</span>
          </label>
          <textarea
            id="pricingNotes"
            name="pricingNotes"
            value={data.pricingNotes}
            onChange={handleChange}
            rows="4"
            className="w-full px-4 py-3 rounded-xl bg-slate-900/50 border border-slate-700 text-white placeholder-slate-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all resize-none"
            placeholder="If you want, tell us more about your pricing challenges."
          />
        </div>
      </div>
    </div>
  );
};

export default Step2PricingContext;

