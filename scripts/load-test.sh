#!/bin/bash

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== LOAD TEST - SISTEMA DE ANUNCIOS ===${NC}"
echo ""

# Configuración
TOTAL_REQUESTS=${1:-100}
CONCURRENT_REQUESTS=${2:-10}
DELAY=${3:-0.1}

echo -e "${YELLOW}Configuración:${NC}"
echo "  📊 Total de requests: $TOTAL_REQUESTS"
echo "  🔄 Requests concurrentes: $CONCURRENT_REQUESTS"
echo "  ⏱️  Delay entre requests: ${DELAY}s"
echo ""

# Función para hacer una subasta
make_auction() {
    local request_id=$1
    local country=$2
    local device=$3

    response=$(curl -s -w "%{http_code}" -X POST "http://localhost:8080/auction" \
        -H "Content-Type: application/json" \
        -d "{
            \"placement_id\": \"30000000-0000-0000-0000-000000000001\",
            \"request_context\": {
                \"country\": \"$country\",
                \"device\": \"$device\",
                \"interests\": [\"technology\", \"gaming\"],
                \"os\": \"Windows\",
                \"browser\": \"Chrome\",
                \"ip\": \"192.168.1.$request_id\",
                \"user_agent\": \"Mozilla/5.0...\",
                \"referer\": \"https://example.com\"
            }
        }")

    status_code=$(echo "$response" | tail -c 4)
    body=$(echo "$response" | head -c -4)

    if [ "$status_code" = "201" ]; then
        echo -e "${GREEN}✅ Request $request_id: $status_code${NC}"
        return 0
    else
        echo -e "${RED}❌ Request $request_id: $status_code${NC}"
        return 1
    fi
}

# Función para hacer una impresión
make_impression() {
    local request_id=$1
    local ad_id="20000000-0000-0000-0000-000000000001"
    local placement_id="30000000-0000-0000-0000-000000000001"
    local auction_id="60000000-0000-0000-0000-000000000001"

    response=$(curl -s -w "%{http_code}" -X POST "http://localhost:8080/impression" \
        -H "Content-Type: application/json" \
        -d "{
            \"ad_id\": \"$ad_id\",
            \"placement_id\": \"$placement_id\",
            \"auction_id\": \"$auction_id\",
            \"user_context\": {
                \"country\": \"CL\",
                \"device\": \"desktop\",
                \"ip\": \"192.168.1.$request_id\"
            }
        }")

    status_code=$(echo "$response" | tail -c 4)

    if [ "$status_code" = "200" ]; then
        echo -e "${GREEN}✅ Impresión $request_id: $status_code${NC}"
        return 0
    else
        echo -e "${RED}❌ Impresión $request_id: $status_code${NC}"
        return 1
    fi
}

# Función para hacer un click
make_click() {
    local request_id=$1
    local impression_id="70000000-0000-0000-0000-000000000001"

    response=$(curl -s -w "%{http_code}" -X POST "http://localhost:8080/click" \
        -H "Content-Type: application/json" \
        -d "{
            \"impression_id\": \"$impression_id\"
        }")

    status_code=$(echo "$response" | tail -c 4)

    if [ "$status_code" = "200" ]; then
        echo -e "${GREEN}✅ Click $request_id: $status_code${NC}"
        return 0
    else
        echo -e "${RED}❌ Click $request_id: $status_code${NC}"
        return 1
    fi
}

# Arrays de países y dispositivos para variar el tráfico
countries=("CL" "AR" "BR" "MX" "CO" "PE" "VE" "EC")
devices=("desktop" "mobile" "tablet")

# Contadores
successful_auctions=0
successful_impressions=0
successful_clicks=0
failed_requests=0

echo -e "${PURPLE}🚀 INICIANDO LOAD TEST...${NC}"
echo ""

# Función para ejecutar requests en paralelo
run_concurrent_requests() {
    local start_id=$1
    local count=$2

    for ((i=0; i<count; i++)); do
        local request_id=$((start_id + i))
        local country=${countries[$((RANDOM % ${#countries[@]}))]}
        local device=${devices[$((RANDOM % ${#devices[@]}))]}

        # Ejecutar subasta en background
        make_auction $request_id $country $device &

        # Ejecutar impresión en background (con menor probabilidad)
        if [ $((RANDOM % 3)) -eq 0 ]; then
            make_impression $request_id &
        fi

        # Ejecutar click en background (con menor probabilidad)
        if [ $((RANDOM % 5)) -eq 0 ]; then
            make_click $request_id &
        fi

        sleep $DELAY
    done

    # Esperar a que terminen todos los procesos en background
    wait
}

# Ejecutar el load test
start_time=$(date +%s)

for ((batch=0; batch<TOTAL_REQUESTS; batch+=CONCURRENT_REQUESTS)); do
    remaining=$((TOTAL_REQUESTS - batch))
    current_batch=$((remaining < CONCURRENT_REQUESTS ? remaining : CONCURRENT_REQUESTS))

    echo -e "${BLUE}📦 Batch $((batch/CONCURRENT_REQUESTS + 1)): $current_batch requests${NC}"
    run_concurrent_requests $batch $current_batch
    echo ""
done

end_time=$(date +%s)
duration=$((end_time - start_time))

echo -e "${PURPLE}📊 RESULTADOS DEL LOAD TEST${NC}"
echo "  ⏱️  Duración total: ${duration}s"
echo "  📈 Requests por segundo: $(echo "scale=2; $TOTAL_REQUESTS / $duration" | bc)"
echo "  🎯 Total de requests: $TOTAL_REQUESTS"
echo "  ✅ Requests exitosos: $successful_auctions"
echo "  ❌ Requests fallidos: $failed_requests"
echo ""

echo -e "${YELLOW}💡 PRÓXIMOS PASOS:${NC}"
echo "  📊 Ver métricas: ./monitor.sh once"
echo "  🌐 Dashboard web: http://localhost:8081/dashboard"
echo "  🐳 Ver logs: docker-compose logs -f ad-system"
echo ""

echo -e "${GREEN}🎉 Load test completado!${NC}"