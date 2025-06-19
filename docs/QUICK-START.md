# 🚀 Sistema de Anuncios - Quick Start

## ✅ Estado Actual
**¡Todo funcionando!** El sistema está completamente operativo con:
- 🐳 **5 servicios Docker** ejecutándose
- 🗄️ **Base de datos** poblada con datos de prueba
- 🚀 **APIs RTB** respondiendo correctamente
- 📊 **Interfaz web** disponible

## 🎯 Comandos Rápidos

### 1. Verificar Estado
```bash
# Estado de servicios
docker-compose ps

# Métricas en tiempo real
./monitor.sh once
```

### 2. Probar APIs
```bash
# Pruebas básicas
./test-api.sh

# Load test (100 requests, 10 concurrentes)
./load-test.sh 100 10 0.1
```

### 3. Interfaces Web
```bash
# Abrir en navegador
open http://localhost:8081/dashboard
open http://localhost:8081/campaigns
open http://localhost:8081/login
```

### 4. Monitoreo
```bash
# Logs en tiempo real
docker-compose logs -f

# Logs específicos
docker-compose logs -f ad-system
docker-compose logs -f auction-engine
```

## 📊 URLs Importantes

| Servicio | URL | Descripción |
|----------|-----|-------------|
| **Dashboard** | http://localhost:8081/dashboard | Métricas en tiempo real |
| **Campañas** | http://localhost:8081/campaigns | Gestión de campañas |
| **Login** | http://localhost:8081/login | Acceso al sistema |
| **API RTB** | http://localhost:8080/auction | Subastas en tiempo real |

## 🔐 Credenciales de Prueba
- **Email**: `acme@example.com`
- **Contraseña**: No requerida (modo demo)

## 🧪 Datos de Prueba Disponibles

### Anunciantes
- **Acme Corp** - Balance: $1000
- **Globex Inc** - Balance: $500

### Campañas Activas
- **Campaña ACME** - Presupuesto: $500

### Anuncios
- **"¡Compra ahora!"** - CTA: "Ver más"

### Placements
- **Sidebar 300x250** - Tamaño: 300x250px

## 🚀 Casos de Uso

### 1. Probar Subasta RTB
```bash
curl -X POST http://localhost:8080/auction \
  -H "Content-Type: application/json" \
  -d '{
    "placement_id": "30000000-0000-0000-0000-000000000001",
    "request_context": {
      "country": "CL",
      "device": "desktop",
      "interests": ["technology"],
      "os": "Windows",
      "browser": "Chrome",
      "ip": "192.168.1.1"
    }
  }'
```

### 2. Crear Nueva Campaña
1. Ir a http://localhost:8081/campaigns
2. Hacer login con `acme@example.com`
3. Click en "Crear Campaña"
4. Llenar formulario y guardar

### 3. Ver Métricas en Tiempo Real
1. Ir a http://localhost:8081/dashboard
2. Ver CTR, impresiones, clicks
3. Monitorear actividad en tiempo real

## 🔧 Troubleshooting

### Problema: APIs devuelven error 500
```bash
# Ver logs
docker-compose logs ad-system

# Reiniciar servicio
docker-compose restart ad-system
```

### Problema: Base de datos no conecta
```bash
# Verificar PostgreSQL
docker-compose logs postgres

# Probar conexión
psql postgres://postgres:postgres@localhost:5432/postgres
```

### Problema: Interfaz web no carga
```bash
# Verificar servicio
docker-compose logs campaigns

# Reiniciar
docker-compose restart campaigns
```

## 📈 Métricas a Observar

### KPIs Principales
- **CTR**: Clicks / Impresiones
- **Tiempo de respuesta**: < 100ms para subastas
- **Throughput**: Subastas por segundo
- **Tasa de éxito**: Subastas con ganadores

### Monitoreo Continuo
```bash
# Monitor automático (actualiza cada 10s)
./monitor.sh

# Load test intensivo
./load-test.sh 1000 50 0.05
```

## 🎯 Próximos Pasos

1. **Explorar interfaz web** - Crear campañas y anuncios
2. **Ejecutar load tests** - Probar rendimiento
3. **Monitorear métricas** - Ver comportamiento en tiempo real
4. **Analizar logs** - Entender el flujo de datos
5. **Modificar datos** - Agregar más anuncios y bids

## 📞 Comandos de Emergencia

```bash
# Parar todo
docker-compose down

# Levantar de nuevo
docker-compose up -d

# Resetear base de datos
docker-compose down -v
docker-compose up -d postgres redis
DATABASE_URL="postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" make migrate-up
DATABASE_URL="postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" make seed
docker-compose up -d
```

---

**¡El sistema está listo para testing y desarrollo! 🚀** 