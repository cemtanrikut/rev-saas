const Reports = () => {
  return (
    <div className="max-w-7xl mx-auto">
      <div className="bg-slate-900/50 backdrop-blur-sm rounded-2xl p-12 border border-slate-800 text-center">
        <div className="w-16 h-16 bg-blue-500/10 rounded-2xl flex items-center justify-center mx-auto mb-6">
          <svg className="w-8 h-8 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
        </div>
        <h2 className="text-2xl font-bold text-white mb-3">
          Reports
        </h2>
        <p className="text-slate-400 mb-6">
          AI-generated pricing reports coming soon
        </p>
        <div className="inline-block px-4 py-1.5 bg-blue-500/10 border border-blue-500/30 rounded-full text-sm font-medium text-blue-400">
          Coming Soon
        </div>
      </div>
    </div>
  );
};

export default Reports;

