#!/bin/bash

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== GENERANDO DATOS DE PRUEBA REALISTAS ===${NC}"
echo ""

# Configuración
TOTAL_IMPRESSIONS=${1:-100}
TOTAL_CLICKS=${2:-15}
PLACEMENT_ID="30000000-0000-0000-0000-000000000001"
AD_ID="20000000-0000-0000-0000-000000000001"

echo -e "${YELLOW}Configuración:${NC}"
echo "  👁️  Impresiones a generar: $TOTAL_IMPRESSIONS"
echo "  🖱️  Clicks a generar: $TOTAL_CLICKS"
echo "  📍 Placement ID: $PLACEMENT_ID"
echo "  🎨 Ad ID: $AD_ID"
echo ""

# Función para generar impresiones
generate_impressions() {
    echo -e "${CYAN}👁️  Generando impresiones...${NC}"
    
    for ((i=1; i<=TOTAL_IMPRESSIONS; i++)); do
        # Generar auction ID único
        auction_id=$(uuidgen)
        
        # Generar impresión
        impression_response=$(curl -s -X POST "http://localhost:8080/impression" \
            -H "Content-Type: application/json" \
            -d "{
                \"ad_id\": \"$AD_ID\",
                \"placement_id\": \"$PLACEMENT_ID\",
                \"auction_id\": \"$auction_id\",
                \"user_context\": {
                    \"country\": \"CL\",
                    \"device\": \"desktop\",
                    \"ip\": \"192.168.1.$i\"
                }
            }")
        
        if [ $((i % 20)) -eq 0 ]; then
            echo -e "${GREEN}  ✅ Generadas $i impresiones${NC}"
        fi
    done
    
    echo -e "${GREEN}  ✅ Total de impresiones generadas: $TOTAL_IMPRESSIONS${NC}"
    echo ""
}

# Función para generar clicks
generate_clicks() {
    echo -e "${CYAN}🖱️  Generando clicks...${NC}"
    
    for ((i=1; i<=TOTAL_CLICKS; i++)); do
        # Generar click
        click_response=$(curl -s -X POST "http://localhost:8080/click" \
            -H "Content-Type: application/json" \
            -d "{
                \"impression_id\": \"70000000-0000-0000-0000-000000000001\"
            }")
        
        if [ $((i % 5)) -eq 0 ]; then
            echo -e "${GREEN}  ✅ Generados $i clicks${NC}"
        fi
    done
    
    echo -e "${GREEN}  ✅ Total de clicks generados: $TOTAL_CLICKS${NC}"
    echo ""
}

# Función para generar subastas
generate_auctions() {
    echo -e "${CYAN}🎯 Generando subastas...${NC}"
    
    local auction_count=$((TOTAL_IMPRESSIONS + 10))
    
    for ((i=1; i<=auction_count; i++)); do
        # Generar subasta
        auction_response=$(curl -s -X POST "http://localhost:8080/auction" \
            -H "Content-Type: application/json" \
            -d "{
                \"placement_id\": \"$PLACEMENT_ID\",
                \"request_context\": {
                    \"country\": \"CL\",
                    \"device\": \"desktop\",
                    \"interests\": [\"technology\", \"gaming\"],
                    \"os\": \"Windows\",
                    \"browser\": \"Chrome\",
                    \"ip\": \"192.168.1.$i\",
                    \"user_agent\": \"Mozilla/5.0...\",
                    \"referer\": \"https://example.com\"
                }
            }")
        
        if [ $((i % 20)) -eq 0 ]; then
            echo -e "${GREEN}  ✅ Generadas $i subastas${NC}"
        fi
    done
    
    echo -e "${GREEN}  ✅ Total de subastas generadas: $auction_count${NC}"
    echo ""
}

# Función para mostrar métricas finales
show_final_metrics() {
    echo -e "${PURPLE}📊 MÉTRICAS FINALES${NC}"
    
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
        echo -e "${RED}  ❌ Error obteniendo métricas${NC}"
    fi
    echo ""
}

# Ejecutar generación de datos
echo -e "${PURPLE}🚀 INICIANDO GENERACIÓN DE DATOS...${NC}"
echo ""

generate_auctions
generate_impressions
generate_clicks

echo -e "${PURPLE}📈 CALCULANDO MÉTRICAS...${NC}"
echo ""

show_final_metrics

echo -e "${YELLOW}💡 PRÓXIMOS PASOS:${NC}"
echo "  📊 Ver dashboard: http://localhost:8081/dashboard"
echo "  📋 Ver campañas: http://localhost:8081/campaigns"
echo "  🔐 Login: http://localhost:8081/login (acme@example.com)"
echo "  📈 Monitoreo: ./monitor.sh once"
echo ""

echo -e "${GREEN}🎉 Datos de prueba generados exitosamente!${NC}"
echo -e "${BLUE}🌐 La nueva interfaz está lista para mostrar métricas realistas${NC}" 