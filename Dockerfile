# Etapa de compilación: se usa una imagen de Go en linux/amd64
FROM --platform=linux/amd64 golang:1.24 AS builder

# Configuramos las variables de entorno necesarias para CGO
ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /app

# Copiamos los archivos de módulos y descargamos dependencias
COPY go.mod go.sum ./
RUN go mod download

# Copiamos el resto del código fuente y el directorio de assets
COPY . .
COPY utils/assets /app/utils/assets

# Compilamos el binario
RUN go build -o out .

# Etapa final: se usa una imagen runtime con glibc actualizada (Debian Bookworm)
FROM --platform=linux/amd64 debian:bookworm-slim

# Instalamos certificados y otras librerías necesarias
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copiamos el binario compilado desde la etapa anterior
COPY --from=builder /app/out .

# Copiamos el directorio de assets desde la etapa builder
COPY --from=builder /app/utils/assets /app/utils/assets

# Exponemos el puerto (modifica si tu app usa otro)
EXPOSE 8080

# Comando para ejecutar la aplicación
CMD ["./out"]
