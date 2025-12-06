import { createContext, useContext, useState, useEffect, useCallback, useMemo } from 'react';
import { analysisApi, LimitError } from '../lib/apiClient';
import { useAuth } from './AuthContext';

const AnalysisContext = createContext();

export const AnalysisProvider = ({ children }) => {
  const { isAuthenticated, isLoading: authLoading } = useAuth();
  
  const [analyses, setAnalyses] = useState([]);
  const [selectedAnalysisId, setSelectedAnalysisId] = useState(null);
  const [isLoading, setIsLoading] = useState(false);
  const [isRunning, setIsRunning] = useState(false);
  const [error, setError] = useState(null);
  const [limitError, setLimitError] = useState(null); // Specific limit error state

  // Derived values
  const lastAnalysis = analyses.length > 0 ? analyses[0] : null;

  const selectedAnalysis = useMemo(() => {
    if (!selectedAnalysisId) return lastAnalysis;
    return analyses.find(a => a.id === selectedAnalysisId) || lastAnalysis;
  }, [analyses, selectedAnalysisId, lastAnalysis]);

  // Fetch analysis history from backend
  const fetchHistory = useCallback(async () => {
    if (!isAuthenticated) {
      setAnalyses([]);
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      const { data } = await analysisApi.list();
      // Transform backend data - already sorted by created_at desc
      const transformedAnalyses = (data || []).map(transformAnalysis);
      setAnalyses(transformedAnalyses);
      
      // Select the first (latest) analysis if none selected
      if (transformedAnalyses.length > 0 && !selectedAnalysisId) {
        setSelectedAnalysisId(transformedAnalyses[0].id);
      }
    } catch (err) {
      console.error('Failed to fetch analyses:', err);
      setError(err.message || 'Failed to load analysis history');
      setAnalyses([]);
    } finally {
      setIsLoading(false);
    }
  }, [isAuthenticated, selectedAnalysisId]);

  // Fetch history when auth state changes
  useEffect(() => {
    if (!authLoading) {
      fetchHistory();
    }
  }, [authLoading, fetchHistory]);

  // Transform backend analysis to frontend format
  const transformAnalysis = (backendAnalysis) => {
    const recommendations = backendAnalysis.recommendations || [];
    
    // Calculate stats from recommendations
    const priceChanges = recommendations
      .filter(r => r.current_price > 0 && r.suggested_new_price > 0)
      .map(r => ((r.suggested_new_price - r.current_price) / r.current_price) * 100);
    
    const averageChangePercent = priceChanges.length > 0
      ? priceChanges.reduce((sum, c) => sum + c, 0) / priceChanges.length
      : 0;
    const maxChangePercent = priceChanges.length > 0 ? Math.max(...priceChanges) : 0;
    const minChangePercent = priceChanges.length > 0 ? Math.min(...priceChanges) : 0;

    return {
      id: backendAnalysis.id,
      createdAt: backendAnalysis.created_at,
      summary: backendAnalysis.summary,
      numPlans: backendAnalysis.num_plans,
      numCompetitors: backendAnalysis.num_competitors,
      recommendations: recommendations.map(r => ({
        planId: r.plan_id,
        planName: r.plan_name,
        currentPrice: r.current_price,
        suggestedPrice: r.suggested_new_price,
        position: r.position,
        suggestedAction: r.suggested_action,
        rationale: r.rationale,
        changeAbsolute: r.suggested_new_price - r.current_price,
        changePercent: r.current_price > 0 
          ? ((r.suggested_new_price - r.current_price) / r.current_price) * 100 
          : 0
      })),
      // Stats for UI compatibility
      stats: {
        averageChangePercent,
        maxChangePercent,
        minChangePercent,
        numPlans: backendAnalysis.num_plans,
        numCompetitors: backendAnalysis.num_competitors
      }
    };
  };

  // Run a new analysis via backend
  const runAnalysis = async () => {
    setIsRunning(true);
    setError(null);
    setLimitError(null);

    try {
      const { data } = await analysisApi.run();
      const transformedAnalysis = transformAnalysis(data);
      
      // Prepend to history
      setAnalyses(prev => [transformedAnalysis, ...prev]);
      setSelectedAnalysisId(transformedAnalysis.id);
      
      return { success: true, analysis: transformedAnalysis };
    } catch (err) {
      console.error('Failed to run analysis:', err);
      
      // Handle limit errors specifically
      if (err instanceof LimitError) {
        setLimitError({
          errorCode: err.errorCode,
          reason: err.reason,
          plan: err.plan,
          limit: err.limit,
          current: err.current,
        });
        return { 
          success: false, 
          error: err.reason,
          isLimitError: true,
          limitError: {
            errorCode: err.errorCode,
            reason: err.reason,
            plan: err.plan,
            limit: err.limit,
            current: err.current,
          }
        };
      }
      
      setError(err.message || 'Failed to run analysis');
      return { success: false, error: err.message };
    } finally {
      setIsRunning(false);
    }
  };

  const clearLimitError = () => {
    setLimitError(null);
  };

  const selectAnalysis = (id) => {
    setSelectedAnalysisId(id);
  };

  const clearAnalysis = () => {
    setSelectedAnalysisId(null);
  };

  const clearError = () => {
    setError(null);
  };

  // Reset local state (for demo reset feature)
  const clearAnalyses = () => {
    setAnalyses([]);
    setSelectedAnalysisId(null);
  };

  const resetAnalyses = () => {
    fetchHistory();
  };

  const reset = () => {
    setAnalyses([]);
    setSelectedAnalysisId(null);
  };

  const value = {
    analyses,
    lastAnalysis,
    selectedAnalysis,
    isLoading,
    isRunning,
    error,
    limitError,
    runAnalysis,
    fetchHistory,
    selectAnalysis,
    clearAnalysis,
    clearError,
    clearLimitError,
    clearAnalyses,
    resetAnalyses,
    reset
  };

  return (
    <AnalysisContext.Provider value={value}>
      {children}
    </AnalysisContext.Provider>
  );
};

export const useAnalysis = () => {
  const context = useContext(AnalysisContext);
  if (!context) {
    throw new Error('useAnalysis must be used within AnalysisProvider');
  }
  return context;
};
