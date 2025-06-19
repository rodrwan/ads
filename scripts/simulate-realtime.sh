#!/bin/bash

# Script para simular actividad en tiempo real del sistema de anuncios
# Este script genera impresiones y clicks continuamente para probar las actualizaciones en tiempo real

echo "🚀 Iniciando simulación de actividad en tiempo real..."
echo "📊 Las métricas del dashboard se actualizarán automáticamente cada 5 segundos"
echo ""

# Función para generar una impresión
generate_impression() {
    local ad_id=$(uuidgen)
    local placement_id=$(uuidgen)
    local auction_id=$(uuidgen)
    
    curl -s -X POST http://localhost:8080/impression \
        -H "Content-Type: application/json" \
        -d "{
            \"ad_id\": \"$ad_id\",
            \"placement_id\": \"$placement_id\",
            \"auction_id\": \"$auction_id\",
            \"user_context\": {
                \"user_agent\": \"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)\",
                \"ip\": \"192.168.1.$((RANDOM % 255))\",
                \"location\": \"ES\"
            }
        }" > /dev/null
    
    echo "👁️  Impresión generada - Ad: ${ad_id:0:8}..."
}

# Función para generar un click
generate_click() {
    local impression_id=$(uuidgen)
    
    curl -s -X POST http://localhost:8080/click \
        -H "Content-Type: application/json" \
        -d "{
            \"impression_id\": \"$impression_id\"
        }" > /dev/null
    
    echo "🖱️  Click generado - Impression: ${impression_id:0:8}..."
}

# Función para mostrar estadísticas
show_stats() {
    echo ""
    echo "📈 Estadísticas actuales:"
    echo "   • Impresiones: $(curl -s http://localhost:8080/metrics | jq -r '.total_impressions // "N/A"')"
    echo "   • Clicks: $(curl -s http://localhost:8080/metrics | jq -r '.total_clicks // "N/A"')"
    echo "   • CTR: $(curl -s http://localhost:8080/metrics | jq -r '.ctr // "N/A"')%"
    echo "   • Anuncios activos: $(curl -s http://localhost:8080/metrics | jq -r '.active_ads // "N/A"')"
    echo ""
}

# Función para mostrar instrucciones
show_instructions() {
    echo "🎯 Instrucciones:"
    echo "   1. Abre el dashboard en: http://localhost:8080/dashboard"
    echo "   2. Observa cómo los números se actualizan en tiempo real"
    echo "   3. Presiona Ctrl+C para detener la simulación"
    echo ""
    echo "⏰ La simulación generará actividad cada 2-5 segundos..."
    echo ""
}

# Mostrar instrucciones
show_instructions

# Contador para alternar entre impresiones y clicks
counter=0

# Bucle principal de simulación
while true; do
    counter=$((counter + 1))
    
    # Generar impresión (más frecuente)
    generate_impression
    
    # Cada 3 impresiones, generar un click
    if [ $((counter % 3)) -eq 0 ]; then
        sleep 1
        generate_click
    fi
    
    # Mostrar estadísticas cada 10 iteraciones
    if [ $((counter % 10)) -eq 0 ]; then
        show_stats
    fi
    
    # Esperar entre 2 y 5 segundos
    sleep $((2 + RANDOM % 4))
done 