# 🧪 ChivoMap API - Guía de Testing con cURL

Esta documentación proporciona flujos de prueba completos para verificar la funcionalidad de la ChivoMap API usando cURL. Está diseñada para ser utilizada por sistemas automatizados, AI, y desarrolladores.

## 📋 Índice

1. [**Setup y Configuración**](./01-setup.md)
2. [**Health Checks y Monitoreo**](./02-health-checks.md)
3. [**Endpoints Geoespaciales**](./03-geo-endpoints.md)
4. [**Endpoints de Sismos**](./04-sismos-endpoints.md)
5. [**Testing de Performance**](./05-performance-tests.md)
6. [**Testing de Seguridad**](./06-security-tests.md)
7. [**Casos de Error y Edge Cases**](./07-error-cases.md)
8. [**Flujos de Integración Completos**](./08-integration-flows.md)

## 🎯 Objetivo

Proporcionar pruebas exhaustivas que verifiquen:
- ✅ **Funcionalidad correcta** de todos los endpoints
- ✅ **Performance y cache** funcionando
- ✅ **Seguridad y rate limiting**
- ✅ **Manejo de errores** apropiado
- ✅ **Documentación Swagger** accesible
- ✅ **Monitoreo y observabilidad**

## 🚀 Quick Start

```bash
# 1. Iniciar la API
export $(grep -v '^#' .env | xargs)
export PORT=8080
./chivomap-api &

# 2. Verificar que está funcionando
curl -s http://localhost:8080/health | grep -o '"status":"[^"]*"'

# 3. Ejecutar suite de pruebas básica
./docs/testing/scripts/basic-test-suite.sh
```

## 📊 Formato de Respuestas

Todas las respuestas siguen este formato estándar:

### ✅ Respuesta Exitosa
```json
{
  "data": { /* contenido específico del endpoint */ },
  "timestamp": "2025-06-28T16:00:00-06:00"
}
```

### ❌ Respuesta de Error
```json
{
  "error": "Código de error",
  "message": "Descripción detallada del error"
}
```

### 🏥 Health Check Response
```json
{
  "status": "UP|DEGRADED|DOWN",
  "version": "1.0.0",
  "timestamp": "2025-06-28T16:00:00-06:00",
  "uptime": "1h30m45s",
  "components": {
    "database": { "status": "UP", "details": {...} },
    "cache": { "status": "UP", "details": {...} },
    "static_files": { "status": "UP", "details": {...} }
  }
}
```

## 🔧 Variables de Entorno para Testing

```bash
# API Configuration
export API_BASE_URL="http://localhost:8080"
export TIMEOUT_SECONDS=30

# Test Configuration
export VERBOSE_OUTPUT=true
export SAVE_RESPONSES=true
export RESPONSE_DIR="./test-responses"
```

## 📝 Convenciones

- **🟢 PASS**: Prueba exitosa
- **🔴 FAIL**: Prueba fallida
- **🟡 WARN**: Advertencia o comportamiento inesperado
- **ℹ️ INFO**: Información adicional

## 🎪 Flujos de Prueba Recomendados

### 1. **Smoke Test** (2 minutos)
Verificación básica de que la API está funcionando.

### 2. **Functional Test** (10 minutos)
Prueba de todos los endpoints con casos típicos.

### 3. **Performance Test** (15 minutos)
Verificación de cache, latencia y throughput.

### 4. **Security Test** (5 minutos)
Rate limiting, validación de input, CORS.

### 5. **Integration Test** (20 minutos)
Flujos completos de usuario simulando casos reales.

---

## 📁 Estructura de Archivos

```
docs/testing/
├── README.md                    # Este archivo
├── 01-setup.md                 # Configuración inicial
├── 02-health-checks.md         # Pruebas de salud
├── 03-geo-endpoints.md         # Endpoints geoespaciales
├── 04-sismos-endpoints.md      # Endpoints sísmicos
├── 05-performance-tests.md     # Pruebas de rendimiento
├── 06-security-tests.md        # Pruebas de seguridad
├── 07-error-cases.md           # Casos de error
├── 08-integration-flows.md     # Flujos de integración
├── scripts/                    # Scripts automatizados
│   ├── basic-test-suite.sh     # Suite básica
│   ├── performance-test.sh     # Pruebas de performance
│   ├── security-test.sh        # Pruebas de seguridad
│   └── helpers/                # Funciones auxiliares
└── responses/                  # Respuestas de ejemplo
    ├── health-up.json
    ├── health-degraded.json
    ├── geo-data-success.json
    └── error-examples.json
```

🚀 **¡Comienza con [Setup y Configuración](./01-setup.md)!**