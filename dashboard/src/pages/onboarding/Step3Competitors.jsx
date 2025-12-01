const Step3Competitors = ({ data, onChange }) => {
  const handleChange = (e) => {
    const { name, value } = e.target;
    onChange({ [name]: value });
  };

  return (
    <div>
      <div className="mb-8 text-center">
        <h2 className="text-3xl font-bold text-white mb-2">
          Competitors
        </h2>
        <p className="text-slate-400 max-w-2xl mx-auto">
          Add 1–3 competitor pricing pages. This helps Revalyze benchmark your pricing. You can skip this now and add them later in the dashboard.
        </p>
      </div>

      <div className="space-y-6">
        {/* Competitor 1 */}
        <div>
          <label htmlFor="competitor1" className="block text-sm font-semibold text-slate-300 mb-2">
            Competitor 1 <span className="text-slate-500 font-normal">(optional)</span>
          </label>
          <input
            id="competitor1"
            name="competitor1"
            type="url"
            value={data.competitor1}
            onChange={handleChange}
            className="w-full px-4 py-3 rounded-xl bg-slate-900/50 border border-slate-700 text-white placeholder-slate-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all"
            placeholder="https://competitor1.com/pricing"
          />
        </div>

        {/* Competitor 2 */}
        <div>
          <label htmlFor="competitor2" className="block text-sm font-semibold text-slate-300 mb-2">
            Competitor 2 <span className="text-slate-500 font-normal">(optional)</span>
          </label>
          <input
            id="competitor2"
            name="competitor2"
            type="url"
            value={data.competitor2}
            onChange={handleChange}
            className="w-full px-4 py-3 rounded-xl bg-slate-900/50 border border-slate-700 text-white placeholder-slate-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all"
            placeholder="https://competitor2.com/pricing"
          />
        </div>

        {/* Competitor 3 */}
        <div>
          <label htmlFor="competitor3" className="block text-sm font-semibold text-slate-300 mb-2">
            Competitor 3 <span className="text-slate-500 font-normal">(optional)</span>
          </label>
          <input
            id="competitor3"
            name="competitor3"
            type="url"
            value={data.competitor3}
            onChange={handleChange}
            className="w-full px-4 py-3 rounded-xl bg-slate-900/50 border border-slate-700 text-white placeholder-slate-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all"
            placeholder="https://competitor3.com/pricing"
          />
        </div>

        {/* Info box */}
        <div className="bg-blue-500/10 border border-blue-500/30 rounded-xl p-4">
          <div className="flex items-start gap-3">
            <svg className="w-5 h-5 text-blue-400 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd" />
            </svg>
            <div>
              <p className="text-sm text-slate-300">
                <span className="font-semibold">Tip:</span> Don't worry if you skip this step. You can always add competitors later from your dashboard.
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Step3Competitors;

