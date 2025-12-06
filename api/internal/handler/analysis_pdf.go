package handler

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"rev-saas-api/internal/middleware"
	"rev-saas-api/internal/service"
	mongorepo "rev-saas-api/internal/repository/mongo"
)

// AnalysisPDFHandler handles PDF export of analysis reports.
type AnalysisPDFHandler struct {
	analysisService     *service.AnalysisService
	businessMetricsRepo *mongorepo.BusinessMetricsRepository
}

// NewAnalysisPDFHandler creates a new AnalysisPDFHandler.
func NewAnalysisPDFHandler(
	analysisService *service.AnalysisService,
	businessMetricsRepo *mongorepo.BusinessMetricsRepository,
) *AnalysisPDFHandler {
	return &AnalysisPDFHandler{
		analysisService:     analysisService,
		businessMetricsRepo: businessMetricsRepo,
	}
}

// ExportPDF handles GET /api/analysis/{id}/export-pdf
func (h *AnalysisPDFHandler) ExportPDF(w http.ResponseWriter, r *http.Request) {
	// Get user ID from JWT context
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Get analysis ID from URL
	vars := mux.Vars(r)
	analysisID := vars["id"]
	if analysisID == "" {
		http.Error(w, "analysis ID is required", http.StatusBadRequest)
		return
	}

	// Fetch the analysis (ensures it belongs to the user)
	analysis, err := h.analysisService.GetAnalysis(r.Context(), userID, analysisID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to fetch analysis: %v", err), http.StatusInternalServerError)
		return
	}
	if analysis == nil {
		http.Error(w, "analysis not found", http.StatusNotFound)
		return
	}

	// Fetch business metrics (optional - may be nil)
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		http.Error(w, "invalid user ID", http.StatusBadRequest)
		return
	}
	
	metrics, err := h.businessMetricsRepo.GetByUserID(r.Context(), uid)
	if err != nil {
		// Log error but continue - metrics are optional
		metrics = nil
	}

	// Generate PDF
	pdfData := service.PDFExportData{
		Analysis: analysis,
		Metrics:  metrics,
	}

	pdfBuffer, err := service.GenerateAnalysisPDF(pdfData)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to generate PDF: %v", err), http.StatusInternalServerError)
		return
	}

	// Generate filename
	filename := fmt.Sprintf("pricing-report-%s.pdf", analysis.CreatedAt.Format("2006-01-02"))

	// Set headers for PDF download
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", pdfBuffer.Len()))

	// Write PDF to response
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(pdfBuffer.Bytes())
}

