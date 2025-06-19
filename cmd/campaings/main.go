package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"time"

	"ads-system/internal/api"
	"ads-system/internal/database"
	"ads-system/internal/logger"
	"ads-system/internal/web/templates"
	adsView "ads-system/internal/web/templates/ads"
	"ads-system/internal/web/templates/campaigns"
	campaignsView "ads-system/internal/web/templates/campaigns"
	"ads-system/internal/web/templates/optimization"
	"ads-system/internal/web/templates/session"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rodrwan/secretly/pkg/secretly"
)

func main() {
	envs := secretly.New(
		secretly.WithBaseURL("http://environment:9000"),
	)
	if err := envs.LoadToEnvironment("campaigns"); err != nil {
		log.Fatalf("Error al cargar el archivo .env: %v", err)
	}

	// Configurar la conexión a la base de datos
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/themenu?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("No se pudo conectar a la base de datos: %v", err)
	}
	defer pool.Close()

	// Configurar Redis
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}

	// Crear el cliente de base de datos
	db := database.New(pool)

	// Crear el WebSocket manager
	wsManager := api.NewWebSocketManager(db)
	go wsManager.Start()
	wsManager.StartMetricsBroadcaster()

	r := gin.Default()

	r.Static("/static", "./static")

	r.GET("/", func(c *gin.Context) {
		http.Redirect(c.Writer, c.Request, "/campaigns", http.StatusSeeOther)
	})

	r.GET("/login", RenderTempl(session.LoginForm()))
	r.POST("/login", HandleLogin(db))
	r.GET("/logout", HandleLogout(db))

	r.GET("/campaigns", RequireAuth(db), ListCampaignsPage(db))
	r.GET("/campaigns/create", RequireAuth(db), RenderTempl(campaigns.CreateForm()))
	r.POST("/campaigns/create", RequireAuth(db), HandleCreateCampaign(db))

	// Nuevas rutas para gestión de campañas
	r.GET("/campaigns/:id", RequireAuth(db), ShowCampaignDetails(db))
	r.PUT("/campaigns/:id", RequireAuth(db), HandleUpdateCampaign(db))
	r.POST("/campaigns/:id/pause", RequireAuth(db), HandlePauseCampaign(db))
	r.POST("/campaigns/:id/activate", RequireAuth(db), HandleActivateCampaign(db))

	r.GET("/campaigns/:id/ads", RequireAuth(db), ListAdsPage(db))
	r.GET("/campaigns/:id/ads/create", RequireAuth(db), RenderTempl(adsView.CreateForm()))
	r.POST("/campaigns/:id/ads/create", RequireAuth(db), HandleCreateAd(db))

	r.GET("/dashboard", RequireAuth(db), ShowDashboard(db))

	// Ruta de optimización de presupuesto
	r.GET("/optimization", RequireAuth(db), ShowOptimizationPage(db))

	// WebSocket endpoint para actualizaciones en tiempo real
	r.GET("/ws", wsManager.WebSocketHandler())

	// page to see all ads
	r.GET("/ads", RequireAuth(db), ListAdsPagePublic(db))
	r.POST("/ads/simulate/impression", RequireAuth(db), HandleSimulateImpression(db))
	r.POST("/ads/simulate/click", RequireAuth(db), HandleSimulateClick(db))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	r.Run(":" + port)
}

func RenderTempl(component templ.Component) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Status(http.StatusOK)
		component.Render(c.Request.Context(), c.Writer)
	}
}

func HandleCreateCampaign(db database.Querier) func(c *gin.Context) {
	return func(c *gin.Context) {
		name := c.PostForm("name")
		budget := c.PostForm("budget")

		var id pgtype.UUID
		_ = id.Scan(uuid.New().String())
		// validar y parsear budget a float
		budgetFloat, err := strconv.ParseFloat(budget, 64)
		if err != nil {
			logger.Error("Presupuesto inválido", "error", err)
			c.String(http.StatusBadRequest, "Presupuesto inválido")
			return
		}

		// Convertir float a int (multiplicando por 100 para preservar 2 decimales)
		budgetInt := int64(budgetFloat * 100)

		_, err = db.CreateCampaign(c.Request.Context(), database.CreateCampaignParams{
			ID:           id,
			AdvertiserID: pgtype.UUID{Bytes: [16]byte(uuid.MustParse("hardcoded-advertiser")), Valid: true},
			Name:         name,
			Budget:       pgtype.Numeric{Int: big.NewInt(budgetInt), Valid: true},
		})
		if err != nil {
			logger.Error("Error al crear campaña", "error", err)
			c.String(http.StatusInternalServerError, "Error al crear campaña")
			return
		}

		c.Redirect(http.StatusSeeOther, "/campaigns")
	}
}

func HandleLogin(db database.Querier) func(c *gin.Context) {
	return func(c *gin.Context) {
		email := c.PostForm("email")

		advertiser, err := db.GetAdvertiserByEmail(c.Request.Context(), email)
		if err != nil {
			logger.Info("Error al obtener el anunciante", "error", err)
			c.String(http.StatusInternalServerError, "Error al obtener el anunciante")
			return
		}

		token := pgtype.UUID{Bytes: [16]byte(uuid.MustParse(uuid.New().String())), Valid: true}
		_, err = db.CreateSession(c.Request.Context(), database.CreateSessionParams{
			Token:        token,
			AdvertiserID: advertiser.ID,
			ExpiresAt:    pgtype.Timestamp{Time: time.Now().Add(7 * 24 * time.Hour), Valid: true},
		})
		if err != nil {
			logger.Info("Error al crear sesión", "error", err)
			c.String(http.StatusInternalServerError, "Error al crear sesión")
			return
		}

		logger.Info("Sesión creada", "token", token)
		c.SetCookie("session_token", token.String(), 3600*24*7, "/", "", false, true)
		c.Redirect(http.StatusSeeOther, "/campaigns")
	}
}

func HandleLogout(db database.Querier) func(c *gin.Context) {
	return func(c *gin.Context) {
		cookie, _ := c.Cookie("session_token")
		err := db.DeleteSession(c.Request.Context(), pgtype.UUID{Bytes: [16]byte(uuid.MustParse(cookie)), Valid: true})
		if err != nil {
			logger.Error("Error al eliminar sesión", "error", err)
			c.String(http.StatusInternalServerError, "Error al eliminar sesión")
			return
		}
		c.SetCookie("session_token", "", -1, "/", "", false, true)
		c.Redirect(http.StatusSeeOther, "/login")
	}
}

func RequireAuth(db database.Querier) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("session_token")
		if err != nil {
			logger.Info("Error al obtener sesión desde cookie", "error", err)
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}

		// transformar token a pgtype.UUID
		tokenUUID := pgtype.UUID{}
		tokenUUID.Scan(token)
		// Buscar en DB
		advertiserID, err := db.GetSessionByToken(c.Request.Context(), tokenUUID)
		if err != nil {
			logger.Info("Error al obtener sesión desde DB", "error", err)
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}

		logger.Info("Sesión encontrada advertiser_id ", advertiserID)

		c.Set("advertiser_id", advertiserID)
		c.Next()
	}
}

func HandleCreateAd(db database.Querier) func(c *gin.Context) {
	return func(c *gin.Context) {
		campaignID := c.Param("id")
		advertiserID := c.MustGet("advertiser_id").(pgtype.UUID)

		campaignIDUUID := pgtype.UUID{Bytes: [16]byte(uuid.MustParse(campaignID)), Valid: true}
		// Validar que esta campaña pertenece al anunciante
		valid, err := db.CheckCampaignBelongsToAdvertiser(c.Request.Context(), database.CheckCampaignBelongsToAdvertiserParams{
			ID:           campaignIDUUID,
			AdvertiserID: advertiserID,
		})
		if err != nil || !valid {
			logger.Error("Campaña no válida o no te pertenece", "error", err)
			c.String(http.StatusUnauthorized, "Campaña no válida o no te pertenece")
			return
		}

		title := c.PostForm("title")
		description := c.PostForm("description")
		imageURL := c.PostForm("image_url")
		destinationURL := c.PostForm("destination_url")
		ctaLabel := c.PostForm("cta_label")
		country := c.PostForm("country")

		targetingJSON := fmt.Sprintf(`{"country": "%s"}`, country)

		_, err = db.CreateAd(c.Request.Context(), database.CreateAdParams{
			ID:             pgtype.UUID{Bytes: [16]byte(uuid.MustParse(uuid.New().String())), Valid: true},
			CampaignID:     campaignIDUUID,
			Title:          pgtype.Text{String: title, Valid: true},
			Description:    pgtype.Text{String: description, Valid: true},
			ImageUrl:       pgtype.Text{String: imageURL, Valid: true},
			DestinationUrl: pgtype.Text{String: destinationURL, Valid: true},
			CtaLabel:       pgtype.Text{String: ctaLabel, Valid: true},
			TargetingJson:  []byte(targetingJSON),
			Status:         pgtype.Text{String: "active", Valid: true},
		})

		if err != nil {
			logger.Error("Error creando anuncio", "error", err)
			c.String(http.StatusInternalServerError, "Error creando anuncio")
			return
		}

		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/campaigns/%s/ads", campaignID))
	}
}

func ListCampaignsPage(db database.Querier) func(c *gin.Context) {
	return func(c *gin.Context) {
		advertiserID := c.MustGet("advertiser_id").(pgtype.UUID)

		campaigns, err := db.GetCampaignsByAdvertiser(c.Request.Context(), advertiserID)
		if err != nil {
			logger.Error("Error obteniendo campañas", "error", err)
			c.String(http.StatusInternalServerError, "Error obteniendo campañas")
			return
		}

		view := campaignsView.List(campaigns)
		view.Render(c.Request.Context(), c.Writer)
	}
}

func ListAdsPage(db database.Querier) func(c *gin.Context) {
	return func(c *gin.Context) {
		campaignID := c.Param("id")
		advertiserID := c.MustGet("advertiser_id").(pgtype.UUID)

		// Validar campaña
		valid, err := db.CheckCampaignBelongsToAdvertiser(c.Request.Context(), database.CheckCampaignBelongsToAdvertiserParams{
			ID:           pgtype.UUID{Bytes: [16]byte(uuid.MustParse(campaignID)), Valid: true},
			AdvertiserID: advertiserID,
		})
		if err != nil || !valid {
			logger.Error("Campaña inválida", "error", err)
			c.String(http.StatusUnauthorized, "Campaña inválida")
			return
		}

		ads, err := db.GetAdsByCampaign(c.Request.Context(), pgtype.UUID{Bytes: [16]byte(uuid.MustParse(campaignID)), Valid: true})
		if err != nil {
			logger.Error("Error obteniendo anuncios", "error", err)
			c.String(http.StatusInternalServerError, "Error al obtener anuncios")
			return
		}

		view := adsView.List(ads)
		view.Render(c.Request.Context(), c.Writer)
	}
}

func ListAdsPagePublic(db database.Querier) func(c *gin.Context) {
	return func(c *gin.Context) {
		ads, err := db.GetAds(c.Request.Context())
		if err != nil {
			c.String(http.StatusInternalServerError, "Error al obtener anuncios")
			return
		}

		var adsTemplates []templates.Ad
		for _, ad := range ads {
			adsTemplates = append(adsTemplates, templates.Ad{
				ID:             ad.ID.String(),
				Title:          ad.Title.String,
				Description:    ad.Description.String,
				ImageURL:       ad.ImageUrl.String,
				DestinationURL: ad.DestinationUrl.String,
				CtaLabel:       ad.CtaLabel.String,
				CampaignID:     ad.CampaignID.String(),
			})
		}

		view := adsView.AdsView(adsTemplates)
		view.Render(c.Request.Context(), c.Writer)
	}
}

func ShowDashboard(db database.Querier) func(c *gin.Context) {
	return func(c *gin.Context) {
		metrics, err := GetDashboardMetrics(c.Request.Context(), db)
		if err != nil {
			c.String(http.StatusInternalServerError, "Error al obtener métricas")
			return
		}

		component := templates.Dashboard(metrics)
		component.Render(c.Request.Context(), c.Writer)
	}
}

func GetDashboardMetrics(ctx context.Context, db database.Querier) (templates.DashboardMetrics, error) {
	impressionsCount, err := db.CountImpressions(ctx)
	if err != nil {
		logger.Error("Error al obtener métricas de impresiones", "error", err)
		return templates.DashboardMetrics{}, err
	}
	logger.Info("Métricas de impresiones", "impressionsCount", impressionsCount)

	clicksCount, err := db.CountClicks(ctx)
	if err != nil {
		logger.Error("Error al obtener métricas de clicks", "error", err)
		return templates.DashboardMetrics{}, err
	}
	logger.Info("Métricas de clicks", "clicksCount", clicksCount)
	ctr, err := db.AverageCTR(ctx)
	if err != nil {
		logger.Error("Error al obtener métricas de CTR", "error", err)
		return templates.DashboardMetrics{}, err
	}
	logger.Info("Métricas de CTR", "ctr", ctr)
	activeAdsCount, err := db.CountActiveAds(ctx)
	if err != nil {
		logger.Error("Error al obtener métricas de anuncios activos", "error", err)
		return templates.DashboardMetrics{}, err
	}
	logger.Info("Métricas de anuncios activos", "activeAdsCount", activeAdsCount)

	var ctrFloat float64
	if ctr.Int == nil {
		ctrFloat = 0
	} else {
		ctrFloat = float64(ctr.Int.Int64()) / 100
	}

	return templates.DashboardMetrics{
		TotalImpressions: int(impressionsCount),
		TotalClicks:      int(clicksCount),
		CTR:              ctrFloat / 100,
		ActiveAds:        int(activeAdsCount),
	}, nil
}

func HandleSimulateImpression(db database.Querier) func(c *gin.Context) {
	return func(c *gin.Context) {
		adID := c.PostForm("ad_id")
		placementID := c.PostForm("placement_id")
		auctionID := uuid.New().String()

		// Inserta auction + impression
		db.InsertAuction(c.Request.Context(), database.InsertAuctionParams{
			ID:             pgtype.UUID{Bytes: [16]byte(uuid.MustParse(auctionID)), Valid: true},
			PlacementID:    pgtype.UUID{Bytes: [16]byte(uuid.MustParse(placementID)), Valid: true},
			RequestContext: []byte("{}"),
		})
		db.CreateImpression(c.Request.Context(), database.CreateImpressionParams{
			ID:          pgtype.UUID{Bytes: [16]byte(uuid.MustParse(uuid.New().String())), Valid: true},
			AdID:        pgtype.UUID{Bytes: [16]byte(uuid.MustParse(adID)), Valid: true},
			PlacementID: pgtype.UUID{Bytes: [16]byte(uuid.MustParse(placementID)), Valid: true},
			AuctionID:   pgtype.UUID{Bytes: [16]byte(uuid.MustParse(auctionID)), Valid: true},
			UserContext: []byte("{}"),
		})
		c.Redirect(http.StatusSeeOther, "/ads")
	}
}

func HandleSimulateClick(db database.Querier) func(c *gin.Context) {
	return func(c *gin.Context) {
		adID := c.PostForm("ad_id")

		// Obtener la última impresión del anuncio
		lastImpressionID, err := db.GetLastImpressionByAdID(c.Request.Context(), pgtype.UUID{Bytes: [16]byte(uuid.MustParse(adID)), Valid: true})
		if err != nil {
			logger.Error("Error al obtener última impresión", "error", err)
			c.String(http.StatusInternalServerError, "Error al obtener última impresión")
			return
		}

		// Crear el click
		clickID := pgtype.UUID{Bytes: [16]byte(uuid.MustParse(uuid.New().String())), Valid: true}
		_, err = db.CreateClick(c.Request.Context(), database.CreateClickParams{
			ID:           clickID,
			ImpressionID: lastImpressionID,
		})
		if err != nil {
			logger.Error("Error al crear click", "error", err)
			c.String(http.StatusInternalServerError, "Error al crear click")
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Click registrado exitosamente"})
	}
}

// ShowCampaignDetails muestra los detalles de una campaña específica
func ShowCampaignDetails(db database.Querier) func(c *gin.Context) {
	return func(c *gin.Context) {
		campaignIDStr := c.Param("id")
		campaignID, err := uuid.Parse(campaignIDStr)
		if err != nil {
			c.String(http.StatusBadRequest, "ID de campaña inválido")
			return
		}

		// Obtener datos de la campaña
		campaign, err := db.GetCampaignByID(c.Request.Context(), pgtype.UUID{Bytes: campaignID, Valid: true})
		if err != nil {
			logger.Error("Error al obtener campaña", "error", err)
			c.String(http.StatusInternalServerError, "Error al obtener campaña")
			return
		}

		// Obtener métricas de la campaña
		metrics, err := db.GetCampaignMetrics(c.Request.Context(), pgtype.UUID{Bytes: campaignID, Valid: true})
		if err != nil {
			logger.Error("Error al obtener métricas", "error", err)
			c.String(http.StatusInternalServerError, "Error al obtener métricas")
			return
		}

		// Obtener gasto de la campaña
		spend, err := db.GetCampaignSpend(c.Request.Context(), pgtype.UUID{Bytes: campaignID, Valid: true})
		if err != nil {
			logger.Error("Error al obtener gasto", "error", err)
			c.String(http.StatusInternalServerError, "Error al obtener gasto")
			return
		}

		// Obtener anuncios de la campaña
		ads, err := db.GetAdsByCampaign(c.Request.Context(), pgtype.UUID{Bytes: campaignID, Valid: true})
		if err != nil {
			logger.Error("Error al obtener anuncios", "error", err)
			c.String(http.StatusInternalServerError, "Error al obtener anuncios")
			return
		}

		data := campaignsView.CampaignDetailData{
			Campaign: campaign,
			Metrics:  metrics,
			Spend:    spend,
			Ads:      ads,
		}

		component := campaignsView.Detail(data)
		component.Render(c.Request.Context(), c.Writer)
	}
}

// HandleUpdateCampaign actualiza una campaña existente
func HandleUpdateCampaign(db database.Querier) func(c *gin.Context) {
	return func(c *gin.Context) {
		campaignIDStr := c.Param("id")
		campaignID, err := uuid.Parse(campaignIDStr)
		if err != nil {
			c.String(http.StatusBadRequest, "ID de campaña inválido")
			return
		}

		// Parsear formulario
		if err := c.Request.ParseForm(); err != nil {
			c.String(http.StatusBadRequest, "Error al procesar formulario")
			return
		}

		// Obtener valores del formulario
		name := c.PostForm("name")
		budgetStr := c.PostForm("budget")
		dailyBudgetStr := c.PostForm("daily_budget")
		startDateStr := c.PostForm("start_date")
		endDateStr := c.PostForm("end_date")

		// Convertir budget
		budgetFloat, err := strconv.ParseFloat(budgetStr, 64)
		if err != nil {
			c.String(http.StatusBadRequest, "Presupuesto inválido")
			return
		}
		budget := pgtype.Numeric{Int: big.NewInt(int64(budgetFloat * 100)), Valid: true}

		// Convertir daily_budget (opcional)
		var dailyBudget pgtype.Numeric
		if dailyBudgetStr != "" {
			dailyBudgetFloat, err := strconv.ParseFloat(dailyBudgetStr, 64)
			if err != nil {
				c.String(http.StatusBadRequest, "Presupuesto diario inválido")
				return
			}
			dailyBudget = pgtype.Numeric{Int: big.NewInt(int64(dailyBudgetFloat * 100)), Valid: true}
		}

		// Convertir fechas (opcionales)
		var startDate pgtype.Date
		if startDateStr != "" {
			startTime, err := time.Parse("2006-01-02", startDateStr)
			if err != nil {
				c.String(http.StatusBadRequest, "Fecha de inicio inválida")
				return
			}
			startDate = pgtype.Date{Time: startTime, Valid: true}
		}

		var endDate pgtype.Date
		if endDateStr != "" {
			endTime, err := time.Parse("2006-01-02", endDateStr)
			if err != nil {
				c.String(http.StatusBadRequest, "Fecha de fin inválida")
				return
			}
			endDate = pgtype.Date{Time: endTime, Valid: true}
		}

		// Actualizar campaña
		_, err = db.UpdateCampaign(c.Request.Context(), database.UpdateCampaignParams{
			ID:           pgtype.UUID{Bytes: campaignID, Valid: true},
			Name:         name,
			Budget:       budget,
			DailyBudget:  dailyBudget,
			StartDate:    startDate,
			EndDate:      endDate,
			Status:       pgtype.Text{String: "active", Valid: true}, // Mantener estado actual
			AdvertiserID: pgtype.UUID{Bytes: [16]byte(uuid.MustParse("hardcoded-advertiser")), Valid: true},
		})
		if err != nil {
			logger.Error("Error al actualizar campaña", "error", err)
			c.String(http.StatusInternalServerError, "Error al actualizar campaña")
			return
		}

		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/campaigns/%s", campaignIDStr))
	}
}

// HandlePauseCampaign pausa una campaña
func HandlePauseCampaign(db database.Querier) func(c *gin.Context) {
	return func(c *gin.Context) {
		campaignIDStr := c.Param("id")
		campaignID, err := uuid.Parse(campaignIDStr)
		if err != nil {
			c.String(http.StatusBadRequest, "ID de campaña inválido")
			return
		}

		// Pausar campaña
		_, err = db.UpdateCampaignStatus(c.Request.Context(), database.UpdateCampaignStatusParams{
			ID:           pgtype.UUID{Bytes: campaignID, Valid: true},
			Status:       pgtype.Text{String: "paused", Valid: true},
			AdvertiserID: pgtype.UUID{Bytes: [16]byte(uuid.MustParse("hardcoded-advertiser")), Valid: true},
		})
		if err != nil {
			logger.Error("Error al pausar campaña", "error", err)
			c.String(http.StatusInternalServerError, "Error al pausar campaña")
			return
		}

		// Responder con el nuevo estado
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, `<span class="text-yellow-400">Pausada</span>`)
	}
}

// HandleActivateCampaign activa una campaña
func HandleActivateCampaign(db database.Querier) func(c *gin.Context) {
	return func(c *gin.Context) {
		campaignIDStr := c.Param("id")
		campaignID, err := uuid.Parse(campaignIDStr)
		if err != nil {
			c.String(http.StatusBadRequest, "ID de campaña inválido")
			return
		}

		// Activar campaña
		_, err = db.UpdateCampaignStatus(c.Request.Context(), database.UpdateCampaignStatusParams{
			ID:           pgtype.UUID{Bytes: campaignID, Valid: true},
			Status:       pgtype.Text{String: "active", Valid: true},
			AdvertiserID: pgtype.UUID{Bytes: [16]byte(uuid.MustParse("hardcoded-advertiser")), Valid: true},
		})
		if err != nil {
			logger.Error("Error al activar campaña", "error", err)
			c.String(http.StatusInternalServerError, "Error al activar campaña")
			return
		}

		// Responder con el nuevo estado
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, `<span class="text-green-400">Activa</span>`)
	}
}

// ShowOptimizationPage muestra la página de optimización de presupuesto
func ShowOptimizationPage(db database.Querier) func(c *gin.Context) {
	return func(c *gin.Context) {
		advertiserID := c.MustGet("advertiser_id").(pgtype.UUID)

		// Obtener análisis de rendimiento de campañas
		campaignAnalysis, err := db.GetCampaignPerformanceAnalysis(c.Request.Context(), advertiserID)
		if err != nil {
			logger.Error("Error al obtener análisis de campañas", "error", err)
			c.String(http.StatusInternalServerError, "Error al obtener análisis de campañas")
			return
		}

		// Obtener análisis de rendimiento de anuncios
		adAnalysis, err := db.GetAdPerformanceAnalysis(c.Request.Context(), advertiserID)
		if err != nil {
			logger.Error("Error al obtener análisis de anuncios", "error", err)
			c.String(http.StatusInternalServerError, "Error al obtener análisis de anuncios")
			return
		}

		// Obtener alertas de presupuesto
		budgetAlerts, err := db.GetBudgetAlerts(c.Request.Context(), advertiserID)
		if err != nil {
			logger.Error("Error al obtener alertas de presupuesto", "error", err)
			c.String(http.StatusInternalServerError, "Error al obtener alertas de presupuesto")
			return
		}

		// Obtener recomendaciones de optimización
		recommendations, err := db.GetOptimizationRecommendations(c.Request.Context(), advertiserID)
		if err != nil {
			logger.Error("Error al obtener recomendaciones", "error", err)
			c.String(http.StatusInternalServerError, "Error al obtener recomendaciones")
			return
		}

		data := optimization.OptimizationData{
			CampaignAnalysis: campaignAnalysis,
			AdAnalysis:       adAnalysis,
			BudgetAlerts:     budgetAlerts,
			Recommendations:  recommendations,
		}

		component := optimization.Optimization(data)
		component.Render(c.Request.Context(), c.Writer)
	}
}
