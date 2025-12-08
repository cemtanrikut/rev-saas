import { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { authApi, setToken, getToken, removeToken } from '../lib/apiClient';

const AuthContext = createContext();

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [token, setTokenState] = useState(getToken());
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(null);

  // Hydrate user from token on mount
  useEffect(() => {
    const hydrateUser = async () => {
      const storedToken = getToken();
      
      if (!storedToken) {
        setIsLoading(false);
        return;
      }

      try {
        const { data } = await authApi.me();
        setUser(data);
        setTokenState(storedToken);
      } catch (err) {
        // Token is invalid or expired
        console.error('Failed to hydrate user:', err);
        removeToken();
        setTokenState(null);
        setUser(null);
      } finally {
        setIsLoading(false);
      }
    };

    hydrateUser();
  }, []);

  // Login function
  const login = useCallback(async (email, password) => {
    setError(null);
    
    try {
      const { data } = await authApi.login(email, password);
      
      // Store token
      setToken(data.token);
      setTokenState(data.token);
      
      // Store user
      setUser(data.user);
      
      return { success: true, user: data.user };
    } catch (err) {
      const errorMessage = err.message || 'Login failed';
      setError(errorMessage);
      return { success: false, error: errorMessage };
    }
  }, []);

  // Signup function - accepts full signup data object
  // Now auto-logs in the user by saving the returned token
  const signup = useCallback(async (signupData) => {
    setError(null);
    
    try {
      const { data } = await authApi.signup(signupData);
      
      // Auto-login: save token and user
      if (data.token) {
        setToken(data.token);
        setTokenState(data.token);
        setUser(data.user);
      }
      
      return { success: true, user: data.user, company: data.company };
    } catch (err) {
      const errorMessage = err.message || 'Signup failed';
      setError(errorMessage);
      return { success: false, error: errorMessage };
    }
  }, []);

  // Logout function
  const logout = useCallback(() => {
    removeToken();
    setTokenState(null);
    setUser(null);
    setError(null);
  }, []);

  // Clear error
  const clearError = useCallback(() => {
    setError(null);
  }, []);

  // Check if user is admin (has unlimited access)
  const isAdmin = user?.role === 'admin' || user?.plan === 'admin';

  const value = {
    user,
    token,
    isLoading,
    isAuthenticated: !!user && !!token,
    isAdmin,
    error,
    login,
    signup,
    logout,
    clearError,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
};

