package web

import (
	"context"
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	"huruf-genetik/services"
)

type WebHandler struct {
	engine *services.EngineService
	tmpl   *template.Template
}

func NewWebHandler(engine *services.EngineService) *WebHandler {
	tmpl := template.Must(template.ParseFiles("web/templates/index.html"))
	return &WebHandler{
		engine: engine,
		tmpl:   tmpl,
	}
}

func (h *WebHandler) ServeHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	h.tmpl.Execute(w, nil)
}

type AnalyzeRequest struct {
	Text string `json:"text"`
}

func (h *WebHandler) HandleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.engine.ProcessText(r.Context(), req.Text)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *WebHandler) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.engine.GetDashboardMetrics(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func (h *WebHandler) HandleImportQuran(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	go func() {
		err := h.engine.ProcessQuranImport(context.Background())
		if err != nil {
			log.Printf("Background Import Error: %v\n", err)
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "Import started in background. Please check server logs."})
}

func (h *WebHandler) HandleAlemHeatmap(w http.ResponseWriter, r *http.Request) {
	matrix, err := h.engine.GetAlemTransitionMatrix(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(matrix)
}

func (h *WebHandler) HandleRootJourney(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	journey, err := h.engine.GetRootWordJourney(r.Context(), req.Text)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(journey)
}

func (h *WebHandler) HandleEsmaCorrelation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	correlations, err := h.engine.CalculateEsmaCorrelation(r.Context(), req.Text)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(correlations)
}

func (h *WebHandler) HandleEsmaMatrix(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	matrix, err := h.engine.CalculateEsmaMatrix(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(matrix)
}
