package interfaces

// MetricsNotifier interfaz para notificar cambios en métricas
type MetricsNotifier interface {
	NotifyMetricsUpdate()
}
