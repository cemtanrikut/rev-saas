import { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { pricingV2Api } from '../lib/apiClient';
import { useAuth } from './AuthContext';

const PlansContext = createContext();

export const PlansProvider = ({ children }) => {
  const { isAuthenticated, isLoading: authLoading } = useAuth();
  
  const [plans, setPlans] = useState([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);

  // Fetch plans from backend when authenticated (using Pricing V2 API)
  const fetchPlans = useCallback(async () => {
    if (!isAuthenticated) {
      setPlans([]);
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const response = await pricingV2Api.list();
      const data = response.data || response;
      // Transform pricing v2 data to match frontend expectations
      const transformedPlans = (data.plans || []).map(plan => ({
        id: plan.id,
        name: plan.plan_name,
        price: plan.price_amount,
        priceString: plan.price_string,
        currency: plan.currency || 'USD',
        interval: plan.billing_period || 'monthly',
        billingPeriod: plan.billing_period,
        monthlyEquivalent: plan.monthly_equivalent_amount,
        annualBilled: plan.annual_billed_amount,
        features: plan.features || [],
        includedUnits: plan.included_units || [],
        sourceUrl: plan.source_url,
        websiteUrl: plan.website_url,
        extractedAt: plan.extracted_at
      }));
      setPlans(transformedPlans);
    } catch (err) {
      console.error('Failed to fetch plans:', err);
      setError(err.message || 'Failed to load plans');
      setPlans([]);
    } finally {
      setIsLoading(false);
    }
  }, [isAuthenticated]);

  // Fetch plans when auth state changes
  useEffect(() => {
    if (!authLoading) {
      fetchPlans();
    }
  }, [authLoading, fetchPlans]);

  // Remove a plan via API (Pricing V2)
  const removePlan = async (id) => {
    setError(null);

    try {
      await pricingV2Api.delete(id);
      setPlans(prev => prev.filter(plan => plan.id !== id));
      return { success: true };
    } catch (err) {
      console.error('Failed to remove plan:', err);
      setError(err.message || 'Failed to remove plan');
      return { success: false, error: err.message };
    }
  };

  // Note: addPlan and updatePlan are not used in V2 - plans are imported via auto-detect
  const addPlan = async () => {
    console.warn('addPlan is deprecated in V2 - use auto-import instead');
    return { success: false, error: 'Use auto-import feature on My Pricing page' };
  };

  const updatePlan = async () => {
    console.warn('updatePlan is deprecated in V2 - plans are read-only');
    return { success: false, error: 'Plans are read-only in V2' };
  };

  // Clear all plans locally (for reset demo data feature)
  const clearPlans = () => {
    setPlans([]);
  };

  // Reset plans by refetching from backend
  const resetPlans = () => {
    fetchPlans();
  };

  // Clear error
  const clearError = () => {
    setError(null);
  };

  const value = {
    plans,
    isLoading,
    error,
    addPlan,
    removePlan,
    updatePlan,
    clearPlans,
    resetPlans,
    clearError,
    refetch: fetchPlans
  };

  return (
    <PlansContext.Provider value={value}>
      {children}
    </PlansContext.Provider>
  );
};

export const usePlans = () => {
  const context = useContext(PlansContext);
  if (!context) {
    throw new Error('usePlans must be used within PlansProvider');
  }
  return context;
};
