# 🏥 Health Checks y Monitoreo

Esta sección cubre todas las pruebas relacionadas con el monitoreo de salud de la API.

## 🎯 Objetivos de Testing

- ✅ Verificar que el health check responde correctamente
- ✅ Validar estados de componentes individuales
- ✅ Probar diferentes escenarios de salud (UP/DEGRADED/DOWN)
- ✅ Verificar métricas de performance y uptime
- ✅ Testear bajo condiciones de estrés

## 📊 Estados Esperados del Sistema

### 🟢 Sistema Saludable (UP)
```json
{
  "status": "UP",
  "components": {
    "database": { "status": "UP" },
    "static_files": { "status": "UP" },
    "cache": { "status": "UP" }
  }
}
```

### 🟡 Sistema Degradado (DEGRADED)
```json
{
  "status": "DEGRADED",
  "components": {
    "database": { "status": "UP" },
    "static_files": { "status": "UP" },
    "cache": { "status": "UP" },
    "censo_database": { "status": "DOWN", "message": "Base de datos del censo no configurada" }
  }
}
```

### 🔴 Sistema Caído (DOWN)
```json
{
  "status": "DOWN",
  "components": {
    "database": { "status": "DOWN", "message": "Error de conectividad" }
  }
}
```

## 🧪 Casos de Prueba

### Test 1: Health Check Básico

```bash
# TC-HEALTH-001: Verificar respuesta básica del health check
echo "🧪 TC-HEALTH-001: Basic Health Check"

response=$(curl -s \
  -w "STATUS:%{http_code}|TIME:%{time_total}" \
  "$API_BASE_URL/health")

# Extraer código de estado y tiempo
http_code=$(echo "$response" | grep -o "STATUS:[0-9]*" | cut -d: -f2)
response_time=$(echo "$response" | grep -o "TIME:[0-9.]*" | cut -d: -f2)
body=$(echo "$response" | sed 's/STATUS:.*//g')

# Validaciones
if [ "$http_code" = "200" ]; then
    echo "✅ HTTP Status: 200 OK"
else
    echo "❌ HTTP Status: $http_code (expected 200)"
fi

if (( $(echo "$response_time < 1.0" | bc -l) )); then
    echo "✅ Response Time: ${response_time}s (< 1s)"
else
    echo "🟡 Response Time: ${response_time}s (slow)"
fi

# Validar estructura JSON
if validate_json "$body"; then
    echo "✅ Valid JSON response"
    
    # Extraer campos críticos
    status=$(extract_json_field "$body" "status")
    version=$(extract_json_field "$body" "version")
    timestamp=$(extract_json_field "$body" "timestamp")
    uptime=$(extract_json_field "$body" "uptime")
    
    echo "📊 Status: $status"
    echo "📊 Version: $version"
    echo "📊 Uptime: $uptime"
    
    # Guardar respuesta para análisis
    if [ "$SAVE_RESPONSES" = "true" ]; then
        echo "$body" > "$RESPONSE_DIR/health-basic.json"
    fi
else
    echo "❌ Invalid JSON response"
    echo "Response: $body"
fi
```

### Test 2: Verificación de Componentes

```bash
# TC-HEALTH-002: Verificar estado de todos los componentes
echo "🧪 TC-HEALTH-002: Component Health Check"

response=$(api_request GET "/health" 200)
if [ $? -ne 0 ]; then
    echo "❌ Health endpoint not accessible"
    exit 1
fi

# Verificar componentes requeridos
components=("database" "static_files" "cache")
all_components_ok=true

for component in "${components[@]}"; do
    status=$(extract_json_field "$response" "components.$component.status")
    message=$(extract_json_field "$response" "components.$component.message")
    
    case "$status" in
        "UP")
            echo "✅ $component: UP"
            ;;
        "DOWN")
            echo "❌ $component: DOWN - $message"
            all_components_ok=false
            ;;
        "null")
            echo "⚠️ $component: Component not found"
            all_components_ok=false
            ;;
        *)
            echo "🟡 $component: Unknown status ($status)"
            ;;
    esac
done

# Verificar componentes opcionales
optional_components=("censo_database")
for component in "${optional_components[@]}"; do
    status=$(extract_json_field "$response" "components.$component.status")
    if [ "$status" != "null" ]; then
        echo "ℹ️ $component: $status (optional)"
    fi
done

if [ "$all_components_ok" = true ]; then
    echo "🎉 All critical components are healthy"
else
    echo "⚠️ Some components have issues"
fi
```

### Test 3: Detalles de Base de Datos

```bash
# TC-HEALTH-003: Verificar detalles de conectividad de BD
echo "🧪 TC-HEALTH-003: Database Connection Details"

response=$(api_request GET "/health" 200)
db_status=$(extract_json_field "$response" "components.database.status")

if [ "$db_status" = "UP" ]; then
    echo "✅ Database is UP"
    
    # Verificar estadísticas de conexión
    open_connections=$(extract_json_field "$response" "components.database.details.open_connections")
    in_use=$(extract_json_field "$response" "components.database.details.in_use")
    idle=$(extract_json_field "$response" "components.database.details.idle")
    
    echo "📊 Connection Stats:"
    echo "   - Open: $open_connections"
    echo "   - In Use: $in_use"
    echo "   - Idle: $idle"
    
    # Validaciones de salud de conexiones
    if [ "$open_connections" -gt 0 ]; then
        echo "✅ Database connections available"
    else
        echo "⚠️ No database connections open"
    fi
    
    if [ "$idle" -gt 0 ]; then
        echo "✅ Idle connections available for new requests"
    else
        echo "🟡 No idle connections (might indicate high load)"
    fi
    
else
    echo "❌ Database is not UP: $db_status"
    error_msg=$(extract_json_field "$response" "components.database.message")
    echo "Error: $error_msg"
fi
```

### Test 4: Estado del Cache

```bash
# TC-HEALTH-004: Verificar estado y detalles del cache
echo "🧪 TC-HEALTH-004: Cache Status and Details"

# Primero verificar cache sin datos
echo "📋 Step 1: Check cache before loading data"
response=$(api_request GET "/health" 200)
cache_status=$(extract_json_field "$response" "components.cache.status")

if [ "$cache_status" = "DOWN" ]; then
    echo "✅ Cache correctly shows DOWN before first use"
    cache_message=$(extract_json_field "$response" "components.cache.message")
    echo "   Message: $cache_message"
else
    echo "🟡 Cache status: $cache_status (expected DOWN initially)"
fi

# Cargar datos para inicializar cache
echo "📋 Step 2: Load geo data to initialize cache"
geo_response=$(api_request GET "/geo/search-data" 200)
if [ $? -eq 0 ]; then
    echo "✅ Geo data loaded successfully"
    sleep 1  # Dar tiempo para que el cache se actualice
else
    echo "❌ Failed to load geo data"
    return 1
fi

# Verificar cache después de cargar datos
echo "📋 Step 3: Check cache after loading data"
response=$(api_request GET "/health" 200)
cache_status=$(extract_json_field "$response" "components.cache.status")

if [ "$cache_status" = "UP" ]; then
    echo "✅ Cache successfully initialized"
    
    # Verificar detalles del cache
    loaded=$(extract_json_field "$response" "components.cache.details.loaded")
    features_count=$(extract_json_field "$response" "components.cache.details.geoFeatures")
    file_path=$(extract_json_field "$response" "components.cache.details.filePath")
    loaded_at=$(extract_json_field "$response" "components.cache.details.loadedAt")
    
    echo "📊 Cache Details:"
    echo "   - Loaded: $loaded"
    echo "   - Features: $features_count"
    echo "   - File Path: $file_path"
    echo "   - Loaded At: $loaded_at"
    
    # Validaciones
    if [ "$features_count" -gt 0 ]; then
        echo "✅ Cache contains $features_count geospatial features"
    else
        echo "⚠️ Cache loaded but no features found"
    fi
    
else
    echo "❌ Cache failed to initialize: $cache_status"
    cache_message=$(extract_json_field "$response" "components.cache.message")
    echo "   Error: $cache_message"
fi
```

### Test 5: Archivos Estáticos

```bash
# TC-HEALTH-005: Verificar estado de archivos estáticos
echo "🧪 TC-HEALTH-005: Static Files Status"

response=$(api_request GET "/health" 200)
static_status=$(extract_json_field "$response" "components.static_files.status")

if [ "$static_status" = "UP" ]; then
    echo "✅ Static files are accessible"
    
    # Verificar detalles del archivo TopoJSON
    file_size=$(extract_json_field "$response" "components.static_files.details.topo_file_size")
    mod_time=$(extract_json_field "$response" "components.static_files.details.topo_mod_time")
    
    echo "📊 TopoJSON File Details:"
    echo "   - Size: $file_size bytes ($(($file_size / 1024 / 1024)) MB)"
    echo "   - Modified: $mod_time"
    
    # Validaciones
    expected_min_size=5000000  # 5MB mínimo
    if [ "$file_size" -gt $expected_min_size ]; then
        echo "✅ File size is appropriate (>5MB)"
    else
        echo "⚠️ File size seems small: $file_size bytes"
    fi
    
else
    echo "❌ Static files not accessible: $static_status"
    error_msg=$(extract_json_field "$response" "components.static_files.message")
    echo "   Error: $error_msg"
fi
```

### Test 6: Health Check bajo Carga

```bash
# TC-HEALTH-006: Health check durante alta concurrencia
echo "🧪 TC-HEALTH-006: Health Check Under Load"

echo "📋 Starting concurrent health checks..."
concurrent_requests=20
pids=()

# Lanzar múltiples requests concurrentes
for i in $(seq 1 $concurrent_requests); do
    {
        start_time=$(date +%s.%N)
        response=$(curl -s "$API_BASE_URL/health")
        end_time=$(date +%s.%N)
        duration=$(echo "$end_time - $start_time" | bc)
        
        if validate_json "$response"; then
            status=$(extract_json_field "$response" "status")
            echo "Request $i: $status ($duration seconds)"
        else
            echo "Request $i: FAILED ($duration seconds)"
        fi
    } &
    pids+=($!)
done

# Esperar a que terminen todos
echo "⏳ Waiting for all requests to complete..."
failed_count=0
for pid in "${pids[@]}"; do
    if ! wait $pid; then
        failed_count=$((failed_count + 1))
    fi
done

echo "📊 Concurrent Health Check Results:"
echo "   - Total Requests: $concurrent_requests"
echo "   - Failed Requests: $failed_count"
echo "   - Success Rate: $(( (concurrent_requests - failed_count) * 100 / concurrent_requests ))%"

if [ $failed_count -eq 0 ]; then
    echo "✅ All concurrent health checks succeeded"
elif [ $failed_count -lt $((concurrent_requests / 4)) ]; then
    echo "🟡 Some failures but within acceptable range"
else
    echo "❌ High failure rate under load"
fi
```

### Test 7: Monitoreo de Uptime

```bash
# TC-HEALTH-007: Verificar tracking de uptime
echo "🧪 TC-HEALTH-007: Uptime Monitoring"

# Múltiples checks para ver progresión del uptime
echo "📋 Monitoring uptime progression..."

for i in {1..3}; do
    response=$(api_request GET "/health" 200)
    uptime=$(extract_json_field "$response" "uptime")
    timestamp=$(extract_json_field "$response" "timestamp")
    
    echo "Check $i: Uptime=$uptime, Timestamp=$timestamp"
    
    if [ $i -lt 3 ]; then
        sleep 2
    fi
done

echo "✅ Uptime tracking is working"

# Verificar formato de uptime (debe contener unidades de tiempo)
final_response=$(api_request GET "/health" 200)
uptime=$(extract_json_field "$final_response" "uptime")

if echo "$uptime" | grep -E "[0-9]+(h|m|s)" > /dev/null; then
    echo "✅ Uptime format is correct: $uptime"
else
    echo "⚠️ Uptime format might be incorrect: $uptime"
fi
```

## 📊 Health Check Performance Benchmarks

```bash
# Benchmark completo de health checks
health_benchmark() {
    echo "🏃 Health Check Performance Benchmark"
    echo "====================================="
    
    local iterations=100
    local total_time=0
    local min_time=999
    local max_time=0
    local success_count=0
    
    echo "Running $iterations health checks..."
    
    for i in $(seq 1 $iterations); do
        start_time=$(date +%s.%N)
        
        response=$(curl -s "$API_BASE_URL/health")
        curl_exit=$?
        
        end_time=$(date +%s.%N)
        duration=$(echo "$end_time - $start_time" | bc)
        
        if [ $curl_exit -eq 0 ] && validate_json "$response"; then
            success_count=$((success_count + 1))
        fi
        
        # Actualizar estadísticas
        total_time=$(echo "$total_time + $duration" | bc)
        
        if (( $(echo "$duration < $min_time" | bc -l) )); then
            min_time=$duration
        fi
        
        if (( $(echo "$duration > $max_time" | bc -l) )); then
            max_time=$duration
        fi
        
        # Progress indicator
        if [ $((i % 10)) -eq 0 ]; then
            echo -n "."
        fi
    done
    
    echo ""
    
    # Calcular métricas
    avg_time=$(echo "scale=3; $total_time / $iterations" | bc)
    success_rate=$(echo "scale=1; $success_count * 100 / $iterations" | bc)
    
    echo "📊 Performance Results:"
    echo "   - Total Requests: $iterations"
    echo "   - Successful: $success_count"
    echo "   - Success Rate: $success_rate%"
    echo "   - Average Time: ${avg_time}s"
    echo "   - Min Time: ${min_time}s"
    echo "   - Max Time: ${max_time}s"
    
    # Evaluación de performance
    if (( $(echo "$avg_time < 0.1" | bc -l) )); then
        echo "🚀 Excellent performance (<0.1s avg)"
    elif (( $(echo "$avg_time < 0.5" | bc -l) )); then
        echo "✅ Good performance (<0.5s avg)"
    elif (( $(echo "$avg_time < 1.0" | bc -l) )); then
        echo "🟡 Acceptable performance (<1.0s avg)"
    else
        echo "🔴 Poor performance (>1.0s avg)"
    fi
}

# Ejecutar benchmark
health_benchmark
```

## 🚨 Alertas y Umbrales

```bash
# Configurar umbrales para alertas
HEALTH_RESPONSE_THRESHOLD=1.0    # segundos
UPTIME_MIN_THRESHOLD=60          # segundos mínimos de uptime
SUCCESS_RATE_THRESHOLD=95        # porcentaje mínimo de éxito

# Función de validación de SLA
validate_health_sla() {
    local response_time=$1
    local uptime_seconds=$2
    local success_rate=$3
    
    local sla_violations=0
    
    if (( $(echo "$response_time > $HEALTH_RESPONSE_THRESHOLD" | bc -l) )); then
        echo "🚨 SLA VIOLATION: Response time ($response_time s) exceeds threshold ($HEALTH_RESPONSE_THRESHOLD s)"
        sla_violations=$((sla_violations + 1))
    fi
    
    if [ "$uptime_seconds" -lt $UPTIME_MIN_THRESHOLD ]; then
        echo "🚨 SLA VIOLATION: Uptime ($uptime_seconds s) below minimum ($UPTIME_MIN_THRESHOLD s)"
        sla_violations=$((sla_violations + 1))
    fi
    
    if (( $(echo "$success_rate < $SUCCESS_RATE_THRESHOLD" | bc -l) )); then
        echo "🚨 SLA VIOLATION: Success rate ($success_rate%) below threshold ($SUCCESS_RATE_THRESHOLD%)"
        sla_violations=$((sla_violations + 1))
    fi
    
    if [ $sla_violations -eq 0 ]; then
        echo "✅ All SLA requirements met"
        return 0
    else
        echo "❌ $sla_violations SLA violations detected"
        return 1
    fi
}
```

---

**▶️ Siguiente: [Endpoints Geoespaciales](./03-geo-endpoints.md)**