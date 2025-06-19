#!/bin/bash

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== SIMULACIÓN DE INTERACCIÓN DE ANUNCIOS ===${NC}"
echo ""

# Configuración
PLACEMENT_ID="30000000-0000-0000-0000-000000000001"
AD_ID="20000000-0000-0000-0000-000000000001"
TOTAL_SIMULATIONS=${1:-10}
DELAY=${2:-1}

echo -e "${YELLOW}Configuración:${NC}"
echo "  🎯 Total de simulaciones: $TOTAL_SIMULATIONS"
echo "  ⏱️  Delay entre simulaciones: ${DELAY}s"
echo "  📍 Placement ID: $PLACEMENT_ID"
echo "  🎨 Ad ID: $AD_ID"
echo ""

# Función para simular una interacción completa
simulate_ad_interaction() {
    local interaction_id=$1
    local country=$2
    local device=$3
    
    echo -e "${CYAN}🔄 Simulación $interaction_id - $country/$device${NC}"
    
    # Paso 1: Subasta RTB
    echo -e "${BLUE}  1️⃣  Creando subasta RTB...${NC}"
    auction_response=$(curl -s -X POST "http://localhost:8080/auction" \
        -H "Content-Type: application/json" \
        -d "{
            \"placement_id\": \"$PLACEMENT_ID\",
            \"request_context\": {
                \"country\": \"$country\",
                \"device\": \"$device\",
                \"interests\": [\"technology\", \"gaming\"],
                \"os\": \"Windows\",
                \"browser\": \"Chrome\",
                \"ip\": \"192.168.1.$interaction_id\",
                \"user_agent\": \"Mozilla/5.0...\",
                \"referer\": \"https://example.com\"
            }
        }")
    
    if echo "$auction_response" | grep -q "ad_id"; then
        echo -e "${GREEN}     ✅ Subasta exitosa${NC}"
        ad_id=$(echo "$auction_response" | jq -r '.ad_id' 2>/dev/null)
        echo -e "${BLUE}     📋 Anuncio ganador: $ad_id${NC}"
    else
        echo -e "${RED}     ❌ Error en subasta${NC}"
        return 1
    fi
    
    # Paso 2: Registrar impresión
    echo -e "${BLUE}  2️⃣  Registrando impresión...${NC}"
    impression_response=$(curl -s -X POST "http://localhost:8080/impression" \
        -H "Content-Type: application/json" \
        -d "{
            \"ad_id\": \"$AD_ID\",
            \"placement_id\": \"$PLACEMENT_ID\",
            \"auction_id\": \"60000000-0000-0000-0000-000000000001\",
            \"user_context\": {
                \"country\": \"$country\",
                \"device\": \"$device\",
                \"ip\": \"192.168.1.$interaction_id\"
            }
        }")
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}     ✅ Impresión registrada${NC}"
    else
        echo -e "${RED}     ❌ Error registrando impresión${NC}"
    fi
    
    # Paso 3: Simular click (con probabilidad)
    if [ $((RANDOM % 3)) -eq 0 ]; then
        echo -e "${BLUE}  3️⃣  Simulando click...${NC}"
        click_response=$(curl -s -X POST "http://localhost:8080/click" \
            -H "Content-Type: application/json" \
            -d "{
                \"impression_id\": \"70000000-0000-0000-0000-000000000001\"
            }")
        
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}     ✅ Click registrado${NC}"
        else
            echo -e "${RED}     ❌ Error registrando click${NC}"
        fi
    else
        echo -e "${YELLOW}  3️⃣  No se simuló click (probabilidad)${NC}"
    fi
    
    echo ""
}

# Arrays para variar el tráfico
countries=("CL" "AR" "BR" "MX" "CO" "PE" "VE" "EC")
devices=("desktop" "mobile" "tablet")

# Contadores
successful_interactions=0
failed_interactions=0

echo -e "${PURPLE}🚀 INICIANDO SIMULACIÓN...${NC}"
echo ""

# Ejecutar simulaciones
for ((i=1; i<=TOTAL_SIMULATIONS; i++)); do
    country=${countries[$((RANDOM % ${#countries[@]}))]}
    device=${devices[$((RANDOM % ${#devices[@]}))]}
    
    if simulate_ad_interaction $i $country $device; then
        ((successful_interactions++))
    else
        ((failed_interactions++))
    fi
    
    sleep $DELAY
done

echo -e "${PURPLE}📊 RESULTADOS DE LA SIMULACIÓN${NC}"
echo "  🎯 Total de simulaciones: $TOTAL_SIMULATIONS"
echo "  ✅ Exitosas: $successful_interactions"
echo "  ❌ Fallidas: $failed_interactions"
echo "  📈 Tasa de éxito: $(echo "scale=1; $successful_interactions * 100 / $TOTAL_SIMULATIONS" | bc)%"
echo ""

# Mostrar métricas actuales
echo -e "${CYAN}📈 MÉTRICAS ACTUALES DEL SISTEMA${NC}"
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
    FROM auctions;
" 2>/dev/null)

if [ $? -eq 0 ]; then
    echo "$metrics" | while read line; do
        if [ -n "$line" ]; then
            echo -e "${GREEN}  $line${NC}"
        fi
    done
else
    echo -e "${RED}  ❌ Error obteniendo métricas${NC}"
fi

echo ""
echo -e "${YELLOW}💡 PRÓXIMOS PASOS:${NC}"
echo "  📊 Ver dashboard: http://localhost:8081/dashboard"
echo "  🌐 Ver anuncios: http://localhost:8081/ads"
echo "  📈 Monitoreo: ./monitor.sh once"
echo "  🧪 Más pruebas: ./load-test.sh 100 10 0.1"
echo ""

echo -e "${GREEN}🎉 Simulación completada!${NC}" 