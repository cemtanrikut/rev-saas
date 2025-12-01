import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import AuthCard from '../components/AuthCard';

const Login = () => {
  const navigate = useNavigate();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');

  const handleSubmit = (e) => {
    e.preventDefault();
    console.log('Login attempted:', { email, password });
    // TODO: API call here
    
    // Redirect to dashboard
    navigate('/app/overview');
  };

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
              className="w-full px-4 py-3 rounded-xl bg-slate-900/50 border border-slate-700 text-white placeholder-slate-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all"
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
              className="w-full px-4 py-3 rounded-xl bg-slate-900/50 border border-slate-700 text-white placeholder-slate-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 outline-none transition-all"
              placeholder="••••••••"
            />
          </div>

          <button
            type="submit"
            className="w-full bg-gradient-to-r from-blue-500 to-indigo-600 text-white font-semibold py-3.5 rounded-xl hover:from-blue-600 hover:to-indigo-700 hover:scale-105 hover:shadow-xl hover:shadow-blue-500/30 transition-all duration-200"
          >
            Sign In
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

