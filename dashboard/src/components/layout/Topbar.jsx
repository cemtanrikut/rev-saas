import { useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';
import { useAiCredits } from '../../hooks/useAiCredits';

// AI Credits Indicator Component
const AICreditsIndicator = ({ credits, loading, onClick }) => {
  if (loading) {
    return (
      <div className="flex items-center gap-2 px-4 py-2 bg-slate-800/50 rounded-xl border border-slate-700/50">
        <div className="w-4 h-4 border-2 border-slate-500 border-t-violet-400 rounded-full animate-spin"></div>
        <span className="text-xs text-slate-400">Loading...</span>
      </div>
    );
  }

  if (!credits) return null;

  // Support both snake_case (from API) and camelCase
  const remainingCredits = credits.remaining_credits ?? credits.remainingCredits ?? 0;
  const monthlyCredits = credits.monthly_credits ?? credits.monthlyCredits ?? 0;
  const percentage = monthlyCredits > 0 ? (remainingCredits / monthlyCredits) * 100 : 0;

  const isLow = remainingCredits <= 2;
  const isEmpty = remainingCredits === 0;

  return (
    <button
      onClick={onClick}
      className={`group flex items-center gap-3 px-4 py-2.5 rounded-xl border transition-all duration-200 hover:scale-[1.02] ${
        isEmpty
          ? 'bg-red-500/10 border-red-500/30 hover:border-red-500/50 hover:bg-red-500/20'
          : isLow
          ? 'bg-amber-500/10 border-amber-500/30 hover:border-amber-500/50 hover:bg-amber-500/20'
          : 'bg-slate-800/50 border-slate-700/50 hover:border-violet-500/50 hover:bg-slate-800'
      }`}
      title={isEmpty ? 'Upgrade your plan to get more AI Insight Credits' : 'AI Insight Credits remaining this month'}
    >
      {/* Sparkle icon */}
      <div className={`w-8 h-8 rounded-lg flex items-center justify-center ${
        isEmpty ? 'bg-red-500/20' : isLow ? 'bg-amber-500/20' : 'bg-violet-500/20'
      }`}>
        <svg className={`w-4 h-4 ${isEmpty ? 'text-red-400' : isLow ? 'text-amber-400' : 'text-violet-400'}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
        </svg>
      </div>
      
      <div className="flex flex-col items-start">
        <span className={`text-xs font-semibold ${isEmpty ? 'text-red-400' : isLow ? 'text-amber-400' : 'text-white'}`}>
          {remainingCredits} credits
        </span>
        <div className="w-16 h-1.5 bg-slate-700 rounded-full overflow-hidden mt-1">
          <div 
            className={`h-full rounded-full transition-all ${
              isEmpty ? 'bg-red-500' : isLow ? 'bg-amber-500' : 'bg-gradient-to-r from-violet-500 to-fuchsia-500'
            }`}
            style={{ width: `${percentage}%` }}
          />
        </div>
      </div>

      {isEmpty && (
        <span className="text-xs text-red-400 font-medium ml-1">Upgrade</span>
      )}
    </button>
  );
};

const Topbar = () => {
  const location = useLocation();
  const navigate = useNavigate();
  const { isAdmin, user } = useAuth();
  const { credits, loading } = useAiCredits();

  const getPageInfo = () => {
    const path = location.pathname;
    if (path.includes('overview')) return { title: 'Overview', subtitle: 'Your pricing intelligence at a glance' };
    if (path.includes('analyses')) return { title: 'Analyses', subtitle: 'AI-powered pricing insights' };
    if (path.includes('simulation')) return { title: 'Pricing Simulation', subtitle: 'Test pricing scenarios' };
    if (path.includes('competitors')) return { title: 'Competitors', subtitle: 'Track competitive landscape' };
    if (path.includes('plans')) return { title: 'My Pricing', subtitle: 'Manage your pricing plans' };
    if (path.includes('reports')) return { title: 'Reports', subtitle: 'Export and share insights' };
    if (path.includes('settings')) return { title: 'Settings', subtitle: 'Configure your workspace' };
    if (path.includes('billing')) return { title: 'Billing', subtitle: 'Manage subscription' };
    return { title: 'Dashboard', subtitle: 'Welcome back' };
  };

  // Get user initials for avatar
  const getUserInitials = () => {
    if (user?.full_name) {
      const names = user.full_name.split(' ');
      if (names.length >= 2) {
        return (names[0][0] + names[1][0]).toUpperCase();
      }
      return names[0].substring(0, 2).toUpperCase();
    }
    if (user?.email) {
      return user.email.substring(0, 2).toUpperCase();
    }
    return 'U';
  };

  const pageInfo = getPageInfo();

  return (
    <div className="h-20 bg-slate-950/80 backdrop-blur-xl border-b border-slate-800/50 flex items-center justify-between px-8 sticky top-0 z-10">
      {/* Page Title */}
      <div>
        <h1 className="text-2xl font-bold text-white">
          {pageInfo.title}
        </h1>
        <p className="text-sm text-slate-500">{pageInfo.subtitle}</p>
      </div>

      {/* Right side */}
      <div className="flex items-center gap-4">
        {/* AI Credits Indicator */}
        {!isAdmin && (
          <AICreditsIndicator 
            credits={credits} 
            loading={loading}
            onClick={() => navigate('/app/billing')}
          />
        )}

        {/* Upgrade Button - hidden for admin users */}
        {!isAdmin && (
          <button 
            onClick={() => navigate('/app/billing')}
            className="px-5 py-2.5 bg-gradient-to-r from-violet-500 to-fuchsia-600 text-white rounded-xl font-semibold hover:from-violet-600 hover:to-fuchsia-700 transition-all duration-200 hover:scale-[1.02] text-sm shadow-lg shadow-violet-500/25"
          >
            Upgrade Plan
          </button>
        )}

        {/* Divider */}
        <div className="h-10 w-px bg-slate-800" />

        {/* User Avatar */}
        <div className="flex items-center gap-3">
          <div className="w-11 h-11 bg-gradient-to-br from-violet-500 to-fuchsia-600 rounded-xl flex items-center justify-center text-white font-bold text-sm shadow-lg shadow-violet-500/20 ring-2 ring-slate-800">
            {getUserInitials()}
          </div>
        </div>
      </div>
    </div>
  );
};

export default Topbar;
