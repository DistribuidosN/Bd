-- En PostgreSQL no se usa "USE database_name;". 
-- Debes conectarte a la base de datos primero (ej. usando \c image_processing_db en psql)
-- CREATE DATABASE image_processing_db;

-- ==========================================================
-- 1. TABLAS DE CATÁLOGO (LOOKUP TABLES)
-- ==========================================================

CREATE TABLE node_status (
    id SERIAL PRIMARY KEY,
    name VARCHAR(20) NOT NULL UNIQUE,
    description VARCHAR(100)
);
COMMENT ON TABLE node_status IS 'Catálogo de estados posibles para los nodos trabajadores';
COMMENT ON COLUMN node_status.name IS 'Nombre del estado (ACTIVE, INACTIVE, ERROR)';
COMMENT ON COLUMN node_status.description IS 'Descripción detallada del estado del nodo';

CREATE TABLE batch_status (
    id SERIAL PRIMARY KEY,
    name VARCHAR(20) NOT NULL UNIQUE,
    description VARCHAR(100)
);
COMMENT ON TABLE batch_status IS 'Catálogo de estados para los lotes de trabajo';
COMMENT ON COLUMN batch_status.name IS 'Nombre del estado (PENDING, PROCESSING, COMPLETED, FAILED)';
COMMENT ON COLUMN batch_status.description IS 'Descripción del estado global de la petición';

CREATE TABLE image_status (
    id SERIAL PRIMARY KEY,
    name VARCHAR(20) NOT NULL UNIQUE,
    description VARCHAR(100)
);
COMMENT ON TABLE image_status IS 'Catálogo de estados para el ciclo de vida de cada imagen';
COMMENT ON COLUMN image_status.name IS 'Nombre del estado (RECEIVED, PROCESSING, CONVERTED, FAILED)';
COMMENT ON COLUMN image_status.description IS 'Descripción del progreso individual de la imagen';

CREATE TABLE transformation_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    price DECIMAL(10,2) NOT NULL,
    description VARCHAR(100)
);
COMMENT ON TABLE transformation_types IS 'Catálogo de transformaciones soportadas por los nodos Python';
COMMENT ON COLUMN transformation_types.name IS 'Nombre técnico de la transformación';
COMMENT ON COLUMN transformation_types.price IS 'Precio de la transformación';
COMMENT ON COLUMN transformation_types.description IS 'Explicación de la operación (ej. Escala de grises, Rotar)';

CREATE TABLE log_levels (
    id SERIAL PRIMARY KEY,
    name VARCHAR(15) NOT NULL UNIQUE,
    description VARCHAR(100)
);
COMMENT ON TABLE log_levels IS 'Catálogo de niveles de severidad para el sistema de logs';
COMMENT ON COLUMN log_levels.name IS 'Severidad (INFO, WARNING, ERROR)';

-- ==========================================================
-- 2. TABLAS PRINCIPALES DEL SISTEMA
-- ==========================================================

CREATE TABLE nodes (
    id SERIAL PRIMARY KEY,
    node_id VARCHAR(100) NOT NULL UNIQUE,
    host VARCHAR(255) NOT NULL,
    port INT NOT NULL,
    status_id INT NOT NULL,
    last_signal TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (status_id) REFERENCES node_status(id)
);
COMMENT ON TABLE nodes IS 'Registro y monitoreo de los nodos trabajadores';
COMMENT ON COLUMN nodes.id IS 'Identificador interno del registro';
COMMENT ON COLUMN nodes.node_id IS 'ID descriptivo del nodo';
COMMENT ON COLUMN nodes.host IS 'Dirección IP o dominio';
COMMENT ON COLUMN nodes.port IS 'Puerto gRPC';
COMMENT ON COLUMN nodes.status_id IS 'FK a node_statuses';
COMMENT ON COLUMN nodes.last_signal IS 'Monitoreo de actividad';

-- Función y Trigger para replicar el "ON UPDATE CURRENT_TIMESTAMP" de MySQL
CREATE OR REPLACE FUNCTION update_last_signal_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.last_signal = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_nodes_last_signal
BEFORE UPDATE ON nodes
FOR EACH ROW
EXECUTE FUNCTION update_last_signal_column();

CREATE TABLE batches (
    batch_uuid CHAR(36) NOT NULL PRIMARY KEY,
    user_uuid CHAR(36) NOT NULL,
    request_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status_id INT NOT NULL,
    FOREIGN KEY (status_id) REFERENCES batch_status(id)
);
COMMENT ON TABLE batches IS 'Cabecera de las solicitudes de procesamiento';
COMMENT ON COLUMN batches.batch_uuid IS 'ID del lote';
COMMENT ON COLUMN batches.user_uuid IS 'Referencia lógica a auth_db';
COMMENT ON COLUMN batches.request_time IS 'Fecha de recepción';
COMMENT ON COLUMN batches.status_id IS 'FK a batch_statuses';

CREATE TABLE images (
    image_uuid CHAR(36) NOT NULL PRIMARY KEY,
    batch_uuid CHAR(36) NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    result_path VARCHAR(500),
    status_id INT NOT NULL,
    node_id INT,
    reception_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    conversion_time TIMESTAMP NULL,
    FOREIGN KEY (batch_uuid) REFERENCES batches(batch_uuid),
    FOREIGN KEY (status_id) REFERENCES image_status(id),
    FOREIGN KEY (node_id) REFERENCES nodes(id)
);
COMMENT ON TABLE images IS 'Trazabilidad detallada de conversión por imagen';
COMMENT ON COLUMN images.image_uuid IS 'ID de la imagen';
COMMENT ON COLUMN images.batch_uuid IS 'Relación con el lote';
COMMENT ON COLUMN images.original_name IS 'Nombre del archivo original';
COMMENT ON COLUMN images.result_path IS 'Ruta del resultado final';
COMMENT ON COLUMN images.status_id IS 'FK a image_statuses';
COMMENT ON COLUMN images.node_id IS 'Nodo asignado para el trabajo';
COMMENT ON COLUMN images.conversion_time IS 'Fecha/hora de finalización';

CREATE TABLE image_transformations (
    id SERIAL PRIMARY KEY,
    image_uuid CHAR(36) NOT NULL,
    type_id INT NOT NULL,
    params JSONB,
    FOREIGN KEY (image_uuid) REFERENCES images(image_uuid),
    FOREIGN KEY (type_id) REFERENCES transformation_types(id)
);
COMMENT ON TABLE image_transformations IS 'Pipeline de transformaciones requeridas por imagen';
COMMENT ON COLUMN image_transformations.type_id IS 'FK a transformation_types';
COMMENT ON COLUMN image_transformations.params IS 'Parámetros técnicos en JSONB';

CREATE TABLE processing_logs (
    id SERIAL PRIMARY KEY,
    node_id INT NOT NULL,
    image_uuid CHAR(36) NOT NULL,
    level_id INT NOT NULL,
    message TEXT NOT NULL,
    log_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (node_id) REFERENCES nodes(id),
    FOREIGN KEY (image_uuid) REFERENCES images(image_uuid),
    FOREIGN KEY (level_id) REFERENCES log_levels(id)
);
COMMENT ON TABLE processing_logs IS 'Registro centralizado de eventos y errores distribuidos';
COMMENT ON COLUMN processing_logs.level_id IS 'FK a log_levels';

-- ==========================================================
-- 3. TABLAS DE TELEMETRÍA Y MONITOREO (PROTOBUF)
-- ==========================================================

CREATE TABLE node_metrics (
    id BIGSERIAL PRIMARY KEY,
    node_id INT NOT NULL,
    image_uuid CHAR(36) NOT NULL,
    ram_used_mb DECIMAL(10,2) NOT NULL,
    ram_total_mb DECIMAL(10,2) NOT NULL,
    cpu_percent DECIMAL(5,2) NOT NULL,
    workers_busy INT NOT NULL,
    workers_total INT NOT NULL,
    queue_size INT NOT NULL,
    queue_capacity INT NOT NULL,
    tasks_done INT NOT NULL,
    steals_performed INT NOT NULL,
    avg_latency_ms DECIMAL(10,2),
    p95_latency_ms DECIMAL(10,2),
    uptime_seconds BIGINT NOT NULL,
    status_id INT NOT NULL,
    reported_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (node_id) REFERENCES nodes(id),
    FOREIGN KEY (status_id) REFERENCES node_status(id),
    FOREIGN KEY (image_uuid) REFERENCES images(image_uuid)
);
COMMENT ON TABLE node_metrics IS 'Historial de telemetría y salud reportado por el mensaje NodeMetrics';
COMMENT ON COLUMN node_metrics.id IS 'BIGSERIAL porque las métricas crecen muy rápido';
COMMENT ON COLUMN node_metrics.node_id IS 'FK a la tabla nodes (identificador relacional)';
COMMENT ON COLUMN node_metrics.image_uuid IS 'FK a la tabla images (identificador relacional)';
COMMENT ON COLUMN node_metrics.ram_used_mb IS 'Memoria RAM utilizada en MB';
COMMENT ON COLUMN node_metrics.ram_total_mb IS 'Memoria RAM total en MB';
COMMENT ON COLUMN node_metrics.cpu_percent IS 'Porcentaje de uso de CPU (0 a 100)';
COMMENT ON COLUMN node_metrics.workers_busy IS 'Cantidad de workers actualmente ocupados';
COMMENT ON COLUMN node_metrics.workers_total IS 'Total de workers';
COMMENT ON COLUMN node_metrics.queue_size IS 'Tareas en cola local ahora mismo';
COMMENT ON COLUMN node_metrics.queue_capacity IS 'Capacidad de la cola';
COMMENT ON COLUMN node_metrics.tasks_done IS 'Acumulado de tareas finalizadas desde el arranque';
COMMENT ON COLUMN node_metrics.steals_performed IS 'Work-steals realizados a otros nodos';
COMMENT ON COLUMN node_metrics.avg_latency_ms IS 'Media de latencia de las últimas 100 tareas en ms';
COMMENT ON COLUMN node_metrics.p95_latency_ms IS 'Percentil 95 de latencia en ms';
COMMENT ON COLUMN node_metrics.uptime_seconds IS 'Tiempo de actividad del nodo en segundos';
COMMENT ON COLUMN node_metrics.status_id IS 'FK a node_status (IDLE, BUSY, STEALING, ERROR)';
COMMENT ON COLUMN node_metrics.reported_at IS 'Timestamp exacto en que se generó la métrica';