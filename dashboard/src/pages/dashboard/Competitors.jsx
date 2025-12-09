import { useState } from 'react';
import { useCompetitors } from '../../context/CompetitorsContext';

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

const Competitors = () => {
  const { competitors, addCompetitor, updateCompetitor, removeCompetitor, isLoading, error, clearError } = useCompetitors();
  
  // Add form state
  const [formData, setFormData] = useState({ name: '', url: '' });
  const [plans, setPlans] = useState([]);
  const [newPlan, setNewPlan] = useState({
    name: '',
    price: '',
    currency: 'USD',
    billingCycle: 'monthly',
  });
  const [formError, setFormError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [removingId, setRemovingId] = useState(null);
  const [expandedId, setExpandedId] = useState(null);

  // Edit state
  const [editingId, setEditingId] = useState(null);
  const [editData, setEditData] = useState({ name: '', url: '', plans: [] });
  const [editNewPlan, setEditNewPlan] = useState({
    name: '',
    price: '',
    currency: 'USD',
    billingCycle: 'monthly',
  });
  const [editError, setEditError] = useState('');
  const [isUpdating, setIsUpdating] = useState(false);

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
    if (formError) setFormError('');
    if (error) clearError();
  };

  const handleAddPlanToForm = () => {
    if (!newPlan.name.trim()) {
      setFormError('Plan name is required');
      return;
    }
    if (!newPlan.price || parseFloat(newPlan.price) < 0) {
      setFormError('Price must be 0 or greater');
      return;
    }

    setPlans([...plans, {
      id: Date.now(),
      name: newPlan.name.trim(),
      price: parseFloat(newPlan.price),
      currency: newPlan.currency,
      billingCycle: newPlan.billingCycle,
    }]);

    setNewPlan({ name: '', price: '', currency: 'USD', billingCycle: 'monthly' });
    setFormError('');
  };

  const handleRemovePlanFromForm = (planId) => {
    setPlans(plans.filter(p => p.id !== planId));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    if (!formData.name.trim()) {
      setFormError('Competitor name is required');
      return;
    }

    if (plans.length === 0) {
      setFormError('Add at least one pricing plan for this competitor');
      return;
    }

    setIsSubmitting(true);

    const result = await addCompetitor({
      name: formData.name.trim(),
      url: formData.url.trim(),
      plans: plans.map(p => ({
        name: p.name,
        price: p.price,
        currency: p.currency,
        billingCycle: p.billingCycle,
      })),
    });

    setIsSubmitting(false);

    if (result.success) {
      setFormData({ name: '', url: '' });
      setPlans([]);
      setFormError('');
    } else {
      setFormError(result.error || 'Failed to add competitor');
    }
  };

  const handleRemove = async (id) => {
    if (!confirm('Are you sure you want to remove this competitor?')) {
      return;
    }

    setRemovingId(id);
    await removeCompetitor(id);
    setRemovingId(null);
  };

  // Edit handlers
  const startEdit = (competitor) => {
    setEditingId(competitor.id);
    setEditData({
      name: competitor.name,
      url: competitor.url || '',
      plans: competitor.plans?.map(p => ({
        id: Date.now() + Math.random(),
        name: p.name,
        price: p.price,
        currency: p.currency,
        billingCycle: p.billingCycle,
      })) || [],
    });
    setEditError('');
    setExpandedId(competitor.id);
  };

  const cancelEdit = () => {
    setEditingId(null);
    setEditData({ name: '', url: '', plans: [] });
    setEditError('');
  };

  const handleEditChange = (e) => {
    const { name, value } = e.target;
    setEditData(prev => ({ ...prev, [name]: value }));
    if (editError) setEditError('');
  };

  const handleAddPlanToEdit = () => {
    if (!editNewPlan.name.trim()) {
      setEditError('Plan name is required');
      return;
    }
    if (!editNewPlan.price || parseFloat(editNewPlan.price) < 0) {
      setEditError('Price must be 0 or greater');
      return;
    }

    setEditData(prev => ({
      ...prev,
      plans: [...prev.plans, {
        id: Date.now(),
        name: editNewPlan.name.trim(),
        price: parseFloat(editNewPlan.price),
        currency: editNewPlan.currency,
        billingCycle: editNewPlan.billingCycle,
      }],
    }));

    setEditNewPlan({ name: '', price: '', currency: 'USD', billingCycle: 'monthly' });
    setEditError('');
  };

  const handleRemovePlanFromEdit = (planId) => {
    setEditData(prev => ({
      ...prev,
      plans: prev.plans.filter(p => p.id !== planId),
    }));
  };

  const handleSaveEdit = async () => {
    if (!editData.name.trim()) {
      setEditError('Competitor name is required');
      return;
    }

    if (editData.plans.length === 0) {
      setEditError('Add at least one pricing plan');
      return;
    }

    setIsUpdating(true);

    const result = await updateCompetitor(editingId, {
      name: editData.name.trim(),
      url: editData.url.trim(),
      plans: editData.plans.map(p => ({
        name: p.name,
        price: p.price,
        currency: p.currency,
        billingCycle: p.billingCycle,
      })),
    });

    setIsUpdating(false);

    if (result.success) {
      setEditingId(null);
      setEditData({ name: '', url: '', plans: [] });
      setEditError('');
    } else {
      setEditError(result.error || 'Failed to update competitor');
    }
  };

  const formatPrice = (price, currency) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: currency || 'USD',
      minimumFractionDigits: 0,
      maximumFractionDigits: 2,
    }).format(price);
  };

  const getTotalPlansCount = () => {
    return competitors.reduce((acc, c) => acc + (c.plans?.length || 0), 0);
  };

  return (
    <div className="max-w-5xl mx-auto space-y-8">
      {/* Page Header */}
      <div>
        <h1 className="text-3xl font-bold text-white mb-2">Competitors</h1>
        <p className="text-slate-400">
          Add competitors and their pricing plans. Revalyze uses this data to benchmark your SaaS pricing.
        </p>
      </div>

      {/* API Error Banner */}
      {error && (
        <div className="bg-red-500/10 border border-red-500/30 rounded-xl p-4 flex items-center justify-between">
          <p className="text-red-400">{error}</p>
          <button onClick={clearError} className="text-red-400 hover:text-red-300 p-1">
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      )}

      {/* Add Competitor Form */}
      <div className="bg-slate-900/50 backdrop-blur-sm rounded-2xl p-6 border border-slate-800">
        <h3 className="text-lg font-semibold text-white mb-4">Add Competitor</h3>

        <form onSubmit={handleSubmit} className="space-y-5">
          {/* Competitor Info */}
          <div className="grid md:grid-cols-2 gap-4">
            <div>
              <label htmlFor="name" className="block text-sm font-semibold text-slate-300 mb-2">
                Competitor Name <span className="text-red-400">*</span>
              </label>
              <input
                id="name"
                name="name"
                type="text"
                value={formData.name}
                onChange={handleChange}
                disabled={isSubmitting}
                className="w-full px-4 py-3 rounded-xl bg-slate-900/50 border border-slate-700 text-white placeholder-slate-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all disabled:opacity-50"
                placeholder="e.g., Notion, Slack"
              />
            </div>
            <div>
              <label htmlFor="url" className="block text-sm font-semibold text-slate-300 mb-2">
                Website <span className="text-slate-500 font-normal">(optional)</span>
              </label>
              <input
                id="url"
                name="url"
                type="url"
                value={formData.url}
                onChange={handleChange}
                disabled={isSubmitting}
                className="w-full px-4 py-3 rounded-xl bg-slate-900/50 border border-slate-700 text-white placeholder-slate-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all disabled:opacity-50"
                placeholder="https://competitor.com"
              />
            </div>
          </div>

          {/* Plans Section */}
          <div className="border-t border-slate-700 pt-5">
            <h4 className="text-sm font-semibold text-slate-300 mb-3">
              Pricing Plans <span className="text-red-400">*</span>
            </h4>

            {/* Plans List */}
            {plans.length > 0 && (
              <div className="mb-4 space-y-2">
                {plans.map((plan) => (
                  <div
                    key={plan.id}
                    className="flex items-center justify-between p-3 rounded-lg bg-slate-800/50 border border-slate-700"
                  >
                    <div>
                      <span className="text-white font-medium">{plan.name}</span>
                      <span className="text-slate-400 ml-2">
                        {formatPrice(plan.price, plan.currency)} / {plan.billingCycle}
                      </span>
                    </div>
                    <button
                      type="button"
                      onClick={() => handleRemovePlanFromForm(plan.id)}
                      className="p-1 text-slate-500 hover:text-red-400 transition-colors"
                    >
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                      </svg>
                    </button>
                  </div>
                ))}
              </div>
            )}

            {/* Add Plan Inputs */}
            <div className="flex flex-wrap gap-3 items-end">
              <div className="flex-1 min-w-[150px]">
                <input
                  type="text"
                  value={newPlan.name}
                  onChange={(e) => setNewPlan({ ...newPlan, name: e.target.value })}
                  disabled={isSubmitting}
                  className="w-full px-3 py-2 rounded-lg bg-slate-900/50 border border-slate-700 text-white placeholder-slate-500 focus:border-blue-500 outline-none text-sm disabled:opacity-50"
                  placeholder="Plan name (e.g., Pro)"
                />
              </div>
              <div className="flex gap-2">
                <select
                  value={newPlan.currency}
                  onChange={(e) => setNewPlan({ ...newPlan, currency: e.target.value })}
                  disabled={isSubmitting}
                  className="w-20 px-2 py-2 rounded-lg bg-slate-900/50 border border-slate-700 text-white text-sm outline-none disabled:opacity-50"
                >
                  {CURRENCIES.map((c) => (
                    <option key={c.value} value={c.value}>{c.label}</option>
                  ))}
                </select>
                <input
                  type="number"
                  min="0"
                  step="0.01"
                  value={newPlan.price}
                  onChange={(e) => setNewPlan({ ...newPlan, price: e.target.value })}
                  disabled={isSubmitting}
                  className="w-24 px-3 py-2 rounded-lg bg-slate-900/50 border border-slate-700 text-white placeholder-slate-500 focus:border-blue-500 outline-none text-sm disabled:opacity-50"
                  placeholder="Price"
                />
              </div>
              <select
                value={newPlan.billingCycle}
                onChange={(e) => setNewPlan({ ...newPlan, billingCycle: e.target.value })}
                disabled={isSubmitting}
                className="px-3 py-2 rounded-lg bg-slate-900/50 border border-slate-700 text-white text-sm outline-none disabled:opacity-50"
              >
                {BILLING_CYCLES.map((b) => (
                  <option key={b.value} value={b.value}>{b.label}</option>
                ))}
              </select>
              <button
                type="button"
                onClick={handleAddPlanToForm}
                disabled={isSubmitting}
                className="px-4 py-2 bg-slate-700 text-white rounded-lg text-sm font-medium hover:bg-slate-600 transition-all disabled:opacity-50"
              >
                + Add Plan
              </button>
            </div>
          </div>

          {formError && <p className="text-sm text-red-400">{formError}</p>}

          <button
            type="submit"
            disabled={isSubmitting}
            className="px-6 py-3 bg-gradient-to-r from-blue-500 to-indigo-600 text-white rounded-xl font-semibold hover:from-blue-600 hover:to-indigo-700 hover:scale-105 transition-all shadow-lg shadow-blue-500/20 disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:scale-100"
          >
            {isSubmitting ? (
              <span className="flex items-center gap-2">
                <svg className="animate-spin h-5 w-5" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                </svg>
                Adding...
              </span>
            ) : (
              'Add Competitor'
            )}
          </button>
        </form>
      </div>

      {/* Competitors List */}
      <div className="bg-slate-900/50 backdrop-blur-sm rounded-2xl border border-slate-800 overflow-hidden">
        <div className="p-6 border-b border-slate-800 flex items-center justify-between">
          <h3 className="text-lg font-semibold text-white">
            Your Competitors ({competitors.length})
          </h3>
          <span className="text-sm text-slate-400">
            {getTotalPlansCount()} pricing plans total
          </span>
        </div>

        {isLoading ? (
          <div className="p-12 text-center">
            <div className="w-12 h-12 border-4 border-blue-500 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
            <p className="text-slate-400">Loading competitors...</p>
          </div>
        ) : competitors.length === 0 ? (
          <div className="p-12 text-center">
            <div className="w-16 h-16 bg-slate-800 rounded-2xl flex items-center justify-center mx-auto mb-4">
              <svg className="w-8 h-8 text-slate-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
              </svg>
            </div>
            <p className="text-slate-400 text-lg">You haven't added any competitors yet.</p>
            <p className="text-slate-500 text-sm mt-2">
              Add your first competitor above to start benchmarking your pricing.
            </p>
          </div>
        ) : (
          <div className="divide-y divide-slate-800">
            {competitors.map((competitor) => (
              <div key={competitor.id} className={`${removingId === competitor.id ? 'opacity-50' : ''}`}>
                {/* View Mode */}
                {editingId !== competitor.id ? (
                  <>
                    {/* Competitor Header */}
                    <div
                      className="flex items-center justify-between p-4 hover:bg-slate-800/30 transition-colors cursor-pointer"
                      onClick={() => setExpandedId(expandedId === competitor.id ? null : competitor.id)}
                    >
                      <div className="flex items-center gap-4">
                        <div className="w-10 h-10 rounded-lg bg-purple-500/20 flex items-center justify-center">
                          <svg className="w-5 h-5 text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
                          </svg>
                        </div>
                        <div>
                          <p className="text-white font-semibold">{competitor.name || 'Unnamed'}</p>
                          <p className="text-slate-500 text-sm">
                            {competitor.plans?.length || 0} plan(s)
                            {competitor.url && (
                              <>
                                {' • '}
                                <a
                                  href={competitor.url}
                                  target="_blank"
                                  rel="noopener noreferrer"
                                  className="text-blue-400 hover:text-blue-300"
                                  onClick={(e) => e.stopPropagation()}
                                >
                                  {new URL(competitor.url).hostname}
                                </a>
                              </>
                            )}
                          </p>
                        </div>
                      </div>
                      <div className="flex items-center gap-2">
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            startEdit(competitor);
                          }}
                          className="p-2 text-slate-400 hover:text-blue-400 hover:bg-blue-500/10 rounded-lg transition-all"
                          title="Edit"
                        >
                          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                          </svg>
                        </button>
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            handleRemove(competitor.id);
                          }}
                          disabled={removingId === competitor.id}
                          className="p-2 text-slate-400 hover:text-red-400 hover:bg-red-500/10 rounded-lg transition-all disabled:opacity-50"
                          title="Delete"
                        >
                          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                          </svg>
                        </button>
                        <svg
                          className={`w-5 h-5 text-slate-400 transition-transform ${expandedId === competitor.id ? 'rotate-180' : ''}`}
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                        </svg>
                      </div>
                    </div>

                    {/* Plans List (expandable) */}
                    {expandedId === competitor.id && competitor.plans?.length > 0 && (
                      <div className="px-4 pb-4">
                        <div className="ml-14 space-y-2">
                          {competitor.plans.map((plan, idx) => (
                            <div
                              key={idx}
                              className="flex items-center justify-between p-3 rounded-lg bg-slate-800/30 border border-slate-700/50"
                            >
                              <span className="text-slate-300">{plan.name}</span>
                              <span className="text-slate-400">
                                {formatPrice(plan.price, plan.currency)} / {plan.billingCycle}
                              </span>
                            </div>
                          ))}
                        </div>
                      </div>
                    )}
                  </>
                ) : (
                  /* Edit Mode */
                  <div className="p-4 bg-slate-800/30">
                    <div className="space-y-4">
                      {/* Edit Header */}
                      <div className="flex items-center justify-between">
                        <h4 className="text-lg font-semibold text-white">Edit Competitor</h4>
                        <div className="flex gap-2">
                          <button
                            onClick={cancelEdit}
                            disabled={isUpdating}
                            className="px-4 py-2 text-slate-400 hover:text-white border border-slate-600 rounded-lg text-sm font-medium hover:bg-slate-700 transition-all disabled:opacity-50"
                          >
                            Cancel
                          </button>
                          <button
                            onClick={handleSaveEdit}
                            disabled={isUpdating}
                            className="px-4 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700 transition-all disabled:opacity-50 flex items-center gap-2"
                          >
                            {isUpdating ? (
                              <>
                                <svg className="animate-spin h-4 w-4" viewBox="0 0 24 24">
                                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                                </svg>
                                Saving...
                              </>
                            ) : (
                              'Save Changes'
                            )}
                          </button>
                        </div>
                      </div>

                      {/* Edit Form */}
                      <div className="grid md:grid-cols-2 gap-4">
                        <div>
                          <label className="block text-sm font-semibold text-slate-300 mb-2">
                            Competitor Name <span className="text-red-400">*</span>
                          </label>
                          <input
                            type="text"
                            name="name"
                            value={editData.name}
                            onChange={handleEditChange}
                            disabled={isUpdating}
                            className="w-full px-4 py-3 rounded-xl bg-slate-900/50 border border-slate-700 text-white placeholder-slate-500 focus:border-blue-500 outline-none disabled:opacity-50"
                            placeholder="e.g., Notion"
                          />
                        </div>
                        <div>
                          <label className="block text-sm font-semibold text-slate-300 mb-2">
                            Website
                          </label>
                          <input
                            type="url"
                            name="url"
                            value={editData.url}
                            onChange={handleEditChange}
                            disabled={isUpdating}
                            className="w-full px-4 py-3 rounded-xl bg-slate-900/50 border border-slate-700 text-white placeholder-slate-500 focus:border-blue-500 outline-none disabled:opacity-50"
                            placeholder="https://competitor.com"
                          />
                        </div>
                      </div>

                      {/* Plans Section */}
                      <div className="border-t border-slate-700 pt-4">
                        <h5 className="text-sm font-semibold text-slate-300 mb-3">Pricing Plans</h5>

                        {/* Existing Plans */}
                        {editData.plans.length > 0 && (
                          <div className="mb-4 space-y-2">
                            {editData.plans.map((plan) => (
                              <div
                                key={plan.id}
                                className="flex items-center justify-between p-3 rounded-lg bg-slate-800/50 border border-slate-700"
                              >
                                <div>
                                  <span className="text-white font-medium">{plan.name}</span>
                                  <span className="text-slate-400 ml-2">
                                    {formatPrice(plan.price, plan.currency)} / {plan.billingCycle}
                                  </span>
                                </div>
                                <button
                                  type="button"
                                  onClick={() => handleRemovePlanFromEdit(plan.id)}
                                  disabled={isUpdating}
                                  className="p-1 text-slate-500 hover:text-red-400 transition-colors disabled:opacity-50"
                                >
                                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                                  </svg>
                                </button>
                              </div>
                            ))}
                          </div>
                        )}

                        {/* Add Plan Inputs */}
                        <div className="flex flex-wrap gap-3 items-end">
                          <div className="flex-1 min-w-[150px]">
                            <input
                              type="text"
                              value={editNewPlan.name}
                              onChange={(e) => setEditNewPlan({ ...editNewPlan, name: e.target.value })}
                              disabled={isUpdating}
                              className="w-full px-3 py-2 rounded-lg bg-slate-900/50 border border-slate-700 text-white placeholder-slate-500 focus:border-blue-500 outline-none text-sm disabled:opacity-50"
                              placeholder="Plan name"
                            />
                          </div>
                          <div className="flex gap-2">
                            <select
                              value={editNewPlan.currency}
                              onChange={(e) => setEditNewPlan({ ...editNewPlan, currency: e.target.value })}
                              disabled={isUpdating}
                              className="w-20 px-2 py-2 rounded-lg bg-slate-900/50 border border-slate-700 text-white text-sm outline-none disabled:opacity-50"
                            >
                              {CURRENCIES.map((c) => (
                                <option key={c.value} value={c.value}>{c.label}</option>
                              ))}
                            </select>
                            <input
                              type="number"
                              min="0"
                              step="0.01"
                              value={editNewPlan.price}
                              onChange={(e) => setEditNewPlan({ ...editNewPlan, price: e.target.value })}
                              disabled={isUpdating}
                              className="w-24 px-3 py-2 rounded-lg bg-slate-900/50 border border-slate-700 text-white placeholder-slate-500 focus:border-blue-500 outline-none text-sm disabled:opacity-50"
                              placeholder="Price"
                            />
                          </div>
                          <select
                            value={editNewPlan.billingCycle}
                            onChange={(e) => setEditNewPlan({ ...editNewPlan, billingCycle: e.target.value })}
                            disabled={isUpdating}
                            className="px-3 py-2 rounded-lg bg-slate-900/50 border border-slate-700 text-white text-sm outline-none disabled:opacity-50"
                          >
                            {BILLING_CYCLES.map((b) => (
                              <option key={b.value} value={b.value}>{b.label}</option>
                            ))}
                          </select>
                          <button
                            type="button"
                            onClick={handleAddPlanToEdit}
                            disabled={isUpdating}
                            className="px-4 py-2 bg-slate-700 text-white rounded-lg text-sm font-medium hover:bg-slate-600 transition-all disabled:opacity-50"
                          >
                            + Add Plan
                          </button>
                        </div>
                      </div>

                      {editError && <p className="text-sm text-red-400">{editError}</p>}
                    </div>
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default Competitors;
