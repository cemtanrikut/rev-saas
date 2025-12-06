import { useState } from 'react';
import { useCompetitors } from '../../context/CompetitorsContext';

const Competitors = () => {
  const { competitors, addCompetitor, removeCompetitor, isLoading, error, clearError } = useCompetitors();
  
  const [formData, setFormData] = useState({
    name: '',
    url: ''
  });
  const [formError, setFormError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [removingId, setRemovingId] = useState(null);

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
    if (formError) setFormError('');
    if (error) clearError();
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    // Validation
    if (!formData.url.trim()) {
      setFormError('Pricing page URL is required');
      return;
    }

    // Basic URL validation
    try {
      new URL(formData.url);
    } catch {
      setFormError('Please enter a valid URL (e.g., https://competitor.com/pricing)');
      return;
    }

    setIsSubmitting(true);

    const result = await addCompetitor({
      name: formData.name.trim(),
      url: formData.url.trim()
    });

    setIsSubmitting(false);

    if (result.success) {
      // Reset form
      setFormData({ name: '', url: '' });
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

  return (
    <div className="max-w-5xl mx-auto space-y-8">
      {/* Page Header */}
      <div>
        <h1 className="text-3xl font-bold text-white mb-2">
          Competitors
        </h1>
        <p className="text-slate-400">
          Add competitor pricing pages. Revalyze will use these URLs for benchmarking your SaaS pricing.
        </p>
      </div>

      {/* API Error Banner */}
      {error && (
        <div className="bg-red-500/10 border border-red-500/30 rounded-xl p-4 flex items-center justify-between">
          <p className="text-red-400">{error}</p>
          <button 
            onClick={clearError}
            className="text-red-400 hover:text-red-300 p-1"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      )}

      {/* Add Competitor Form */}
      <div className="bg-slate-900/50 backdrop-blur-sm rounded-2xl p-6 border border-slate-800">
        <h3 className="text-lg font-semibold text-white mb-4">
          Add Competitor
        </h3>
        
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid md:grid-cols-2 gap-4">
            {/* Competitor Name */}
            <div>
              <label htmlFor="name" className="block text-sm font-semibold text-slate-300 mb-2">
                Competitor Name <span className="text-slate-500 font-normal">(optional)</span>
              </label>
              <input
                id="name"
                name="name"
                type="text"
                value={formData.name}
                onChange={handleChange}
                disabled={isSubmitting}
                className="w-full px-4 py-3 rounded-xl bg-slate-900/50 border border-slate-700 text-white placeholder-slate-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all disabled:opacity-50"
                placeholder="e.g., Competitor Inc."
              />
            </div>

            {/* Pricing Page URL */}
            <div>
              <label htmlFor="url" className="block text-sm font-semibold text-slate-300 mb-2">
                Pricing Page URL
              </label>
              <input
                id="url"
                name="url"
                type="url"
                value={formData.url}
                onChange={handleChange}
                disabled={isSubmitting}
                className={`w-full px-4 py-3 rounded-xl bg-slate-900/50 border ${
                  formError ? 'border-red-500' : 'border-slate-700'
                } text-white placeholder-slate-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all disabled:opacity-50`}
                placeholder="https://competitor.com/pricing"
                required
              />
            </div>
          </div>

          {formError && (
            <p className="text-sm text-red-400">{formError}</p>
          )}

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
        <div className="p-6 border-b border-slate-800">
          <h3 className="text-lg font-semibold text-white">
            Your Competitors ({competitors.length})
          </h3>
        </div>

        {isLoading ? (
          /* Loading State */
          <div className="p-12 text-center">
            <div className="w-12 h-12 border-4 border-blue-500 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
            <p className="text-slate-400">Loading competitors...</p>
          </div>
        ) : competitors.length === 0 ? (
          /* Empty State */
          <div className="p-12 text-center">
            <div className="w-16 h-16 bg-slate-800 rounded-2xl flex items-center justify-center mx-auto mb-4">
              <svg className="w-8 h-8 text-slate-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
              </svg>
            </div>
            <p className="text-slate-400 text-lg">
              You haven't added any competitors yet.
            </p>
            <p className="text-slate-500 text-sm mt-2">
              Add your first competitor above to start benchmarking your pricing.
            </p>
          </div>
        ) : (
          /* Competitors Table */
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-slate-800/50">
                <tr>
                  <th className="text-left px-6 py-3 text-xs font-semibold text-slate-400 uppercase tracking-wider">
                    Name
                  </th>
                  <th className="text-left px-6 py-3 text-xs font-semibold text-slate-400 uppercase tracking-wider">
                    Pricing Page URL
                  </th>
                  <th className="text-right px-6 py-3 text-xs font-semibold text-slate-400 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-800">
                {competitors.map((competitor) => (
                  <tr 
                    key={competitor.id} 
                    className={`hover:bg-slate-800/30 transition-colors ${
                      removingId === competitor.id ? 'opacity-50' : ''
                    }`}
                  >
                    <td className="px-6 py-4">
                      <span className="text-slate-300 font-medium">
                        {competitor.name || 'Unnamed'}
                      </span>
                    </td>
                    <td className="px-6 py-4">
                      <a
                        href={competitor.url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-blue-400 hover:text-blue-300 transition-colors flex items-center gap-2"
                      >
                        {competitor.url}
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
                        </svg>
                      </a>
                    </td>
                    <td className="px-6 py-4 text-right">
                      <button
                        onClick={() => handleRemove(competitor.id)}
                        disabled={removingId === competitor.id}
                        className="px-3 py-2 text-sm text-red-400 hover:text-red-300 hover:bg-red-500/10 rounded-lg transition-all disabled:opacity-50"
                      >
                        {removingId === competitor.id ? (
                          <span className="flex items-center gap-2">
                            <svg className="animate-spin h-4 w-4" viewBox="0 0 24 24">
                              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                            </svg>
                            Removing...
                          </span>
                        ) : (
                          'Remove'
                        )}
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
};

export default Competitors;
