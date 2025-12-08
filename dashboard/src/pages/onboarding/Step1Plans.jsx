import { useState } from 'react';

const CURRENCIES = [
  { value: 'USD', label: 'USD ($)' },
  { value: 'EUR', label: 'EUR (€)' },
  { value: 'GBP', label: 'GBP (£)' },
  { value: 'TRY', label: 'TRY (₺)' },
];

const BILLING_CYCLES = [
  { value: 'monthly', label: 'Monthly' },
  { value: 'yearly', label: 'Yearly' },
  { value: 'one-time', label: 'One-time' },
];

const Step1Plans = ({ data, onChange }) => {
  const [newPlan, setNewPlan] = useState({
    name: '',
    price: '',
    currency: 'USD',
    billingCycle: 'monthly',
  });
  const [error, setError] = useState('');

  const handleNewPlanChange = (e) => {
    const { name, value } = e.target;
    setNewPlan((prev) => ({ ...prev, [name]: value }));
    setError('');
  };

  const handleAddPlan = () => {
    if (!newPlan.name.trim()) {
      setError('Plan name is required');
      return;
    }
    if (!newPlan.price || parseFloat(newPlan.price) <= 0) {
      setError('Price must be greater than 0');
      return;
    }

    const plan = {
      id: Date.now(),
      name: newPlan.name.trim(),
      price: parseFloat(newPlan.price),
      currency: newPlan.currency,
      billingCycle: newPlan.billingCycle,
    };

    onChange([...data, plan]);
    setNewPlan({ name: '', price: '', currency: 'USD', billingCycle: 'monthly' });
    setError('');
  };

  const handleRemovePlan = (id) => {
    onChange(data.filter((p) => p.id !== id));
  };

  const formatPrice = (price, currency) => {
    const symbols = { USD: '$', EUR: '€', GBP: '£', TRY: '₺' };
    return `${symbols[currency] || currency}${price}`;
  };

  return (
    <div>
      <div className="mb-8 text-center">
        <h2 className="text-3xl font-bold text-white mb-2">Your Pricing Plans</h2>
        <p className="text-slate-400">
          Add the pricing plans you currently offer. At least one plan is required.
        </p>
      </div>

      {/* Plan List */}
      {data.length > 0 && (
        <div className="mb-6 space-y-3">
          {data.map((plan) => (
            <div
              key={plan.id}
              className="flex items-center justify-between p-4 rounded-xl bg-slate-900/50 border border-slate-700"
            >
              <div className="flex items-center gap-4">
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
                <div>
                  <p className="text-white font-semibold">{plan.name}</p>
                  <p className="text-slate-400 text-sm">
                    {formatPrice(plan.price, plan.currency)} / {plan.billingCycle}
                  </p>
                </div>
              </div>
              <button
                onClick={() => handleRemovePlan(plan.id)}
                className="p-2 text-slate-400 hover:text-red-400 hover:bg-red-500/10 rounded-lg transition-all"
              >
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                  />
                </svg>
              </button>
            </div>
          ))}
        </div>
      )}

      {/* Add Plan Form */}
      <div className="bg-slate-900/30 rounded-xl p-6 border border-slate-700/50">
        <h3 className="text-lg font-semibold text-white mb-4">Add a Plan</h3>

        <div className="space-y-4">
          {/* Plan Name */}
          <div>
            <label htmlFor="name" className="block text-sm font-semibold text-slate-300 mb-2">
              Plan Name <span className="text-red-400">*</span>
            </label>
            <input
              id="name"
              name="name"
              type="text"
              value={newPlan.name}
              onChange={handleNewPlanChange}
              className="w-full px-4 py-3 rounded-xl bg-slate-900/50 border border-slate-700 text-white placeholder-slate-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all"
              placeholder="e.g. Starter, Pro, Enterprise"
            />
          </div>

          {/* Price and Currency */}
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label htmlFor="price" className="block text-sm font-semibold text-slate-300 mb-2">
                Price <span className="text-red-400">*</span>
              </label>
              <input
                id="price"
                name="price"
                type="number"
                min="0"
                step="0.01"
                value={newPlan.price}
                onChange={handleNewPlanChange}
                className="w-full px-4 py-3 rounded-xl bg-slate-900/50 border border-slate-700 text-white placeholder-slate-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all"
                placeholder="29"
              />
            </div>
            <div>
              <label htmlFor="currency" className="block text-sm font-semibold text-slate-300 mb-2">
                Currency
              </label>
              <select
                id="currency"
                name="currency"
                value={newPlan.currency}
                onChange={handleNewPlanChange}
                className="w-full px-4 py-3 pr-10 rounded-xl bg-slate-900/50 border border-slate-700 text-white focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all appearance-none bg-[url('data:image/svg+xml;charset=UTF-8,%3csvg xmlns=%27http://www.w3.org/2000/svg%27 viewBox=%270 0 24 24%27 fill=%27none%27 stroke=%27rgb(148 163 184)%27 stroke-width=%272%27 stroke-linecap=%27round%27 stroke-linejoin=%27round%27%3e%3cpolyline points=%276 9 12 15 18 9%27%3e%3c/polyline%3e%3c/svg%3e')] bg-[length:1.25rem_1.25rem] bg-[position:right_0.75rem_center] bg-no-repeat"
              >
                {CURRENCIES.map((c) => (
                  <option key={c.value} value={c.value}>
                    {c.label}
                  </option>
                ))}
              </select>
            </div>
          </div>

          {/* Billing Cycle */}
          <div>
            <label htmlFor="billingCycle" className="block text-sm font-semibold text-slate-300 mb-2">
              Billing Cycle
            </label>
            <select
              id="billingCycle"
              name="billingCycle"
              value={newPlan.billingCycle}
              onChange={handleNewPlanChange}
              className="w-full px-4 py-3 pr-10 rounded-xl bg-slate-900/50 border border-slate-700 text-white focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all appearance-none bg-[url('data:image/svg+xml;charset=UTF-8,%3csvg xmlns=%27http://www.w3.org/2000/svg%27 viewBox=%270 0 24 24%27 fill=%27none%27 stroke=%27rgb(148 163 184)%27 stroke-width=%272%27 stroke-linecap=%27round%27 stroke-linejoin=%27round%27%3e%3cpolyline points=%276 9 12 15 18 9%27%3e%3c/polyline%3e%3c/svg%3e')] bg-[length:1.25rem_1.25rem] bg-[position:right_0.75rem_center] bg-no-repeat"
            >
              {BILLING_CYCLES.map((b) => (
                <option key={b.value} value={b.value}>
                  {b.label}
                </option>
              ))}
            </select>
          </div>

          {/* Error */}
          {error && <p className="text-red-400 text-sm">{error}</p>}

          {/* Add Button */}
          <button
            onClick={handleAddPlan}
            className="w-full px-4 py-3 bg-blue-500/20 text-blue-400 border border-blue-500/30 rounded-xl font-semibold hover:bg-blue-500/30 hover:border-blue-500/50 transition-all flex items-center justify-center gap-2"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
            </svg>
            Add Plan
          </button>
        </div>
      </div>

      {/* Requirement Info */}
      {data.length === 0 && (
        <div className="mt-4 bg-amber-500/10 border border-amber-500/30 rounded-xl p-4">
          <div className="flex items-start gap-3">
            <svg
              className="w-5 h-5 text-amber-400 flex-shrink-0 mt-0.5"
              fill="currentColor"
              viewBox="0 0 20 20"
            >
              <path
                fillRule="evenodd"
                d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z"
                clipRule="evenodd"
              />
            </svg>
            <p className="text-sm text-amber-300">
              At least one pricing plan is required to continue.
            </p>
          </div>
        </div>
      )}
    </div>
  );
};

export default Step1Plans;

