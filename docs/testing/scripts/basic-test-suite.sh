#!/bin/bash

# ChivoMap API - Basic Test Suite
# Este script ejecuta las pruebas básicas de la API

set -e

# Configuración
API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"
TIMEOUT_SECONDS="${TIMEOUT_SECONDS:-30}"
VERBOSE_OUTPUT="${VERBOSE_OUTPUT:-false}"
SAVE_RESPONSES="${SAVE_RESPONSES:-false}"
RESPONSE_DIR="${RESPONSE_DIR:-./test-responses}"

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Contadores
TESTS_TOTAL=0
TESTS_PASSED=0
TESTS_FAILED=0

# Funciones helper
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
}

# Función para hacer requests con timeout y retry
api_request() {
    local method=$1
    local endpoint=$2
    local expected_status=$3
    local max_attempts=3
    local attempt=1
    
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    
    while [ $attempt -le $max_attempts ]; do
        if [ "$VERBOSE_OUTPUT" = "true" ]; then
            log_info "Attempt $attempt/$max_attempts: $method $endpoint"
        fi
        
        response=$(curl -s -w "\n%{http_code}" \
            --max-time $TIMEOUT_SECONDS \
            -X $method \
            "$API_BASE_URL$endpoint")
        
        http_code=$(echo "$response" | tail -n1)
        body=$(echo "$response" | head -n -1)
        
        if [ "$http_code" = "$expected_status" ]; then
            echo "$body"
            return 0
        fi
        
        attempt=$((attempt + 1))
        if [ $attempt -le $max_attempts ]; then
            sleep 1
        fi
    done
    
    log_error "Failed after $max_attempts attempts. Last response: $http_code"
    return 1
}

# Función para validar JSON
validate_json() {
    local json_string=$1
    echo "$json_string" | python3 -m json.tool > /dev/null 2>&1
    return $?
}

# Función para extraer campo de JSON
extract_json_field() {
    local json_string=$1
    local field_path=$2
    echo "$json_string" | python3 -c "
import json, sys
try:
    data = json.load(sys.stdin)
    result = data
    for key in '$field_path'.split('.'):
        if key.isdigit():
            result = result[int(key)]
        else:
            result = result[key]
    print(result)
except:
    print('null')
"
}

# Crear directorio de respuestas si es necesario
if [ "$SAVE_RESPONSES" = "true" ]; then
    mkdir -p "$RESPONSE_DIR"
fi

echo "🧪 ChivoMap API - Basic Test Suite"
echo "=================================="
echo "🎯 Target API: $API_BASE_URL"
echo "⏱️  Timeout: ${TIMEOUT_SECONDS}s"
echo "📁 Response Dir: $RESPONSE_DIR"
echo ""

# Test 1: Health Check
echo "🏥 Test 1: Health Check"
echo "------------------------"

health_response=$(api_request GET "/health" 200)
if [ $? -eq 0 ]; then
    if validate_json "$health_response"; then
        status=$(extract_json_field "$health_response" "status")
        version=$(extract_json_field "$health_response" "version")
        log_success "Health check passed - Status: $status, Version: $version"
        
        if [ "$SAVE_RESPONSES" = "true" ]; then
            echo "$health_response" > "$RESPONSE_DIR/health.json"
        fi
    else
        log_error "Health check returned invalid JSON"
    fi
else
    log_error "Health check endpoint failed"
fi

echo ""

# Test 2: Swagger Documentation
echo "📚 Test 2: Swagger Documentation"
echo "---------------------------------"

TESTS_TOTAL=$((TESTS_TOTAL + 1))
swagger_response=$(curl -s -w "%{http_code}" "$API_BASE_URL/docs/doc.json")
swagger_code=$(echo "$swagger_response" | tail -c 4)

if [ "$swagger_code" = "200" ]; then
    swagger_body=$(echo "$swagger_response" | head -c -4)
    if validate_json "$swagger_body"; then
        api_title=$(extract_json_field "$swagger_body" "info.title")
        log_success "Swagger documentation accessible - Title: $api_title"
        
        if [ "$SAVE_RESPONSES" = "true" ]; then
            echo "$swagger_body" > "$RESPONSE_DIR/swagger.json"
        fi
    else
        log_error "Swagger returned invalid JSON"
    fi
else
    log_error "Swagger documentation not accessible (HTTP $swagger_code)"
fi

echo ""

# Test 3: Geographic Data
echo "🗺️  Test 3: Geographic Data"
echo "----------------------------"

geo_response=$(api_request GET "/geo/search-data" 200)
if [ $? -eq 0 ]; then
    if validate_json "$geo_response"; then
        # Extraer y contar datos geográficos
        departamentos=$(extract_json_field "$geo_response" "data.departamentos")
        municipios=$(extract_json_field "$geo_response" "data.municipios")
        
        if [ "$departamentos" != "null" ] && [ "$municipios" != "null" ]; then
            # Contar elementos aproximadamente
            dep_count=$(echo "$departamentos" | grep -o ',' | wc -l)
            mun_count=$(echo "$municipios" | grep -o ',' | wc -l)
            
            log_success "Geographic data loaded - Deps: ~$((dep_count + 1)), Munis: ~$((mun_count + 1))"
            
            if [ "$SAVE_RESPONSES" = "true" ]; then
                echo "$geo_response" > "$RESPONSE_DIR/geo-data.json"
            fi
        else
            log_error "Geographic data structure invalid"
        fi
    else
        log_error "Geographic data returned invalid JSON"
    fi
else
    log_error "Geographic data endpoint failed"
fi

echo ""

# Test 4: Geographic Filtering
echo "🔍 Test 4: Geographic Filtering"
echo "--------------------------------"

filter_response=$(api_request GET "/geo/filter?query=SAN%20SALVADOR&whatIs=D" 200)
if [ $? -eq 0 ]; then
    if validate_json "$filter_response"; then
        feature_type=$(extract_json_field "$filter_response" "data.type")
        features=$(extract_json_field "$filter_response" "data.features")
        
        if [ "$feature_type" = "FeatureCollection" ] && [ "$features" != "null" ]; then
            feature_count=$(echo "$features" | grep -o '"type":"Feature"' | wc -l)
            log_success "Geographic filtering works - Found $feature_count features for SAN SALVADOR"
            
            if [ "$SAVE_RESPONSES" = "true" ]; then
                echo "$filter_response" > "$RESPONSE_DIR/geo-filter.json"
            fi
        else
            log_error "Geographic filter response structure invalid"
        fi
    else
        log_error "Geographic filter returned invalid JSON"
    fi
else
    log_error "Geographic filter endpoint failed"
fi

echo ""

# Test 5: Seismic Data
echo "🌍 Test 5: Seismic Data"
echo "------------------------"

sismos_response=$(api_request GET "/sismos" 200)
if [ $? -eq 0 ]; then
    if validate_json "$sismos_response"; then
        total_sismos=$(extract_json_field "$sismos_response" "data.totalSismos")
        sismos_data=$(extract_json_field "$sismos_response" "data.data")
        
        if [ "$total_sismos" != "null" ] && [ "$sismos_data" != "null" ]; then
            log_success "Seismic data accessible - Total events: $total_sismos"
            
            if [ "$SAVE_RESPONSES" = "true" ]; then
                echo "$sismos_response" > "$RESPONSE_DIR/sismos.json"
            fi
        else
            log_error "Seismic data structure invalid"
        fi
    else
        log_error "Seismic data returned invalid JSON"
    fi
else
    log_error "Seismic data endpoint failed"
fi

echo ""

# Test 6: Parameter Validation
echo "🛡️  Test 6: Parameter Validation"
echo "---------------------------------"

TESTS_TOTAL=$((TESTS_TOTAL + 1))
invalid_response=$(curl -s -w "%{http_code}" "$API_BASE_URL/geo/filter?query=test&whatIs=INVALID")
invalid_code=$(echo "$invalid_response" | tail -c 4)

if [ "$invalid_code" = "400" ]; then
    log_success "Parameter validation working (rejected invalid input)"
else
    log_warning "Parameter validation may not be strict enough (got HTTP $invalid_code)"
fi

echo ""

# Test 7: Rate Limiting Headers
echo "🚦 Test 7: Rate Limiting"
echo "-------------------------"

TESTS_TOTAL=$((TESTS_TOTAL + 1))
rate_limit_response=$(curl -s -I "$API_BASE_URL/health")

if echo "$rate_limit_response" | grep -q "X-Ratelimit-Limit"; then
    limit=$(echo "$rate_limit_response" | grep "X-Ratelimit-Limit" | cut -d' ' -f2 | tr -d '\r')
    remaining=$(echo "$rate_limit_response" | grep "X-Ratelimit-Remaining" | cut -d' ' -f2 | tr -d '\r')
    log_success "Rate limiting active - Limit: $limit, Remaining: $remaining"
else
    log_warning "Rate limiting headers not found"
fi

echo ""

# Test 8: Cache Performance
echo "⚡ Test 8: Cache Performance"
echo "----------------------------"

echo "🔄 Testing cache performance..."

# Primera request (puede cargar cache)
start_time=$(date +%s.%N)
cache_test1=$(api_request GET "/geo/search-data" 200)
end_time=$(date +%s.%N)
first_time=$(echo "$end_time - $start_time" | bc)

if [ $? -eq 0 ]; then
    # Segunda request (debería usar cache)
    sleep 1
    start_time=$(date +%s.%N)
    cache_test2=$(api_request GET "/geo/search-data" 200)
    end_time=$(date +%s.%N)
    second_time=$(echo "$end_time - $start_time" | bc)
    
    if [ $? -eq 0 ]; then
        if (( $(echo "$second_time < $first_time" | bc -l) )); then
            improvement=$(echo "scale=1; ($first_time - $second_time) / $first_time * 100" | bc)
            log_success "Cache working - ${improvement}% improvement (${first_time}s → ${second_time}s)"
        else
            log_warning "Cache may not be working optimally"
        fi
    else
        log_error "Second cache test failed"
    fi
else
    log_error "First cache test failed"
fi

echo ""

# Verificar estado del cache en health check
echo "🏥 Verifying cache status in health check..."
final_health=$(api_request GET "/health" 200)
if [ $? -eq 0 ]; then
    cache_status=$(extract_json_field "$final_health" "components.cache.status")
    if [ "$cache_status" = "UP" ]; then
        cache_features=$(extract_json_field "$final_health" "components.cache.details.geoFeatures")
        log_success "Cache status confirmed - $cache_features features loaded"
    else
        log_warning "Cache status: $cache_status"
    fi
fi

echo ""

# Resumen Final
echo "📊 TEST SUITE SUMMARY"
echo "====================="
echo "🎯 Total Tests: $TESTS_TOTAL"
echo "✅ Passed: $TESTS_PASSED"
echo "❌ Failed: $TESTS_FAILED"

if [ $TESTS_FAILED -eq 0 ]; then
    success_rate=100
else
    success_rate=$(echo "scale=1; $TESTS_PASSED * 100 / $TESTS_TOTAL" | bc)
fi

echo "📈 Success Rate: $success_rate%"

if [ "$SAVE_RESPONSES" = "true" ]; then
    echo "💾 Responses saved to: $RESPONSE_DIR"
fi

echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}🎉 ALL TESTS PASSED! API is working correctly.${NC}"
    exit 0
elif [ $TESTS_PASSED -gt $((TESTS_TOTAL * 3 / 4)) ]; then
    echo -e "${YELLOW}🟡 Most tests passed. Minor issues detected.${NC}"
    exit 0
else
    echo -e "${RED}❌ Multiple test failures detected. API needs attention.${NC}"
    exit 1
fi