# Documentación de la API ChivoMap

## Endpoints Principales

### Datos Sísmicos

#### GET /sismos
Obtiene información de sismos recientes en El Salvador.

**Respuesta**: 
```json
{
  "timestamp": "2023-05-25T12:34:56Z",
  "data": {
    "totalSismos": 10,
    "data": [
      {
        "fecha": "2023-05-25 10:30:00",
        "fases": "P,S",
        "latitud": "13.6894",
        "longitud": "-89.1872",
        "profundidad": "5.5",
        "magnitud": "4.2",
        "localizacion": "5 km al Este de San Salvador",
        "rms": "0.3",
        "estado": "Revisado"
      }
      // Más sismos...
    ]
  }
}
```

#### GET /sismos/refresh
Fuerza la actualización de datos sísmicos.

**Respuesta**: 
```json
{
  "timestamp": "2023-05-25T12:34:56Z",
  "data": {
    "message": "Cache actualizada exitosamente",
    "totalSismos": 10,
    "data": [
      // Datos de sismos actualizados
    ]
  }
}
```

### Datos Geoespaciales

#### GET /geo/search-data
Obtiene datos geográficos de El Salvador.

**Respuesta**: 
```json
{
  "timestamp": "2023-05-25T12:34:56Z",
  "data": {
    "departamentos": ["San Salvador", "La Libertad", "Santa Ana", ...],
    "municipios": ["San Salvador", "Santa Tecla", "Mejicanos", ...],
    "distritos": ["Centro", "Norte", "Sur", ...]
  }
}
```

#### GET /geo/filter
Filtra datos geoespaciales según parámetros.

**Parámetros**:
- `query`: Cadena de búsqueda
- `whatIs`: Tipo de filtro (departamento, municipio, etc.)

**Respuesta**: GeoJSON con los resultados filtrados.

### Otros Endpoints

#### GET /health
Verifica el estado de la API.

**Respuesta**:
```json
{
  "timestamp": "2023-05-25T12:34:56Z",
  "data": {
    "status": "UP",
    "version": "1.0.0"
  }
}
```

#### GET /scrape
Obtiene datos scrapeados.

**Respuesta**:
```json
{
  "timestamp": "2023-05-25T12:34:56Z",
  "data": {
    "totalItems": 5,
    "items": [
      {
        "id": 1,
        "title": "Título ejemplo 1"
      }
      // Más datos...
    ]
  }
}
```

## Formatos de Respuesta

Todas las respuestas tienen un formato estándar:

### Respuestas Exitosas
```json
{
  "timestamp": "2023-05-25T12:34:56Z",  // Marca de tiempo (RFC3339)
  "data": { ... }                       // Datos específicos del endpoint
}
```

### Respuestas de Error
```json
{
  "timestamp": "2023-05-25T12:34:56Z",  // Marca de tiempo (RFC3339)
  "error": "Mensaje descriptivo del error"
}
```

## Códigos de Estado HTTP

- **200**: Solicitud exitosa
- **400**: Error en la solicitud (parámetros inválidos)
- **404**: Recurso no encontrado
- **500**: Error interno del servidor 