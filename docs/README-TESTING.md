# 🚀 Sistema de Anuncios - Guía de Testing

## 📋 Estado Actual del Sistema

✅ **Servicios Levantados:**
- 🐳 **Docker Compose**: Todos los servicios funcionando
- 🗄️ **PostgreSQL**: Base de datos con datos de prueba
- 🔴 **Redis**: Cache y streams funcionando
- 🚀 **Ad-Server**: API RTB en puerto 8080
- ⚙️ **Auction-Engine**: Procesamiento asíncrono
- 📊 **Campaigns**: Interfaz web en puerto 8081

## 🧪 Herramientas de Testing Disponibles

### 1. Script de Pruebas API (`test-api.sh`)
```bash
# Hacer ejecutable
chmod +x test-api.sh

# Ejecutar todas las pruebas
./test-api.sh
```

**Endpoints probados:**
- `POST /auction` - Subastas RTB
- `POST /impression` - Registro de impresiones  
- `POST /click` - Registro de clicks

### 2. Monitor en Tiempo Real (`monitor.sh`)
```bash
# Hacer ejecutable
chmod +x monitor.sh

# Monitoreo continuo (actualiza cada 10s)
./monitor.sh

# Monitoreo una sola vez
./monitor.sh once
```

**Métricas mostradas:**
- 📊 Estado de servicios
- 📈 Impresiones, clicks, CTR
- 🚀 Simulación de tráfico automática
- 📝 Logs recientes

## 🌐 Interfaces Web

### Dashboard de Campañas
- **URL**: http://localhost:8081/dashboard
- **Login**: acme@example.com
- **Funcionalidades**:
  - 📊 Métricas en tiempo real
  - 📈 CTR y estadísticas
  - 📋 Lista de campañas activas

### Gestión de Campañas
- **URL**: http://localhost:8081/campaigns
- **Funcionalidades**:
  - ➕ Crear nuevas campañas
  - 📝 Editar campañas existentes
  - 🗑️ Eliminar campañas

### Login
- **URL**: http://localhost:8081/login
- **Credenciales de prueba**:
  - Email: `acme@example.com`
  - (No requiere contraseña en modo demo)

## 🔧 Comandos de Desarrollo

### Gestión de Docker
```bash
# Ver estado de servicios
docker-compose ps

# Ver logs en tiempo real
docker-compose logs -f

# Ver logs de un servicio específico
docker-compose logs -f ad-system
docker-compose logs -f auction-engine
docker-compose logs -f campaigns

# Reiniciar un servicio
docker-compose restart ad-system

# Parar todos los servicios
docker-compose down

# Levantar todo de nuevo
docker-compose up -d
```

### Base de Datos
```bash
# Conectar a PostgreSQL
psql postgres://postgres:postgres@localhost:5432/postgres

# Ejecutar migraciones
DATABASE_URL="postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" make migrate-up

# Poblar datos de prueba
DATABASE_URL="postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" make seed
```

### Generación de Código
```bash
# Generar código SQL
make generate

# Generar templates web
make generate-web
```

## 📊 Datos de Prueba Disponibles

### Anunciantes
- **Acme Corp** (acme@example.com) - Balance: $1000
- **Globex Inc** (globex@example.com) - Balance: $500

### Campañas
- **Campaña ACME** - Presupuesto: $500, Estado: Activa

### Anuncios
- **"¡Compra ahora!"** - CTA: "Ver más", Targeting: Chile

### Placements
- **Sidebar 300x250** - Tamaño: 300x250px

### Bids
- **Bid ACME** - Precio: $0.75, Máximo diario: $10

## 🧪 Casos de Prueba Manuales

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

### 2. Registrar Impresión
```bash
curl -X POST http://localhost:8080/impression \
  -H "Content-Type: application/json" \
  -d '{
    "ad_id": "20000000-0000-0000-0000-000000000001",
    "placement_id": "30000000-0000-0000-0000-000000000001",
    "auction_id": "60000000-0000-0000-0000-000000000001",
    "user_context": {
      "country": "CL",
      "device": "mobile"
    }
  }'
```

### 3. Registrar Click
```bash
curl -X POST http://localhost:8080/click \
  -H "Content-Type: application/json" \
  -d '{
    "impression_id": "70000000-0000-0000-0000-000000000001"
  }'
```

## 🔍 Troubleshooting

### Problemas Comunes

**1. Servicios no responden**
```bash
# Verificar estado
docker-compose ps

# Ver logs de error
docker-compose logs ad-system
```

**2. Base de datos no conecta**
```bash
# Verificar PostgreSQL
docker-compose logs postgres

# Probar conexión
psql postgres://postgres:postgres@localhost:5432/postgres
```

**3. Redis no funciona**
```bash
# Verificar Redis
docker-compose logs redis

# Probar conexión
redis-cli -h localhost -p 6379 ping
```

**4. APIs devuelven error 500**
```bash
# Ver logs del ad-server
docker-compose logs -f ad-system

# Verificar base de datos
psql postgres://postgres:postgres@localhost:5432/postgres -c "SELECT COUNT(*) FROM ads;"
```

## 📈 Métricas a Observar

### KPIs del Sistema
- **CTR (Click-Through Rate)**: Clicks / Impresiones
- **Tiempo de respuesta**: Latencia de subastas RTB
- **Throughput**: Subastas por segundo
- **Tasa de éxito**: Subastas con anuncios ganadores

### Monitoreo en Tiempo Real
```bash
# Dashboard web
open http://localhost:8081/dashboard

# Monitor CLI
./monitor.sh

# Logs en tiempo real
docker-compose logs -f
```

## 🎯 Próximos Pasos

1. **Ejecutar pruebas básicas**: `./test-api.sh`
2. **Explorar interfaz web**: http://localhost:8081/dashboard
3. **Monitorear métricas**: `./monitor.sh`
4. **Crear campañas de prueba** desde la interfaz web
5. **Simular tráfico real** con múltiples subastas

¡El sistema está listo para testing! 🚀 