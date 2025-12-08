import { useState } from 'react';

const CURRENCIES = [
  { value: 'USD', label: 'USD' },
  { value: 'EUR', label: 'EUR' },
  { value: 'GBP', label: 'GBP' },
  { value: 'TRY', label: 'TRY' },
];

const BILLING_CYCLES = [
  { value: 'monthly', label: 'Monthly' },
  { value: 'yearly', label: 'Yearly' },
];

const Step2Competitors = ({ data, onChange }) => {
  const [editingCompetitor, setEditingCompetitor] = useState(null);
  const [newCompetitorName, setNewCompetitorName] = useState('');
  const [newCompetitorUrl, setNewCompetitorUrl] = useState('');
  const [newPlan, setNewPlan] = useState({
    name: '',
    price: '',
    currency: 'USD',
    billingCycle: 'monthly',
  });
  const [error, setError] = useState('');

  // Add a new competitor (empty, to be filled with plans)
  const handleAddCompetitor = () => {
    if (!newCompetitorName.trim()) {
      setError('Competitor name is required');
      return;
    }

    const competitor = {
      id: Date.now(),
      name: newCompetitorName.trim(),
      url: newCompetitorUrl.trim(),
      plans: [],
    };

    onChange([...data, competitor]);
    setNewCompetitorName('');
    setNewCompetitorUrl('');
    setError('');
    setEditingCompetitor(competitor.id);
  };

  // Remove a competitor
  const handleRemoveCompetitor = (id) => {
    onChange(data.filter((c) => c.id !== id));
    if (editingCompetitor === id) {
      setEditingCompetitor(null);
    }
  };

  // Add a plan to a competitor
  const handleAddPlan = (competitorId) => {
    if (!newPlan.name.trim()) {
      setError('Plan name is required');
      return;
    }
    if (!newPlan.price || parseFloat(newPlan.price) < 0) {
      setError('Price must be 0 or greater');
      return;
    }

    const plan = {
      id: Date.now(),
      name: newPlan.name.trim(),
      price: parseFloat(newPlan.price),
      currency: newPlan.currency,
      billingCycle: newPlan.billingCycle,
    };

    onChange(
      data.map((c) =>
        c.id === competitorId ? { ...c, plans: [...c.plans, plan] } : c
      )
    );

    setNewPlan({ name: '', price: '', currency: 'USD', billingCycle: 'monthly' });
    setError('');
  };

  // Remove a plan from a competitor
  const handleRemovePlan = (competitorId, planId) => {
    onChange(
      data.map((c) =>
        c.id === competitorId
          ? { ...c, plans: c.plans.filter((p) => p.id !== planId) }
          : c
      )
    );
  };

  const formatPrice = (price, currency) => {
    const symbols = { USD: '$', EUR: '€', GBP: '£', TRY: '₺' };
    return `${symbols[currency] || currency}${price}`;
  };

  return (
    <div>
      <div className="mb-8 text-center">
        <h2 className="text-3xl font-bold text-white mb-2">Competitors</h2>
        <p className="text-slate-400 max-w-2xl mx-auto">
          Add your competitors and their pricing plans. This step is optional.
        </p>
      </div>

      {/* Competitor List */}
      {data.length > 0 && (
        <div className="mb-6 space-y-4">
          {data.map((competitor) => (
            <div
              key={competitor.id}
              className="rounded-xl bg-slate-900/50 border border-slate-700 overflow-hidden"
            >
              {/* Competitor Header */}
              <div className="flex items-center justify-between p-4 bg-slate-800/50">
                <div className="flex items-center gap-3">
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
                  <div>
                    <p className="text-white font-semibold">{competitor.name}</p>
                    <p className="text-slate-500 text-sm">
                      {competitor.plans.length} plan(s)
                    </p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <button
                    onClick={() =>
                      setEditingCompetitor(
                        editingCompetitor === competitor.id ? null : competitor.id
                      )
                    }
                    className="p-2 text-slate-400 hover:text-blue-400 hover:bg-blue-500/10 rounded-lg transition-all"
                  >
                    <svg
                      className="w-5 h-5"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M12 4v16m8-8H4"
                      />
                    </svg>
                  </button>
                  <button
                    onClick={() => handleRemoveCompetitor(competitor.id)}
                    className="p-2 text-slate-400 hover:text-red-400 hover:bg-red-500/10 rounded-lg transition-all"
                  >
                    <svg
                      className="w-5 h-5"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                      />
                    </svg>
                  </button>
                </div>
              </div>

              {/* Plans List */}
              {competitor.plans.length > 0 && (
                <div className="p-4 pt-0">
                  <div className="mt-3 space-y-2">
                    {competitor.plans.map((plan) => (
                      <div
                        key={plan.id}
                        className="flex items-center justify-between p-3 rounded-lg bg-slate-800/30 border border-slate-700/50"
                      >
                        <div>
                          <span className="text-white">{plan.name}</span>
                          <span className="text-slate-400 ml-2">
                            {formatPrice(plan.price, plan.currency)} /{' '}
                            {plan.billingCycle}
                          </span>
                        </div>
                        <button
                          onClick={() => handleRemovePlan(competitor.id, plan.id)}
                          className="p-1 text-slate-500 hover:text-red-400 transition-colors"
                        >
                          <svg
                            className="w-4 h-4"
                            fill="none"
                            stroke="currentColor"
                            viewBox="0 0 24 24"
                          >
                            <path
                              strokeLinecap="round"
                              strokeLinejoin="round"
                              strokeWidth={2}
                              d="M6 18L18 6M6 6l12 12"
                            />
                          </svg>
                        </button>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* Add Plan Form (expanded) */}
              {editingCompetitor === competitor.id && (
                <div className="p-4 border-t border-slate-700/50 bg-slate-800/20">
                  <p className="text-sm font-semibold text-slate-300 mb-3">
                    Add a plan to {competitor.name}
                  </p>
                  <div className="grid grid-cols-2 gap-3 mb-3">
                    <input
                      type="text"
                      value={newPlan.name}
                      onChange={(e) =>
                        setNewPlan({ ...newPlan, name: e.target.value })
                      }
                      className="px-3 py-2 rounded-lg bg-slate-900/50 border border-slate-700 text-white placeholder-slate-500 focus:border-blue-500 outline-none text-sm"
                      placeholder="Plan name (e.g., Pro)"
                    />
                    <div className="flex gap-2">
                      <select
                        value={newPlan.currency}
                        onChange={(e) =>
                          setNewPlan({ ...newPlan, currency: e.target.value })
                        }
                        className="w-20 px-2 py-2 rounded-lg bg-slate-900/50 border border-slate-700 text-white text-sm outline-none"
                      >
                        {CURRENCIES.map((c) => (
                          <option key={c.value} value={c.value}>
                            {c.label}
                          </option>
                        ))}
                      </select>
                      <input
                        type="number"
                        min="0"
                        step="0.01"
                        value={newPlan.price}
                        onChange={(e) =>
                          setNewPlan({ ...newPlan, price: e.target.value })
                        }
                        className="flex-1 px-3 py-2 rounded-lg bg-slate-900/50 border border-slate-700 text-white placeholder-slate-500 focus:border-blue-500 outline-none text-sm"
                        placeholder="Price"
                      />
                    </div>
                  </div>
                  <div className="flex gap-3 items-center">
                    <select
                      value={newPlan.billingCycle}
                      onChange={(e) =>
                        setNewPlan({ ...newPlan, billingCycle: e.target.value })
                      }
                      className="px-3 py-2 rounded-lg bg-slate-900/50 border border-slate-700 text-white text-sm outline-none"
                    >
                      {BILLING_CYCLES.map((b) => (
                        <option key={b.value} value={b.value}>
                          {b.label}
                        </option>
                      ))}
                    </select>
                    <button
                      onClick={() => handleAddPlan(competitor.id)}
                      className="px-4 py-2 bg-purple-500/20 text-purple-400 border border-purple-500/30 rounded-lg text-sm font-medium hover:bg-purple-500/30 transition-all"
                    >
                      Add Plan
                    </button>
                  </div>
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      {/* Add Competitor Form */}
      <div className="bg-slate-900/30 rounded-xl p-6 border border-slate-700/50">
        <h3 className="text-lg font-semibold text-white mb-4">Add a Competitor</h3>

        <div className="space-y-4">
          <div className="grid md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-semibold text-slate-300 mb-2">
                Competitor Name <span className="text-red-400">*</span>
              </label>
              <input
                type="text"
                value={newCompetitorName}
                onChange={(e) => {
                  setNewCompetitorName(e.target.value);
                  setError('');
                }}
                className="w-full px-4 py-3 rounded-xl bg-slate-900/50 border border-slate-700 text-white placeholder-slate-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all"
                placeholder="e.g., Notion, Slack"
              />
            </div>
            <div>
              <label className="block text-sm font-semibold text-slate-300 mb-2">
                Website <span className="text-slate-500 font-normal">(optional)</span>
              </label>
              <input
                type="url"
                value={newCompetitorUrl}
                onChange={(e) => setNewCompetitorUrl(e.target.value)}
                className="w-full px-4 py-3 rounded-xl bg-slate-900/50 border border-slate-700 text-white placeholder-slate-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all"
                placeholder="https://competitor.com"
              />
            </div>
          </div>

          {error && <p className="text-red-400 text-sm">{error}</p>}

          <button
            onClick={handleAddCompetitor}
            className="px-6 py-3 bg-purple-500/20 text-purple-400 border border-purple-500/30 rounded-xl font-semibold hover:bg-purple-500/30 hover:border-purple-500/50 transition-all flex items-center gap-2"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 4v16m8-8H4"
              />
            </svg>
            Add Competitor
          </button>
        </div>
      </div>

      {/* Info box */}
      <div className="mt-4 bg-blue-500/10 border border-blue-500/30 rounded-xl p-4">
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
              <span className="font-semibold">Tip:</span> Add each competitor once, then add their pricing plans. You can skip this step and add competitors later.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Step2Competitors;
