import { useState, useEffect } from 'react';
import { useSettings } from '../../context/SettingsContext';
import { usePlans } from '../../context/PlansContext';
import { useCompetitors } from '../../context/CompetitorsContext';
import { useAnalysis } from '../../context/AnalysisContext';
import { useBusinessMetrics } from '../../context/BusinessMetricsContext';

const Settings = () => {
  const { profile, updateProfile } = useSettings();
  const { clearPlans } = usePlans();
  const { clearCompetitors } = useCompetitors();
  const { clearAnalyses } = useAnalysis();
  const { metrics, saveMetrics, isSaving, error: metricsError, clearError } = useBusinessMetrics();

  const [showSavedMessage, setShowSavedMessage] = useState(false);
  const [metricsForm, setMetricsForm] = useState({
    currency: 'USD',
    mrr: '',
    customers: '',
    monthly_churn_rate: ''
  });
  const [metricsSaved, setMetricsSaved] = useState(false);

  // Initialize form with existing metrics
  useEffect(() => {
    if (metrics) {
      setMetricsForm({
        currency: metrics.currency || 'USD',
        mrr: metrics.mrr?.toString() || '',
        customers: metrics.customers?.toString() || '',
        monthly_churn_rate: metrics.monthly_churn_rate?.toString() || ''
      });
    }
  }, [metrics]);

  const handleInputChange = (field, value) => {
    updateProfile({ [field]: value });
    
    // Show "saved" message briefly
    setShowSavedMessage(true);
    setTimeout(() => {
      setShowSavedMessage(false);
    }, 2000);
  };

  const handleMetricsChange = (field, value) => {
    setMetricsForm(prev => ({
      ...prev,
      [field]: value
    }));
    if (metricsError) clearError();
    if (metricsSaved) setMetricsSaved(false);
  };

  const handleMetricsSubmit = async (e) => {
    e.preventDefault();
    
    const result = await saveMetrics({
      currency: metricsForm.currency,
      mrr: parseFloat(metricsForm.mrr) || 0,
      customers: parseInt(metricsForm.customers) || 0,
      monthly_churn_rate: parseFloat(metricsForm.monthly_churn_rate) || 0
    });

    if (result.success) {
      setMetricsSaved(true);
      setTimeout(() => setMetricsSaved(false), 3000);
    }
  };

  const handleResetDemoData = () => {
    const confirmed = window.confirm(
      'This will clear all local plans, competitors, and analyses. Are you sure?'
    );

    if (confirmed) {
      clearPlans();
      clearCompetitors();
      clearAnalyses();
      
      // Show success message or feedback
      alert('Demo data has been reset successfully.');
    }
  };

  return (
    <div className="max-w-5xl mx-auto space-y-8">
      {/* Page Header */}
      <div>
        <h1 className="text-3xl font-bold text-white mb-2">
          Settings
        </h1>
        <p className="text-slate-400">
          Manage your profile, business metrics, and workspace preferences.
        </p>
      </div>

      {/* Profile & Workspace Section */}
      <div className="bg-slate-900/50 backdrop-blur-sm rounded-2xl p-8 border border-slate-800">
        <div className="mb-6">
          <h2 className="text-xl font-semibold text-white mb-2">
            Profile & Workspace
          </h2>
          <p className="text-sm text-slate-400">
            Your personal information and company details.
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {/* Full Name */}
          <div>
            <label htmlFor="fullName" className="block text-sm font-semibold text-slate-300 mb-2">
              Full Name
            </label>
            <input
              id="fullName"
              type="text"
              value={profile.fullName}
              onChange={(e) => handleInputChange('fullName', e.target.value)}
              className="w-full px-4 py-3 rounded-xl bg-slate-900/50 border border-slate-700 text-white placeholder-slate-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all"
              placeholder="John Doe"
            />
          </div>

          {/* Email */}
          <div>
            <label htmlFor="email" className="block text-sm font-semibold text-slate-300 mb-2">
              Email
            </label>
            <input
              id="email"
              type="email"
              value={profile.email}
              onChange={(e) => handleInputChange('email', e.target.value)}
              className="w-full px-4 py-3 rounded-xl bg-slate-900/50 border border-slate-700 text-white placeholder-slate-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all"
              placeholder="john@example.com"
            />
          </div>

          {/* Role */}
          <div>
            <label htmlFor="role" className="block text-sm font-semibold text-slate-300 mb-2">
              Role
            </label>
            <input
              id="role"
              type="text"
              value={profile.role}
              onChange={(e) => handleInputChange('role', e.target.value)}
              className="w-full px-4 py-3 rounded-xl bg-slate-900/50 border border-slate-700 text-white placeholder-slate-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all"
              placeholder="Founder, Head of Product, etc."
            />
          </div>

          {/* Currency */}
          <div>
            <label htmlFor="currency" className="block text-sm font-semibold text-slate-300 mb-2">
              Display Currency
            </label>
            <select
              id="currency"
              value={profile.currency}
              onChange={(e) => handleInputChange('currency', e.target.value)}
              className="w-full px-4 py-3 rounded-xl bg-slate-900/50 border border-slate-700 text-white focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all appearance-none pr-10"
              style={{
                backgroundImage: `url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 24 24' stroke='%23cbd5e1'%3E%3Cpath stroke-linecap='round' stroke-linejoin='round' stroke-width='2' d='M19 9l-7 7-7-7'%3E%3C/path%3E%3C/svg%3E")`,
                backgroundRepeat: 'no-repeat',
                backgroundPosition: 'right 0.75rem center',
                backgroundSize: '1.5rem'
              }}
            >
              <option value="EUR">EUR (€)</option>
              <option value="USD">USD ($)</option>
              <option value="GBP">GBP (£)</option>
            </select>
          </div>

          {/* Company Name */}
          <div>
            <label htmlFor="companyName" className="block text-sm font-semibold text-slate-300 mb-2">
              Company Name
            </label>
            <input
              id="companyName"
              type="text"
              value={profile.companyName}
              onChange={(e) => handleInputChange('companyName', e.target.value)}
              className="w-full px-4 py-3 rounded-xl bg-slate-900/50 border border-slate-700 text-white placeholder-slate-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all"
              placeholder="Acme Inc."
            />
          </div>

          {/* Company Website */}
          <div>
            <label htmlFor="companyWebsite" className="block text-sm font-semibold text-slate-300 mb-2">
              Company Website <span className="text-slate-500 font-normal">(optional)</span>
            </label>
            <input
              id="companyWebsite"
              type="url"
              value={profile.companyWebsite}
              onChange={(e) => handleInputChange('companyWebsite', e.target.value)}
              className="w-full px-4 py-3 rounded-xl bg-slate-900/50 border border-slate-700 text-white placeholder-slate-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all"
              placeholder="https://acme.com"
            />
          </div>
        </div>

        {/* Saved Indicator */}
        {showSavedMessage && (
          <div className="mt-6 p-3 bg-blue-500/10 border border-blue-500/20 rounded-xl">
            <p className="text-sm text-blue-400">
              Changes saved locally
            </p>
          </div>
        )}
      </div>

      {/* Business Metrics Section */}
      <div className="bg-slate-900/50 backdrop-blur-sm rounded-2xl p-8 border border-slate-800">
        <div className="mb-6">
          <h2 className="text-xl font-semibold text-white mb-2">
            Business Metrics
          </h2>
          <p className="text-sm text-slate-400">
            Your current SaaS metrics. These will be used to provide more accurate pricing recommendations.
          </p>
        </div>

        <form onSubmit={handleMetricsSubmit}>
          {/* Error Banner */}
          {metricsError && (
            <div className="mb-6 p-4 bg-red-500/10 border border-red-500/30 rounded-xl flex items-center justify-between">
              <p className="text-sm text-red-400">{metricsError}</p>
              <button 
                type="button"
                onClick={clearError}
                className="text-red-400 hover:text-red-300 p-1"
              >
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
          )}

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {/* Currency */}
            <div>
              <label htmlFor="metricsCurrency" className="block text-sm font-semibold text-slate-300 mb-2">
                Revenue Currency
              </label>
              <select
                id="metricsCurrency"
                value={metricsForm.currency}
                onChange={(e) => handleMetricsChange('currency', e.target.value)}
                disabled={isSaving}
                className="w-full px-4 py-3 rounded-xl bg-slate-900/50 border border-slate-700 text-white focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all appearance-none pr-10 disabled:opacity-50"
                style={{
                  backgroundImage: `url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 24 24' stroke='%23cbd5e1'%3E%3Cpath stroke-linecap='round' stroke-linejoin='round' stroke-width='2' d='M19 9l-7 7-7-7'%3E%3C/path%3E%3C/svg%3E")`,
                  backgroundRepeat: 'no-repeat',
                  backgroundPosition: 'right 0.75rem center',
                  backgroundSize: '1.5rem'
                }}
              >
                <option value="USD">USD ($)</option>
                <option value="EUR">EUR (€)</option>
                <option value="GBP">GBP (£)</option>
              </select>
            </div>

            {/* MRR */}
            <div>
              <label htmlFor="mrr" className="block text-sm font-semibold text-slate-300 mb-2">
                Monthly Recurring Revenue (MRR)
              </label>
              <input
                id="mrr"
                type="number"
                step="0.01"
                min="0"
                value={metricsForm.mrr}
                onChange={(e) => handleMetricsChange('mrr', e.target.value)}
                disabled={isSaving}
                className="w-full px-4 py-3 rounded-xl bg-slate-900/50 border border-slate-700 text-white placeholder-slate-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all disabled:opacity-50"
                placeholder="e.g., 15000"
              />
            </div>

            {/* Customers */}
            <div>
              <label htmlFor="customers" className="block text-sm font-semibold text-slate-300 mb-2">
                Number of Customers
              </label>
              <input
                id="customers"
                type="number"
                min="0"
                value={metricsForm.customers}
                onChange={(e) => handleMetricsChange('customers', e.target.value)}
                disabled={isSaving}
                className="w-full px-4 py-3 rounded-xl bg-slate-900/50 border border-slate-700 text-white placeholder-slate-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all disabled:opacity-50"
                placeholder="e.g., 120"
              />
            </div>

            {/* Monthly Churn Rate */}
            <div>
              <label htmlFor="churnRate" className="block text-sm font-semibold text-slate-300 mb-2">
                Monthly Churn Rate (%)
              </label>
              <input
                id="churnRate"
                type="number"
                step="0.1"
                min="0"
                max="100"
                value={metricsForm.monthly_churn_rate}
                onChange={(e) => handleMetricsChange('monthly_churn_rate', e.target.value)}
                disabled={isSaving}
                className="w-full px-4 py-3 rounded-xl bg-slate-900/50 border border-slate-700 text-white placeholder-slate-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all disabled:opacity-50"
                placeholder="e.g., 2.5"
              />
            </div>
          </div>

          <div className="mt-6 flex items-center gap-4">
            <button
              type="submit"
              disabled={isSaving}
              className="px-6 py-3 bg-gradient-to-r from-blue-500 to-indigo-600 text-white rounded-xl font-semibold hover:from-blue-600 hover:to-indigo-700 hover:scale-105 transition-all shadow-lg shadow-blue-500/20 disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:scale-100"
            >
              {isSaving ? (
                <span className="flex items-center gap-2">
                  <svg className="animate-spin h-5 w-5" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                  </svg>
                  Saving...
                </span>
              ) : (
                'Save Metrics'
              )}
            </button>

            {metricsSaved && (
              <span className="text-sm text-emerald-400 flex items-center gap-1">
                <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                </svg>
                Metrics saved successfully
              </span>
            )}
          </div>
        </form>
      </div>

      {/* Demo Environment Section */}
      <div className="bg-slate-900/50 backdrop-blur-sm rounded-2xl p-8 border border-slate-800">
        <div className="mb-6">
          <h2 className="text-xl font-semibold text-white mb-2">
            Demo Environment
          </h2>
          <p className="text-sm text-slate-400">
            This environment is currently running in demo mode. You can clear all local pricing data (plans, competitors, and analyses) to start over.
          </p>
        </div>

        <div className="flex items-start gap-4 p-4 bg-slate-800/30 rounded-xl border border-slate-700 mb-6">
          <svg className="w-5 h-5 text-amber-400 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
          </svg>
          <div>
            <h3 className="text-white font-medium mb-1">
              Warning
            </h3>
            <p className="text-sm text-slate-400 leading-relaxed">
              Resetting demo data will permanently delete all plans, competitors, and pricing analyses from your browser. This action cannot be undone.
            </p>
          </div>
        </div>

        <button
          onClick={handleResetDemoData}
          className="px-6 py-3 bg-slate-800 text-red-400 border border-red-500/30 rounded-xl font-semibold hover:bg-red-500/10 hover:border-red-500/50 transition-all"
        >
          Reset Demo Data
        </button>

        <p className="mt-4 text-sm text-slate-500">
          This only affects local demo data in your browser. No real accounts or billing data are involved yet.
        </p>
      </div>

      {/* Additional Info */}
      <div className="bg-blue-500/5 backdrop-blur-sm rounded-2xl p-6 border border-blue-500/20">
        <div className="flex items-start gap-3">
          <svg className="w-5 h-5 text-blue-400 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <div>
            <h3 className="text-white font-semibold mb-1">
              About Demo Mode
            </h3>
            <p className="text-sm text-slate-300 leading-relaxed">
              You're currently using Revalyze in demo mode. All data is stored locally in your browser. Once we launch full accounts and billing, you'll be able to sync your data across devices and collaborate with your team.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Settings;
