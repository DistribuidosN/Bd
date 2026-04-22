# Enfok BD — Backend API de Procesamiento Distribuido de Imágenes

**Enfok BD** es la API REST central del sistema distribuido Enfok. Actúa como el cerebro de coordinación: registra imágenes, gestiona nodos de trabajo, almacena logs y métricas, y persiste el estado de todo el pipeline de procesamiento.

---

## ¿Qué hace este servicio?

Cuando un cliente sube una imagen al sistema, este backend:

1. **Crea un Batch** — agrupa una o más imágenes en una unidad de trabajo.
2. **Sube la imagen a MinIO** — almacenamiento de objetos (S3-compatible).
3. **Registra la imagen en PostgreSQL** — con estado inicial `RECEIVED`.
4. **Los nodos workers** consultan la BD, toman imágenes y actualizan su estado a `PROCESSING` → `CONVERTED` / `FAILED`.
5. **Cada nodo** envía heartbeats periódicos, reporta sus métricas de recursos (CPU, RAM, workers) y registra logs de cada operación.

Este servicio **no procesa imágenes** — solo coordina, persiste y expone el estado del sistema.

---

## Arquitectura — Hexagonal (Ports & Adapters)

El proyecto sigue estrictamente la **Arquitectura Hexagonal**, separando el negocio de la infraestructura mediante interfaces (ports).

```
┌─────────────────────────────────────────────────────────────┐
│              API REST (Driving Side — Entrada)               │
│           src/handlers/http/  →  src/routes/                │
└──────────────────────┬──────────────────────────────────────┘
                       │ usa driving ports
┌──────────────────────▼──────────────────────────────────────┐
│           domain/ports/driving/  (contratos de entrada)      │
│   ImageServicePort, NodeServicePort, LogServicePort...       │
└──────────────────────┬──────────────────────────────────────┘
                       │ implementados por
┌──────────────────────▼──────────────────────────────────────┐
│                    src/service/                              │
│       ImageService, NodeService, LogService...               │
└──────────────────────┬──────────────────────────────────────┘
                       │ usa driven ports
┌──────────────────────▼──────────────────────────────────────┐
│           domain/ports/driven/  (contratos de salida)        │
│   ImageRepository, NodeRepository, StorageRepository...      │
└──────────────────────┬──────────────────────────────────────┘
                       │ implementados por
┌──────────────────────▼──────────────────────────────────────┐
│             src/infrastructure/repository/                   │
│       PostgreSQL (sqlx + pgx) · MinIO (minio-go)            │
└─────────────────────────────────────────────────────────────┘
```

### Flujo de datos

```
Cliente HTTP
    │
    ▼
handlers/http/        ← Valida request, llama al service via driving port
    │
    ▼
service/              ← Lógica de negocio pura (orquesta el caso de uso)
    │
    ▼
domain/ports/driven/  ← Interfaz: el service no sabe cómo se persiste
    │
    ▼
infrastructure/
  ├── repository/     ← Implementa los ports: queries SQL + MinIO
  └── dto/            ← Structs que mapean 1:1 con las tablas de BD

utils/status_mapper.go ← Convierte "PENDING" ↔ 1, "ACTIVE" ↔ 1, etc.
```

---

## Estructura de Directorios

```
src/
├── main.go                          # Bootstrap: carga config, conecta BD, inyecta dependencias
│
├── domain/                          # Núcleo del negocio — sin dependencias externas
│   ├── image/
│   │   ├── image.go                 # Entidad Image (estado como string: "RECEIVED", "CONVERTED"...)
│   │   ├── batch.go                 # Entidad Batch (agrupa imágenes)
│   │   └── transformation.go        # Entidades Transformation e ImageTransformation
│   ├── node/
│   │   └── node.go                  # Entidad Node (worker registrado: host, puerto, estado)
│   ├── logs/
│   │   └── log.go                   # Entidad ProcessingLog (nivel: INFO/WARNING/ERROR)
│   ├── metrics/
│   │   └── metric.go                # Entidad NodeMetrics (CPU, RAM, workers, latencia...)
│   └── ports/
│       ├── driving/                 # Contratos de ENTRADA (API → Servicios)
│       │   ├── image_service_port.go
│       │   ├── node_service_port.go
│       │   ├── log_service_port.go
│       │   └── metric_service_port.go
│       └── driven/                  # Contratos de SALIDA (Servicios → Repositorios)
│           ├── image_repository.go
│           ├── node_repository.go
│           ├── log_repository.go
│           ├── metric_repository.go
│           └── storage_repository.go
│
├── handlers/http/                   # Adaptadores de entrada — controladores Gin
│   ├── image_handler.go
│   ├── batch_handler.go
│   ├── node_handler.go
│   ├── log_handler.go
│   └── metrics_handler.go
│
├── routes/
│   └── routes.go                    # Registro centralizado de todas las rutas API v1
│
├── service/                         # Casos de uso — lógica de negocio pura
│   ├── image_service.go
│   ├── batch_service.go
│   ├── node_service.go
│   ├── log_service.go
│   └── metrics_service.go
│
├── infrastructure/                  # Adaptadores de salida — detalles técnicos
│   ├── config/
│   │   └── config.go                # Carga variables de entorno (panic si falta alguna crítica)
│   ├── db/
│   │   └── db.go                    # Inicializa pool PostgreSQL (sqlx/pgx) y cliente MinIO
│   ├── dto/                         # Data Transfer Objects — mapeo 1:1 con tablas de BD
│   │   ├── image_dto.go             # ImageDTO, BatchDTO + constantes ImageStatusReceived=1...
│   │   ├── node_dto.go              # NodeDTO + NodeStatusActive=1, NodeStatusInactive=2...
│   │   ├── log_dto.go               # LogDTO + LogLevelInfo=1, LogLevelWarning=2...
│   │   ├── metric_dto.go            # MetricDTO
│   │   └── transformation_dto.go    # TransformationDTO, ImageTransformationDTO
│   └── repository/                  # Implementaciones concretas de los driven ports
│       ├── postgres_image.go        # Image + Batch: INSERT, SELECT, UPDATE status
│       ├── postgres_node.go         # Node: registro, heartbeat, listado, UPSERT
│       ├── postgres_log.go          # ProcessingLog: INSERT y SELECT por imagen
│       ├── postgres_metric.go       # NodeMetrics: INSERT y SELECT por nodo
│       └── minio_storage.go         # Upload de objetos + URLs presignadas
│
└── utils/
    └── status_mapper.go             # GetIDFromStatus / GetStatusFromID + catálogos de estados
```

---

## API REST — Endpoints

| Método | Ruta | Descripción |
|--------|------|-------------|
| `GET` | `/health` | Health check del servicio |
| **Imágenes** | | |
| `POST` | `/api/v1/images` | Sube una imagen (form-data: `image` + `user_uuid`) |
| `GET` | `/api/v1/images/:id` | Consulta una imagen por UUID |
| `PATCH` | `/api/v1/images/:id/status` | Actualiza estado de imagen |
| **Batches** | | |
| `GET` | `/api/v1/batches/:id` | Consulta un batch por UUID |
| `PATCH` | `/api/v1/batches/:id/status` | Actualiza estado de batch |
| **Nodos** | | |
| `POST` | `/api/v1/nodes` | Registra un nodo worker |
| `POST` | `/api/v1/nodes/:node_id/heartbeat` | Señal de vida del nodo |
| `GET` | `/api/v1/nodes` | Lista todos los nodos registrados |
| **Logs** | | |
| `POST` | `/api/v1/logs` | Registra un log de procesamiento |
| `GET` | `/api/v1/logs/:image_uuid` | Obtiene logs de una imagen |
| **Métricas** | | |
| `POST` | `/api/v1/metrics` | Registra métricas de un nodo |
| `GET` | `/api/v1/metrics/:node_id` | Obtiene métricas de un nodo |

---

## Estados del Sistema

Los estados se almacenan como `int` en BD y se mapean a `string` en el dominio vía `utils/status_mapper.go`.

| Entidad | Estado | ID en BD |
|---------|--------|----------|
| **Image** | `RECEIVED` | 1 |
| | `PROCESSING` | 2 |
| | `CONVERTED` | 3 |
| | `FAILED` | 4 |
| **Batch** | `PENDING` | 1 |
| | `PROCESSING` | 2 |
| | `COMPLETED` | 3 |
| | `FAILED` | 4 |
| **Node** | `ACTIVE` | 1 |
| | `INACTIVE` | 2 |
| | `ERROR` | 3 |
| **Log Level** | `INFO` | 1 |
| | `WARNING` | 2 |
| | `ERROR` | 3 |

---

## Configuración — Variables de Entorno

Crear un archivo `.env` en la raíz del proyecto (o exportar las variables):

```env
# Servidor
SERVER_PORT=8080
SERVER_HOST=localhost

# PostgreSQL
DB_USER=postgres
DB_PASSWORD=secret
DB_HOST=localhost
DB_PORT=5432
DB_NAME=enfok

# MinIO
MINIO_URL=localhost:9000
MINIO_USER=minioadmin
MINIO_PWD=minioadmin
MINIO_BUCKET=enfok-images
MINIO_SSL=false
```

> ⚠️ Si alguna variable marcada como requerida está ausente, el servidor **no arrancará** (panic en startup — falla rápida en lugar de error silencioso).

---

## Ejecutar el Proyecto

```bash
# 1. Instalar dependencias
go mod tidy

# 2. Ejecutar
go run ./src/main.go

# 3. O compilar primero
go build -o enfok_bd ./src/main.go
./enfok_bd
```

---

## Dependencias Principales

| Librería | Uso |
|----------|-----|
| `gin-gonic/gin` | Framework HTTP / Router |
| `jmoiron/sqlx` | SQL tipado sobre `database/sql` |
| `jackc/pgx/v5` | Driver PostgreSQL de alto rendimiento |
| `minio/minio-go/v7` | Cliente MinIO / S3 |
| `joho/godotenv` | Carga de archivos `.env` |
| `google/uuid` | Generación de UUIDs para imágenes y batches |
| `log/slog` | Logging estructurado (JSON en producción) |

---

## Diseño de la Separación DTO / Dominio

Un principio clave de este proyecto es que **los objetos de la base de datos nunca salen del repositorio**:

```
BD (int status_id=1)
       │
   postgres_image.go  ←  convierte usando utils.GetStatusFromID
       │
   domain/image.Image  (string Status="RECEIVED")
       │
   handlers/http       ←  responde directamente el modelo de dominio
       │
   Cliente HTTP (json: "status": "RECEIVED")
```

Esto garantiza que un cambio en el esquema de BD solo afecta los DTOs y repositorios, nunca la lógica de negocio ni los contratos del API.