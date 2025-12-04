import { createContext, useContext, useState } from 'react';

const CompetitorsContext = createContext();

export const CompetitorsProvider = ({ children }) => {
  const [competitors, setCompetitors] = useState([]);

  const addCompetitor = ({ name, url }) => {
    const newCompetitor = {
      id: Date.now().toString(),
      name: name || '',
      url
    };
    setCompetitors(prev => [...prev, newCompetitor]);
  };

  const removeCompetitor = (id) => {
    setCompetitors(prev => prev.filter(comp => comp.id !== id));
  };

  const clearCompetitors = () => {
    setCompetitors([]);
  };

  const resetCompetitors = () => {
    setCompetitors([]);
  };

  const value = {
    competitors,
    addCompetitor,
    removeCompetitor,
    clearCompetitors,
    resetCompetitors
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

