
### Requisitos

- **Docker instalado.**
- Archivo `.env` con las variables:
  - `TURSO_DATABASE_URL`
  - `TURSO_AUTH_TOKEN`
- Directorio `utils/assets` (con `topo.json`) incluido en el proyecto.

---

### Build

1. Desde la raíz del proyecto, construir la imagen:
   ```bash
   sudo docker build -t chivomap-api-v2 .
   ```

---

### Run

1. Ejecutar el contenedor inyectando las variables de entorno (o asegurándote de que la aplicación cargue el archivo `.env`):
   ```bash
   sudo docker run -it -p 8080:8080 --env-file .env chivomap-api-v2
   ```
   o, pasando las variables manualmente:
   ```bash
   sudo docker run -it -p 8080:8080 --env-file .env chivomap-api-v2
   ```

---

### Notas

- El Dockerfile copia el directorio `utils/assets` para que el JSON esté disponible.
- La imagen usa CGO activado y se basa en `debian:bookworm-slim` para compatibilidad con GLIBC.

Con estos pasos, deberías poder compilar y ejecutar la aplicación tanto en local como en producción.