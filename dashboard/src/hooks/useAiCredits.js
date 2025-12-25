import { useState, useEffect, useCallback } from 'react';
import { aiCreditsApi } from '../lib/apiClient';

/**
 * Hook to fetch and manage AI Insight Credits data.
 * 
 * @returns {Object} - { credits, loading, error, refetch }
 *   - credits: { planType, monthlyCredits, usedCredits, remainingCredits, simulationsEnabled, monthKey }
 *   - loading: boolean
 *   - error: Error | null
 *   - refetch: Function to refresh credits data
 */
export const useAiCredits = () => {
  const [credits, setCredits] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const fetchCredits = useCallback(async () => {
    setLoading(true);
    setError(null);
    
    try {
      const { data } = await aiCreditsApi.get();
      setCredits(data);
    } catch (err) {
      console.error('Failed to fetch AI credits:', err);
      setError(err);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchCredits();
  }, [fetchCredits]);

  return {
    credits,
    loading,
    error,
    refetch: fetchCredits,
  };
};

export default useAiCredits;





