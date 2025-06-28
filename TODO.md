# TODO - ChivoMap API

Lista de tareas pendientes para mejorar la estabilidad, seguridad y rendimiento de la API.

## ✅ Problemas Críticos Resueltos

### 1. ✅ Resource Leak - Database Connection
- **Archivo:** `main.go:87-98`
- **Problema:** ~~La conexión global `DB` nunca se cierra durante shutdown~~ **RESUELTO**
- **Solución:** Agregado cleanup de DB y CensoDB en graceful shutdown
- **Commit:** `9c8aaa7` - fix/critical-issues

### 2. ✅ Variables de Entorno Faltantes
- **Archivo:** `.env.example`
- **Problema:** ~~Faltan `TURSO_DATABASE_URL` y `TURSO_AUTH_TOKEN` requeridas~~ **RESUELTO**
- **Solución:** Agregadas todas las variables necesarias con documentación
- **Commit:** `9c8aaa7` - fix/critical-issues

### 3. ✅ Vulnerabilidad de Seguridad - Browser Headless
- **Archivo:** `services/scraping/sismos.go:35-37`
- **Problema:** ~~`Headless: false` en producción~~ **RESUELTO**
- **Solución:** Configurable vía `BROWSER_HEADLESS` (default: true)
- **Commit:** `9c8aaa7` - fix/critical-issues

## 🟠 Problemas Importantes (Media Prioridad)

### 4. Exceso de Fatal Calls
- **Archivos:** `config/config.go:36`, `config/database.go:25,33,38,51`, `services/scraping/sismos.go:28,39,49,56,71,81,107`
- **Problema:** App termina inmediatamente en cualquier error
- **Impacto:** Mala experiencia de usuario, sin recuperación de errores
- **Solución:** Reemplazar con proper error handling y logging

### 5. Falta de Validación de Input
- **Archivo:** `handlers/geo.go:36-42`
- **Problema:** Query parameters sin validación
- **Impacto:** Riesgo de SQL injection, crashes
- **Solución:** Implementar validación y sanitización

### 6. Rutas Hardcodeadas
- **Archivo:** `services/geospatial/search.go:238`
- **Problema:** `"./utils/assets/topo.json"` hardcodeado
- **Impacto:** Falla si se ejecuta desde otro directorio
- **Solución:** Usar rutas absolutas o configurables

### 7. Race Condition en Cache
- **Archivo:** `handlers/geo.go:57-59`
- **Problema:** Cache map sin sincronización
- **Impacto:** Corrupción de datos en concurrencia
- **Solución:** Implementar proper synchronization

## 🟡 Problemas Medios (Baja Prioridad)

### 8. Error Handling Inconsistente
- **Archivo:** `services/geospatial/search.go:208-211`
- **Problema:** Errores de geometría ignorados silenciosamente
- **Impacto:** Pérdida de datos sin notificación
- **Solución:** Proper error logging y handling

### 9. Falta de Context Usage
- **Archivo:** `handlers/scrape.go:20`
- **Problema:** DB queries sin context para timeout/cancellation
- **Impacto:** Requests colgados, resource leaks
- **Solución:** Usar request context en todas las operaciones

### 10. Estrategia de Cache Poco Clara
- **Archivo:** `handlers/geo.go:44-49`
- **Problema:** Posibles key collisions, lookups ineficientes
- **Impacto:** Cache hits incorrectos, degradación de performance
- **Solución:** Rediseñar cache key strategy

### 11. Validación de Configuración Faltante
- **Archivo:** `main.go:67-69`
- **Problema:** Lógica de fallback de puerto primitiva
- **Impacto:** App puede iniciar en puertos inesperados
- **Solución:** Validación robusta de configuración

### 12. Uso de ioutil Deprecado
- **Archivo:** `docs/scripts/convert.go:7`
- **Problema:** `ioutil` está deprecado desde Go 1.16
- **Impacto:** Código legacy, warnings
- **Solución:** Usar `os.ReadFile` directamente

## 🔒 Mejoras de Seguridad

### 13. Límites de Request Size
- **Archivo:** `main.go`
- **Problema:** Sin límites de body size en Fiber
- **Impacto:** Posibles ataques DoS
- **Solución:** Configurar límites apropiados

### 14. CORS Configuration
- **Archivo:** `main.go:44-49`
- **Problema:** CORS permite orígenes específicos pero podría ser más restrictivo
- **Solución:** Configuración CORS por ambiente

### 15. Database Connection Exposure
- **Archivo:** `config/database.go:28`
- **Problema:** Auth token en URL parameters
- **Impacto:** Posible exposición en logs
- **Solución:** Manejo seguro de credenciales

## ⚡ Mejoras de Performance

### 16. I/O Excesivo
- **Archivo:** `services/geospatial/search.go`
- **Problema:** TopoJSON se lee en cada request
- **Impacto:** Alta latencia, overhead de I/O
- **Solución:** Cache del archivo en startup

### 17. Memory Allocation en Loops
- **Archivo:** `services/geospatial/search.go:275-291`
- **Problema:** Múltiples allocations sin preallocation
- **Impacto:** Presión en garbage collector
- **Solución:** Preallocate maps y slices

## 📝 Mejoras de Código

### 18. Error Wrapping Faltante
- **Archivos:** Todo el codebase
- **Problema:** Errores sin contexto
- **Impacto:** Debugging difícil
- **Solución:** Usar `fmt.Errorf("operation failed: %w", err)`

### 19. Global State Dependencies
- **Archivos:** `config/database.go`, `config/censo_db.go`
- **Problema:** Variables globales dificultan testing
- **Solución:** Dependency injection

### 20. Health Check Incompleto
- **Archivo:** `handlers/health.go`
- **Problema:** No verifica conectividad de DB
- **Impacto:** False positive en health status
- **Solución:** Agregar verificación de servicios

## 🏗️ Tareas de Refactoring

### 21. Documentación Swagger Inconsistente
- **Archivos:** `handlers/models.go`, `models/swagger.go`
- **Problema:** Modelos duplicados
- **Impacto:** Inconsistencias en docs
- **Solución:** Consolidar definiciones

### 22. Configuración Estructurada
- **Problema:** Configuración dispersa y sin validación
- **Solución:** Sistema de configuración centralizado con validación

### 23. Sistema de Logging Mejorado
- **Problema:** Logging básico sin niveles apropiados
- **Solución:** Structured logging con niveles configurables

### 24. Testing Framework
- **Problema:** Sin tests unitarios
- **Solución:** Implementar testing suite completo

## 📊 Monitoring y Observabilidad

### 25. Métricas de Performance
- **Tarea:** Implementar métricas de latencia, throughput, errores
- **Solución:** Integrar Prometheus/OpenTelemetry

### 26. Request Tracing
- **Tarea:** Implementar distributed tracing
- **Solución:** Usar OpenTracing/Jaeger

### 27. Alerting
- **Tarea:** Sistema de alertas para errores críticos
- **Solución:** Integrar alertas por logs y métricas

---

## 🎯 Roadmap de Implementación

### Fase 1 - Críticos (Semana 1)
- [x] Fix database connection leak
- [x] Completar variables de entorno
- [x] Seguridad browser headless
- [x] Validación de input básica

### Fase 2 - Importantes (Semana 2-3)
- [x] Reemplazar log.Fatal con error handling
- [x] Implementar cache synchronization
- [x] Fix rutas hardcodeadas
- [x] Context usage en DB operations

### Fase 3 - Performance (Semana 4-5)
- [x] Cache de archivos estáticos
- [x] Optimizar memory allocations
- [x] Request size limits
- [x] Health checks completos

### Fase 4 - Código y Testing (Semana 6-8)
- [ ] Error wrapping consistente
- [ ] Refactor global state
- [ ] Implementar testing suite
- [ ] Documentación consolidada

### Fase 5 - Observabilidad (Semana 9-10)
- [ ] Métricas y monitoring
- [ ] Request tracing
- [ ] Sistema de alertas
- [ ] Performance benchmarks