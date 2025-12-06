import { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { competitorsApi } from '../lib/apiClient';
import { useAuth } from './AuthContext';

const CompetitorsContext = createContext();

export const CompetitorsProvider = ({ children }) => {
  const { isAuthenticated, isLoading: authLoading } = useAuth();
  
  const [competitors, setCompetitors] = useState([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);

  // Fetch competitors from backend when authenticated
  const fetchCompetitors = useCallback(async () => {
    if (!isAuthenticated) {
      setCompetitors([]);
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const { data } = await competitorsApi.list();
      // Transform backend data to match frontend expectations
      const transformedCompetitors = (data || []).map(comp => ({
        id: comp.id,
        name: comp.name || '',
        url: comp.url || '',
        basePrice: comp.base_price || 0,
        createdAt: comp.created_at
      }));
      setCompetitors(transformedCompetitors);
    } catch (err) {
      console.error('Failed to fetch competitors:', err);
      setError(err.message || 'Failed to load competitors');
      setCompetitors([]);
    } finally {
      setIsLoading(false);
    }
  }, [isAuthenticated]);

  // Fetch competitors when auth state changes
  useEffect(() => {
    if (!authLoading) {
      fetchCompetitors();
    }
  }, [authLoading, fetchCompetitors]);

  // Add a new competitor via API
  const addCompetitor = async ({ name, url, basePrice = 0 }) => {
    setError(null);

    try {
      const { data } = await competitorsApi.create(name || '', url, basePrice);
      
      // Add the new competitor to state
      const newCompetitor = {
        id: data.id,
        name: data.name || '',
        url: data.url || '',
        basePrice: data.base_price || 0,
        createdAt: data.created_at
      };
      
      setCompetitors(prev => [...prev, newCompetitor]);
      return { success: true, competitor: newCompetitor };
    } catch (err) {
      console.error('Failed to add competitor:', err);
      setError(err.message || 'Failed to add competitor');
      return { success: false, error: err.message };
    }
  };

  // Remove a competitor via API
  const removeCompetitor = async (id) => {
    setError(null);

    try {
      await competitorsApi.delete(id);
      setCompetitors(prev => prev.filter(comp => comp.id !== id));
      return { success: true };
    } catch (err) {
      console.error('Failed to remove competitor:', err);
      setError(err.message || 'Failed to remove competitor');
      return { success: false, error: err.message };
    }
  };

  // Clear all competitors locally (for reset demo data feature)
  const clearCompetitors = () => {
    setCompetitors([]);
  };

  // Reset competitors by refetching from backend
  const resetCompetitors = () => {
    fetchCompetitors();
  };

  // Clear error
  const clearError = () => {
    setError(null);
  };

  const value = {
    competitors,
    isLoading,
    error,
    addCompetitor,
    removeCompetitor,
    clearCompetitors,
    resetCompetitors,
    clearError,
    refetch: fetchCompetitors
  };

  return (
    <CompetitorsContext.Provider value={value}>
      {children}
    </CompetitorsContext.Provider>
  );
};

export const useCompetitors = () => {
  const context = useContext(CompetitorsContext);
  if (!context) {
    throw new Error('useCompetitors must be used within CompetitorsProvider');
  }
  return context;
};
