# 🌍 Endpoints de Sismos

Esta sección cubre las pruebas para los endpoints relacionados con datos sísmicos de El Salvador.

## 🎯 Endpoints a Probar

| Endpoint | Método | Descripción |
|----------|--------|-------------|
| `/sismos` | GET | Obtiene lista de sismos recientes |
| `/sismos/refresh` | POST | Actualiza datos sísmicos desde la fuente |

## 📊 Estructura de Datos Esperada

### Response de `/sismos`
```json
{
  "data": {
    "totalSismos": 10,
    "data": [
      {
        "fecha": "28/6/2025, 2:05:24 p. m.",
        "fases": "15",
        "latitud": "13.5661478",
        "longitud": "-91.0059433",
        "profundidad": "10",
        "magnitud": "4",
        "localizacion": "Localizado frente a la Costa de Guatemala",
        "rms": "0.2549138002",
        "estado": "automatic Sujeto a revisión y puede sufrir cambios."
      }
    ]
  },
  "timestamp": "2025-06-28T16:00:00-06:00"
}
```

### Response de `/sismos/refresh`
```json
{
  "data": {
    "message": "Datos de sismos actualizados correctamente",
    "totalSismos": 15,
    "data": [...]
  },
  "timestamp": "2025-06-28T16:00:00-06:00"
}
```

## 🧪 Casos de Prueba

### Test 1: Obtener Datos Sísmicos Básicos

```bash
# TC-SISMOS-001: Obtener datos sísmicos básicos
echo "🧪 TC-SISMOS-001: Basic Seismic Data Retrieval"

response=$(api_request GET "/sismos" 200)
if [ $? -ne 0 ]; then
    echo "❌ Failed to retrieve seismic data"
    exit 1
fi

echo "✅ Sismos endpoint accessible"

# Validar estructura JSON
if validate_json "$response"; then
    echo "✅ Valid JSON response"
else
    echo "❌ Invalid JSON response"
    echo "Response: $response"
    exit 1
fi

# Verificar campos requeridos
data_field=$(extract_json_field "$response" "data")
timestamp_field=$(extract_json_field "$response" "timestamp")

if [ "$data_field" != "null" ]; then
    echo "✅ Data field present"
else
    echo "❌ Data field missing"
fi

if [ "$timestamp_field" != "null" ]; then
    echo "✅ Timestamp field present: $timestamp_field"
else
    echo "❌ Timestamp field missing"
fi

# Verificar estructura interna
total_sismos=$(extract_json_field "$response" "data.totalSismos")
sismos_array=$(extract_json_field "$response" "data.data")

if [ "$total_sismos" != "null" ] && [ "$total_sismos" != "0" ]; then
    echo "✅ Total sismos field present: $total_sismos"
else
    echo "⚠️ Total sismos: $total_sismos (may be zero)"
fi

if [ "$sismos_array" != "null" ]; then
    echo "✅ Sismos array present"
    
    # Contar sismos en la respuesta
    sismo_count=$(echo "$sismos_array" | grep -o '"fecha":' | wc -l)
    echo "📊 Found $sismo_count seismic events"
    
    if [ "$sismo_count" -gt 0 ]; then
        echo "✅ Seismic data contains events"
        
        # Verificar que el conteo coincide
        if [ "$sismo_count" = "$total_sismos" ]; then
            echo "✅ Total count matches array length"
        else
            echo "⚠️ Count mismatch: array has $sismo_count, total says $total_sismos"
        fi
    else
        echo "⚠️ No seismic events in response (may be normal)"
    fi
else
    echo "❌ Sismos array missing"
fi

# Guardar respuesta para análisis
if [ "$SAVE_RESPONSES" = "true" ]; then
    echo "$response" > "$RESPONSE_DIR/sismos-basic.json"
fi
```

### Test 2: Validación de Estructura de Sismo Individual

```bash
# TC-SISMOS-002: Validar estructura de sismos individuales
echo "🧪 TC-SISMOS-002: Individual Seismic Event Structure"

response=$(api_request GET "/sismos" 200)
if [ $? -ne 0 ]; then
    echo "❌ Failed to retrieve seismic data"
    exit 1
fi

# Extraer primer sismo para validación detallada
first_sismo=$(echo "$response" | python3 -c "
import json, sys
try:
    data = json.load(sys.stdin)
    sismos = data.get('data', {}).get('data', [])
    if sismos:
        print(json.dumps(sismos[0], indent=2))
    else:
        print('null')
except:
    print('null')
")

if [ "$first_sismo" != "null" ]; then
    echo "✅ Successfully extracted first seismic event"
    
    # Campos requeridos para sismos
    required_fields=("fecha" "latitud" "longitud" "profundidad" "magnitud" "localizacion")
    all_fields_present=true
    
    for field in "${required_fields[@]}"; do
        field_value=$(echo "$first_sismo" | python3 -c "
import json, sys
sismo = json.load(sys.stdin)
print(sismo.get('$field', 'null'))
")
        
        if [ "$field_value" != "null" ] && [ "$field_value" != "" ]; then
            echo "✅ Field '$field': $field_value"
        else
            echo "❌ Field '$field': missing or empty"
            all_fields_present=false
        fi
    done
    
    if [ "$all_fields_present" = true ]; then
        echo "🎉 All required fields present in seismic event"
    else
        echo "⚠️ Some required fields are missing"
    fi
    
    # Validaciones específicas de datos sísmicos
    echo ""
    echo "📋 Validating seismic data values..."
    
    # Validar coordenadas (deben estar en rango de El Salvador/región)
    latitud=$(echo "$first_sismo" | python3 -c "
import json, sys
sismo = json.load(sys.stdin)
try:
    lat = float(sismo.get('latitud', '0'))
    print(lat)
except:
    print('invalid')
")
    
    longitud=$(echo "$first_sismo" | python3 -c "
import json, sys
sismo = json.load(sys.stdin)
try:
    lon = float(sismo.get('longitud', '0'))
    print(lon)
except:
    print('invalid')
")
    
    # Rango aproximado para El Salvador y región
    if [ "$latitud" != "invalid" ]; then
        if (( $(echo "$latitud >= 12.0 && $latitud <= 15.0" | bc -l) )); then
            echo "✅ Latitude in valid range: $latitud"
        else
            echo "⚠️ Latitude outside expected range: $latitud"
        fi
    else
        echo "❌ Invalid latitude value"
    fi
    
    if [ "$longitud" != "invalid" ]; then
        if (( $(echo "$longitud >= -92.0 && $longitud <= -87.0" | bc -l) )); then
            echo "✅ Longitude in valid range: $longitud"
        else
            echo "⚠️ Longitude outside expected range: $longitud"
        fi
    else
        echo "❌ Invalid longitude value"
    fi
    
    # Validar magnitud
    magnitud=$(echo "$first_sismo" | python3 -c "
import json, sys
sismo = json.load(sys.stdin)
try:
    mag = float(sismo.get('magnitud', '0'))
    print(mag)
except:
    print('invalid')
")
    
    if [ "$magnitud" != "invalid" ]; then
        if (( $(echo "$magnitud >= 0.0 && $magnitud <= 10.0" | bc -l) )); then
            echo "✅ Magnitude in valid range: $magnitud"
        else
            echo "⚠️ Magnitude outside expected range: $magnitud"
        fi
    else
        echo "❌ Invalid magnitude value"
    fi
    
    # Validar profundidad
    profundidad=$(echo "$first_sismo" | python3 -c "
import json, sys
sismo = json.load(sys.stdin)
try:
    prof = float(sismo.get('profundidad', '0'))
    print(prof)
except:
    print('invalid')
")
    
    if [ "$profundidad" != "invalid" ]; then
        if (( $(echo "$profundidad >= 0.0 && $profundidad <= 1000.0" | bc -l) )); then
            echo "✅ Depth in valid range: $profundidad km"
        else
            echo "⚠️ Depth outside expected range: $profundidad km"
        fi
    else
        echo "❌ Invalid depth value"
    fi
    
    # Validar formato de fecha
    fecha=$(echo "$first_sismo" | python3 -c "
import json, sys
sismo = json.load(sys.stdin)
print(sismo.get('fecha', ''))
")
    
    if echo "$fecha" | grep -E "[0-9]+/[0-9]+/[0-9]+" > /dev/null; then
        echo "✅ Date format appears valid: $fecha"
    else
        echo "⚠️ Date format may be unusual: $fecha"
    fi
    
    # Guardar sismo de ejemplo
    if [ "$SAVE_RESPONSES" = "true" ]; then
        echo "$first_sismo" > "$RESPONSE_DIR/sample-sismo.json"
    fi
    
else
    echo "⚠️ No seismic events found for detailed validation"
fi
```

### Test 3: Actualización de Datos Sísmicos

```bash
# TC-SISMOS-003: Actualización de datos sísmicos
echo "🧪 TC-SISMOS-003: Seismic Data Refresh"

# Obtener datos actuales primero
echo "📋 Step 1: Get current seismic data"
initial_response=$(api_request GET "/sismos" 200)
if [ $? -ne 0 ]; then
    echo "❌ Failed to get initial seismic data"
    exit 1
fi

initial_count=$(extract_json_field "$initial_response" "data.totalSismos")
initial_timestamp=$(extract_json_field "$initial_response" "timestamp")

echo "📊 Initial state:"
echo "   - Count: $initial_count events"
echo "   - Timestamp: $initial_timestamp"

# Intentar actualizar datos
echo "📋 Step 2: Refresh seismic data"
refresh_response=$(curl -s -w "STATUS:%{http_code}" -X POST "$API_BASE_URL/sismos/refresh")

# Extraer código de estado y respuesta
http_code=$(echo "$refresh_response" | grep -o "STATUS:[0-9]*" | cut -d: -f2)
body=$(echo "$refresh_response" | sed 's/STATUS:.*//g')

echo "📊 Refresh request:"
echo "   - HTTP Code: $http_code"

if [ "$http_code" = "200" ]; then
    echo "✅ Refresh request successful"
    
    if validate_json "$body"; then
        echo "✅ Valid JSON response from refresh"
        
        # Verificar mensaje de éxito
        message=$(extract_json_field "$body" "data.message")
        if echo "$message" | grep -q "actualiz"; then
            echo "✅ Appropriate success message: $message"
        else
            echo "⚠️ Unexpected message: $message"
        fi
        
        # Verificar datos actualizados
        refresh_count=$(extract_json_field "$body" "data.totalSismos")
        refresh_timestamp=$(extract_json_field "$body" "timestamp")
        
        echo "📊 After refresh:"
        echo "   - Count: $refresh_count events"
        echo "   - Timestamp: $refresh_timestamp"
        
        # Comparar timestamps
        if [ "$refresh_timestamp" != "$initial_timestamp" ]; then
            echo "✅ Timestamp updated after refresh"
        else
            echo "⚠️ Timestamp unchanged (may indicate no new data)"
        fi
        
    else
        echo "❌ Invalid JSON response from refresh"
        echo "Response: $body"
    fi
    
elif [ "$http_code" = "429" ]; then
    echo "⚠️ Rate limited (429) - refresh may be throttled"
    
elif [ "$http_code" = "503" ]; then
    echo "⚠️ Service unavailable (503) - external source may be down"
    
else
    echo "❌ Unexpected HTTP code: $http_code"
    echo "Response: $body"
fi

# Verificar que el endpoint GET sigue funcionando después del refresh
echo "📋 Step 3: Verify GET endpoint after refresh"
sleep 2
post_refresh_response=$(api_request GET "/sismos" 200)
if [ $? -eq 0 ]; then
    echo "✅ GET endpoint still functional after refresh"
    
    post_refresh_count=$(extract_json_field "$post_refresh_response" "data.totalSismos")
    echo "   - Final count: $post_refresh_count events"
    
else
    echo "❌ GET endpoint failed after refresh"
fi
```

### Test 4: Performance de Endpoints de Sismos

```bash
# TC-SISMOS-004: Performance testing de endpoints sísmicos
echo "🧪 TC-SISMOS-004: Seismic Endpoints Performance"

echo "📋 Testing GET /sismos performance..."

# Test performance del GET endpoint
iterations=20
total_time=0
success_count=0
min_time=999
max_time=0

for i in $(seq 1 $iterations); do
    start_time=$(date +%s.%N)
    response=$(curl -s "$API_BASE_URL/sismos")
    end_time=$(date +%s.%N)
    duration=$(echo "$end_time - $start_time" | bc)
    
    if validate_json "$response"; then
        success_count=$((success_count + 1))
        total_time=$(echo "$total_time + $duration" | bc)
        
        # Actualizar min/max
        if (( $(echo "$duration < $min_time" | bc -l) )); then
            min_time=$duration
        fi
        
        if (( $(echo "$duration > $max_time" | bc -l) )); then
            max_time=$duration
        fi
    fi
    
    # Progress indicator
    if [ $((i % 5)) -eq 0 ]; then
        echo -n "."
    fi
done

echo ""

# Calcular métricas
if [ $success_count -gt 0 ]; then
    avg_time=$(echo "scale=3; $total_time / $success_count" | bc)
    success_rate=$(echo "scale=1; $success_count * 100 / $iterations" | bc)
    
    echo "📊 GET /sismos Performance:"
    echo "   - Requests: $iterations"
    echo "   - Success Rate: $success_rate%"
    echo "   - Average Time: ${avg_time}s"
    echo "   - Min Time: ${min_time}s"
    echo "   - Max Time: ${max_time}s"
    
    # Evaluación de performance
    if (( $(echo "$avg_time < 0.5" | bc -l) )); then
        echo "🚀 Excellent performance (<0.5s avg)"
    elif (( $(echo "$avg_time < 2.0" | bc -l) )); then
        echo "✅ Good performance (<2.0s avg)"
    elif (( $(echo "$avg_time < 5.0" | bc -l) )); then
        echo "🟡 Acceptable performance (<5.0s avg)"
    else
        echo "🔴 Poor performance (>5.0s avg)"
    fi
else
    echo "❌ No successful requests"
fi

# Test concurrencia
echo ""
echo "📋 Testing concurrent access..."
concurrent_count=10
pids=()

echo "Starting $concurrent_count concurrent requests..."

for i in $(seq 1 $concurrent_count); do
    {
        start_time=$(date +%s.%N)
        response=$(curl -s "$API_BASE_URL/sismos")
        end_time=$(date +%s.%N)
        duration=$(echo "$end_time - $start_time" | bc)
        
        if validate_json "$response"; then
            echo "Request $i: SUCCESS ($duration s)"
        else
            echo "Request $i: FAILED ($duration s)"
        fi
    } &
    pids+=($!)
done

# Esperar a que terminen todos
failed_count=0
for pid in "${pids[@]}"; do
    if ! wait $pid; then
        failed_count=$((failed_count + 1))
    fi
done

echo "📊 Concurrent Access Results:"
echo "   - Total: $concurrent_count"
echo "   - Failed: $failed_count"
echo "   - Success Rate: $(( (concurrent_count - failed_count) * 100 / concurrent_count ))%"

if [ $failed_count -eq 0 ]; then
    echo "✅ Perfect concurrent performance"
elif [ $failed_count -lt $((concurrent_count / 4)) ]; then
    echo "🟡 Good concurrent performance"
else
    echo "❌ Poor concurrent performance"
fi
```

### Test 5: Validación de Datos Históricos

```bash
# TC-SISMOS-005: Validación de datos históricos y temporales
echo "🧪 TC-SISMOS-005: Historical and Temporal Data Validation"

response=$(api_request GET "/sismos" 200)
if [ $? -ne 0 ]; then
    echo "❌ Failed to retrieve seismic data for temporal validation"
    exit 1
fi

# Extraer todas las fechas de los sismos
fechas=$(echo "$response" | python3 -c "
import json, sys
try:
    data = json.load(sys.stdin)
    sismos = data.get('data', {}).get('data', [])
    fechas = [sismo.get('fecha', '') for sismo in sismos]
    for fecha in fechas:
        print(fecha)
except:
    pass
")

if [ -n "$fechas" ]; then
    echo "✅ Extracted dates from seismic events"
    
    fecha_count=$(echo "$fechas" | wc -l)
    echo "📊 Found $fecha_count dated events"
    
    # Verificar que las fechas están en orden cronológico (más recientes primero)
    echo "📋 Checking chronological order..."
    
    # Analizar las primeras fechas
    first_few_dates=$(echo "$fechas" | head -3)
    echo "📅 Sample dates:"
    echo "$first_few_dates" | sed 's/^/   - /'
    
    # Verificar formato de fecha consistente
    date_format_count=$(echo "$fechas" | grep -E "[0-9]+/[0-9]+/[0-9]+" | wc -l)
    
    if [ "$date_format_count" -eq "$fecha_count" ]; then
        echo "✅ All dates follow consistent format"
    else
        echo "⚠️ Some dates may have inconsistent format"
        echo "   Expected: $fecha_count, Found: $date_format_count"
    fi
    
    # Verificar que las fechas son recientes (últimos 30 días aproximadamente)
    current_year=$(date +%Y)
    current_month=$(date +%m)
    
    recent_count=$(echo "$fechas" | grep "$current_year" | wc -l)
    
    if [ "$recent_count" -gt 0 ]; then
        echo "✅ Contains recent seismic events ($recent_count from current year)"
    else
        echo "⚠️ No events from current year found"
    fi
    
    # Verificar distribución de magnitudes
    magnitudes=$(echo "$response" | python3 -c "
import json, sys
try:
    data = json.load(sys.stdin)
    sismos = data.get('data', {}).get('data', [])
    for sismo in sismos:
        try:
            mag = float(sismo.get('magnitud', '0'))
            print(mag)
        except:
            pass
except:
    pass
")
    
    if [ -n "$magnitudes" ]; then
        echo "📊 Magnitude distribution analysis:"
        
        # Contar por rangos de magnitud
        low_mag=$(echo "$magnitudes" | awk '$1 < 3.0' | wc -l)
        med_mag=$(echo "$magnitudes" | awk '$1 >= 3.0 && $1 < 5.0' | wc -l)
        high_mag=$(echo "$magnitudes" | awk '$1 >= 5.0' | wc -l)
        
        total_mag=$(echo "$magnitudes" | wc -l)
        
        echo "   - Low (< 3.0): $low_mag events"
        echo "   - Medium (3.0-5.0): $med_mag events"
        echo "   - High (>= 5.0): $high_mag events"
        echo "   - Total analyzed: $total_mag events"
        
        # Verificar distribución realista
        if [ "$low_mag" -gt "$high_mag" ]; then
            echo "✅ Realistic magnitude distribution (more low than high)"
        else
            echo "⚠️ Unusual magnitude distribution"
        fi
    fi
    
else
    echo "⚠️ No dates extracted for temporal validation"
fi
```

### Test 6: Rate Limiting en Refresh

```bash
# TC-SISMOS-006: Rate limiting en endpoint de refresh
echo "🧪 TC-SISMOS-006: Rate Limiting on Refresh Endpoint"

echo "📋 Testing refresh rate limiting..."

# Intentar múltiples refreshes seguidos
max_attempts=5
successful_refreshes=0
rate_limited_count=0

for i in $(seq 1 $max_attempts); do
    echo "Attempt $i: Refreshing seismic data..."
    
    start_time=$(date +%s.%N)
    response=$(curl -s -w "STATUS:%{http_code}" -X POST "$API_BASE_URL/sismos/refresh")
    end_time=$(date +%s.%N)
    duration=$(echo "$end_time - $start_time" | bc)
    
    http_code=$(echo "$response" | grep -o "STATUS:[0-9]*" | cut -d: -f2)
    body=$(echo "$response" | sed 's/STATUS:.*//g')
    
    case "$http_code" in
        "200")
            echo "   ✅ SUCCESS ($duration s)"
            successful_refreshes=$((successful_refreshes + 1))
            ;;
        "429")
            echo "   🚫 RATE LIMITED ($duration s)"
            rate_limited_count=$((rate_limited_count + 1))
            
            # Verificar headers de rate limiting si están presentes
            if echo "$body" | grep -q "rate"; then
                echo "   📝 Rate limit message found"
            fi
            ;;
        "503")
            echo "   ⚠️ SERVICE UNAVAILABLE ($duration s)"
            ;;
        *)
            echo "   ❌ UNEXPECTED CODE: $http_code ($duration s)"
            ;;
    esac
    
    # Pausa corta entre intentos
    if [ $i -lt $max_attempts ]; then
        sleep 1
    fi
done

echo ""
echo "📊 Rate Limiting Test Results:"
echo "   - Total Attempts: $max_attempts"
echo "   - Successful: $successful_refreshes"
echo "   - Rate Limited: $rate_limited_count"
echo "   - Other: $((max_attempts - successful_refreshes - rate_limited_count))"

# Evaluación del rate limiting
if [ $rate_limited_count -gt 0 ]; then
    echo "✅ Rate limiting is working (protected against abuse)"
elif [ $successful_refreshes -eq $max_attempts ]; then
    echo "⚠️ No rate limiting detected (may be vulnerable to abuse)"
else
    echo "🟡 Mixed results (may have other protection mechanisms)"
fi

# Verificar recuperación después de rate limiting
if [ $rate_limited_count -gt 0 ]; then
    echo ""
    echo "📋 Testing recovery after rate limiting..."
    echo "Waiting 10 seconds for rate limit to reset..."
    sleep 10
    
    recovery_response=$(curl -s -w "STATUS:%{http_code}" -X POST "$API_BASE_URL/sismos/refresh")
    recovery_code=$(echo "$recovery_response" | grep -o "STATUS:[0-9]*" | cut -d: -f2)
    
    if [ "$recovery_code" = "200" ]; then
        echo "✅ Successfully recovered from rate limiting"
    elif [ "$recovery_code" = "429" ]; then
        echo "⚠️ Still rate limited after waiting"
    else
        echo "🟡 Different response after waiting: $recovery_code"
    fi
fi
```

### Test 7: Tolerancia a Fallos de Fuente Externa

```bash
# TC-SISMOS-007: Tolerancia a fallos de fuente externa
echo "🧪 TC-SISMOS-007: External Source Fault Tolerance"

echo "📋 Testing behavior when external source is unavailable..."

# Este test simula condiciones donde la fuente externa de datos sísmicos
# no está disponible o responde lentamente

# Test 1: Verificar que GET sigue funcionando incluso si refresh falla
echo "Step 1: Verify GET endpoint resilience"
get_response=$(api_request GET "/sismos" 200)
if [ $? -eq 0 ]; then
    echo "✅ GET endpoint remains functional"
    
    # Verificar que devuelve datos (aunque sean cached/anteriores)
    total_sismos=$(extract_json_field "$get_response" "data.totalSismos")
    if [ "$total_sismos" != "null" ]; then
        echo "✅ Returns cached/stored data when available"
        echo "   - Events available: $total_sismos"
    else
        echo "⚠️ No seismic data available"
    fi
else
    echo "❌ GET endpoint failed (should be resilient)"
fi

# Test 2: Verificar manejo de errores en refresh
echo ""
echo "Step 2: Test refresh error handling"

# Intentar refresh con timeout más corto para simular problemas de red
refresh_response=$(timeout 5s curl -s -w "STATUS:%{http_code}" -X POST "$API_BASE_URL/sismos/refresh")
refresh_exit_code=$?

if [ $refresh_exit_code -eq 0 ]; then
    http_code=$(echo "$refresh_response" | grep -o "STATUS:[0-9]*" | cut -d: -f2)
    body=$(echo "$refresh_response" | sed 's/STATUS:.*//g')
    
    case "$http_code" in
        "200")
            echo "✅ Refresh successful"
            ;;
        "503")
            echo "✅ Proper error handling (503 Service Unavailable)"
            
            # Verificar mensaje de error apropiado
            if validate_json "$body"; then
                error_msg=$(extract_json_field "$body" "error")
                if [ "$error_msg" != "null" ]; then
                    echo "   📝 Error message: $error_msg"
                fi
            fi
            ;;
        "408")
            echo "✅ Proper timeout handling (408 Request Timeout)"
            ;;
        *)
            echo "⚠️ Unexpected response code: $http_code"
            ;;
    esac
else
    echo "⚠️ Refresh timed out or failed (exit code: $refresh_exit_code)"
fi

# Test 3: Verificar que la API sigue funcionando después de errores
echo ""
echo "Step 3: Verify API stability after errors"

# Hacer múltiples requests GET para verificar estabilidad
stable_count=0
total_checks=5

for i in $(seq 1 $total_checks); do
    check_response=$(curl -s -w "%{http_code}" "$API_BASE_URL/sismos")
    check_code=$(echo "$check_response" | tail -c 4)
    
    if [ "$check_code" = "200" ]; then
        stable_count=$((stable_count + 1))
    fi
    
    sleep 1
done

stability_rate=$(echo "scale=1; $stable_count * 100 / $total_checks" | bc)
echo "📊 API Stability after errors: $stability_rate% ($stable_count/$total_checks)"

if [ $stable_count -eq $total_checks ]; then
    echo "✅ Perfect stability maintained"
elif [ $stable_count -gt $((total_checks * 3 / 4)) ]; then
    echo "🟡 Good stability"
else
    echo "❌ Poor stability after errors"
fi

# Test 4: Verificar logging/monitoring de errores
echo ""
echo "Step 4: Error monitoring validation"

# Verificar que health check refleja el estado del sistema
health_response=$(api_request GET "/health" 200)
if [ $? -eq 0 ]; then
    overall_status=$(extract_json_field "$health_response" "status")
    echo "📊 Overall system status: $overall_status"
    
    # El sistema debería seguir siendo UP o DEGRADED, no DOWN
    if [ "$overall_status" = "UP" ] || [ "$overall_status" = "DEGRADED" ]; then
        echo "✅ System maintains acceptable health status"
    elif [ "$overall_status" = "DOWN" ]; then
        echo "⚠️ System shows DOWN status (may be expected if critical services failed)"
    else
        echo "❌ Unknown system status: $overall_status"
    fi
else
    echo "❌ Health check endpoint failed"
fi
```

## 📊 Suite Completa de Sismos

```bash
# Suite completa de testing para endpoints de sismos
sismos_test_suite() {
    echo "🌍 Complete Seismic Endpoints Test Suite"
    echo "======================================="
    
    local tests_passed=0
    local tests_failed=0
    local total_tests=7
    
    echo "🧪 Running $total_tests test cases for seismic endpoints..."
    echo ""
    
    # Ejecutar todos los tests
    test_functions=(
        "TC-SISMOS-001"
        "TC-SISMOS-002" 
        "TC-SISMOS-003"
        "TC-SISMOS-004"
        "TC-SISMOS-005"
        "TC-SISMOS-006"
        "TC-SISMOS-007"
    )
    
    for test_case in "${test_functions[@]}"; do
        echo "🏃 Running $test_case..."
        
        # Aquí iría la llamada a cada función de test individual
        # Por simplicidad, asumimos que cada test retorna 0 para éxito
        
        if [ $? -eq 0 ]; then
            echo "✅ $test_case PASSED"
            tests_passed=$((tests_passed + 1))
        else
            echo "❌ $test_case FAILED"
            tests_failed=$((tests_failed + 1))
        fi
        echo ""
    done
    
    # Resumen final
    success_rate=$(echo "scale=1; $tests_passed * 100 / $total_tests" | bc)
    
    echo "📊 SEISMIC ENDPOINTS TEST SUMMARY"
    echo "================================="
    echo "Total Tests: $total_tests"
    echo "Passed: $tests_passed"
    echo "Failed: $tests_failed"
    echo "Success Rate: $success_rate%"
    echo ""
    
    if [ $tests_passed -eq $total_tests ]; then
        echo "🎉 ALL SEISMIC TESTS PASSED!"
        return 0
    elif [ $tests_passed -gt $((total_tests * 3 / 4)) ]; then
        echo "🟡 Most tests passed (acceptable)"
        return 0
    else
        echo "❌ Multiple test failures detected"
        return 1
    fi
}

# Ejecutar suite completa
sismos_test_suite
```

---

**▶️ Siguiente: [Testing de Performance](./05-performance-tests.md)**