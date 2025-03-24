-- 000001_create_metric.up.sql

CREATE TABLE IF NOT EXISTS metric (
    time TIMESTAMPTZ NOT NULL,
    service_url TEXT NOT NULL,
    metric_name TEXT NOT NULL,
    pod_name TEXT NOT NULL,
    metric_value DOUBLE PRECISION NOT NULL,
    is_anomaly BOOLEAN NOT NULL DEFAULT false
);

-- Уникальный индекс для метрик по времени, сервису, поду и имени метрики
CREATE UNIQUE INDEX idx_metric_unique_composite ON metric (time DESC, service_url, metric_name, pod_name);

-- Преобразуем таблицу в hypertable с партиционированием по времени
-- Устанавливаем маленький интервал чанка (1 час), так как данные за 5 мин - 3 часа берутся
SELECT create_hypertable('metric', 'time', chunk_time_interval => INTERVAL '1 hour', if_not_exists => true);

-- Включаем сжатие для старых данных
ALTER TABLE metric SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'service_url, metric_name',
    timescaledb.compress_orderby = 'time DESC'
);

-- Настраиваем политику сжатия: сжимаем данные старше 12 часов
SELECT add_compression_policy('metric', INTERVAL '12 hours');

-- Настраиваем политику хранения: удаляем данные старше 7 дней
SELECT add_retention_policy('metric', INTERVAL '7 days');
