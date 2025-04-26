# Guía de Instalación y Uso

## Requisitos Previos

- Go 1.18 o superior
- Cuenta en [Turso](https://turso.tech/) para la base de datos

## Instalación

1. **Clonar el repositorio**

```bash
git clone https://github.com/oclazi/chivomap-api.git
cd chivomap-api
```

2. **Configurar variables de entorno**

Crea un archivo `.env` en la raíz del proyecto:

```
TURSO_DATABASE_URL=libsql://tu-base-de-datos.turso.io
TURSO_AUTH_TOKEN=tu-token-de-acceso
PORT=8080  # Opcional, por defecto es 8080
```

3. **Instalar dependencias**

```bash
go mod tidy
```

## Ejecución

### Modo Desarrollo

```bash
go run main.go
```

### Compilar y Ejecutar

```bash
go build -o chivomap-api
./chivomap-api
```

### Usando Docker

Construir la imagen:

```bash
docker build -t chivomap-api .
```

Ejecutar el contenedor:

```bash
docker run -p 8080:8080 --env-file .env chivomap-api
```

## Verificación de Funcionamiento

Una vez en ejecución, la API estará disponible en:

```
http://localhost:8080
```

Puedes verificar que funciona correctamente accediendo a:

```
http://localhost:8080/health
```

Deberías recibir una respuesta como:

```json
{
  "timestamp": "2023-05-25T12:34:56Z",
  "data": {
    "status": "UP",
    "version": "1.0.0"
  }
}
```

## Endpoints Principales

- **Sismos**: `GET /sismos`
- **Datos Geoespaciales**: `GET /geo/search-data`

Para más detalles sobre los endpoints disponibles, consulta la [documentación de la API](API.md).

## Logs

La aplicación genera logs estructurados con diferentes niveles (INFO, ERROR, DEBUG, FATAL) que se muestran en la salida estándar.

Ejemplo de log:

```
[INFO] 2023-05-25 10:30:15 - 🚀 Servidor corriendo en http://localhost:8080
[HTTP] 2023-05-25 10:31:20 - GET /health 200 2.5ms
```

## Cierre

Para detener la aplicación de forma segura, envía una señal SIGINT (Ctrl+C) o SIGTERM. La aplicación cerrará correctamente todas las conexiones antes de finalizar. 