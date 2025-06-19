# 🚀 Dashboard en Tiempo Real - Sistema de Anuncios

## ✨ Nuevas Funcionalidades Implementadas

### 🔄 Actualizaciones en Tiempo Real
- **WebSockets**: Conexión bidireccional para actualizaciones instantáneas
- **Broadcast Automático**: Métricas se actualizan cada 5 segundos automáticamente
- **Notificaciones Inmediatas**: Los handlers de impresiones y clicks notifican cambios en tiempo real
- **Reconexión Automática**: El cliente se reconecta automáticamente si se pierde la conexión

### 🎨 Interfaz Mejorada
- **Animaciones Suaves**: Los números cambian con animaciones visuales
- **Indicador de Estado**: Punto verde/rojo que muestra el estado de la conexión
- **Notificaciones Toast**: Mensajes temporales cuando se actualizan las métricas
- **Transiciones Fluidas**: Efectos visuales para cambios de valores

### 📊 Métricas en Tiempo Real
- **Impresiones Totales**: Se actualiza automáticamente
- **Clicks Totales**: Se actualiza automáticamente  
- **Click-Through Rate (CTR)**: Se calcula y actualiza en tiempo real
- **Anuncios Activos**: Se actualiza automáticamente

## 🏗️ Arquitectura Técnica

### Backend (Go)
```
internal/api/
├── websocket.go          # Manager de WebSockets
├── interfaces/
│   └── metrics.go        # Interfaz MetricsNotifier
└── handlers/
    ├── impressions/      # Handler con notificaciones
    └── clicks/          # Handler con notificaciones
```

### Frontend (JavaScript)
- **WebSocket Client**: Conexión automática al endpoint `/ws`
- **Animaciones**: Efectos visuales para cambios de valores
- **Reconexión**: Lógica de reconexión con backoff exponencial
- **Notificaciones**: Sistema de notificaciones toast

## 🚀 Cómo Usar

### 1. Iniciar el Sistema
```bash
# Levantar todos los servicios
docker-compose up -d

# Ejecutar migraciones
make migrate

# Cargar datos de prueba
./generate-test-data.sh

# Compilar y ejecutar el servidor
go build -o ad-server cmd/ad-server/main.go
./ad-server
```

### 2. Acceder al Dashboard
- **URL**: http://localhost:8080/dashboard
- **Credenciales**: admin@example.com / password123

### 3. Probar Tiempo Real
```bash
# Ejecutar simulación de actividad
./simulate-realtime.sh
```

## 🔧 Configuración

### WebSocket Manager
- **Intervalo de Actualización**: 5 segundos
- **Timeout de Conexión**: 5 segundos
- **Reconexión Máxima**: 5 intentos con backoff exponencial

### Animaciones
- **Duración**: 1 segundo
- **Color de Cambio**: Verde (#10B981)
- **Efecto**: Pulse + cambio de color

## 📈 Flujo de Datos

1. **Evento**: Se crea una impresión o click
2. **Handler**: Notifica al WebSocket Manager
3. **Broadcast**: Se envían métricas actualizadas a todos los clientes
4. **Cliente**: Recibe datos y actualiza la interfaz con animaciones
5. **Usuario**: Ve los cambios en tiempo real

## 🎯 Características Destacadas

### ✅ Implementado
- [x] WebSockets para comunicación en tiempo real
- [x] Actualizaciones automáticas cada 5 segundos
- [x] Notificaciones inmediatas en eventos
- [x] Animaciones visuales para cambios
- [x] Reconexión automática del cliente
- [x] Indicador de estado de conexión
- [x] Notificaciones toast
- [x] Script de simulación de actividad

### 🔮 Futuras Mejoras
- [ ] Gráficos en tiempo real con Chart.js
- [ ] Filtros por campaña en tiempo real
- [ ] Alertas configurables
- [ ] Métricas por geolocalización
- [ ] Exportación de datos en tiempo real

## 🛠️ Comandos Útiles

```bash
# Ver logs del servidor
docker-compose logs -f ad-server

# Probar WebSocket manualmente
wscat -c ws://localhost:8080/ws

# Ver métricas actuales
curl http://localhost:8080/metrics

# Simular actividad
./simulate-realtime.sh

# Monitorear en tiempo real
./monitor.sh
```

## 🎨 Personalización

### Cambiar Intervalo de Actualización
En `internal/api/websocket.go`:
```go
func (manager *WebSocketManager) StartMetricsBroadcaster() {
    ticker := time.NewTicker(5 * time.Second) // Cambiar aquí
    // ...
}
```

### Cambiar Animaciones
En `internal/web/templates/dashboard.templ`:
```javascript
// Duración de animación
setTimeout(() => {
    element.classList.remove('animate-pulse', 'text-green-400');
}, 1000); // Cambiar aquí
```

## 🔍 Troubleshooting

### WebSocket no conecta
1. Verificar que el servidor esté corriendo
2. Revisar logs: `docker-compose logs ad-server`
3. Probar conexión: `wscat -c ws://localhost:8080/ws`

### Métricas no se actualizan
1. Verificar que los handlers estén notificando
2. Revisar logs de WebSocket
3. Verificar conexión a base de datos

### Animaciones no funcionan
1. Verificar que Tailwind CSS esté cargado
2. Revisar consola del navegador
3. Verificar IDs de elementos

## 📞 Soporte

Para reportar problemas o solicitar mejoras:
1. Revisar logs del sistema
2. Verificar configuración
3. Probar con datos de ejemplo
4. Documentar el problema

---

**¡El dashboard ahora es completamente interactivo y se actualiza en tiempo real! 🎉** 