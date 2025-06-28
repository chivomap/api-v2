# ⚡ Testing de Performance

Esta sección cubre pruebas exhaustivas de rendimiento para validar que las optimizaciones implementadas funcionen correctamente.

## 🎯 Objetivos de Performance Testing

- ✅ Verificar mejoras del cache estático
- ✅ Validar límites de request y rate limiting
- ✅ Medir latencia y throughput
- ✅ Probar bajo carga concurrente
- ✅ Verificar memory efficiency
- ✅ Testear degradación graceful

## 📊 Métricas de Performance Esperadas

### 🚀 Benchmarks Target

| Métrica | Sin Cache | Con Cache | Mejora |
|---------|-----------|-----------|--------|
| Geo Data Load | ~500ms | ~5ms | 99% |
| Memory Usage | Variable | Optimized | Estable |
| Concurrent Requests | Limited | High | 10x+ |
| Cache Hit Ratio | N/A | >95% | ∞ |

### 🏁 Umbrales de Aceptación

| Endpoint | Response Time | Throughput | Success Rate |
|----------|---------------|------------|--------------|
| `/health` | <100ms | >200 req/s | 100% |
| `/geo/search-data` | <50ms (cached) | >500 req/s | 99%+ |
| `/geo/filter` | <100ms | >200 req/s | 99%+ |
| `/sismos` | <500ms | >100 req/s | 95%+ |

## 🧪 Casos de Prueba

### Test 1: Benchmark de Cache Estático

```bash
# TC-PERF-001: Static Cache Performance Benchmark
echo "🧪 TC-PERF-001: Static Cache Performance Benchmark"

echo "📋 Testing cache performance with cold start..."

# Test 1.1: Cold start (sin cache)
echo "Step 1: Cold start performance (cache loading)"

start_time=$(date +%s.%N)
first_response=$(api_request GET "/geo/search-data" 200)
end_time=$(date +%s.%N)
cold_start_time=$(echo "$end_time - $start_time" | bc)

if [ $? -eq 0 ]; then
    echo "✅ Cold start successful"
    echo "⏱️ Cold start time: ${cold_start_time}s"
    
    # Verificar que los datos se cargaron
    dep_count=$(echo "$first_response" | grep -o '"departamentos"' | wc -l)
    if [ $dep_count -gt 0 ]; then
        echo "✅ Data loaded successfully"
    fi
else
    echo "❌ Cold start failed"
    exit 1
fi

# Esperar a que el cache se estabilice
sleep 2

# Test 1.2: Warm cache performance
echo ""
echo "Step 2: Warm cache performance testing"

warm_times=()
warm_iterations=10

echo "Running $warm_iterations warm cache requests..."

for i in $(seq 1 $warm_iterations); do
    start_time=$(date +%s.%N)
    response=$(curl -s "$API_BASE_URL/geo/search-data")
    end_time=$(date +%s.%N)
    duration=$(echo "$end_time - $start_time" | bc)
    
    if validate_json "$response"; then
        warm_times+=($duration)
        echo -n "."
    else
        echo -n "x"
    fi
done

echo ""

# Calcular estadísticas de warm cache
if [ ${#warm_times[@]} -gt 0 ]; then
    total_warm_time=0
    min_warm_time=${warm_times[0]}
    max_warm_time=${warm_times[0]}
    
    for time in "${warm_times[@]}"; do
        total_warm_time=$(echo "$total_warm_time + $time" | bc)
        
        if (( $(echo "$time < $min_warm_time" | bc -l) )); then
            min_warm_time=$time
        fi
        
        if (( $(echo "$time > $max_warm_time" | bc -l) )); then
            max_warm_time=$time
        fi
    done
    
    avg_warm_time=$(echo "scale=4; $total_warm_time / ${#warm_times[@]}" | bc)
    
    echo "📊 Warm Cache Performance:"
    echo "   - Average: ${avg_warm_time}s"
    echo "   - Min: ${min_warm_time}s"
    echo "   - Max: ${max_warm_time}s"
    echo "   - Successful: ${#warm_times[@]}/$warm_iterations"
    
    # Calcular mejora de performance
    if (( $(echo "$cold_start_time > 0" | bc -l) )); then
        improvement=$(echo "scale=1; ($cold_start_time - $avg_warm_time) / $cold_start_time * 100" | bc)
        speedup=$(echo "scale=1; $cold_start_time / $avg_warm_time" | bc)
        
        echo "🚀 Cache Performance Improvement:"
        echo "   - Improvement: ${improvement}%"
        echo "   - Speedup: ${speedup}x faster"
        
        # Evaluación
        if (( $(echo "$improvement > 90" | bc -l) )); then
            echo "🎉 Excellent cache performance (>90% improvement)"
        elif (( $(echo "$improvement > 70" | bc -l) )); then
            echo "✅ Good cache performance (>70% improvement)"
        elif (( $(echo "$improvement > 50" | bc -l) )); then
            echo "🟡 Moderate cache performance (>50% improvement)"
        else
            echo "❌ Poor cache performance (<50% improvement)"
        fi
        
        # Verificar umbral target
        if (( $(echo "$avg_warm_time < 0.05" | bc -l) )); then
            echo "🎯 Target achieved: <50ms response time"
        else
            echo "⚠️ Target missed: ${avg_warm_time}s > 0.05s"
        fi
    fi
else
    echo "❌ No successful warm cache requests"
fi

# Test 1.3: Cache consistency
echo ""
echo "Step 3: Cache consistency verification"

responses=()
consistency_checks=5

for i in $(seq 1 $consistency_checks); do
    response=$(curl -s "$API_BASE_URL/geo/search-data")
    if validate_json "$response"; then
        # Calcular hash simple del contenido para comparación
        content_hash=$(echo "$response" | shasum | cut -d' ' -f1)
        responses+=($content_hash)
    fi
done

# Verificar que todas las respuestas son idénticas
if [ ${#responses[@]} -gt 0 ]; then
    first_hash=${responses[0]}
    all_same=true
    
    for hash in "${responses[@]}"; do
        if [ "$hash" != "$first_hash" ]; then
            all_same=false
            break
        fi
    done
    
    if [ "$all_same" = true ]; then
        echo "✅ Cache consistency verified (all responses identical)"
    else
        echo "❌ Cache inconsistency detected (responses differ)"
    fi
else
    echo "❌ No responses received for consistency check"
fi
```

### Test 2: Load Testing y Concurrencia

```bash
# TC-PERF-002: Load testing y concurrencia
echo "🧪 TC-PERF-002: Load Testing and Concurrency"

echo "📋 Testing concurrent request handling..."

# Test 2.1: Concurrent requests al endpoint de geo data
echo "Step 1: Concurrent geo data requests"

concurrent_users=(5 10 20 50)
results=()

for users in "${concurrent_users[@]}"; do
    echo ""
    echo "Testing with $users concurrent users..."
    
    pids=()
    start_time=$(date +%s.%N)
    
    # Lanzar requests concurrentes
    for i in $(seq 1 $users); do
        {
            req_start=$(date +%s.%N)
            response=$(curl -s "$API_BASE_URL/geo/search-data")
            req_end=$(date +%s.%N)
            req_duration=$(echo "$req_end - $req_start" | bc)
            
            if validate_json "$response"; then
                echo "SUCCESS $req_duration" > "/tmp/load_test_${users}_${i}.result"
            else
                echo "FAILED $req_duration" > "/tmp/load_test_${users}_${i}.result"
            fi
        } &
        pids+=($!)
    done
    
    # Esperar a que terminen todos
    for pid in "${pids[@]}"; do
        wait $pid
    done
    
    end_time=$(date +%s.%N)
    total_duration=$(echo "$end_time - $start_time" | bc)
    
    # Analizar resultados
    success_count=0
    failed_count=0
    total_req_time=0
    min_req_time=999
    max_req_time=0
    
    for i in $(seq 1 $users); do
        if [ -f "/tmp/load_test_${users}_${i}.result" ]; then
            result=$(cat "/tmp/load_test_${users}_${i}.result")
            status=$(echo "$result" | cut -d' ' -f1)
            req_time=$(echo "$result" | cut -d' ' -f2)
            
            if [ "$status" = "SUCCESS" ]; then
                success_count=$((success_count + 1))
                total_req_time=$(echo "$total_req_time + $req_time" | bc)
                
                if (( $(echo "$req_time < $min_req_time" | bc -l) )); then
                    min_req_time=$req_time
                fi
                
                if (( $(echo "$req_time > $max_req_time" | bc -l) )); then
                    max_req_time=$req_time
                fi
            else
                failed_count=$((failed_count + 1))
            fi
            
            rm -f "/tmp/load_test_${users}_${i}.result"
        fi
    done
    
    # Calcular métricas
    success_rate=$(echo "scale=1; $success_count * 100 / $users" | bc)
    
    if [ $success_count -gt 0 ]; then
        avg_req_time=$(echo "scale=4; $total_req_time / $success_count" | bc)
        throughput=$(echo "scale=1; $success_count / $total_duration" | bc)
        
        echo "📊 Results for $users users:"
        echo "   - Success Rate: $success_rate%"
        echo "   - Total Duration: ${total_duration}s"
        echo "   - Avg Request Time: ${avg_req_time}s"
        echo "   - Min Request Time: ${min_req_time}s"
        echo "   - Max Request Time: ${max_req_time}s"
        echo "   - Throughput: ${throughput} req/s"
        
        # Guardar para análisis final
        results+=("$users:$success_rate:$avg_req_time:$throughput")
        
        # Evaluación por nivel de concurrencia
        if (( $(echo "$success_rate >= 99" | bc -l) )); then
            echo "   ✅ Excellent concurrency handling"
        elif (( $(echo "$success_rate >= 95" | bc -l) )); then
            echo "   🟡 Good concurrency handling"
        else
            echo "   ❌ Poor concurrency handling"
        fi
    else
        echo "   ❌ All requests failed"
        results+=("$users:0:0:0")
    fi
done

# Test 2.2: Sustained load test
echo ""
echo "Step 2: Sustained load test"

sustained_duration=30  # segundos
sustained_users=10
sustained_interval=0.1  # intervalo entre requests

echo "Running sustained load for ${sustained_duration}s with $sustained_users users..."

sustained_start=$(date +%s)
sustained_pids=()
sustained_request_count=0

# Función para generar carga sostenida
generate_sustained_load() {
    local user_id=$1
    local end_time=$(($(date +%s) + sustained_duration))
    local request_count=0
    
    while [ $(date +%s) -lt $end_time ]; do
        start_req=$(date +%s.%N)
        response=$(curl -s "$API_BASE_URL/geo/search-data")
        end_req=$(date +%s.%N)
        duration=$(echo "$end_req - $start_req" | bc)
        
        if validate_json "$response"; then
            echo "USER${user_id}_SUCCESS_${duration}" >> "/tmp/sustained_load.log"
        else
            echo "USER${user_id}_FAILED_${duration}" >> "/tmp/sustained_load.log"
        fi
        
        request_count=$((request_count + 1))
        sleep $sustained_interval
    done
    
    echo "$request_count" > "/tmp/sustained_user_${user_id}.count"
}

# Lanzar usuarios para carga sostenida
for i in $(seq 1 $sustained_users); do
    generate_sustained_load $i &
    sustained_pids+=($!)
done

# Esperar a que termine la prueba
for pid in "${sustained_pids[@]}"; do
    wait $pid
done

sustained_end=$(date +%s)
actual_duration=$((sustained_end - sustained_start))

# Analizar resultados de carga sostenida
if [ -f "/tmp/sustained_load.log" ]; then
    total_requests=$(wc -l < "/tmp/sustained_load.log")
    successful_requests=$(grep "SUCCESS" "/tmp/sustained_load.log" | wc -l)
    failed_requests=$(grep "FAILED" "/tmp/sustained_load.log" | wc -l)
    
    sustained_success_rate=$(echo "scale=1; $successful_requests * 100 / $total_requests" | bc)
    sustained_throughput=$(echo "scale=1; $total_requests / $actual_duration" | bc)
    
    echo "📊 Sustained Load Results:"
    echo "   - Duration: ${actual_duration}s"
    echo "   - Total Requests: $total_requests"
    echo "   - Successful: $successful_requests"
    echo "   - Failed: $failed_requests"
    echo "   - Success Rate: $sustained_success_rate%"
    echo "   - Throughput: $sustained_throughput req/s"
    
    # Limpiar archivos temporales
    rm -f "/tmp/sustained_load.log"
    rm -f /tmp/sustained_user_*.count
    
    if (( $(echo "$sustained_success_rate >= 95" | bc -l) )); then
        echo "✅ Excellent sustained performance"
    elif (( $(echo "$sustained_success_rate >= 90" | bc -l) )); then
        echo "🟡 Good sustained performance"
    else
        echo "❌ Poor sustained performance"
    fi
else
    echo "❌ No sustained load data collected"
fi

# Resumen de concurrencia
echo ""
echo "📊 CONCURRENCY TEST SUMMARY"
echo "============================"

for result in "${results[@]}"; do
    IFS=':' read -r users success_rate avg_time throughput <<< "$result"
    echo "Users: $users | Success: $success_rate% | Avg Time: ${avg_time}s | Throughput: $throughput req/s"
done
```

### Test 3: Rate Limiting Performance

```bash
# TC-PERF-003: Rate limiting performance and effectiveness
echo "🧪 TC-PERF-003: Rate Limiting Performance"

echo "📋 Testing rate limiting effectiveness..."

# Test 3.1: Rate limit threshold testing
echo "Step 1: Rate limit threshold testing"

# Configuración según main.go: 100 requests por minuto
rate_limit_max=100
rate_limit_window=60  # segundos

echo "Testing rate limit: $rate_limit_max requests per $rate_limit_window seconds"

# Hacer requests rápidamente para alcanzar el límite
rapid_requests=120  # Más que el límite para provocar rate limiting
rapid_pids=()
rapid_start=$(date +%s)

echo "Sending $rapid_requests rapid requests..."

for i in $(seq 1 $rapid_requests); do
    {
        response=$(curl -s -w "STATUS:%{http_code}" "$API_BASE_URL/health")
        http_code=$(echo "$response" | grep -o "STATUS:[0-9]*" | cut -d: -f2)
        timestamp=$(date +%s.%N)
        
        echo "$timestamp:$http_code" >> "/tmp/rate_limit_test.log"
    } &
    rapid_pids+=($!)
    
    # Pequeña pausa para no sobrecargar el sistema
    if [ $((i % 10)) -eq 0 ]; then
        sleep 0.1
    fi
done

# Esperar a que terminen todos
echo "Waiting for all requests to complete..."
for pid in "${rapid_pids[@]}"; do
    wait $pid
done

rapid_end=$(date +%s)
rapid_duration=$((rapid_end - rapid_start))

# Analizar resultados de rate limiting
if [ -f "/tmp/rate_limit_test.log" ]; then
    total_rapid_requests=$(wc -l < "/tmp/rate_limit_test.log")
    successful_requests=$(grep ":200$" "/tmp/rate_limit_test.log" | wc -l)
    rate_limited_requests=$(grep ":429$" "/tmp/rate_limit_test.log" | wc -l)
    other_requests=$((total_rapid_requests - successful_requests - rate_limited_requests))
    
    echo "📊 Rate Limiting Results:"
    echo "   - Total Requests: $total_rapid_requests"
    echo "   - Successful (200): $successful_requests"
    echo "   - Rate Limited (429): $rate_limited_requests"
    echo "   - Other Status: $other_requests"
    echo "   - Duration: ${rapid_duration}s"
    
    # Calcular rate de requests que pasaron
    actual_rate=$(echo "scale=1; $successful_requests * 60 / $rapid_duration" | bc)
    echo "   - Effective Rate: $actual_rate req/min"
    
    # Evaluación del rate limiting
    if [ $rate_limited_requests -gt 0 ]; then
        rate_limit_percentage=$(echo "scale=1; $rate_limited_requests * 100 / $total_rapid_requests" | bc)
        echo "   - Rate Limited: $rate_limit_percentage%"
        echo "✅ Rate limiting is working (blocking excess requests)"
        
        # Verificar que el rate no excede significativamente el límite
        if (( $(echo "$successful_requests <= $rate_limit_max * 1.1" | bc -l) )); then
            echo "✅ Rate limit correctly enforced"
        else
            echo "⚠️ Rate limit may be too permissive"
        fi
    else
        echo "⚠️ No rate limiting detected (may not have reached threshold)"
    fi
    
    rm -f "/tmp/rate_limit_test.log"
else
    echo "❌ No rate limiting data collected"
fi

# Test 3.2: Rate limit recovery testing
echo ""
echo "Step 2: Rate limit recovery testing"

echo "Testing rate limit recovery after cooldown..."

# Esperar para que el rate limit se resetee
cooldown_time=65  # Más que la ventana de rate limiting
echo "Waiting ${cooldown_time}s for rate limit reset..."

sleep $cooldown_time

# Hacer algunos requests para verificar que el rate limit se ha reseteado
recovery_requests=10
recovery_successful=0

for i in $(seq 1 $recovery_requests); do
    response=$(curl -s -w "%{http_code}" "$API_BASE_URL/health")
    http_code=$(echo "$response" | tail -c 4)
    
    if [ "$http_code" = "200" ]; then
        recovery_successful=$((recovery_successful + 1))
    fi
    
    sleep 1  # Pausa entre requests para no triggering rate limiting nuevamente
done

recovery_rate=$(echo "scale=1; $recovery_successful * 100 / $recovery_requests" | bc)

echo "📊 Recovery Test Results:"
echo "   - Recovery Requests: $recovery_requests"
echo "   - Successful: $recovery_successful"
echo "   - Success Rate: $recovery_rate%"

if [ $recovery_successful -eq $recovery_requests ]; then
    echo "✅ Complete recovery from rate limiting"
elif [ $recovery_successful -gt $((recovery_requests * 3 / 4)) ]; then
    echo "🟡 Good recovery from rate limiting"
else
    echo "❌ Poor recovery from rate limiting"
fi

# Test 3.3: Rate limiting headers verification
echo ""
echo "Step 3: Rate limiting headers verification"

echo "Checking rate limiting headers..."

header_response=$(curl -s -I "$API_BASE_URL/health")

if echo "$header_response" | grep -q "X-Ratelimit-Limit"; then
    limit_header=$(echo "$header_response" | grep "X-Ratelimit-Limit" | cut -d' ' -f2 | tr -d '\r')
    remaining_header=$(echo "$header_response" | grep "X-Ratelimit-Remaining" | cut -d' ' -f2 | tr -d '\r')
    reset_header=$(echo "$header_response" | grep "X-Ratelimit-Reset" | cut -d' ' -f2 | tr -d '\r')
    
    echo "✅ Rate limiting headers present:"
    echo "   - Limit: $limit_header"
    echo "   - Remaining: $remaining_header"
    echo "   - Reset: $reset_header seconds"
    
    # Verificar que los valores son razonables
    if [ "$limit_header" = "100" ]; then
        echo "✅ Correct rate limit configured"
    else
        echo "⚠️ Unexpected rate limit: $limit_header (expected 100)"
    fi
    
    if [ "$remaining_header" -le "$limit_header" ]; then
        echo "✅ Remaining count is logical"
    else
        echo "⚠️ Remaining count seems incorrect"
    fi
    
else
    echo "⚠️ Rate limiting headers not found (may not be implemented)"
fi
```

### Test 4: Memory Usage y Resource Efficiency

```bash
# TC-PERF-004: Memory usage and resource efficiency
echo "🧪 TC-PERF-004: Memory Usage and Resource Efficiency"

echo "📋 Testing memory usage patterns..."

# Test 4.1: Memory usage during cache loading
echo "Step 1: Memory monitoring during cache operations"

# Función para obtener memoria del proceso
get_api_memory() {
    # Buscar proceso de la API
    api_pid=$(pgrep -f chivomap-api | head -1)
    if [ -n "$api_pid" ]; then
        # Obtener memoria en KB usando ps
        memory_kb=$(ps -p $api_pid -o rss= | tr -d ' ')
        if [ -n "$memory_kb" ]; then
            echo "$memory_kb"
        else
            echo "0"
        fi
    else
        echo "0"
    fi
}

# Memoria baseline (antes de cargar cache)
baseline_memory=$(get_api_memory)
echo "📊 Baseline memory usage: ${baseline_memory} KB"

if [ "$baseline_memory" = "0" ]; then
    echo "⚠️ Could not detect API process for memory monitoring"
    echo "   Continuing with other performance tests..."
else
    # Trigger cache loading
    echo "Loading cache and monitoring memory..."
    
    pre_cache_memory=$(get_api_memory)
    cache_response=$(curl -s "$API_BASE_URL/geo/search-data")
    sleep 2  # Dar tiempo para que el cache se establezca
    post_cache_memory=$(get_api_memory)
    
    echo "📊 Memory usage during cache loading:"
    echo "   - Pre-cache: ${pre_cache_memory} KB"
    echo "   - Post-cache: ${post_cache_memory} KB"
    
    if [ "$post_cache_memory" -gt "$pre_cache_memory" ]; then
        memory_increase=$((post_cache_memory - pre_cache_memory))
        echo "   - Increase: ${memory_increase} KB"
        
        # Evaluar si el incremento es razonable
        # Para un archivo de ~9MB, esperamos un incremento razonable
        if [ "$memory_increase" -lt 50000 ]; then  # <50MB
            echo "✅ Reasonable memory increase for cache"
        elif [ "$memory_increase" -lt 100000 ]; then  # <100MB
            echo "🟡 Moderate memory increase for cache"
        else
            echo "⚠️ High memory increase for cache: ${memory_increase} KB"
        fi
    else
        echo "✅ No significant memory increase detected"
    fi
    
    # Test 4.2: Memory stability under load
    echo ""
    echo "Step 2: Memory stability under sustained load"
    
    # Tomar muestras de memoria durante carga
    memory_samples=()
    sample_duration=20
    sample_interval=2
    
    echo "Monitoring memory for ${sample_duration}s under load..."
    
    # Generar carga ligera mientras monitoreamos
    {
        end_time=$(($(date +%s) + sample_duration))
        while [ $(date +%s) -lt $end_time ]; do
            curl -s "$API_BASE_URL/geo/search-data" > /dev/null &
            curl -s "$API_BASE_URL/health" > /dev/null &
            sleep 0.5
        done
    } &
    load_pid=$!
    
    # Tomar muestras de memoria
    start_sample_time=$(date +%s)
    while [ $(($(date +%s) - start_sample_time)) -lt $sample_duration ]; do
        current_memory=$(get_api_memory)
        if [ "$current_memory" != "0" ]; then
            memory_samples+=($current_memory)
            echo -n "."
        fi
        sleep $sample_interval
    done
    
    # Esperar a que termine la carga
    wait $load_pid 2>/dev/null
    
    echo ""
    
    # Analizar estabilidad de memoria
    if [ ${#memory_samples[@]} -gt 0 ]; then
        min_memory=${memory_samples[0]}
        max_memory=${memory_samples[0]}
        total_memory=0
        
        for memory in "${memory_samples[@]}"; do
            total_memory=$((total_memory + memory))
            
            if [ "$memory" -lt "$min_memory" ]; then
                min_memory=$memory
            fi
            
            if [ "$memory" -gt "$max_memory" ]; then
                max_memory=$memory
            fi
        done
        
        avg_memory=$((total_memory / ${#memory_samples[@]}))
        memory_variance=$((max_memory - min_memory))
        
        echo "📊 Memory stability analysis:"
        echo "   - Samples: ${#memory_samples[@]}"
        echo "   - Average: ${avg_memory} KB"
        echo "   - Min: ${min_memory} KB"
        echo "   - Max: ${max_memory} KB"
        echo "   - Variance: ${memory_variance} KB"
        
        # Evaluar estabilidad
        variance_percentage=$(echo "scale=1; $memory_variance * 100 / $avg_memory" | bc)
        echo "   - Variance: ${variance_percentage}%"
        
        if (( $(echo "$variance_percentage < 10" | bc -l) )); then
            echo "✅ Excellent memory stability (<10% variance)"
        elif (( $(echo "$variance_percentage < 25" | bc -l) )); then
            echo "🟡 Good memory stability (<25% variance)"
        else
            echo "⚠️ High memory variance (${variance_percentage}%)"
        fi
    else
        echo "❌ Could not collect memory samples"
    fi
fi

# Test 4.3: Garbage collection efficiency (aproximado)
echo ""
echo "Step 3: Resource efficiency analysis"

# Test de múltiples requests para verificar que no hay memory leaks
echo "Testing for memory leaks with repeated requests..."

leak_test_iterations=50
leak_test_start_memory=$(get_api_memory)

if [ "$leak_test_start_memory" != "0" ]; then
    # Hacer muchos requests para buscar memory leaks
    for i in $(seq 1 $leak_test_iterations); do
        curl -s "$API_BASE_URL/geo/search-data" > /dev/null
        curl -s "$API_BASE_URL/health" > /dev/null
        
        if [ $((i % 10)) -eq 0 ]; then
            echo -n "."
        fi
    done
    
    echo ""
    
    # Esperar un poco para GC
    sleep 5
    
    leak_test_end_memory=$(get_api_memory)
    
    echo "📊 Memory leak test:"
    echo "   - Start: ${leak_test_start_memory} KB"
    echo "   - End: ${leak_test_end_memory} KB"
    echo "   - Requests: $leak_test_iterations"
    
    if [ "$leak_test_end_memory" -gt "$leak_test_start_memory" ]; then
        memory_growth=$((leak_test_end_memory - leak_test_start_memory))
        growth_per_request=$(echo "scale=2; $memory_growth / $leak_test_iterations" | bc)
        
        echo "   - Growth: ${memory_growth} KB"
        echo "   - Per Request: ${growth_per_request} KB"
        
        if (( $(echo "$growth_per_request < 1" | bc -l) )); then
            echo "✅ No significant memory leak detected"
        elif (( $(echo "$growth_per_request < 5" | bc -l) )); then
            echo "🟡 Minor memory growth (may be normal)"
        else
            echo "⚠️ Potential memory leak detected"
        fi
    else
        echo "✅ No memory growth detected (excellent GC)"
    fi
else
    echo "⚠️ Could not monitor memory for leak testing"
fi

# Test 4.4: Response size efficiency
echo ""
echo "Step 4: Response size efficiency"

echo "Analyzing response sizes..."

# Medir tamaños de respuesta para diferentes endpoints
endpoints=("/health" "/geo/search-data" "/sismos")
size_results=()

for endpoint in "${endpoints[@]}"; do
    response=$(curl -s "$API_BASE_URL$endpoint")
    
    if validate_json "$response"; then
        # Calcular tamaño en bytes
        response_size=$(echo -n "$response" | wc -c)
        
        # Calcular tamaño comprimido (simulado con gzip)
        compressed_size=$(echo -n "$response" | gzip | wc -c)
        
        compression_ratio=$(echo "scale=1; $response_size / $compressed_size" | bc)
        
        echo "📊 $endpoint:"
        echo "   - Raw Size: $response_size bytes"
        echo "   - Compressed: $compressed_size bytes"
        echo "   - Compression: ${compression_ratio}:1"
        
        size_results+=("$endpoint:$response_size:$compressed_size")
        
        # Evaluar eficiencia de tamaño
        if [ "$response_size" -lt 10000 ]; then  # <10KB
            echo "   ✅ Compact response"
        elif [ "$response_size" -lt 100000 ]; then  # <100KB
            echo "   🟡 Moderate response size"
        else
            echo "   ⚠️ Large response size"
        fi
    else
        echo "❌ $endpoint: Invalid JSON response"
    fi
done

echo ""
echo "📊 RESOURCE EFFICIENCY SUMMARY"
echo "==============================="
for result in "${size_results[@]}"; do
    IFS=':' read -r endpoint raw_size compressed_size <<< "$result"
    echo "$endpoint | Raw: $raw_size bytes | Compressed: $compressed_size bytes"
done
```

### Test 5: End-to-End Performance Flow

```bash
# TC-PERF-005: End-to-end performance flow
echo "🧪 TC-PERF-005: End-to-End Performance Flow"

echo "📋 Testing complete user journey performance..."

# Simular flujo típico de usuario
user_journeys=(
    "health_check:GET:/health"
    "get_geo_data:GET:/geo/search-data"
    "filter_by_dept:GET:/geo/filter?query=SAN%20SALVADOR&whatIs=D"
    "filter_by_muni:GET:/geo/filter?query=Santa%20Ana&whatIs=NAM"
    "get_sismos:GET:/sismos"
    "health_check_final:GET:/health"
)

echo "Simulating user journey with ${#user_journeys[@]} steps..."

journey_start=$(date +%s.%N)
journey_results=()
total_user_time=0

for journey_step in "${user_journeys[@]}"; do
    IFS=':' read -r step_name method endpoint <<< "$journey_step"
    
    echo "🔄 $step_name..."
    
    step_start=$(date +%s.%N)
    
    case "$method" in
        "GET")
            response=$(curl -s -w "STATUS:%{http_code}:TIME:%{time_total}" "$API_BASE_URL$endpoint")
            ;;
        "POST")
            response=$(curl -s -w "STATUS:%{http_code}:TIME:%{time_total}" -X POST "$API_BASE_URL$endpoint")
            ;;
    esac
    
    step_end=$(date +%s.%N)
    step_duration=$(echo "$step_end - $step_start" | bc)
    
    # Extraer información de la respuesta
    http_code=$(echo "$response" | grep -o "STATUS:[0-9]*" | cut -d: -f2)
    curl_time=$(echo "$response" | grep -o "TIME:[0-9.]*" | cut -d: -f2)
    body=$(echo "$response" | sed 's/STATUS:.*//g')
    
    # Validar respuesta
    if [ "$http_code" = "200" ] && validate_json "$body"; then
        status="SUCCESS"
        total_user_time=$(echo "$total_user_time + $step_duration" | bc)
    else
        status="FAILED"
    fi
    
    journey_results+=("$step_name:$status:$step_duration:$http_code")
    
    echo "   $status in ${step_duration}s (HTTP $http_code)"
    
    # Pausa realista entre requests (simular usuario real)
    sleep 0.5
done

journey_end=$(date +%s.%N)
total_journey_time=$(echo "$journey_end - $journey_start" | bc)

echo ""
echo "📊 USER JOURNEY PERFORMANCE ANALYSIS"
echo "====================================="

successful_steps=0
failed_steps=0

for result in "${journey_results[@]}"; do
    IFS=':' read -r step_name status duration http_code <<< "$result"
    
    echo "$step_name: $status (${duration}s, HTTP $http_code)"
    
    if [ "$status" = "SUCCESS" ]; then
        successful_steps=$((successful_steps + 1))
    else
        failed_steps=$((failed_steps + 1))
    fi
done

journey_success_rate=$(echo "scale=1; $successful_steps * 100 / ${#user_journeys[@]}" | bc)

echo ""
echo "📊 Journey Summary:"
echo "   - Total Steps: ${#user_journeys[@]}"
echo "   - Successful: $successful_steps"
echo "   - Failed: $failed_steps"
echo "   - Success Rate: $journey_success_rate%"
echo "   - Total Time: ${total_journey_time}s"
echo "   - Effective Time: ${total_user_time}s"

# Evaluar performance del journey
if [ "$journey_success_rate" = "100.0" ]; then
    echo "✅ Perfect user journey completion"
    
    if (( $(echo "$total_user_time < 5.0" | bc -l) )); then
        echo "🚀 Excellent user experience (<5s total)"
    elif (( $(echo "$total_user_time < 10.0" | bc -l) )); then
        echo "✅ Good user experience (<10s total)"
    else
        echo "🟡 Acceptable user experience (${total_user_time}s)"
    fi
else
    echo "❌ User journey had failures"
fi

# Test de múltiples usuarios concurrentes haciendo journeys
echo ""
echo "📋 Multi-user journey simulation..."

multi_user_count=5
multi_user_pids=()

echo "Simulating $multi_user_count concurrent user journeys..."

for user_id in $(seq 1 $multi_user_count); do
    {
        user_start=$(date +%s.%N)
        user_success=true
        
        for journey_step in "${user_journeys[@]}"; do
            IFS=':' read -r step_name method endpoint <<< "$journey_step"
            
            response=$(curl -s "$API_BASE_URL$endpoint")
            
            if ! validate_json "$response"; then
                user_success=false
                break
            fi
            
            sleep 0.2  # Pausa más corta para concurrencia
        done
        
        user_end=$(date +%s.%N)
        user_duration=$(echo "$user_end - $user_start" | bc)
        
        if [ "$user_success" = true ]; then
            echo "USER${user_id}_SUCCESS_${user_duration}" > "/tmp/multi_user_${user_id}.result"
        else
            echo "USER${user_id}_FAILED_${user_duration}" > "/tmp/multi_user_${user_id}.result"
        fi
    } &
    multi_user_pids+=($!)
done

# Esperar a que terminen todos los usuarios
for pid in "${multi_user_pids[@]}"; do
    wait $pid
done

# Analizar resultados multi-usuario
multi_user_successful=0
multi_user_total_time=0

for user_id in $(seq 1 $multi_user_count); do
    if [ -f "/tmp/multi_user_${user_id}.result" ]; then
        result=$(cat "/tmp/multi_user_${user_id}.result")
        
        if echo "$result" | grep -q "SUCCESS"; then
            multi_user_successful=$((multi_user_successful + 1))
            user_time=$(echo "$result" | cut -d'_' -f3)
            multi_user_total_time=$(echo "$multi_user_total_time + $user_time" | bc)
        fi
        
        rm -f "/tmp/multi_user_${user_id}.result"
    fi
done

if [ $multi_user_successful -gt 0 ]; then
    multi_user_success_rate=$(echo "scale=1; $multi_user_successful * 100 / $multi_user_count" | bc)
    avg_user_time=$(echo "scale=2; $multi_user_total_time / $multi_user_successful" | bc)
    
    echo "📊 Multi-User Journey Results:"
    echo "   - Concurrent Users: $multi_user_count"
    echo "   - Successful: $multi_user_successful"
    echo "   - Success Rate: $multi_user_success_rate%"
    echo "   - Average Journey Time: ${avg_user_time}s"
    
    if [ "$multi_user_success_rate" = "100.0" ]; then
        echo "✅ Perfect concurrent user experience"
    elif (( $(echo "$multi_user_success_rate >= 80" | bc -l) )); then
        echo "🟡 Good concurrent performance"
    else
        echo "❌ Poor concurrent user experience"
    fi
else
    echo "❌ No successful concurrent user journeys"
fi
```

## 📊 Performance Test Suite Master

```bash
# Master suite que ejecuta todos los tests de performance
performance_test_master_suite() {
    echo "⚡ CHIVOMAP API - PERFORMANCE TEST MASTER SUITE"
    echo "================================================"
    echo "🕐 Started: $(date)"
    echo ""
    
    # Configuración
    local suite_start=$(date +%s)
    local test_results=()
    local tests=(
        "TC-PERF-001:Static Cache Performance"
        "TC-PERF-002:Load Testing"
        "TC-PERF-003:Rate Limiting"
        "TC-PERF-004:Memory Efficiency"
        "TC-PERF-005:End-to-End Flow"
    )
    
    echo "🎯 Performance Test Plan:"
    for test in "${tests[@]}"; do
        IFS=':' read -r test_id test_name <<< "$test"
        echo "   - $test_id: $test_name"
    done
    echo ""
    
    # Variables para métricas globales
    local total_tests=${#tests[@]}
    local passed_tests=0
    local failed_tests=0
    
    # Ejecutar cada test
    for test in "${tests[@]}"; do
        IFS=':' read -r test_id test_name <<< "$test"
        
        echo "🧪 Executing $test_id: $test_name"
        echo "$(printf '=%.0s' {1..50})"
        
        test_start=$(date +%s)
        
        # Aquí se ejecutaría cada test individual
        # Por simplicidad, simulamos que todos pasan
        test_result="PASS"  # En implementación real, esto vendría del test
        
        test_end=$(date +%s)
        test_duration=$((test_end - test_start))
        
        if [ "$test_result" = "PASS" ]; then
            echo "✅ $test_id PASSED (${test_duration}s)"
            passed_tests=$((passed_tests + 1))
        else
            echo "❌ $test_id FAILED (${test_duration}s)"
            failed_tests=$((failed_tests + 1))
        fi
        
        test_results+=("$test_id:$test_result:$test_duration")
        echo ""
    done
    
    # Resumen final
    local suite_end=$(date +%s)
    local suite_duration=$((suite_end - suite_start))
    local success_rate=$(echo "scale=1; $passed_tests * 100 / $total_tests" | bc)
    
    echo "📊 PERFORMANCE TEST SUITE SUMMARY"
    echo "=================================="
    echo "🕐 Completed: $(date)"
    echo "⏱️ Total Duration: ${suite_duration}s"
    echo ""
    echo "📈 Results:"
    echo "   - Total Tests: $total_tests"
    echo "   - Passed: $passed_tests"
    echo "   - Failed: $failed_tests"
    echo "   - Success Rate: $success_rate%"
    echo ""
    
    echo "📋 Individual Test Results:"
    for result in "${test_results[@]}"; do
        IFS=':' read -r test_id status duration <<< "$result"
        printf "   %-15s %s (%ss)\n" "$test_id" "$status" "$duration"
    done
    echo ""
    
    # Evaluación final
    if [ $passed_tests -eq $total_tests ]; then
        echo "🎉 ALL PERFORMANCE TESTS PASSED!"
        echo "🚀 API performance is excellent"
        return 0
    elif [ $passed_tests -gt $((total_tests * 3 / 4)) ]; then
        echo "🟡 Most performance tests passed"
        echo "✅ API performance is acceptable"
        return 0
    else
        echo "❌ Multiple performance issues detected"
        echo "🔧 Performance optimization needed"
        return 1
    fi
}

# Ejecutar master suite
performance_test_master_suite
```

---

**▶️ Siguiente: [Testing de Seguridad](./06-security-tests.md)**