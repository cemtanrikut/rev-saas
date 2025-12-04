import { createContext, useContext, useState, useMemo } from 'react';

const AnalysisContext = createContext();

export const AnalysisProvider = ({ children }) => {
  const [analyses, setAnalyses] = useState([]);
  const [selectedAnalysisId, setSelectedAnalysisId] = useState(null);

  // Derived values
  const lastAnalysis = analyses.length > 0 ? analyses[0] : null;

  const selectedAnalysis = useMemo(() => {
    if (!selectedAnalysisId) return lastAnalysis;
    return analyses.find(a => a.id === selectedAnalysisId) || lastAnalysis;
  }, [analyses, selectedAnalysisId, lastAnalysis]);

  const generateSuggestions = (plans, competitors) => {
    return plans.map((plan, index) => {
      // Base multiplier: 10% increase
      // Add 2% for each plan tier (so higher plans get slightly more increase)
      // Add 1% for each competitor (more competition = more pricing power)
      const competitorBonus = Math.min(competitors.length * 0.01, 0.05); // cap at 5%
      const tierBonus = index * 0.02; // 0%, 2%, 4%, etc.
      const multiplier = 1.10 + tierBonus + competitorBonus;

      const suggestedPrice = Math.round(plan.price * multiplier * 100) / 100; // round to 2 decimals

      const changeAbsolute = suggestedPrice - plan.price;
      const changePercent = (changeAbsolute / plan.price) * 100;

      return {
        planId: plan.id,
        planName: plan.name,
        planInterval: plan.interval,
        currentPrice: plan.price,
        suggestedPrice,
        changeAbsolute,
        changePercent
      };
    });
  };

  const buildSummary = (plans, competitors, suggestions) => {
    if (!plans.length || !suggestions.length) {
      return "We could not compute a meaningful analysis because there are no plans or suggestions.";
    }

    const numPlans = plans.length;
    const numCompetitors = competitors.length;

    const totalChangePercent = suggestions.reduce((sum, s) => sum + s.changePercent, 0);
    const averageChangePercent = totalChangePercent / suggestions.length;

    const maxChangePercent = Math.max(...suggestions.map(s => s.changePercent));

    const sortedByPrice = [...plans].sort((a, b) => a.price - b.price);
    const entryPlan = sortedByPrice[0];
    const highestPlan = sortedByPrice[sortedByPrice.length - 1];

    // Determine pricing position narrative
    let pricingPosition;
    if (averageChangePercent < 5) {
      pricingPosition = "roughly in line with typical SaaS pricing norms";
    } else if (averageChangePercent < 15) {
      pricingPosition = "slightly underpriced compared to typical SaaS pricing norms";
    } else {
      pricingPosition = "materially underpriced for your category";
    }

    const competitorsText = numCompetitors === 0
      ? "We did not use any explicit competitors for this analysis."
      : `We used ${numCompetitors} competitor${numCompetitors > 1 ? "s" : ""} as a benchmark.`;

    const entryPriceFormatted = new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 0,
      maximumFractionDigits: 2
    }).format(entryPlan.price);

    return [
      `Your current pricing appears ${pricingPosition}. On average, we suggest a ${averageChangePercent.toFixed(1)}% uplift across your paid plans to move you closer to where similar SaaS products typically sit.`,
      `Your entry plan ("${entryPlan.name}") is currently at ${entryPriceFormatted} and is likely the most sensitive to underpricing. Your top plan ("${highestPlan.name}") can usually sustain a stronger increase; in this analysis the highest recommended uplift on a single plan is ${maxChangePercent.toFixed(1)}%.`,
      competitorsText + " These recommendations are not final prices, but a data-driven starting point for your next pricing iteration."
    ].join(" ");
  };

  const runAnalysis = ({ plans, competitors }) => {
    if (!plans || plans.length === 0) {
      throw new Error('Cannot run analysis without plans');
    }

    if (!competitors || competitors.length === 0) {
      throw new Error('Cannot run analysis without competitors');
    }

    const suggestions = generateSuggestions(plans, competitors);
    const summary = buildSummary(plans, competitors, suggestions);

    // Calculate stats for insights
    const totalChangePercent = suggestions.reduce((sum, s) => sum + s.changePercent, 0);
    const averageChangePercent = totalChangePercent / suggestions.length;
    const maxChangePercent = Math.max(...suggestions.map(s => s.changePercent));
    const minChangePercent = Math.min(...suggestions.map(s => s.changePercent));

    const newAnalysis = {
      id: Date.now().toString(),
      createdAt: new Date().toISOString(),
      plans: plans.map(p => ({
        id: p.id,
        name: p.name,
        price: p.price,
        interval: p.interval,
        description: p.description
      })),
      competitors: competitors.map(c => ({
        id: c.id,
        name: c.name,
        url: c.url
      })),
      suggestions,
      summary,
      // Stats for insights
      stats: {
        averageChangePercent,
        maxChangePercent,
        minChangePercent,
        numPlans: plans.length,
        numCompetitors: competitors.length
      }
    };

    setAnalyses(prev => [newAnalysis, ...prev]); // prepend to history
    setSelectedAnalysisId(newAnalysis.id); // select the new analysis

    return newAnalysis;
  };

  const selectAnalysis = (id) => {
    setSelectedAnalysisId(id);
  };

  const clearAnalysis = () => {
    setSelectedAnalysisId(null);
  };

  const reset = () => {
    setAnalyses([]);
    setSelectedAnalysisId(null);
  };

  const clearAnalyses = () => {
    setAnalyses([]);
    setSelectedAnalysisId(null);
  };

  const resetAnalyses = () => {
    setAnalyses([]);
    setSelectedAnalysisId(null);
  };

  const value = {
    analyses,
    lastAnalysis,
    selectedAnalysis,
    runAnalysis,
    selectAnalysis,
    clearAnalysis,
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

