# 🔧 Setup y Configuración

Esta guía cubre la configuración inicial necesaria para ejecutar pruebas de la ChivoMap API.

## 📋 Prerequisitos

### 1. Variables de Entorno Requeridas

```bash
# Crear archivo .env si no existe
cat > .env << EOF
TURSO_DATABASE_URL=libsql://chivomap-api-oclazi.turso.io
TURSO_AUTH_TOKEN=your_token_here
TURSO_DATABASE_URL_CENSO=libsql://censo2024-oclazi.aws-us-east-1.turso.io
TURSO_AUTH_TOKEN_CENSO=your_censo_token_here
EOF

# Cargar variables
export $(grep -v '^#' .env | xargs)
```

### 2. Compilar la API

```bash
# Compilar binary
go build -o chivomap-api .

# Verificar que el binary existe
ls -la chivomap-api
```

### 3. Verificar Assets

```bash
# Verificar que el archivo TopoJSON existe
ls -la utils/assets/topo.json

# Debería mostrar algo como:
# -rw-r--r-- 1 user user 9014869 Feb 22 08:52 utils/assets/topo.json
```

## 🚀 Iniciar la API para Testing

### Opción 1: Puerto por Defecto (8080)

```bash
export PORT=8080
./chivomap-api &

# Guardar PID para terminar después
API_PID=$!
echo $API_PID > .api.pid
```

### Opción 2: Puerto Personalizado

```bash
export PORT=8081
./chivomap-api &
API_PID=$!
echo $API_PID > .api.pid
```

## ✅ Verificación de Setup

### 1. Verificar que la API está corriendo

```bash
# Test básico de conectividad
curl -s -I http://localhost:8080/health

# Respuesta esperada:
# HTTP/1.1 200 OK
# Content-Type: application/json
# X-Ratelimit-Limit: 100
# X-Ratelimit-Remaining: 99
```

### 2. Verificar logs de inicio

```bash
# Los logs deberían mostrar:
# ✅ Conectado a la base de datos Turso
# 🚀 Servidor corriendo en http://localhost:8080
# 📚 Documentación Swagger disponible en http://localhost:8080/docs/
```

### 3. Verificar componentes críticos

```bash
# Health check detallado
curl -s http://localhost:8080/health | python3 -m json.tool
```

**Respuesta esperada:**
```json
{
  "status": "UP",
  "version": "1.0.0",
  "timestamp": "2025-06-28T16:00:00-06:00",
  "uptime": "5.123456789s",
  "components": {
    "database": {
      "status": "UP",
      "details": {
        "open_connections": 1,
        "in_use": 0,
        "idle": 1
      }
    },
    "static_files": {
      "status": "UP",
      "details": {
        "topo_file_size": 9014869,
        "topo_mod_time": "2025-02-22T08:52:32.78106399-06:00"
      }
    },
    "cache": {
      "status": "DOWN",
      "message": "Cache estático no inicializado"
    },
    "censo_database": {
      "status": "DOWN",
      "message": "Base de datos del censo no configurada"
    }
  }
}
```

## 🔧 Configuración de Variables para Testing

```bash
# Variables base para todos los tests
export API_BASE_URL="http://localhost:8080"
export TIMEOUT_SECONDS=30
export MAX_RETRIES=3

# Para testing de performance
export CONCURRENT_REQUESTS=10
export PERFORMANCE_DURATION=60

# Para debugging
export VERBOSE_OUTPUT=true
export SAVE_RESPONSES=true
export RESPONSE_DIR="./test-responses"

# Crear directorio para respuestas si no existe
mkdir -p $RESPONSE_DIR
```

## 🧪 Funciones Helper para Testing

```bash
# Función para hacer requests con timeout y retry
api_request() {
    local method=$1
    local endpoint=$2
    local expected_status=$3
    local max_attempts=${MAX_RETRIES:-3}
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if [ "$VERBOSE_OUTPUT" = "true" ]; then
            echo "🔄 Attempt $attempt/$max_attempts: $method $endpoint"
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
        sleep 1
    done
    
    echo "❌ Failed after $max_attempts attempts. Last response: $http_code"
    return 1
}

# Función para verificar JSON válido
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
data = json.load(sys.stdin)
try:
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
```

## 🎯 Test de Smoke (Verificación Rápida)

```bash
#!/bin/bash
# smoke_test.sh - Verificación básica de 30 segundos

echo "🧪 ChivoMap API - Smoke Test"
echo "================================"

# 1. Health Check
echo "1️⃣ Testing Health Check..."
health_response=$(api_request GET "/health" 200)
if [ $? -eq 0 ]; then
    status=$(extract_json_field "$health_response" "status")
    echo "✅ Health Check: $status"
else
    echo "❌ Health Check failed"
    exit 1
fi

# 2. Swagger Documentation
echo "2️⃣ Testing Swagger Docs..."
swagger_response=$(curl -s -w "%{http_code}" "$API_BASE_URL/docs/doc.json")
http_code=$(echo "$swagger_response" | tail -c 4)
if [ "$http_code" = "200" ]; then
    echo "✅ Swagger Docs accessible"
else
    echo "❌ Swagger Docs failed: $http_code"
fi

# 3. Geo Data (triggers cache loading)
echo "3️⃣ Testing Geo Data..."
geo_response=$(api_request GET "/geo/search-data" 200)
if [ $? -eq 0 ]; then
    departamentos_count=$(extract_json_field "$geo_response" "data.departamentos" | wc -w)
    echo "✅ Geo Data loaded ($departamentos_count departments)"
else
    echo "❌ Geo Data failed"
fi

# 4. Verificar cache después de carga
echo "4️⃣ Verifying Cache..."
sleep 2
health_response=$(api_request GET "/health" 200)
cache_status=$(extract_json_field "$health_response" "components.cache.status")
if [ "$cache_status" = "UP" ]; then
    features_count=$(extract_json_field "$health_response" "components.cache.details.geoFeatures")
    echo "✅ Cache loaded ($features_count features)"
else
    echo "🟡 Cache status: $cache_status"
fi

echo "🎉 Smoke Test Completed!"
```

## 🛑 Limpieza después del Testing

```bash
#!/bin/bash
# cleanup.sh

echo "🧹 Cleaning up test environment..."

# Terminar API si está corriendo
if [ -f .api.pid ]; then
    PID=$(cat .api.pid)
    if kill -0 $PID 2>/dev/null; then
        echo "🛑 Stopping API (PID: $PID)..."
        kill $PID
        sleep 2
        
        # Force kill si es necesario
        if kill -0 $PID 2>/dev/null; then
            kill -9 $PID
            echo "💀 Force killed API"
        fi
    fi
    rm -f .api.pid
fi

# Limpiar archivos temporales
rm -rf $RESPONSE_DIR
echo "🗑️ Cleaned temporary files"

echo "✅ Cleanup completed"
```

---

**▶️ Siguiente: [Health Checks y Monitoreo](./02-health-checks.md)**