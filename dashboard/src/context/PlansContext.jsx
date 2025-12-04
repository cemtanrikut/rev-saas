import { createContext, useContext, useState } from 'react';

const PlansContext = createContext();

export const PlansProvider = ({ children }) => {
  const [plans, setPlans] = useState([]);

  const addPlan = ({ name, price, interval, description }) => {
    const newPlan = {
      id: Date.now().toString(),
      name,
      price: Number(price),
      interval,
      description: description || ''
    };
    setPlans(prev => [...prev, newPlan]);
  };

  const removePlan = (id) => {
    setPlans(prev => prev.filter(plan => plan.id !== id));
  };

  const updatePlan = (id, partialData) => {
    setPlans(prev =>
      prev.map(plan =>
        plan.id === id ? { ...plan, ...partialData } : plan
      )
    );
  };

  const clearPlans = () => {
    setPlans([]);
  };

  const resetPlans = () => {
    setPlans([]);
  };

  const value = {
    plans,
    addPlan,
    removePlan,
    updatePlan,
    clearPlans,
    resetPlans
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

