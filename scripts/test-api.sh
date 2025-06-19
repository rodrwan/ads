#!/bin/bash

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== SISTEMA DE ANUNCIOS - PRUEBAS DE API ===${NC}"
echo ""

# Función para hacer requests y mostrar resultados
test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local description=$4
    
    echo -e "${YELLOW}🔍 Probando: $description${NC}"
    echo -e "${BLUE}Endpoint: $method $endpoint${NC}"
    
    if [ -n "$data" ]; then
        echo -e "${BLUE}Data: $data${NC}"
        response=$(curl -s -w "\n%{http_code}" -X $method "$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data")
    else
        response=$(curl -s -w "\n%{http_code}" -X $method "$endpoint")
    fi
    
    # Separar response body y status code
    body=$(echo "$response" | head -n -1)
    status=$(echo "$response" | tail -n 1)
    
    if [ "$status" -ge 200 ] && [ "$status" -lt 300 ]; then
        echo -e "${GREEN}✅ Status: $status${NC}"
    else
        echo -e "${RED}❌ Status: $status${NC}"
    fi
    
    echo -e "${BLUE}Response:${NC}"
    echo "$body" | jq '.' 2>/dev/null || echo "$body"
    echo ""
    echo "----------------------------------------"
    echo ""
}

# 1. Probar subasta (RTB)
echo -e "${GREEN}📊 1. PROBANDO SUBASTA RTB${NC}"
test_endpoint "POST" "http://localhost:8080/auction" \
    '{
        "placement_id": "30000000-0000-0000-0000-000000000001",
        "request_context": {
            "country": "CL",
            "device": "desktop",
            "interests": ["technology", "gaming"],
            "os": "Windows",
            "browser": "Chrome",
            "ip": "192.168.1.1",
            "user_agent": "Mozilla/5.0...",
            "referer": "https://example.com"
        }
    }' \
    "Crear subasta RTB"

# 2. Probar registro de impresión
echo -e "${GREEN}👁️  2. PROBANDO REGISTRO DE IMPRESIÓN${NC}"
test_endpoint "POST" "http://localhost:8080/impression" \
    '{
        "ad_id": "20000000-0000-0000-0000-000000000001",
        "placement_id": "30000000-0000-0000-0000-000000000001",
        "auction_id": "60000000-0000-0000-0000-000000000001",
        "user_context": {
            "country": "CL",
            "device": "mobile",
            "ip": "192.168.1.100"
        }
    }' \
    "Registrar impresión"

# 3. Probar registro de click
echo -e "${GREEN}🖱️  3. PROBANDO REGISTRO DE CLICK${NC}"
test_endpoint "POST" "http://localhost:8080/click" \
    '{
        "impression_id": "70000000-0000-0000-0000-000000000001"
    }' \
    "Registrar click"

# 4. Probar múltiples subastas para ver diferentes resultados
echo -e "${GREEN}🔄 4. PROBANDO MÚLTIPLES SUBASTAS${NC}"
for i in {1..3}; do
    test_endpoint "POST" "http://localhost:8080/auction" \
        "{
            \"placement_id\": \"30000000-0000-0000-0000-000000000001\",
            \"request_context\": {
                \"country\": \"AR\",
                \"device\": \"mobile\",
                \"interests\": [\"sports\", \"news\"],
                \"os\": \"iOS\",
                \"browser\": \"Safari\",
                \"ip\": \"192.168.1.$i\",
                \"user_agent\": \"Mozilla/5.0 iPhone...\",
                \"referer\": \"https://news.com\"
            }
        }" \
        "Subasta $i con contexto diferente"
done

echo -e "${GREEN}🎉 PRUEBAS COMPLETADAS${NC}"
echo ""
echo -e "${BLUE}📋 RESUMEN DE ENDPOINTS PROBADOS:${NC}"
echo "✅ POST /auction - Subasta RTB"
echo "✅ POST /impression - Registro de impresiones"
echo "✅ POST /click - Registro de clicks"
echo ""
echo -e "${BLUE}🌐 INTERFACES WEB DISPONIBLES:${NC}"
echo "📊 Dashboard: http://localhost:8081/dashboard"
echo "📋 Campañas: http://localhost:8081/campaigns"
echo "🔐 Login: http://localhost:8081/login"
echo ""
echo -e "${BLUE}📊 MONITOREO:${NC}"
echo "🐳 Docker logs: docker-compose logs -f"
echo "🗄️  Base de datos: localhost:5432"
echo "🔴 Redis: localhost:6379" 