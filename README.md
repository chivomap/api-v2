
# 🌍 Chivomap Api

## 📌 Descripción
API en **Go** usando **Fiber** para manejar rutas, **Colly** para scraping y **Turso DB** como base de datos.


## 🚀 **Instalación**

### 1️⃣ Clonar el repositorio
```bash
git clone https://github.com/oclazi/chivomap-api.git
cd chivomap-api
```

### 2️⃣ Configurar variables de entorno  
Crear un archivo **`.env`** con:
```ini
TURSO_DATABASE_URL=libsql://your-database.turso.io
TURSO_AUTH_TOKEN=your-auth-token
```

Si `.env` no se carga, usa:
```bash
export $(grep -v '^#' .env | xargs)
```

### 3️⃣ Instalar dependencias
```bash
go mod tidy
```

### 4️⃣ Levantar la API
```bash
go run main.go
```
Disponible en: `http://localhost:8080`

---


## 🛠 **Comandos Útiles**
```bash
go mod tidy  # Instalar dependencias
export $(grep -v '^#' .env | xargs) && go run main.go  # Cargar .env y ejecutar
```
