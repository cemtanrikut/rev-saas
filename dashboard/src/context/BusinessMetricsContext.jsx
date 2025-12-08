import { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { businessMetricsApi } from '../lib/apiClient';
import { useAuth } from './AuthContext';

const BusinessMetricsContext = createContext();

export const BusinessMetricsProvider = ({ children }) => {
  const { isAuthenticated, isLoading: authLoading } = useAuth();
  
  const [metrics, setMetrics] = useState(null);
  const [isLoading, setIsLoading] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState(null);

  // Load metrics from backend
  const loadMetrics = useCallback(async () => {
    if (!isAuthenticated) {
      setMetrics(null);
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const { data } = await businessMetricsApi.get();
      setMetrics(data); // Can be null if not set yet
    } catch (err) {
      console.error('Failed to load business metrics:', err);
      setError(err.message || 'Failed to load business metrics');
      setMetrics(null);
    } finally {
      setIsLoading(false);
    }
  }, [isAuthenticated]);

  // Load metrics when auth state changes
  useEffect(() => {
    if (!authLoading) {
      loadMetrics();
    }
  }, [authLoading, loadMetrics]);

  // Save metrics to backend
  const saveMetrics = async (metricsInput) => {
    setIsSaving(true);
    setError(null);

    try {
      const { data } = await businessMetricsApi.set(metricsInput);
      setMetrics(data);
      return { success: true, metrics: data };
    } catch (err) {
      console.error('Failed to save business metrics:', err);
      setError(err.message || 'Failed to save business metrics');
      return { success: false, error: err.message };
    } finally {
      setIsSaving(false);
    }
  };

  // Clear error
  const clearError = () => {
    setError(null);
  };

  // Reset local state
  const resetMetrics = () => {
    loadMetrics();
  };

  const value = {
    metrics,
    isLoading,
    isSaving,
    error,
    loadMetrics,
    saveMetrics,
    clearError,
    resetMetrics,
    hasMetrics: metrics !== null,
  };

  return (
    <BusinessMetricsContext.Provider value={value}>
      {children}
    </BusinessMetricsContext.Provider>
  );
};

export const useBusinessMetrics = () => {
  const context = useContext(BusinessMetricsContext);
  if (!context) {
    throw new Error('useBusinessMetrics must be used within BusinessMetricsProvider');
  }
  return context;
};


