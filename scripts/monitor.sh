#!/bin/bash

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== MONITOREO EN TIEMPO REAL - SISTEMA DE ANUNCIOS ===${NC}"
echo ""

# Función para obtener métricas de la base de datos
get_db_metrics() {
    echo -e "${CYAN}📊 MÉTRICAS DE BASE DE DATOS${NC}"
    
    # Conectar a PostgreSQL y obtener métricas
    metrics=$(psql "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" -t -c "
        SELECT 
            'Impresiones: ' || COUNT(*) as impressions
        FROM impressions
        UNION ALL
        SELECT 
            'Clicks: ' || COUNT(*) as clicks
        FROM clicks
        UNION ALL
        SELECT 
            'CTR: ' || ROUND(COUNT(c.id)::decimal / NULLIF(COUNT(i.id), 0) * 100, 2) || '%' as ctr
        FROM impressions i
        LEFT JOIN clicks c ON c.impression_id = i.id
        UNION ALL
        SELECT 
            'Subastas: ' || COUNT(*) as auctions
        FROM auctions
        UNION ALL
        SELECT 
            'Anuncios Activos: ' || COUNT(*) as active_ads
        FROM ads WHERE status = 'active';
    " 2>/dev/null)
    
    if [ $? -eq 0 ]; then
        echo "$metrics" | while read line; do
            if [ -n "$line" ]; then
                echo -e "${GREEN}  $line${NC}"
            fi
        done
    else
        echo -e "${RED}  ❌ Error conectando a la base de datos${NC}"
    fi
    echo ""
}

# Función para verificar estado de servicios
check_services() {
    echo -e "${CYAN}🔍 ESTADO DE SERVICIOS${NC}"
    
    # Verificar Ad-Server
    if curl -s http://localhost:8080/health >/dev/null 2>&1; then
        echo -e "${GREEN}  ✅ Ad-Server (8080): Activo${NC}"
    else
        echo -e "${RED}  ❌ Ad-Server (8080): Inactivo${NC}"
    fi
    
    # Verificar Campaigns
    if curl -s http://localhost:8081/ >/dev/null 2>&1; then
        echo -e "${GREEN}  ✅ Campaigns (8081): Activo${NC}"
    else
        echo -e "${RED}  ❌ Campaigns (8081): Inactivo${NC}"
    fi
    
    # Verificar PostgreSQL
    if pg_isready -h localhost -p 5432 >/dev/null 2>&1; then
        echo -e "${GREEN}  ✅ PostgreSQL (5432): Activo${NC}"
    else
        echo -e "${RED}  ❌ PostgreSQL (5432): Inactivo${NC}"
    fi
    
    # Verificar Redis
    if redis-cli -h localhost -p 6379 ping >/dev/null 2>&1; then
        echo -e "${GREEN}  ✅ Redis (6379): Activo${NC}"
    else
        echo -e "${RED}  ❌ Redis (6379): Inactivo${NC}"
    fi
    echo ""
}

# Función para mostrar logs recientes
show_recent_logs() {
    echo -e "${CYAN}📝 LOGS RECIENTES${NC}"
    echo -e "${YELLOW}  Últimos 5 logs del Ad-Server:${NC}"
    docker-compose logs --tail=5 ad-system 2>/dev/null | grep -v "level=info" || echo "  No hay logs disponibles"
    echo ""
}

# Función para simular tráfico
simulate_traffic() {
    echo -e "${CYAN}🚀 SIMULANDO TRÁFICO${NC}"
    
    # Simular subasta
    auction_response=$(curl -s -X POST "http://localhost:8080/auction" \
        -H "Content-Type: application/json" \
        -d '{
            "placement_id": "30000000-0000-0000-0000-000000000001",
            "request_context": {
                "country": "CL",
                "device": "desktop",
                "interests": ["technology"],
                "os": "Windows",
                "browser": "Chrome",
                "ip": "192.168.1.1",
                "user_agent": "Mozilla/5.0...",
                "referer": "https://example.com"
            }
        }')
    
    if echo "$auction_response" | grep -q "ad_id"; then
        echo -e "${GREEN}  ✅ Subasta simulada exitosamente${NC}"
        ad_id=$(echo "$auction_response" | jq -r '.ad_id' 2>/dev/null)
        if [ "$ad_id" != "null" ] && [ -n "$ad_id" ]; then
            echo -e "${BLUE}  📋 Anuncio ganador: $ad_id${NC}"
        fi
    else
        echo -e "${RED}  ❌ Error en subasta simulada${NC}"
    fi
    echo ""
}

# Función principal de monitoreo
monitor() {
    clear
    echo -e "${BLUE}🕐 $(date)${NC}"
    echo ""
    
    check_services
    get_db_metrics
    simulate_traffic
    show_recent_logs
    
    echo -e "${PURPLE}💡 COMANDOS ÚTILES:${NC}"
    echo "  📊 Dashboard: http://localhost:8081/dashboard"
    echo "  📋 Campañas: http://localhost:8081/campaigns"
    echo "  🔐 Login: http://localhost:8081/login (email: acme@example.com)"
    echo "  🐳 Logs completos: docker-compose logs -f"
    echo "  🧪 Pruebas API: ./test-api.sh"
    echo ""
    echo -e "${YELLOW}⏰ Actualizando en 10 segundos... (Ctrl+C para salir)${NC}"
}

# Ejecutar monitoreo continuo
if [ "$1" = "once" ]; then
    monitor
else
    while true; do
        monitor
        sleep 10
    done
fi 