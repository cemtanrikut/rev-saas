import { useState, useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import AuthCard from '../components/AuthCard';

const Login = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { login, isAuthenticated, isLoading, error, clearError } = useAuth();
  
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [localError, setLocalError] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Redirect if already authenticated
  useEffect(() => {
    if (isAuthenticated && !isLoading) {
      const from = location.state?.from?.pathname || '/app/overview';
      navigate(from, { replace: true });
    }
  }, [isAuthenticated, isLoading, navigate, location]);

  // Clear errors when inputs change
  useEffect(() => {
    if (localError) setLocalError('');
    if (error) clearError();
  }, [email, password]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLocalError('');
    
    if (!email.trim() || !password.trim()) {
      setLocalError('Please enter email and password');
      return;
    }

    setIsSubmitting(true);

    try {
      const result = await login(email, password);
      
      if (result.success) {
        const from = location.state?.from?.pathname || '/app/overview';
        navigate(from, { replace: true });
      } else {
        setLocalError(result.error);
      }
    } catch (err) {
      setLocalError('An unexpected error occurred');
    } finally {
      setIsSubmitting(false);
    }
  };

  const displayError = localError || error;

  return (
    <>
      <div className="min-h-screen bg-gradient-to-b from-slate-900 via-slate-800 to-slate-900 flex items-center justify-center px-4 py-12">
        {/* Logo at top */}
        <div className="w-full max-w-md animate-fadeIn">
          <div className="flex justify-center mb-8">
            <img 
              src="/revalyze-logo.png" 
              alt="Revalyze" 
              className="h-14 w-auto"
            />
          </div>

          <AuthCard 
            title="Sign in to Revalyze"
            subtitle="Access your SaaS pricing intelligence dashboard"
          >
            <form onSubmit={handleSubmit} className="space-y-6">
              {/* Error Message */}
              {displayError && (
                <div className="p-3 rounded-lg bg-red-500/10 border border-red-500/30">
                  <p className="text-sm text-red-400">{displayError}</p>
                </div>
              )}

              <div>
                <label 
                  htmlFor="email" 
                  className="block text-sm font-semibold text-slate-300 mb-2"
                >
                  Email
                </label>
                <input
                  id="email"
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                  disabled={isSubmitting}
                  className="w-full px-4 py-3 rounded-xl bg-slate-900/50 border border-slate-700 text-white placeholder-slate-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all disabled:opacity-50"
                  placeholder="you@company.com"
                />
              </div>

              <div>
                <label 
                  htmlFor="password" 
                  className="block text-sm font-semibold text-slate-300 mb-2"
                >
                  Password
                </label>
                <input
                  id="password"
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                  disabled={isSubmitting}
                  className="w-full px-4 py-3 rounded-xl bg-slate-900/50 border border-slate-700 text-white placeholder-slate-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all disabled:opacity-50"
                  placeholder="••••••••"
                />
              </div>

              <button
                type="submit"
                disabled={isSubmitting}
                className="w-full bg-gradient-to-r from-blue-500 to-indigo-600 text-white font-semibold py-3.5 rounded-xl hover:from-blue-600 hover:to-indigo-700 hover:scale-105 hover:shadow-xl hover:shadow-blue-500/30 transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:scale-100"
              >
                {isSubmitting ? (
                  <span className="flex items-center justify-center gap-2">
                    <svg className="animate-spin h-5 w-5" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                    </svg>
                    Signing in...
                  </span>
                ) : (
                  'Sign In'
                )}
              </button>

              <div className="text-center pt-4">
                <p className="text-sm text-slate-400">
                  Don't have an account?{' '}
                  <a 
                    href="/signup" 
                    className="font-semibold text-blue-400 hover:text-blue-300 transition-colors"
                  >
                    Sign Up
                  </a>
                </p>
              </div>
            </form>
          </AuthCard>
        </div>
      </div>

      {/* Footer */}
      <div className="fixed bottom-0 left-0 right-0 pb-6 text-center">
        <p className="text-slate-400 text-sm">
          © {new Date().getFullYear()} Revalyze B.V.
        </p>
      </div>
    </>
  );
};

export default Login;
