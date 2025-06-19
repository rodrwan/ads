package campaigns

import (
	"ads-system/internal/database"
	"ads-system/internal/web/templates/campaigns"
	"ads-system/internal/web/templates/optimization"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Handler struct {
	db database.Querier
}

func NewHandler(db database.Querier) *Handler {
	return &Handler{db: db}
}

// ShowCampaign muestra los detalles de una campaña específica
func (h *Handler) ShowCampaign(w http.ResponseWriter, r *http.Request) {
	// Obtener advertiser_id de la sesión
	advertiserID, err := getAdvertiserIDFromSession(r)
	if err != nil {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	// Extraer campaign ID de la URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, "ID de campaña requerido", http.StatusBadRequest)
		return
	}
	campaignIDStr := pathParts[2]
	campaignID, err := uuid.Parse(campaignIDStr)
	if err != nil {
		http.Error(w, "ID de campaña inválido", http.StatusBadRequest)
		return
	}

	// Verificar que la campaña pertenece al advertiser
	belongs, err := h.db.CheckCampaignBelongsToAdvertiser(r.Context(), database.CheckCampaignBelongsToAdvertiserParams{
		ID:           pgtype.UUID{Bytes: campaignID, Valid: true},
		AdvertiserID: advertiserID,
	})
	if err != nil || !belongs {
		http.Error(w, "Campaña no encontrada", http.StatusNotFound)
		return
	}

	// Obtener datos de la campaña
	campaign, err := h.db.GetCampaignByID(r.Context(), pgtype.UUID{Bytes: campaignID, Valid: true})
	if err != nil {
		http.Error(w, "Error al obtener campaña", http.StatusInternalServerError)
		return
	}

	// Obtener métricas de la campaña
	metrics, err := h.db.GetCampaignMetrics(r.Context(), pgtype.UUID{Bytes: campaignID, Valid: true})
	if err != nil {
		http.Error(w, "Error al obtener métricas", http.StatusInternalServerError)
		return
	}

	// Obtener gasto de la campaña
	spend, err := h.db.GetCampaignSpend(r.Context(), pgtype.UUID{Bytes: campaignID, Valid: true})
	if err != nil {
		http.Error(w, "Error al obtener gasto", http.StatusInternalServerError)
		return
	}

	// Obtener anuncios de la campaña
	ads, err := h.db.GetAdsByCampaign(r.Context(), pgtype.UUID{Bytes: campaignID, Valid: true})
	if err != nil {
		http.Error(w, "Error al obtener anuncios", http.StatusInternalServerError)
		return
	}

	data := campaigns.CampaignDetailData{
		Campaign: campaign,
		Metrics:  metrics,
		Spend:    spend,
		Ads:      ads,
	}

	component := campaigns.Detail(data)
	component.Render(r.Context(), w)
}

// UpdateCampaign actualiza una campaña existente
func (h *Handler) UpdateCampaign(w http.ResponseWriter, r *http.Request) {
	// Obtener advertiser_id de la sesión
	advertiserID, err := getAdvertiserIDFromSession(r)
	if err != nil {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	// Extraer campaign ID de la URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, "ID de campaña requerido", http.StatusBadRequest)
		return
	}
	campaignIDStr := pathParts[2]
	campaignID, err := uuid.Parse(campaignIDStr)
	if err != nil {
		http.Error(w, "ID de campaña inválido", http.StatusBadRequest)
		return
	}

	// Parsear formulario
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error al procesar formulario", http.StatusBadRequest)
		return
	}

	// Obtener valores del formulario
	name := r.FormValue("name")
	budgetStr := r.FormValue("budget")
	dailyBudgetStr := r.FormValue("daily_budget")
	startDateStr := r.FormValue("start_date")
	endDateStr := r.FormValue("end_date")

	// Convertir budget
	budgetFloat, err := strconv.ParseFloat(budgetStr, 64)
	if err != nil {
		http.Error(w, "Presupuesto inválido", http.StatusBadRequest)
		return
	}
	budget := pgtype.Numeric{Int: big.NewInt(int64(budgetFloat * 100)), Valid: true}

	// Convertir daily_budget (opcional)
	var dailyBudget pgtype.Numeric
	if dailyBudgetStr != "" {
		dailyBudgetFloat, err := strconv.ParseFloat(dailyBudgetStr, 64)
		if err != nil {
			http.Error(w, "Presupuesto diario inválido", http.StatusBadRequest)
			return
		}
		dailyBudget = pgtype.Numeric{Int: big.NewInt(int64(dailyBudgetFloat * 100)), Valid: true}
	}

	// Convertir fechas (opcionales)
	var startDate pgtype.Date
	if startDateStr != "" {
		startTime, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			http.Error(w, "Fecha de inicio inválida", http.StatusBadRequest)
			return
		}
		startDate = pgtype.Date{Time: startTime, Valid: true}
	}

	var endDate pgtype.Date
	if endDateStr != "" {
		endTime, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			http.Error(w, "Fecha de fin inválida", http.StatusBadRequest)
			return
		}
		endDate = pgtype.Date{Time: endTime, Valid: true}
	}

	// Actualizar campaña
	_, err = h.db.UpdateCampaign(r.Context(), database.UpdateCampaignParams{
		ID:           pgtype.UUID{Bytes: campaignID, Valid: true},
		Name:         name,
		Budget:       budget,
		DailyBudget:  dailyBudget,
		StartDate:    startDate,
		EndDate:      endDate,
		Status:       pgtype.Text{String: "active", Valid: true}, // Mantener estado actual
		AdvertiserID: advertiserID,
	})
	if err != nil {
		http.Error(w, "Error al actualizar campaña", http.StatusInternalServerError)
		return
	}

	// Redirigir a la página de detalles
	http.Redirect(w, r, fmt.Sprintf("/campaigns/%s", campaignIDStr), http.StatusSeeOther)
}

// PauseCampaign pausa una campaña
func (h *Handler) PauseCampaign(w http.ResponseWriter, r *http.Request) {
	// Obtener advertiser_id de la sesión
	advertiserID, err := getAdvertiserIDFromSession(r)
	if err != nil {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	// Extraer campaign ID de la URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "ID de campaña requerido", http.StatusBadRequest)
		return
	}
	campaignIDStr := pathParts[2]
	campaignID, err := uuid.Parse(campaignIDStr)
	if err != nil {
		http.Error(w, "ID de campaña inválido", http.StatusBadRequest)
		return
	}

	// Pausar campaña
	_, err = h.db.UpdateCampaignStatus(r.Context(), database.UpdateCampaignStatusParams{
		ID:           pgtype.UUID{Bytes: campaignID, Valid: true},
		Status:       pgtype.Text{String: "paused", Valid: true},
		AdvertiserID: advertiserID,
	})
	if err != nil {
		http.Error(w, "Error al pausar campaña", http.StatusInternalServerError)
		return
	}

	// Responder con el nuevo estado
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<span class="text-yellow-400">Pausada</span>`))
}

// ActivateCampaign activa una campaña
func (h *Handler) ActivateCampaign(w http.ResponseWriter, r *http.Request) {
	// Obtener advertiser_id de la sesión
	advertiserID, err := getAdvertiserIDFromSession(r)
	if err != nil {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	// Extraer campaign ID de la URL
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "ID de campaña requerido", http.StatusBadRequest)
		return
	}
	campaignIDStr := pathParts[2]
	campaignID, err := uuid.Parse(campaignIDStr)
	if err != nil {
		http.Error(w, "ID de campaña inválido", http.StatusBadRequest)
		return
	}

	// Activar campaña
	_, err = h.db.UpdateCampaignStatus(r.Context(), database.UpdateCampaignStatusParams{
		ID:           pgtype.UUID{Bytes: campaignID, Valid: true},
		Status:       pgtype.Text{String: "active", Valid: true},
		AdvertiserID: advertiserID,
	})
	if err != nil {
		http.Error(w, "Error al activar campaña", http.StatusInternalServerError)
		return
	}

	// Responder con el nuevo estado
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<span class="text-green-400">Activa</span>`))
}

// getAdvertiserIDFromSession extrae el advertiser_id de la sesión
func getAdvertiserIDFromSession(r *http.Request) (pgtype.UUID, error) {
	// Obtener token de la cookie
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return pgtype.UUID{}, err
	}

	_, err = uuid.Parse(cookie.Value)
	if err != nil {
		return pgtype.UUID{}, err
	}

	// Aquí deberías obtener el advertiser_id de la sesión
	// Por ahora, usaremos un ID hardcodeado para pruebas
	return pgtype.UUID{Bytes: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"), Valid: true}, nil
}

// Helper para convertir interface{} a float64
func toFloat64(val interface{}) float64 {
	if val == nil {
		return 0
	}
	switch v := val.(type) {
	case float64:
		return v
	case int64:
		return float64(v)
	case int:
		return float64(v)
	case string:
		f, _ := strconv.ParseFloat(v, 64)
		return f
	case *big.Float:
		f, _ := v.Float64()
		return f
	case *big.Int:
		return float64(v.Int64())
	}
	return 0
}

// Handler para la página de optimización de presupuesto
func (h *Handler) Optimization(w http.ResponseWriter, r *http.Request) {
	advertiserID, err := getAdvertiserIDFromSession(r)
	if err != nil {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	// Obtener datos de análisis y recomendaciones
	campaignAnalysis, _ := h.db.GetCampaignPerformanceAnalysis(r.Context(), advertiserID)
	adAnalysis, _ := h.db.GetAdPerformanceAnalysis(r.Context(), advertiserID)
	budgetAlerts, _ := h.db.GetBudgetAlerts(r.Context(), advertiserID)
	recommendations, _ := h.db.GetOptimizationRecommendations(r.Context(), advertiserID)

	// Convertir los campos interface{} a float64 para el template
	for i := range campaignAnalysis {
		if f := toFloat64(campaignAnalysis[i].TotalSpend); f > 0 {
			campaignAnalysis[i].TotalSpend = f / 100 // Asumimos centavos
		} else {
			campaignAnalysis[i].TotalSpend = float64(0)
		}
	}
	for i := range adAnalysis {
		if f := toFloat64(adAnalysis[i].TotalSpend); f > 0 {
			adAnalysis[i].TotalSpend = f / 100
		} else {
			adAnalysis[i].TotalSpend = float64(0)
		}
	}
	for i := range budgetAlerts {
		if f := toFloat64(budgetAlerts[i].CurrentSpend); f > 0 {
			budgetAlerts[i].CurrentSpend = f / 100
		} else {
			budgetAlerts[i].CurrentSpend = float64(0)
		}
	}
	for i := range recommendations {
		if f := toFloat64(recommendations[i].TotalSpend); f > 0 {
			recommendations[i].TotalSpend = f / 100
		} else {
			recommendations[i].TotalSpend = float64(0)
		}
	}

	data := struct {
		CampaignAnalysis []database.GetCampaignPerformanceAnalysisRow
		AdAnalysis       []database.GetAdPerformanceAnalysisRow
		BudgetAlerts     []database.GetBudgetAlertsRow
		Recommendations  []database.GetOptimizationRecommendationsRow
	}{
		CampaignAnalysis: campaignAnalysis,
		AdAnalysis:       adAnalysis,
		BudgetAlerts:     budgetAlerts,
		Recommendations:  recommendations,
	}

	component := optimization.Optimization(data)
	component.Render(r.Context(), w)
}
