# Arquitectura de ChivoMap API

Este documento describe la arquitectura y estructura del proyecto ChivoMap API.

## Estructura del Proyecto

```
chivomap-api/
├── config/            # Configuración de la aplicación
│   ├── config.go      # Gestión de variables de entorno
│   └── database.go    # Conexión a la base de datos
├── handlers/          # Controladores HTTP
│   ├── geo.go         # Endpoints de servicios geoespaciales
│   ├── health.go      # Endpoint de verificación de estado
│   ├── routes.go      # Configuración centralizada de rutas
│   ├── scrape.go      # Endpoints para web scraping
│   └── sismos.go      # Endpoints de información sísmica
├── models/            # Modelos de datos
│   └── scraped_data.go # Estructura para datos scrapeados
├── services/          # Lógica de negocio
│   ├── cache.go       # Servicio genérico de caché
│   ├── geospatial/    # Servicios geoespaciales
│   └── scraping/      # Servicios de scraping
├── utils/             # Utilidades comunes
│   ├── logger.go      # Sistema de logging
│   └── response.go    # Formato de respuestas HTTP
├── main.go            # Punto de entrada de la aplicación
├── docs/              # Documentación
├── go.mod             # Dependencias de Go
└── go.sum             # Checksums de dependencias
```

## Componentes Principales

### Configuración (config/)

- **config.go**: Gestiona la configuración de la aplicación, cargando variables desde el entorno.
- **database.go**: Establece la conexión con la base de datos Turso y crea las tablas necesarias.

### Controladores (handlers/)

Los controladores manejan las solicitudes HTTP y coordinan la lógica de negocio.

- **routes.go**: Centraliza la configuración de todas las rutas de la API.
- **geo.go**: Implementa endpoints para datos geoespaciales.
- **sismos.go**: Gestiona endpoints para información de sismos, con caché integrada.
- **scrape.go**: Maneja funcionalidades de web scraping.
- **health.go**: Proporciona endpoint para verificar el estado de la API.

### Servicios (services/)

Los servicios implementan la lógica de negocio principal.

- **cache.go**: Implementa un sistema de caché genérico con TTL.
- **geospatial/**: Servicios para procesar y filtrar datos geoespaciales.
- **scraping/**: Servicios para obtener datos mediante web scraping.

### Utilidades (utils/)

- **logger.go**: Sistema de logging con niveles (INFO, ERROR, DEBUG, FATAL).
- **response.go**: Formato estandarizado para respuestas HTTP.

## Patrones de Diseño

1. **Patrón Handler**: Los controladores están implementados como estructuras con métodos, permitiendo la encapsulación de estado.

2. **Caché con TTL**: Implementación de caché con tiempo de vida para reducir consultas costosas.

3. **Gestor de Configuración**: Sistema centralizado para gestionar configuraciones desde variables de entorno.

4. **Respuestas Estandarizadas**: Formato consistente para todas las respuestas de la API.

## Flujo de Datos

1. Las solicitudes HTTP llegan al servidor (main.go).
2. El router (configurado en routes.go) redirige a los controladores adecuados.
3. Los controladores utilizan servicios para ejecutar la lógica de negocio.
4. Los resultados se devuelven en un formato estandarizado.

## Sistema de Caché

El proyecto implementa un servicio de caché genérico con control de TTL (Time To Live):

- Reduce la carga en servicios externos (APIs, web scraping).
- Actualización asíncrona en segundo plano.
- TTL configurable según tipo de datos.

## Cierre Graceful

La aplicación implementa un mecanismo de cierre graceful que:

1. Captura señales de sistema (SIGINT, SIGTERM).
2. Completa las solicitudes en curso.
3. Cierra correctamente las conexiones a la base de datos.
4. Finaliza limpiamente la aplicación. 