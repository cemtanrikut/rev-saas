import { createContext, useContext, useState, useEffect } from 'react';

const SettingsContext = createContext();

const STORAGE_KEY = 'revalyze_settings_profile';

export const SettingsProvider = ({ children }) => {
  const [profile, setProfile] = useState(() => {
    // Try to load from localStorage
    try {
      const stored = localStorage.getItem(STORAGE_KEY);
      if (stored) {
        return JSON.parse(stored);
      }
    } catch (error) {
      console.error('Failed to load profile from localStorage:', error);
    }
    
    // Default profile
    return {
      fullName: '',
      email: '',
      role: '',
      companyName: '',
      companyWebsite: '',
      currency: 'EUR'
    };
  });

  // Persist to localStorage whenever profile changes
  useEffect(() => {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(profile));
    } catch (error) {
      console.error('Failed to save profile to localStorage:', error);
    }
  }, [profile]);

  const updateProfile = (partial) => {
    setProfile(prev => ({
      ...prev,
      ...partial
    }));
  };

  const resetProfile = () => {
    const defaultProfile = {
      fullName: '',
      email: '',
      role: '',
      companyName: '',
      companyWebsite: '',
      currency: 'EUR'
    };
    setProfile(defaultProfile);
  };

  const value = {
    profile,
    updateProfile,
    resetProfile
  };

  return (
    <SettingsContext.Provider value={value}>
      {children}
    </SettingsContext.Provider>
  );
};

export const useSettings = () => {
  const context = useContext(SettingsContext);
  if (!context) {
    throw new Error('useSettings must be used within SettingsProvider');
  }
  return context;
};


