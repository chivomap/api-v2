#!/usr/bin/env python3
import os
import csv
import sys
import time
import argparse
from pathlib import Path
from dotenv import load_dotenv
import tqdm
import libsql_experimental as libsql

# Ruta por defecto a los archivos CSV del censo
DEFAULT_CSV_PATH = "/home/devel/chivomap/api/utils/assets/Bases-Finales-CPV2024SV-CSV"

# Constantes
BATCH_SIZE = 200  # Tamaño del lote para inserciones
MAX_ERRORS = 100  # Número máximo de errores antes de abortar

# Cargar variables de entorno
load_dotenv()

# Obtener conexión a la base de datos Turso
def connect_db():
    turso_url = 'libsql://censo2024-oclazi.aws-us-east-1.turso.io'
    turso_token = 'eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9.eyJhIjoicnciLCJpYXQiOjE3NDU2ODEwMDcsImlkIjoiYTE2MDFjMzEtNzcwZi00NzdmLWIxOTctNWEyY2M0N2IxZjQ4IiwicmlkIjoiYTU3YjE2ZDUtYTkxMS00MzZmLTllMGItMjljNDBiNWQ5NTA2In0.OsHLprgGoDjazQceQmRd1BJ_I0KuEbH2pllZS1GOUZi-FCDEYZ0yQ6Xs58Uln2uXS9vDYsub9IQLqPp7t7k_Aw'
    
    if not turso_url:
        print("❌ Error: No se encontró la variable de entorno TURSO_DATABASE_URL")
        sys.exit(1)
    
    if not turso_token:
        print("❌ Error: No se encontró la variable de entorno TURSO_AUTH_TOKEN")
        sys.exit(1)
    
    print(f"🔍 Conectando a Turso usando el SDK oficial")
    
    try:
        # Corregido: La sintaxis correcta para libsql_experimental
        conn = libsql.connect(database_url=turso_url, auth_token=turso_token)
        return conn
    except Exception as e:
        print(f"❌ Error al conectar con Turso: {e}")
        sys.exit(1)

# Eliminar acentos y caracteres especiales
def remove_accents(text):
    replacements = {
        'á': 'a', 'é': 'e', 'í': 'i', 'ó': 'o', 'ú': 'u',
        'Á': 'A', 'É': 'E', 'Í': 'I', 'Ó': 'O', 'Ú': 'U',
        'ñ': 'n', 'Ñ': 'N',
        'ü': 'u', 'Ü': 'U',
    }
    
    for old, new in replacements.items():
        text = text.replace(old, new)
    
    return text

# Verificar si una tabla existe
def table_exists(conn, table_name):
    try:
        cursor = conn.execute("SELECT name FROM sqlite_master WHERE type='table' AND name=?", (table_name,))
        result = cursor.fetchone()
        return result is not None
    except Exception as e:
        print(f"⚠️ Error verificando si la tabla {table_name} existe: {e}")
        return False

# Crear tabla con columnas genéricas
def create_generic_table(conn, table_name, columns):
    # Construir sentencia SQL para crear tabla
    sql = f"CREATE TABLE IF NOT EXISTS {table_name} (\n"
    sql += "  id INTEGER PRIMARY KEY AUTOINCREMENT"
    
    for col in columns:
        sql += f",\n  `{col}` TEXT"
    
    sql += "\n);"
    
    try:
        conn.execute(sql)
        conn.commit()
        print(f"✅ Tabla {table_name} creada correctamente con {len(columns)} columnas")
    except Exception as e:
        print(f"❌ Error al crear tabla {table_name}: {e}")
        sys.exit(1)

# Procesar un archivo CSV
def process_csv_file(conn, file_path, table_name, create_tables_only=False, skip_existing=False):
    print(f"📊 Procesando archivo: {os.path.basename(file_path)} -> tabla: {table_name}")
    
    # Verificar si la tabla ya existe
    exists = table_exists(conn, table_name)
    if exists and skip_existing:
        print(f"⏭️ Tabla {table_name} ya existe. Saltando...")
        return
    
    # Obtener información del archivo
    file_size = os.path.getsize(file_path)
    print(f"📏 Tamaño del archivo: {file_size / (1024 * 1024):.2f} MB")
    
    try:
        # Leer encabezados y procesar en una sola pasada
        with open(file_path, 'r', newline='', encoding='utf-8') as csvfile:
            csv_reader = csv.reader(csvfile)
            headers = next(csv_reader)
            
            # Normalizar nombres de columnas
            column_names = []
            for i, header in enumerate(headers):
                header = str(header).strip()
                
                # Si el encabezado está vacío o es numérico, usar col_N
                if not header or header.isdigit():
                    column_names.append(f"col_{i+1}")
                else:
                    # Normalizar el nombre
                    header = header.lower()
                    header = header.replace(" ", "_")
                    header = header.replace("-", "_")
                    header = header.replace(".", "_")
                    header = header.replace("(", "")
                    header = header.replace(")", "")
                    header = remove_accents(header)
                    column_names.append(header)
            
            print(f"🔍 Número de columnas detectadas: {len(column_names)}")
            
            # Recrear tabla si existe o crearla si no existe
            if exists:
                try:
                    conn.execute(f"DROP TABLE IF EXISTS {table_name}")
                    conn.commit()
                    print(f"🔄 Tabla {table_name} eliminada para recrearla")
                except Exception as e:
                    print(f"⚠️ Error al eliminar tabla existente: {e}")
            
            # Crear tabla con columnas
            create_generic_table(conn, table_name, column_names)
            
            if create_tables_only:
                print(f"✅ Tabla {table_name} creada correctamente. Omitiendo importación de datos.")
                return
            
            # Preparar consulta para inserción
            placeholders = ", ".join(["?" for _ in column_names])
            columns_str = ", ".join([f"`{col}`" for col in column_names])
            insert_sql = f"INSERT INTO {table_name} ({columns_str}) VALUES ({placeholders})"
            
            print(f"🔧 SQL Statement: {insert_sql}")
            
            # Procesar datos
            start_time = time.time()
            row_count = 0
            error_count = 0
            
            # Inicializar barra de progreso
            with tqdm.tqdm(total=file_size, unit='B', unit_scale=True, desc=f"Importing {table_name}") as progress_bar:
                batch = []
                last_position = csvfile.tell()
                
                # Procesar filas
                for record in csv_reader:
                    try:
                        # Asegurar que la fila tiene el número correcto de columnas
                        if len(record) != len(column_names):
                            if len(record) < len(column_names):
                                # Completar con valores vacíos
                                record = record + [''] * (len(column_names) - len(record))
                            else:
                                # Truncar
                                record = record[:len(column_names)]
                        
                        # Agregar a lote para inserción por lotes
                        batch.append(record)
                        row_count += 1
                        
                        # Insertar lote cuando alcanza el tamaño de lote
                        if len(batch) >= BATCH_SIZE:
                            try:
                                # Ejecutar consultas en un lote
                                for row in batch:
                                    conn.execute(insert_sql, row)
                                conn.commit()
                            except Exception as e:
                                error_count += 1
                                print(f"⚠️ Error en inserción por lotes: {e}")
                                
                                # Intentar una por una con manejo individual de errores
                                for row in batch:
                                    try:
                                        conn.execute(insert_sql, row)
                                        conn.commit()
                                    except Exception as inner_e:
                                        error_count += 1
                                        if error_count >= MAX_ERRORS:
                                            print(f"❌ Demasiados errores ({error_count}), abortando importación")
                                            return
                            
                            batch = []
                        
                        # Actualizar barra de progreso basado en la posición del archivo
                        current_position = csvfile.tell()
                        progress_bar.update(current_position - last_position)
                        last_position = current_position
                        
                        # Mostrar progreso cada 1000 filas
                        if row_count % 1000 == 0:
                            print(f"📈 Procesadas {row_count} filas")
                            # Sincronizar con la base de datos remota
                            try:
                                conn.sync()
                            except Exception as sync_e:
                                print(f"⚠️ Error al sincronizar con la base de datos remota: {sync_e}")
                            
                    except Exception as e:
                        error_count += 1
                        print(f"⚠️ Error al procesar fila {row_count}: {e}")
                        if error_count >= MAX_ERRORS:
                            print(f"❌ Demasiados errores ({error_count}), abortando importación")
                            return
                
                # Procesar las filas restantes
                if batch:
                    try:
                        # Ejecutar consultas en un lote
                        for row in batch:
                            conn.execute(insert_sql, row)
                        conn.commit()
                    except Exception as e:
                        # Si falla, intentar insertar fila por fila
                        for row in batch:
                            try:
                                conn.execute(insert_sql, row)
                                conn.commit()
                            except Exception as inner_e:
                                error_count += 1
                                if error_count >= MAX_ERRORS:
                                    print(f"❌ Demasiados errores ({error_count}), abortando importación")
                                    return
                    
                    # Sincronizar con la base de datos remota al final
                    try:
                        conn.sync()
                    except Exception as sync_e:
                        print(f"⚠️ Error al sincronizar con la base de datos remota: {sync_e}")
    except Exception as e:
        print(f"❌ Error al procesar archivo: {e}")
        return
        
    duration = time.time() - start_time
    print(f"✅ Importación completada: {table_name} | {row_count} filas | {duration:.2f} segundos | {error_count} errores")

# Migrar un directorio completo
def migrate_directory(conn, dir_path, create_tables_only=False, skip_existing=False):
    print(f"🔍 Escaneando directorio: {dir_path}")
    
    # Comprobar si el directorio existe
    if not os.path.exists(dir_path):
        print(f"❌ Error: El directorio {dir_path} no existe")
        sys.exit(1)
    
    # Obtener archivos CSV en el directorio
    csv_files = []
    for file in os.listdir(dir_path):
        if file.lower().endswith('.csv') and os.path.isfile(os.path.join(dir_path, file)):
            csv_files.append(file)
    
    if not csv_files:
        print(f"❌ No se encontraron archivos CSV en el directorio: {dir_path}")
        return
    
    print(f"🧰 Se encontraron {len(csv_files)} archivos CSV para procesar")
    
    for file_name in csv_files:
        # Extraer nombre de tabla del nombre de archivo
        base_name = os.path.splitext(file_name)[0]
        parts = base_name.split(" - ")
        
        if len(parts) < 2:
            # Usar el nombre del archivo sin extensión como nombre de tabla
            table_name_part = base_name.lower().replace(" ", "_")
            table_name_part = remove_accents(table_name_part)
        else:
            entity_type = parts[0]
            table_name_part = entity_type.replace("Base de Datos de ", "").lower().replace(" ", "_")
            table_name_part = remove_accents(table_name_part)
        
        table_name = f"censo_{table_name_part}"
        
        file_path = os.path.join(dir_path, file_name)
        process_csv_file(conn, file_path, table_name, create_tables_only, skip_existing)

def main():
    parser = argparse.ArgumentParser(description='Migrar datos del Censo 2024 SV desde archivos CSV a base de datos Turso.')
    parser.add_argument('-dir', '--directory', default=DEFAULT_CSV_PATH, help=f'Directorio que contiene archivos CSV (por defecto: {DEFAULT_CSV_PATH})')
    parser.add_argument('-file', '--file', help='Ruta a un archivo CSV individual (alternativa a -dir)')
    parser.add_argument('-table', '--table', help='Nombre de tabla para migración de un solo archivo (requerido cuando se usa -file)')
    parser.add_argument('-create-tables', '--create-tables-only', action='store_true', help='Solo crear tablas sin importar datos')
    parser.add_argument('-skip-existing', '--skip-existing', action='store_true', help='Saltar tablas que ya existen')
    
    args = parser.parse_args()
    
    # Validar argumentos
    if args.file and not args.table:
        print("❌ Error: -table debe especificarse cuando se usa -file")
        sys.exit(1)
    
    # Conectar a la base de datos Turso
    conn = connect_db()
    
    print("🚀 Iniciando migración a base de datos Turso...")
    
    try:
        # Procesar según los argumentos
        if args.file:
            # Migración de un solo archivo
            process_csv_file(conn, args.file, args.table, args.create_tables_only, args.skip_existing)
        else:
            # Migración de directorio
            migrate_directory(conn, args.directory, args.create_tables_only, args.skip_existing)
            
        print("✅ Migración completada correctamente")
    except Exception as e:
        print(f"❌ Error durante la migración: {e}")
    finally:
        # Cerrar conexión
        conn.close()

if __name__ == "__main__":
    main()