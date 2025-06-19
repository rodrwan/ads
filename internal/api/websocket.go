package api

import (
	"ads-system/internal/database"
	"ads-system/internal/interfaces"
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WebSocketManager maneja todas las conexiones WebSocket
type WebSocketManager struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan interface{}
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mutex      sync.RWMutex
	querier    database.Querier
}

// Asegurar que WebSocketManager implementa MetricsNotifier
var _ interfaces.MetricsNotifier = (*WebSocketManager)(nil)

// MetricsData estructura para enviar métricas en tiempo real
type MetricsData struct {
	Type             string  `json:"type"`
	TotalImpressions int     `json:"total_impressions"`
	TotalClicks      int     `json:"total_clicks"`
	CTR              float64 `json:"ctr"`
	ActiveAds        int     `json:"active_ads"`
	Timestamp        int64   `json:"timestamp"`
}

// NewWebSocketManager crea un nuevo manager de WebSockets
func NewWebSocketManager(querier database.Querier) *WebSocketManager {
	return &WebSocketManager{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan interface{}),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		querier:    querier,
	}
}

// Start inicia el manager de WebSockets
func (manager *WebSocketManager) Start() {
	for {
		select {
		case client := <-manager.register:
			manager.mutex.Lock()
			manager.clients[client] = true
			manager.mutex.Unlock()
			log.Printf("Cliente WebSocket conectado. Total: %d", len(manager.clients))

		case client := <-manager.unregister:
			manager.mutex.Lock()
			if _, ok := manager.clients[client]; ok {
				delete(manager.clients, client)
				client.Close()
			}
			manager.mutex.Unlock()
			log.Printf("Cliente WebSocket desconectado. Total: %d", len(manager.clients))

		case message := <-manager.broadcast:
			manager.mutex.RLock()
			for client := range manager.clients {
				err := client.WriteJSON(message)
				if err != nil {
					log.Printf("Error enviando mensaje: %v", err)
					client.Close()
					delete(manager.clients, client)
				}
			}
			manager.mutex.RUnlock()
		}
	}
}

// BroadcastMetrics envía métricas actualizadas a todos los clientes
func (manager *WebSocketManager) BroadcastMetrics() {
	log.Printf("📊 Obteniendo métricas actuales...")
	metrics, err := manager.getCurrentMetrics()
	if err != nil {
		log.Printf("❌ Error obteniendo métricas: %v", err)
		return
	}

	log.Printf("✅ Métricas obtenidas: %+v", metrics)
	log.Printf("📤 Enviando a %d clientes conectados", len(manager.clients))
	manager.broadcast <- metrics
}

// getCurrentMetrics obtiene las métricas actuales de la base de datos
func (manager *WebSocketManager) getCurrentMetrics() (MetricsData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	impressionsCount, err := manager.querier.CountImpressions(ctx)
	if err != nil {
		return MetricsData{}, err
	}

	clicksCount, err := manager.querier.CountClicks(ctx)
	if err != nil {
		return MetricsData{}, err
	}

	ctr, err := manager.querier.AverageCTR(ctx)
	if err != nil {
		return MetricsData{}, err
	}

	activeAdsCount, err := manager.querier.CountActiveAds(ctx)
	if err != nil {
		return MetricsData{}, err
	}

	var ctrFloat float64
	if ctr.Int != nil {
		ctrFloat = float64(ctr.Int.Int64()) / 10000 // Convertir de centésimas a porcentaje
	}

	return MetricsData{
		Type:             "metrics_update",
		TotalImpressions: int(impressionsCount),
		TotalClicks:      int(clicksCount),
		CTR:              ctrFloat,
		ActiveAds:        int(activeAdsCount),
		Timestamp:        time.Now().Unix(),
	}, nil
}

// WebSocketHandler maneja las conexiones WebSocket
func (manager *WebSocketManager) WebSocketHandler() gin.HandlerFunc {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Permitir todas las conexiones en desarrollo
		},
	}

	return func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("Error upgrading connection: %v", err)
			return
		}

		// Enviar métricas iniciales
		initialMetrics, err := manager.getCurrentMetrics()
		if err == nil {
			conn.WriteJSON(initialMetrics)
		}

		manager.register <- conn

		// Mantener la conexión viva
		go func() {
			defer func() {
				manager.unregister <- conn
			}()

			for {
				// Leer mensajes del cliente (para mantener la conexión)
				_, _, err := conn.ReadMessage()
				if err != nil {
					break
				}
			}
		}()
	}
}

// StartMetricsBroadcaster inicia el broadcast automático de métricas
func (manager *WebSocketManager) StartMetricsBroadcaster() {
	ticker := time.NewTicker(5 * time.Second) // Actualizar cada 5 segundos
	log.Printf("🚀 Iniciando broadcast automático de métricas cada 5 segundos")
	go func() {
		for range ticker.C {
			log.Printf("📊 Enviando broadcast automático de métricas...")
			manager.BroadcastMetrics()
		}
	}()
}

// NotifyMetricsUpdate notifica una actualización inmediata de métricas
func (manager *WebSocketManager) NotifyMetricsUpdate() {
	manager.BroadcastMetrics()
}
