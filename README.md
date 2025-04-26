# 🌍 ChivoMap API

API en Go que proporciona datos geoespaciales y sísmicos de El Salvador.

## 📌 Descripción

ChivoMap API es una aplicación backend desarrollada con Go y Fiber que proporciona:

- Información actualizada de sismos en El Salvador
- Datos geoespaciales del territorio salvadoreño
- Sistema de caché para optimizar el rendimiento

La API utiliza **Fiber** para el manejo de rutas, **Colly/Playwright** para web scraping, y **Turso DB** como base de datos.

## 🚀 Instalación Rápida

```bash
# Clonar el repositorio
git clone https://github.com/oclazi/chivomap-api.git
cd chivomap-api

# Configurar variables de entorno
# Crea un archivo .env con:
# TURSO_DATABASE_URL=libsql://your-database.turso.io
# TURSO_AUTH_TOKEN=your-auth-token

Si `.env` no se carga, usa:
```bash
export $(grep -v '^#' .env | xargs)
```

# Instalar dependencias
go mod tidy

# Ejecutar la API
go run main.go
```

La API estará disponible en: http://localhost:8080

## 🛠️ Características

- **Sistema de Caché**: Reducción de carga en servicios externos mediante caché con TTL
- **Respuestas Estandarizadas**: Formato JSON consistente para todas las respuestas
- **Logs Estructurados**: Sistema de logging con niveles INFO, ERROR, DEBUG, FATAL
- **Cierre Graceful**: Manejo adecuado de finalización de la aplicación

## 📘 Documentación

La documentación completa está disponible en la carpeta [`docs/`](docs/):

- [Guía de Instalación y Uso](docs/GUIA.md)
- [Documentación de la API](docs/API.md)
- [Arquitectura del Proyecto](docs/ARQUITECTURA.md)
- [Documentación Swagger](docs/SWAGGER.md)

## 🔍 Endpoints Principales

- **GET /sismos**: Información de sismos recientes
- **GET /geo/search-data**: Datos geográficos de El Salvador
- **GET /geo/filter**: Filtra datos geoespaciales

Puedes explorar todos los endpoints a través de la interfaz Swagger disponible en:
```
http://localhost:8080/swagger/
```

## 📄 Licencia

Este proyecto está licenciado bajo MIT License - ver [LICENSE.md](LICENSE.md) para más detalles.

## 📝 Aviso de Propiedad Intelectual

Copyright © 2024 ChivoMap. Todos los derechos reservados.

Aunque este proyecto es de código abierto, se aplican las siguientes restricciones:
- El nombre y logo de ChivoMap son marcas registradas y no pueden utilizarse sin permiso
- El uso comercial requiere autorización escrita
- Los trabajos derivados deben mantener todos los avisos de copyright y licencia