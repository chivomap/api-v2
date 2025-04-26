# Documentación Swagger en ChivoMap API

Este proyecto utiliza Swagger para documentar y probar la API. A continuación se describe cómo acceder y utilizar la documentación Swagger.

## Acceso a la Documentación

La documentación Swagger está disponible en la siguiente URL cuando el servidor está en ejecución:

```
http://localhost:8080/swagger/
```

## Características

- **Exploración Interactiva**: Prueba los endpoints directamente desde la interfaz Swagger
- **Documentación Detallada**: Descripción de todos los endpoints, parámetros y respuestas
- **Modelos**: Visualización de las estructuras de datos utilizadas en la API

## Endpoints Documentados

La documentación incluye información detallada sobre:

- **Sismos**: `/sismos` y `/sismos/refresh`
- **Geo**: `/geo/filter` y `/geo/search-data`
- **Health**: `/health`
- **Scraping**: `/scrape`

## Actualización de la Documentación

Si realizas cambios en los endpoints o en la estructura de la API, debes regenerar la documentación Swagger:

```bash
# Asegúrate de tener la herramienta swag instalada
go install github.com/swaggo/swag/cmd/swag@latest

# Genera la documentación
swag init
```

## Comentarios de Swagger

Para documentar correctamente un endpoint, se deben incluir comentarios especiales en el código:

```go
// @Summary Título breve
// @Description Descripción más detallada
// @Tags categoría
// @Produce json
// @Param nombre tipo ubicación requerido "descripción"
// @Success código {tipo} modelo "descripción"
// @Failure código {tipo} modelo "descripción"
// @Router /ruta [método]
func MiHandler(c *fiber.Ctx) error {
    // Implementación
}
```

## Modelos

Los modelos utilizados en la API están definidos en `models/swagger.go` para facilitar la documentación. Estos modelos incluyen ejemplos que se muestran en la interfaz Swagger.

## Beneficios

- **Comunicación Clara**: Facilita la comunicación entre desarrolladores backend y frontend
- **Pruebas Rápidas**: Permite probar endpoints sin necesidad de herramientas externas
- **Documentación Actualizada**: La documentación se genera a partir del código fuente 