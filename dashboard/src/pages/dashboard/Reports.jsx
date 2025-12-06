import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAnalysis } from '../../context/AnalysisContext';
import { analysisApi, downloadBlob } from '../../lib/apiClient';

const Reports = () => {
  const navigate = useNavigate();
  const { analyses, isLoading } = useAnalysis();
  
  // PDF export state
  const [exportingId, setExportingId] = useState(null);
  const [exportError, setExportError] = useState(null);
  const [exportSuccess, setExportSuccess] = useState(null);

  // Handle PDF export
  const handleExportPdf = async (analysisId, analysisDate) => {
    setExportingId(analysisId);
    setExportError(null);
    setExportSuccess(null);
    
    try {
      const { ok, blob } = await analysisApi.exportPdf(analysisId);
      
      if (ok && blob) {
        const dateStr = new Date(analysisDate).toISOString().split('T')[0];
        const filename = `pricing-report-${dateStr}.pdf`;
        downloadBlob(blob, filename);
        setExportSuccess(analysisId);
        setTimeout(() => setExportSuccess(null), 3000);
      }
    } catch (error) {
      console.error('Failed to export PDF:', error);
      setExportError(error.message || 'Failed to export PDF');
      setTimeout(() => setExportError(null), 5000);
    } finally {
      setExportingId(null);
    }
  };

  // Format date
  const formatDate = (isoString) => {
    const date = new Date(isoString);
    return new Intl.DateTimeFormat('en-US', {
      month: 'long',
      day: 'numeric',
      year: 'numeric',
    }).format(date);
  };

  const formatDateShort = (isoString) => {
    const date = new Date(isoString);
    return new Intl.DateTimeFormat('en-US', {
      month: 'short',
      day: '2-digit',
      year: 'numeric',
    }).format(date);
  };

  const formatTime = (isoString) => {
    const date = new Date(isoString);
    return new Intl.DateTimeFormat('en-US', {
      hour: 'numeric',
      minute: '2-digit',
    }).format(date);
  };

  // Get action counts
  const getActionCounts = (recommendations = []) => {
    const raiseCount = recommendations.filter(r => 
      r.suggestedAction === 'raise_price' || 
      r.suggestedAction === 'raise_price_conservative' || 
      r.suggestedAction === 'raise_price_aggressive'
    ).length;
    const lowerCount = recommendations.filter(r => r.suggestedAction === 'lower_price').length;
    const keepCount = recommendations.filter(r => 
      r.suggestedAction === 'keep' || r.suggestedAction === 'keep_for_growth'
    ).length;
    return { raiseCount, lowerCount, keepCount };
  };

  // Loading state
  if (isLoading) {
    return (
      <div className="max-w-6xl mx-auto">
        <div className="text-center py-20">
          <div className="w-12 h-12 border-4 border-blue-500 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
          <p className="text-slate-400">Loading reports...</p>
        </div>
      </div>
    );
  }

  // Empty state
  if (!analyses || analyses.length === 0) {
    return (
      <div className="max-w-6xl mx-auto">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-2xl font-bold text-white mb-2">Reports</h1>
          <p className="text-slate-400">Download and manage your pricing analysis reports</p>
        </div>

        <div className="bg-slate-900/50 backdrop-blur-sm rounded-2xl p-12 border border-slate-800 text-center">
          <div className="w-20 h-20 bg-slate-800 rounded-2xl flex items-center justify-center mx-auto mb-6">
            <svg className="w-10 h-10 text-slate-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
          </div>
          <h2 className="text-2xl font-bold text-white mb-3">No reports yet</h2>
          <p className="text-slate-400 mb-8 max-w-md mx-auto leading-relaxed">
            Run your first pricing analysis to generate a report. Once you have analysis data, you can download professional PDF reports from here.
          </p>
          <button
            onClick={() => navigate('/app/overview')}
            className="px-6 py-3 bg-gradient-to-r from-blue-500 to-indigo-600 text-white rounded-xl font-semibold hover:from-blue-600 hover:to-indigo-700 hover:scale-105 transition-all shadow-lg shadow-blue-500/20"
          >
            Go to Overview
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-6xl mx-auto space-y-6">
      {/* Error Toast */}
      {exportError && (
        <div className="fixed bottom-6 right-6 z-50 animate-fade-in-up">
          <div className="flex items-center gap-3 px-4 py-3 bg-red-500/10 border border-red-500/30 rounded-xl shadow-lg backdrop-blur-sm">
            <svg className="w-5 h-5 text-red-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <span className="text-sm text-red-300">{exportError}</span>
            <button 
              onClick={() => setExportError(null)}
              className="text-red-400 hover:text-red-300 ml-2"
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>
      )}

      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white mb-2">Reports</h1>
          <p className="text-slate-400">Download and manage your pricing analysis reports</p>
        </div>
        <div className="flex items-center gap-3">
          <span className="text-sm text-slate-400 bg-slate-800 px-4 py-2 rounded-lg">
            {analyses.length} report{analyses.length !== 1 ? 's' : ''} available
          </span>
          <button
            onClick={() => navigate('/app/analyses')}
            className="flex items-center gap-2 px-4 py-2 bg-blue-500/10 text-blue-400 rounded-lg font-medium hover:bg-blue-500/20 transition-all border border-blue-500/20"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
            </svg>
            Run New Analysis
          </button>
        </div>
      </div>

      {/* Info Card */}
      <div className="bg-gradient-to-r from-blue-500/5 to-indigo-500/5 border border-blue-500/20 rounded-xl p-4 flex items-start gap-4">
        <div className="w-10 h-10 bg-blue-500/10 rounded-lg flex items-center justify-center flex-shrink-0">
          <svg className="w-5 h-5 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </div>
        <div>
          <h3 className="text-sm font-semibold text-white mb-1">PDF Reports</h3>
          <p className="text-sm text-slate-400 leading-relaxed">
            Each report includes your business snapshot, executive summary, and detailed pricing recommendations. 
            Perfect for sharing with your team or stakeholders.
          </p>
        </div>
      </div>

      {/* Reports Table */}
      <div className="bg-slate-900/50 backdrop-blur-sm rounded-2xl border border-slate-800 overflow-hidden">
        {/* Table Header */}
        <div className="grid grid-cols-12 gap-4 px-6 py-4 bg-slate-800/50 border-b border-slate-800">
          <div className="col-span-4">
            <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Report Date</span>
          </div>
          <div className="col-span-2">
            <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Data</span>
          </div>
          <div className="col-span-3">
            <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Insights</span>
          </div>
          <div className="col-span-3 text-right">
            <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Actions</span>
          </div>
        </div>

        {/* Table Body */}
        <div className="divide-y divide-slate-800/50">
          {analyses.map((analysis, index) => {
            const { raiseCount, lowerCount, keepCount } = getActionCounts(analysis.recommendations);
            const isExporting = exportingId === analysis.id;
            const justExported = exportSuccess === analysis.id;
            const isLatest = index === 0;

            return (
              <div
                key={analysis.id}
                className="grid grid-cols-12 gap-4 px-6 py-5 items-center hover:bg-slate-800/20 transition-colors group"
              >
                {/* Report Date */}
                <div className="col-span-4">
                  <div className="flex items-center gap-3">
                    <div className={`w-10 h-10 rounded-xl flex items-center justify-center flex-shrink-0 ${
                      isLatest ? 'bg-blue-500/10' : 'bg-slate-800'
                    }`}>
                      <svg className={`w-5 h-5 ${isLatest ? 'text-blue-400' : 'text-slate-500'}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                      </svg>
                    </div>
                    <div>
                      <div className="flex items-center gap-2">
                        <span className="text-sm font-semibold text-white">
                          {formatDate(analysis.createdAt)}
                        </span>
                        {isLatest && (
                          <span className="text-xs font-medium text-blue-400 bg-blue-500/10 px-2 py-0.5 rounded-full">
                            Latest
                          </span>
                        )}
                      </div>
                      <span className="text-xs text-slate-500">
                        {formatTime(analysis.createdAt)}
                      </span>
                    </div>
                  </div>
                </div>

                {/* Data (Plans & Competitors) */}
                <div className="col-span-2">
                  <div className="flex items-center gap-4">
                    <div className="text-center">
                      <div className="text-lg font-bold text-white">{analysis.numPlans}</div>
                      <div className="text-xs text-slate-500">Plans</div>
                    </div>
                    <div className="w-px h-8 bg-slate-700"></div>
                    <div className="text-center">
                      <div className="text-lg font-bold text-white">{analysis.numCompetitors}</div>
                      <div className="text-xs text-slate-500">Competitors</div>
                    </div>
                  </div>
                </div>

                {/* Insights */}
                <div className="col-span-3">
                  <div className="flex items-center gap-2 flex-wrap">
                    {raiseCount > 0 && (
                      <span className="inline-flex items-center gap-1 text-xs font-medium text-emerald-400 bg-emerald-500/10 px-2.5 py-1 rounded-lg">
                        <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 10l7-7m0 0l7 7m-7-7v18" />
                        </svg>
                        {raiseCount} Raise
                      </span>
                    )}
                    {lowerCount > 0 && (
                      <span className="inline-flex items-center gap-1 text-xs font-medium text-amber-400 bg-amber-500/10 px-2.5 py-1 rounded-lg">
                        <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 14l-7 7m0 0l-7-7m7 7V3" />
                        </svg>
                        {lowerCount} Lower
                      </span>
                    )}
                    {keepCount > 0 && (
                      <span className="inline-flex items-center gap-1 text-xs font-medium text-blue-400 bg-blue-500/10 px-2.5 py-1 rounded-lg">
                        <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 12h14" />
                        </svg>
                        {keepCount} Keep
                      </span>
                    )}
                    {raiseCount === 0 && lowerCount === 0 && keepCount === 0 && (
                      <span className="text-xs text-slate-500">No recommendations</span>
                    )}
                  </div>
                </div>

                {/* Actions */}
                <div className="col-span-3">
                  <div className="flex items-center justify-end gap-2">
                    {/* View Details */}
                    <button
                      onClick={() => navigate('/app/analyses')}
                      className="flex items-center gap-1.5 px-3 py-2 text-xs font-medium text-slate-400 hover:text-white hover:bg-slate-800 rounded-lg transition-all"
                    >
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                      </svg>
                      View
                    </button>

                    {/* Download PDF */}
                    <button
                      onClick={() => handleExportPdf(analysis.id, analysis.createdAt)}
                      disabled={isExporting}
                      className={`flex items-center gap-1.5 px-4 py-2 text-xs font-semibold rounded-lg transition-all ${
                        justExported
                          ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/30'
                          : isExporting
                          ? 'bg-slate-800 text-slate-500 cursor-not-allowed'
                          : 'bg-blue-500/10 text-blue-400 hover:bg-blue-500/20 border border-blue-500/20'
                      }`}
                    >
                      {isExporting ? (
                        <>
                          <svg className="animate-spin w-4 h-4" viewBox="0 0 24 24">
                            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                          </svg>
                          Exporting...
                        </>
                      ) : justExported ? (
                        <>
                          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                          </svg>
                          Downloaded!
                        </>
                      ) : (
                        <>
                          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                          </svg>
                          Download PDF
                        </>
                      )}
                    </button>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      </div>

      {/* Footer Info */}
      <div className="flex items-center justify-between text-xs text-slate-500 px-2">
        <span>Reports are generated from your pricing analysis data</span>
        <span>PDF files include full recommendations and rationale</span>
      </div>
    </div>
  );
};

export default Reports;
