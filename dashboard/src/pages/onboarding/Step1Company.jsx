const Step1Company = ({ data, onChange }) => {
  const handleChange = (e) => {
    const { name, value } = e.target;
    onChange({ [name]: value });
  };

  return (
    <div>
      <div className="mb-8 text-center">
        <h2 className="text-3xl font-bold text-white mb-2">
          Company Basics
        </h2>
        <p className="text-slate-400">
          Let's start with some information about your company
        </p>
      </div>

      <div className="space-y-6">
        {/* Company Name */}
        <div>
          <label htmlFor="companyName" className="block text-sm font-semibold text-slate-300 mb-2">
            Company Name
          </label>
          <input
            id="companyName"
            name="companyName"
            type="text"
            value={data.companyName}
            onChange={handleChange}
            className="w-full px-4 py-3 rounded-xl bg-slate-900/50 border border-slate-700 text-white placeholder-slate-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all"
            placeholder="Acme Inc."
            required
          />
        </div>

        {/* Website */}
        <div>
          <label htmlFor="website" className="block text-sm font-semibold text-slate-300 mb-2">
            Website
          </label>
          <input
            id="website"
            name="website"
            type="url"
            value={data.website}
            onChange={handleChange}
            className="w-full px-4 py-3 rounded-xl bg-slate-900/50 border border-slate-700 text-white placeholder-slate-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all"
            placeholder="https://acme.com"
            required
          />
        </div>

        {/* Two columns */}
        <div className="grid md:grid-cols-2 gap-6">
          {/* Country */}
          <div>
            <label htmlFor="country" className="block text-sm font-semibold text-slate-300 mb-2">
              Country
            </label>
            <select
              id="country"
              name="country"
              value={data.country}
              onChange={handleChange}
              className="w-full px-4 py-3 pr-10 rounded-xl bg-slate-900/50 border border-slate-700 text-white focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all appearance-none bg-[url('data:image/svg+xml;charset=UTF-8,%3csvg xmlns=%27http://www.w3.org/2000/svg%27 viewBox=%270 0 24 24%27 fill=%27none%27 stroke=%27rgb(148 163 184)%27 stroke-width=%272%27 stroke-linecap=%27round%27 stroke-linejoin=%27round%27%3e%3cpolyline points=%276 9 12 15 18 9%27%3e%3c/polyline%3e%3c/svg%3e')] bg-[length:1.25rem_1.25rem] bg-[position:right_0.75rem_center] bg-no-repeat"
              required
            >
              <option value="">Select country</option>
              <option value="US">United States</option>
              <option value="UK">United Kingdom</option>
              <option value="NL">Netherlands</option>
              <option value="DE">Germany</option>
              <option value="FR">France</option>
              <option value="CA">Canada</option>
              <option value="OTHER">Other</option>
            </select>
          </div>

          {/* Team Size */}
          <div>
            <label htmlFor="teamSize" className="block text-sm font-semibold text-slate-300 mb-2">
              Team Size
            </label>
            <select
              id="teamSize"
              name="teamSize"
              value={data.teamSize}
              onChange={handleChange}
              className="w-full px-4 py-3 pr-10 rounded-xl bg-slate-900/50 border border-slate-700 text-white focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all appearance-none bg-[url('data:image/svg+xml;charset=UTF-8,%3csvg xmlns=%27http://www.w3.org/2000/svg%27 viewBox=%270 0 24 24%27 fill=%27none%27 stroke=%27rgb(148 163 184)%27 stroke-width=%272%27 stroke-linecap=%27round%27 stroke-linejoin=%27round%27%3e%3cpolyline points=%276 9 12 15 18 9%27%3e%3c/polyline%3e%3c/svg%3e')] bg-[length:1.25rem_1.25rem] bg-[position:right_0.75rem_center] bg-no-repeat"
              required
            >
              <option value="">Select team size</option>
              <option value="just-me">Just me</option>
              <option value="2-10">2–10</option>
              <option value="11-50">11–50</option>
              <option value="51-200">51–200</option>
              <option value="200+">200+</option>
            </select>
          </div>
        </div>

        {/* Product Category */}
        <div>
          <label htmlFor="productCategory" className="block text-sm font-semibold text-slate-300 mb-2">
            Product Category
          </label>
          <select
            id="productCategory"
            name="productCategory"
            value={data.productCategory}
            onChange={handleChange}
            className="w-full px-4 py-3 pr-10 rounded-xl bg-slate-900/50 border border-slate-700 text-white focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all appearance-none bg-[url('data:image/svg+xml;charset=UTF-8,%3csvg xmlns=%27http://www.w3.org/2000/svg%27 viewBox=%270 0 24 24%27 fill=%27none%27 stroke=%27rgb(148 163 184)%27 stroke-width=%272%27 stroke-linecap=%27round%27 stroke-linejoin=%27round%27%3e%3cpolyline points=%276 9 12 15 18 9%27%3e%3c/polyline%3e%3c/svg%3e')] bg-[length:1.25rem_1.25rem] bg-[position:right_0.75rem_center] bg-no-repeat"
            required
          >
            <option value="">Select category</option>
            <option value="analytics">Analytics</option>
            <option value="devtool">DevTool</option>
            <option value="crm">CRM</option>
            <option value="marketing">Marketing</option>
            <option value="fintech">Fintech</option>
            <option value="ai-tool">AI Tool</option>
            <option value="other">Other</option>
          </select>
        </div>
      </div>
    </div>
  );
};

export default Step1Company;

