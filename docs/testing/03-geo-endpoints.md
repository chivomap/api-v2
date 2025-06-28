# 🗺️ Endpoints Geoespaciales

Esta sección cubre las pruebas para los endpoints relacionados con datos geoespaciales de El Salvador.

## 🎯 Endpoints a Probar

| Endpoint | Método | Descripción |
|----------|--------|-------------|
| `/geo/search-data` | GET | Obtiene listas de departamentos, municipios y distritos |
| `/geo/filter` | GET | Filtra features geoespaciales por query y tipo |

## 📊 Estructura de Datos Esperada

### Response de `/geo/search-data`
```json
{
  "data": {
    "departamentos": ["SAN SALVADOR", "SANTA ANA", "LA LIBERTAD", ...],
    "municipios": ["San Salvador Centro", "Santa Ana Centro", ...],
    "distritos": ["San Salvador", "Santa Ana", "Antiguo Cuscatlán", ...]
  },
  "timestamp": "2025-06-28T16:00:00-06:00"
}
```

### Response de `/geo/filter`
```json
{
  "data": {
    "type": "FeatureCollection",
    "features": [
      {
        "type": "Feature",
        "geometry": [[...]],
        "properties": {
          "D": "SAN SALVADOR",
          "M": "San Salvador Centro", 
          "NAM": "San Salvador"
        }
      }
    ]
  },
  "timestamp": "2025-06-28T16:00:00-06:00"
}
```

## 🧪 Casos de Prueba

### Test 1: Datos Geográficos Básicos

```bash
# TC-GEO-001: Obtener datos geográficos básicos
echo "🧪 TC-GEO-001: Basic Geographic Data Retrieval"

response=$(api_request GET "/geo/search-data" 200)
if [ $? -ne 0 ]; then
    echo "❌ Failed to retrieve geo data"
    exit 1
fi

echo "✅ Geo data endpoint accessible"

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

# Verificar arrays de datos geográficos
departamentos_str=$(extract_json_field "$response" "data.departamentos")
municipios_str=$(extract_json_field "$response" "data.municipios")
distritos_str=$(extract_json_field "$response" "data.distritos")

# Contar elementos (aproximado)
dep_count=$(echo "$departamentos_str" | grep -o ',' | wc -l)
mun_count=$(echo "$municipios_str" | grep -o ',' | wc -l)
dist_count=$(echo "$distritos_str" | grep -o ',' | wc -l)

echo "📊 Geographic Data Counts:"
echo "   - Departamentos: ~$((dep_count + 1))"
echo "   - Municipios: ~$((mun_count + 1))" 
echo "   - Distritos: ~$((dist_count + 1))"

# Validaciones de contenido
if [ $dep_count -gt 10 ]; then
    echo "✅ Reasonable number of departamentos"
else
    echo "⚠️ Low number of departamentos: $dep_count"
fi

if [ $mun_count -gt 20 ]; then
    echo "✅ Reasonable number of municipios"
else
    echo "⚠️ Low number of municipios: $mun_count"
fi

if [ $dist_count -gt 100 ]; then
    echo "✅ Reasonable number of distritos"
else
    echo "⚠️ Low number of distritos: $dist_count"
fi

# Verificar que contiene datos conocidos de El Salvador
if echo "$departamentos_str" | grep -q "SAN SALVADOR"; then
    echo "✅ Contains expected departamento: SAN SALVADOR"
else
    echo "⚠️ SAN SALVADOR departamento not found"
fi

if echo "$municipios_str" | grep -q "San Salvador"; then
    echo "✅ Contains expected municipio pattern"
else
    echo "⚠️ Expected municipio patterns not found"
fi
```

### Test 2: Filtrado por Departamento

```bash
# TC-GEO-002: Filtrar por departamento específico
echo "🧪 TC-GEO-002: Filter by Department"

test_department="SAN SALVADOR"
encoded_dept=$(echo "$test_department" | sed 's/ /%20/g')

response=$(api_request GET "/geo/filter?query=${encoded_dept}&whatIs=D" 200)
if [ $? -ne 0 ]; then
    echo "❌ Failed to filter by department"
    exit 1
fi

echo "✅ Department filter endpoint accessible"

# Validar estructura de respuesta
if validate_json "$response"; then
    echo "✅ Valid JSON response"
else
    echo "❌ Invalid JSON response"
    exit 1
fi

# Verificar estructura GeoJSON
type_field=$(extract_json_field "$response" "data.type")
features_field=$(extract_json_field "$response" "data.features")

if [ "$type_field" = "FeatureCollection" ]; then
    echo "✅ Correct GeoJSON type: FeatureCollection"
else
    echo "❌ Incorrect type: $type_field (expected FeatureCollection)"
fi

if [ "$features_field" != "null" ]; then
    echo "✅ Features array present"
    
    # Contar features aproximadamente
    feature_count=$(echo "$features_field" | grep -o '"type":"Feature"' | wc -l)
    echo "📊 Found $feature_count features for department: $test_department"
    
    if [ $feature_count -gt 0 ]; then
        echo "✅ Department has geographic features"
        
        # Verificar que todas las features pertenecen al departamento correcto
        if echo "$features_field" | grep -q "\"D\":\"$test_department\""; then
            echo "✅ Features belong to correct department"
        else
            echo "⚠️ Some features may not belong to the requested department"
        fi
        
    else
        echo "⚠️ No features found for department: $test_department"
    fi
    
else
    echo "❌ Features array missing"
fi

# Guardar respuesta para análisis
if [ "$SAVE_RESPONSES" = "true" ]; then
    echo "$response" > "$RESPONSE_DIR/geo-filter-department.json"
fi
```

### Test 3: Filtrado por Municipio

```bash
# TC-GEO-003: Filtrar por municipio específico  
echo "🧪 TC-GEO-003: Filter by Municipality"

test_municipality="San Salvador Centro"
encoded_mun=$(echo "$test_municipality" | sed 's/ /%20/g')

response=$(api_request GET "/geo/filter?query=${encoded_mun}&whatIs=M" 200)
if [ $? -ne 0 ]; then
    echo "❌ Failed to filter by municipality"
    exit 1
fi

echo "✅ Municipality filter endpoint accessible"

# Validar estructura
if validate_json "$response"; then
    echo "✅ Valid JSON response"
    
    feature_count=$(echo "$response" | grep -o '"type":"Feature"' | wc -l)
    echo "📊 Found $feature_count features for municipality: $test_municipality"
    
    if [ $feature_count -gt 0 ]; then
        echo "✅ Municipality has geographic features"
        
        # Verificar que las features tienen el municipio correcto
        if echo "$response" | grep -q "\"M\":\"$test_municipality\""; then
            echo "✅ Features belong to correct municipality"
        else
            echo "⚠️ Features may not match requested municipality"
        fi
    else
        echo "⚠️ No features found for municipality: $test_municipality"
    fi
    
else
    echo "❌ Invalid JSON response"
fi
```

### Test 4: Filtrado por Nombre/Ubicación

```bash
# TC-GEO-004: Filtrar por nombre/ubicación específica
echo "🧪 TC-GEO-004: Filter by Name/Location"

test_location="Santa Ana"
encoded_loc=$(echo "$test_location" | sed 's/ /%20/g')

response=$(api_request GET "/geo/filter?query=${encoded_loc}&whatIs=NAM" 200)
if [ $? -ne 0 ]; then
    echo "❌ Failed to filter by location name"
    exit 1
fi

echo "✅ Location name filter endpoint accessible"

if validate_json "$response"; then
    echo "✅ Valid JSON response"
    
    feature_count=$(echo "$response" | grep -o '"type":"Feature"' | wc -l)
    echo "📊 Found $feature_count features for location: $test_location"
    
    if [ $feature_count -gt 0 ]; then
        echo "✅ Location has geographic features"
        
        # Verificar propiedades específicas
        if echo "$response" | grep -q "\"NAM\":\"$test_location\""; then
            echo "✅ Features match requested location name"
        else
            echo "⚠️ Features may not match exact location name"
        fi
        
        # Extraer información adicional de la primera feature
        first_feature=$(echo "$response" | grep -o '"type":"Feature"[^}]*}[^}]*}' | head -1)
        if [ -n "$first_feature" ]; then
            echo "📋 Sample feature properties found"
        fi
        
    else
        echo "⚠️ No features found for location: $test_location"
    fi
    
else
    echo "❌ Invalid JSON response"
fi
```

### Test 5: Validación de Parámetros

```bash
# TC-GEO-005: Validación de parámetros de entrada
echo "🧪 TC-GEO-005: Input Parameter Validation"

echo "📋 Testing parameter validation..."

# Test 5.1: Parámetro whatIs inválido
echo "5.1: Invalid whatIs parameter"
response=$(curl -s -w "%{http_code}" "$API_BASE_URL/geo/filter?query=test&whatIs=INVALID")
http_code=$(echo "$response" | tail -c 4)
body=$(echo "$response" | head -c -4)

if [ "$http_code" = "400" ]; then
    echo "✅ Correctly rejects invalid whatIs parameter (400)"
    if echo "$body" | grep -q "inválidos"; then
        echo "✅ Appropriate error message returned"
    fi
else
    echo "❌ Should return 400 for invalid whatIs, got: $http_code"
fi

# Test 5.2: Query vacío
echo "5.2: Empty query parameter"
response=$(curl -s -w "%{http_code}" "$API_BASE_URL/geo/filter?query=&whatIs=D")
http_code=$(echo "$response" | tail -c 4)

if [ "$http_code" = "400" ]; then
    echo "✅ Correctly rejects empty query (400)"
else
    echo "❌ Should return 400 for empty query, got: $http_code"
fi

# Test 5.3: Parámetros faltantes
echo "5.3: Missing parameters"
response=$(curl -s -w "%{http_code}" "$API_BASE_URL/geo/filter")
http_code=$(echo "$response" | tail -c 4)

if [ "$http_code" = "400" ]; then
    echo "✅ Correctly rejects missing parameters (400)"
else
    echo "❌ Should return 400 for missing parameters, got: $http_code"
fi

# Test 5.4: Query muy largo
echo "5.4: Very long query parameter"
long_query=$(python3 -c "print('x' * 200)")
encoded_long=$(echo "$long_query" | sed 's/x/%78/g')
response=$(curl -s -w "%{http_code}" "$API_BASE_URL/geo/filter?query=${encoded_long}&whatIs=D")
http_code=$(echo "$response" | tail -c 4)

if [ "$http_code" = "400" ]; then
    echo "✅ Correctly rejects overly long query (400)"
else
    echo "⚠️ Long query handling: $http_code (may be acceptable)"
fi

# Test 5.5: Caracteres especiales peligrosos
echo "5.5: Dangerous characters"
dangerous_chars="<script>alert('xss')</script>"
encoded_dangerous=$(python3 -c "import urllib.parse; print(urllib.parse.quote('$dangerous_chars'))")
response=$(curl -s -w "%{http_code}" "$API_BASE_URL/geo/filter?query=${encoded_dangerous}&whatIs=D")
http_code=$(echo "$response" | tail -c 4)

if [ "$http_code" = "400" ]; then
    echo "✅ Correctly rejects dangerous characters (400)"
elif [ "$http_code" = "200" ]; then
    body=$(echo "$response" | head -c -4)
    if echo "$body" | grep -q "script"; then
        echo "❌ SECURITY ISSUE: Dangerous characters not filtered"
    else
        echo "✅ Dangerous characters filtered but request accepted"
    fi
else
    echo "⚠️ Unexpected response to dangerous characters: $http_code"
fi
```

### Test 6: Performance del Cache

```bash
# TC-GEO-006: Verificar performance del cache
echo "🧪 TC-GEO-006: Cache Performance Testing"

echo "📋 Testing cache performance..."

# Primera request - debería cargar el cache
echo "6.1: First request (cache loading)"
start_time=$(date +%s.%N)
response1=$(api_request GET "/geo/search-data" 200)
end_time=$(date +%s.%N)
first_request_time=$(echo "$end_time - $start_time" | bc)

if [ $? -eq 0 ]; then
    echo "✅ First request successful"
    echo "⏱️ Time: ${first_request_time}s (includes cache loading)"
else
    echo "❌ First request failed"
    exit 1
fi

# Esperar un momento para que el cache se establezca
sleep 1

# Segunda request - debería usar cache
echo "6.2: Second request (from cache)"
start_time=$(date +%s.%N)
response2=$(api_request GET "/geo/search-data" 200)
end_time=$(date +%s.%N)
second_request_time=$(echo "$end_time - $start_time" | bc)

if [ $? -eq 0 ]; then
    echo "✅ Second request successful"
    echo "⏱️ Time: ${second_request_time}s (from cache)"
    
    # Comparar tiempos
    if (( $(echo "$second_request_time < $first_request_time" | bc -l) )); then
        improvement=$(echo "scale=2; ($first_request_time - $second_request_time) / $first_request_time * 100" | bc)
        echo "🚀 Cache improved performance by ${improvement}%"
    else
        echo "⚠️ Second request not faster (cache may not be working)"
    fi
    
    # Verificar que las respuestas son idénticas
    if [ "$response1" = "$response2" ]; then
        echo "✅ Responses are identical (cache consistency)"
    else
        echo "⚠️ Responses differ (potential cache issue)"
    fi
    
else
    echo "❌ Second request failed"
fi

# Test múltiples requests concurrentes desde cache
echo "6.3: Concurrent requests from cache"
concurrent_count=10
echo "Testing $concurrent_count concurrent requests..."

pids=()
times=()

for i in $(seq 1 $concurrent_count); do
    {
        start_time=$(date +%s.%N)
        response=$(curl -s "$API_BASE_URL/geo/search-data")
        end_time=$(date +%s.%N)
        duration=$(echo "$end_time - $start_time" | bc)
        echo "$duration" > "/tmp/geo_test_$i.time"
    } &
    pids+=($!)
done

# Esperar a que terminen todos
for pid in "${pids[@]}"; do
    wait $pid
done

# Calcular estadísticas
total_time=0
count=0
for i in $(seq 1 $concurrent_count); do
    if [ -f "/tmp/geo_test_$i.time" ]; then
        time_val=$(cat "/tmp/geo_test_$i.time")
        total_time=$(echo "$total_time + $time_val" | bc)
        count=$((count + 1))
        rm -f "/tmp/geo_test_$i.time"
    fi
done

if [ $count -gt 0 ]; then
    avg_time=$(echo "scale=3; $total_time / $count" | bc)
    echo "📊 Concurrent Requests Results:"
    echo "   - Successful: $count/$concurrent_count"
    echo "   - Average Time: ${avg_time}s"
    
    if (( $(echo "$avg_time < 0.1" | bc -l) )); then
        echo "🚀 Excellent cache performance (<0.1s avg)"
    elif (( $(echo "$avg_time < 0.5" | bc -l) )); then
        echo "✅ Good cache performance (<0.5s avg)"
    else
        echo "⚠️ Cache performance could be better (>${avg_time}s avg)"
    fi
else
    echo "❌ No concurrent requests completed successfully"
fi
```

### Test 7: Casos Edge de Filtrado

```bash
# TC-GEO-007: Casos edge de filtrado
echo "🧪 TC-GEO-007: Edge Cases for Filtering"

# Test 7.1: Query que no existe
echo "7.1: Non-existent location query"
response=$(api_request GET "/geo/filter?query=NONEXISTENT_LOCATION&whatIs=D" 200)
if [ $? -eq 0 ]; then
    feature_count=$(echo "$response" | grep -o '"type":"Feature"' | wc -l)
    if [ $feature_count -eq 0 ]; then
        echo "✅ Correctly returns empty results for non-existent location"
    else
        echo "⚠️ Unexpected features returned for non-existent location"
    fi
else
    echo "❌ Request failed for non-existent location"
fi

# Test 7.2: Query con acentos y caracteres especiales
echo "7.2: Query with accents and special characters"
accented_query="São Paulo"  # Intentionally not from El Salvador
encoded_accented=$(python3 -c "import urllib.parse; print(urllib.parse.quote('$accented_query'))")
response=$(curl -s -w "%{http_code}" "$API_BASE_URL/geo/filter?query=${encoded_accented}&whatIs=D")
http_code=$(echo "$response" | tail -c 4)

if [ "$http_code" = "200" ]; then
    echo "✅ Handles accented characters properly"
    body=$(echo "$response" | head -c -4)
    feature_count=$(echo "$body" | grep -o '"type":"Feature"' | wc -l)
    echo "   Found $feature_count features"
else
    echo "⚠️ Issue with accented characters: $http_code"
fi

# Test 7.3: Case sensitivity
echo "7.3: Case sensitivity testing"
test_cases=("san salvador" "SAN SALVADOR" "San Salvador")

for test_case in "${test_cases[@]}"; do
    encoded_case=$(echo "$test_case" | sed 's/ /%20/g')
    response=$(api_request GET "/geo/filter?query=${encoded_case}&whatIs=D" 200)
    if [ $? -eq 0 ]; then
        feature_count=$(echo "$response" | grep -o '"type":"Feature"' | wc -l)
        echo "   '$test_case': $feature_count features"
    else
        echo "   '$test_case': request failed"
    fi
done

# Test 7.4: Partial matching
echo "7.4: Partial matching behavior"
partial_queries=("San" "Salvador" "Ana")

for query in "${partial_queries[@]}"; do
    response=$(api_request GET "/geo/filter?query=${query}&whatIs=NAM" 200)
    if [ $? -eq 0 ]; then
        feature_count=$(echo "$response" | grep -o '"type":"Feature"' | wc -l)
        echo "   Partial '$query': $feature_count features"
    else
        echo "   Partial '$query': request failed"
    fi
done
```

### Test 8: Validación de Geometrías

```bash
# TC-GEO-008: Validación de geometrías GeoJSON
echo "🧪 TC-GEO-008: GeoJSON Geometry Validation"

# Obtener algunas features para validar geometrías
response=$(api_request GET "/geo/filter?query=SAN%20SALVADOR&whatIs=D" 200)
if [ $? -ne 0 ]; then
    echo "❌ Failed to get features for geometry validation"
    exit 1
fi

# Extraer primera feature para análisis detallado
feature=$(echo "$response" | python3 -c "
import json, sys
data = json.load(sys.stdin)
features = data.get('data', {}).get('features', [])
if features:
    print(json.dumps(features[0], indent=2))
else:
    print('null')
")

if [ "$feature" != "null" ]; then
    echo "✅ Successfully extracted feature for validation"
    
    # Verificar estructura de feature
    feature_type=$(echo "$feature" | python3 -c "
import json, sys
f = json.load(sys.stdin)
print(f.get('type', 'null'))
")
    
    if [ "$feature_type" = "Feature" ]; then
        echo "✅ Correct feature type"
    else
        echo "❌ Incorrect feature type: $feature_type"
    fi
    
    # Verificar propiedades requeridas
    properties=$(echo "$feature" | python3 -c "
import json, sys
f = json.load(sys.stdin)
props = f.get('properties', {})
print('D' in props, 'M' in props, 'NAM' in props)
")
    
    if echo "$properties" | grep -q "True True True"; then
        echo "✅ All required properties (D, M, NAM) present"
    else
        echo "⚠️ Some properties may be missing: $properties"
    fi
    
    # Verificar que geometry existe
    geometry=$(echo "$feature" | python3 -c "
import json, sys
f = json.load(sys.stdin)
geom = f.get('geometry')
if geom is not None:
    print('present')
else:
    print('null')
")
    
    if [ "$geometry" = "present" ]; then
        echo "✅ Geometry field present"
    else
        echo "⚠️ Geometry field missing or null"
    fi
    
    # Guardar feature de ejemplo
    if [ "$SAVE_RESPONSES" = "true" ]; then
        echo "$feature" > "$RESPONSE_DIR/sample-feature.json"
        echo "💾 Sample feature saved for analysis"
    fi
    
else
    echo "❌ No features found for geometry validation"
fi
```

## 📊 Suite de Performance Geoespacial

```bash
# Suite completa de performance para endpoints geo
geo_performance_suite() {
    echo "🏃 Geospatial Endpoints Performance Suite"
    echo "========================================"
    
    local test_scenarios=(
        "/geo/search-data"
        "/geo/filter?query=SAN%20SALVADOR&whatIs=D"
        "/geo/filter?query=Santa%20Ana&whatIs=NAM"
        "/geo/filter?query=San%20Salvador%20Centro&whatIs=M"
    )
    
    for scenario in "${test_scenarios[@]}"; do
        echo ""
        echo "📊 Testing: $scenario"
        echo "----------------------------------------"
        
        # Warm-up request
        curl -s "$API_BASE_URL$scenario" > /dev/null
        
        # Performance measurement
        local iterations=20
        local total_time=0
        local success_count=0
        
        for i in $(seq 1 $iterations); do
            start_time=$(date +%s.%N)
            response=$(curl -s "$API_BASE_URL$scenario")
            end_time=$(date +%s.%N)
            duration=$(echo "$end_time - $start_time" | bc)
            
            if validate_json "$response"; then
                success_count=$((success_count + 1))
                total_time=$(echo "$total_time + $duration" | bc)
            fi
        done
        
        if [ $success_count -gt 0 ]; then
            avg_time=$(echo "scale=3; $total_time / $success_count" | bc)
            success_rate=$(echo "scale=1; $success_count * 100 / $iterations" | bc)
            
            echo "   ✅ Success Rate: $success_rate%"
            echo "   ⏱️ Average Time: ${avg_time}s"
            
            # Performance rating
            if (( $(echo "$avg_time < 0.05" | bc -l) )); then
                echo "   🚀 Excellent (<0.05s)"
            elif (( $(echo "$avg_time < 0.2" | bc -l) )); then
                echo "   ✅ Good (<0.2s)"
            elif (( $(echo "$avg_time < 1.0" | bc -l) )); then
                echo "   🟡 Acceptable (<1.0s)"
            else
                echo "   🔴 Poor (>1.0s)"
            fi
        else
            echo "   ❌ All requests failed"
        fi
    done
}

# Ejecutar suite de performance
geo_performance_suite
```

---

**▶️ Siguiente: [Endpoints de Sismos](./04-sismos-endpoints.md)**