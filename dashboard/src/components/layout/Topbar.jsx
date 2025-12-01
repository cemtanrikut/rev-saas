import { useLocation } from 'react-router-dom';

const Topbar = () => {
  const location = useLocation();

  const getPageTitle = () => {
    const path = location.pathname;
    if (path.includes('overview')) return 'Overview';
    if (path.includes('analyses')) return 'Analyses';
    if (path.includes('competitors')) return 'Competitors';
    if (path.includes('reports')) return 'Reports';
    if (path.includes('settings')) return 'Settings';
    return 'Dashboard';
  };

  return (
    <div className="h-16 bg-slate-900/50 backdrop-blur-sm border-b border-slate-800 flex items-center justify-between px-6 sticky top-0 z-10">
      {/* Page Title */}
      <div>
        <h1 className="text-xl font-semibold text-white">
          {getPageTitle()}
        </h1>
      </div>

      {/* Right side */}
      <div className="flex items-center gap-4">
        {/* Upgrade Button */}
        <button className="px-4 py-2 bg-gradient-to-r from-blue-500 to-indigo-600 text-white rounded-lg font-medium hover:from-blue-600 hover:to-indigo-700 transition-all hover:scale-105 text-sm shadow-lg shadow-blue-500/20">
          Upgrade
        </button>

        {/* User Avatar */}
        <div className="w-10 h-10 bg-gradient-to-br from-blue-500 to-indigo-600 rounded-full flex items-center justify-center text-white font-semibold text-sm shadow-lg">
          CT
        </div>
      </div>
    </div>
  );
};

export default Topbar;

