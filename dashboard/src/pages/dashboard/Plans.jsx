import { useState } from 'react';
import { usePlans } from '../../context/PlansContext';

const Plans = () => {
  const { plans, addPlan, removePlan } = usePlans();
  
  const [formData, setFormData] = useState({
    name: '',
    price: '',
    interval: 'monthly',
    description: ''
  });
  const [errors, setErrors] = useState({});

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
    // Clear error for this field
    if (errors[name]) {
      setErrors(prev => ({ ...prev, [name]: '' }));
    }
  };

  const validate = () => {
    const newErrors = {};

    if (!formData.name.trim()) {
      newErrors.name = 'Plan name is required';
    }

    if (!formData.price) {
      newErrors.price = 'Price is required';
    } else if (isNaN(formData.price) || Number(formData.price) <= 0) {
      newErrors.price = 'Price must be a positive number';
    }

    if (!formData.interval) {
      newErrors.interval = 'Billing interval is required';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    
    if (!validate()) {
      return;
    }

    // Add plan
    addPlan({
      name: formData.name.trim(),
      price: formData.price,
      interval: formData.interval,
      description: formData.description.trim()
    });

    // Reset form
    setFormData({
      name: '',
      price: '',
      interval: 'monthly',
      description: ''
    });
    setErrors({});
  };

  const handleRemove = (id) => {
    if (confirm('Are you sure you want to remove this plan?')) {
      removePlan(id);
    }
  };

  const formatPrice = (price, interval) => {
    const formattedPrice = new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 0,
      maximumFractionDigits: 2
    }).format(price);

    return `${formattedPrice} / ${interval === 'monthly' ? 'month' : 'year'}`;
  };

  return (
    <div className="max-w-6xl mx-auto space-y-8">
      {/* Page Header */}
      <div>
        <h1 className="text-3xl font-bold text-white mb-2">
          My Pricing
        </h1>
        <p className="text-slate-400">
          Define your current SaaS plans and prices. Revalyze will use these as the baseline for pricing analysis.
        </p>
      </div>

      {/* Add Plan Form */}
      <div className="bg-slate-900/50 backdrop-blur-sm rounded-2xl p-6 border border-slate-800">
        <h3 className="text-lg font-semibold text-white mb-4">
          Add New Plan
        </h3>
        
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid md:grid-cols-2 gap-4">
            {/* Plan Name */}
            <div>
              <label htmlFor="name" className="block text-sm font-semibold text-slate-300 mb-2">
                Plan Name
              </label>
              <input
                id="name"
                name="name"
                type="text"
                value={formData.name}
                onChange={handleChange}
                className={`w-full px-4 py-3 rounded-xl bg-slate-900/50 border ${
                  errors.name ? 'border-red-500' : 'border-slate-700'
                } text-white placeholder-slate-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all`}
                placeholder="e.g., Starter, Pro, Business"
              />
              {errors.name && (
                <p className="text-sm text-red-400 mt-1">{errors.name}</p>
              )}
            </div>

            {/* Price */}
            <div>
              <label htmlFor="price" className="block text-sm font-semibold text-slate-300 mb-2">
                Price (USD)
              </label>
              <input
                id="price"
                name="price"
                type="number"
                step="0.01"
                min="0"
                value={formData.price}
                onChange={handleChange}
                className={`w-full px-4 py-3 rounded-xl bg-slate-900/50 border ${
                  errors.price ? 'border-red-500' : 'border-slate-700'
                } text-white placeholder-slate-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all`}
                placeholder="29.99"
              />
              {errors.price && (
                <p className="text-sm text-red-400 mt-1">{errors.price}</p>
              )}
            </div>
          </div>

          <div className="grid md:grid-cols-2 gap-4">
            {/* Billing Interval */}
            <div>
              <label htmlFor="interval" className="block text-sm font-semibold text-slate-300 mb-2">
                Billing Interval
              </label>
              <select
                id="interval"
                name="interval"
                value={formData.interval}
                onChange={handleChange}
                className="w-full px-4 py-3 rounded-xl bg-slate-900/50 border border-slate-700 text-white focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all appearance-none pr-10"
                style={{
                  backgroundImage: `url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 24 24' stroke='%23cbd5e1'%3E%3Cpath stroke-linecap='round' stroke-linejoin='round' stroke-width='2' d='M19 9l-7 7-7-7'%3E%3C/path%3E%3C/svg%3E")`,
                  backgroundRepeat: 'no-repeat',
                  backgroundPosition: 'right 0.75rem center',
                  backgroundSize: '1.5rem'
                }}
              >
                <option value="monthly">Monthly</option>
                <option value="yearly">Yearly</option>
              </select>
            </div>

            {/* Description (optional) */}
            <div>
              <label htmlFor="description" className="block text-sm font-semibold text-slate-300 mb-2">
                Description <span className="text-slate-500 font-normal">(optional)</span>
              </label>
              <input
                id="description"
                name="description"
                type="text"
                value={formData.description}
                onChange={handleChange}
                className="w-full px-4 py-3 rounded-xl bg-slate-900/50 border border-slate-700 text-white placeholder-slate-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all"
                placeholder="Brief description of what's included"
              />
            </div>
          </div>

          <button
            type="submit"
            className="px-6 py-3 bg-gradient-to-r from-blue-500 to-indigo-600 text-white rounded-xl font-semibold hover:from-blue-600 hover:to-indigo-700 hover:scale-105 transition-all shadow-lg shadow-blue-500/20"
          >
            Add Plan
          </button>
        </form>
      </div>

      {/* Plans List */}
      <div className="bg-slate-900/50 backdrop-blur-sm rounded-2xl border border-slate-800 overflow-hidden">
        <div className="p-6 border-b border-slate-800">
          <h3 className="text-lg font-semibold text-white">
            Your Plans ({plans.length})
          </h3>
        </div>

        {plans.length === 0 ? (
          /* Empty State */
          <div className="p-12 text-center">
            <div className="w-16 h-16 bg-slate-800 rounded-2xl flex items-center justify-center mx-auto mb-4">
              <svg className="w-8 h-8 text-slate-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 14l6-6m-5.5.5h.01m4.99 5h.01M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16l3.5-2 3.5 2 3.5-2 3.5 2zM10 8.5a.5.5 0 11-1 0 .5.5 0 011 0zm5 5a.5.5 0 11-1 0 .5.5 0 011 0z" />
              </svg>
            </div>
            <p className="text-slate-400 text-lg">
              You haven't defined any plans yet.
            </p>
            <p className="text-slate-500 text-sm mt-2">
              Add at least one plan to get meaningful pricing suggestions.
            </p>
          </div>
        ) : (
          /* Plans Grid */
          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-4 p-6">
            {plans.map((plan) => (
              <div
                key={plan.id}
                className="bg-slate-800/30 rounded-xl p-6 border border-slate-700 hover:border-slate-600 transition-all group"
              >
                <div className="flex items-start justify-between mb-4">
                  <div>
                    <h4 className="text-xl font-bold text-white mb-1">
                      {plan.name}
                    </h4>
                    <div className="text-2xl font-bold text-blue-400">
                      {formatPrice(plan.price, plan.interval)}
                    </div>
                  </div>
                  <button
                    onClick={() => handleRemove(plan.id)}
                    className="opacity-0 group-hover:opacity-100 p-2 text-slate-400 hover:text-red-400 hover:bg-red-500/10 rounded-lg transition-all"
                    title="Remove plan"
                  >
                    <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                  </button>
                </div>

                {plan.description && (
                  <p className="text-sm text-slate-400 leading-relaxed">
                    {plan.description}
                  </p>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default Plans;


