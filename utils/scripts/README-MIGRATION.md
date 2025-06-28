# Script de Migración de Datos del Censo 2024 SV

Este script permite migrar los datos del Censo de Población y Vivienda 2024 de El Salvador desde archivos CSV a una base de datos Turso.

## Requisitos previos

- Go 1.21 o superior
- Variables de entorno configuradas en el archivo `.env`:
  - `TURSO_DATABASE_URL_CENSO`: URL de la base de datos Turso para el censo
  - `TURSO_AUTH_TOKEN_CENSO`: Token de autenticación para la base de datos del censo

## Instalación de dependencias

El script utiliza las siguientes dependencias que ya deberían estar en el proyecto:

```bash
go get github.com/joho/godotenv
go get github.com/schollz/progressbar/v3
```

## Uso

### Migrar todos los archivos CSV de un directorio

Para migrar todos los archivos CSV de un directorio, ejecuta:

```bash
go run utils/scripts/censo_migration.go -dir="/path/to/csv/directory"
```

Por ejemplo, para migrar los archivos de la carpeta assets:

```bash
go run utils/scripts/censo_migration.go -dir="utils/assets/Bases-Finales-CPV2024SV-CSV"
```

### Migrar un solo archivo CSV

Para migrar un solo archivo CSV, ejecuta:

```bash
go run utils/scripts/censo_migration.go -file="/path/to/file.csv" -table="nombre_tabla"
```

Por ejemplo:

```bash
go run utils/scripts/censo_migration.go -file="utils/assets/Bases-Finales-CPV2024SV-CSV/Base de Datos de Población - CPV 2024 SV.csv" -table="censo_poblacion"
```

### Solo crear las tablas sin importar datos

Para crear las tablas sin importar los datos (útil para pruebas o para verificar la estructura):

```bash
go run utils/scripts/censo_migration.go -dir="/path/to/csv/directory" -create-tables
```

### Saltar tablas que ya existen

Para omitir la migración de tablas que ya existen en la base de datos:

```bash
go run utils/scripts/censo_migration.go -dir="/path/to/csv/directory" -skip-existing
```

## Estructura de las tablas

El script creará las siguientes tablas en la base de datos:

- `censo_poblacion`: Datos de población
- `censo_hogares`: Datos de hogares
- `censo_viviendas`: Datos de viviendas
- `censo_emigracion_internacional`: Datos de emigración internacional
- `censo_mortalidad`: Datos de mortalidad

Cada tabla contendrá:
- Una columna `id` como clave primaria autoincremental
- Columnas que corresponden a los encabezados de los archivos CSV, normalizados (convertidos a minúsculas, espacios reemplazados por guiones bajos, etc.)
- Si los encabezados son numéricos o están vacíos, se generan nombres de columna automáticamente (`col_1`, `col_2`, etc.)

## Consideraciones

- Los archivos CSV son grandes (especialmente el de población que es de 1.5GB), por lo que el proceso puede tardar bastante tiempo.
- La migración se realiza fila por fila para garantizar la compatibilidad con Turso.
- Se recomienda ejecutar el script en un entorno con suficiente memoria RAM y espacio de almacenamiento.
- El script muestra una barra de progreso para cada archivo que está procesando.
- Las tablas existentes se recrean por defecto, a menos que se use la opción `-skip-existing`.

## Solución de problemas

Si encuentras errores durante la migración, verifica lo siguiente:

1. Que las variables de entorno estén correctamente configuradas en el archivo `.env`
2. Que los permisos de acceso a los archivos CSV sean correctos
3. Que la base de datos Turso esté disponible y accesible
4. Que haya suficiente espacio disponible en la base de datos Turso

## Limitaciones de Turso

Ten en cuenta que Turso puede tener limitaciones en cuanto al tamaño de la base de datos y el número de transacciones. Verifica los límites de tu plan antes de iniciar la migración completa. 