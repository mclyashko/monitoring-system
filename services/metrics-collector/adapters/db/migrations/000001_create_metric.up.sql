-- 000001_create_metric.up.sql

CREATE TABLE IF NOT EXISTS metric (
    id BIGSERIAL PRIMARY KEY,
    service_url TEXT NOT NULL,
    metric_name TEXT NOT NULL,
    metric_value DOUBLE PRECISION NOT NULL,
    time TIMESTAMPTZ NOT NULL,
    is_anomaly BOOLEAN NOT NULL DEFAULT false
);

-- Преобразуем таблицу в hypertable с партиционированием по времени
-- Устанавливаем маленький интервал чанка (например, 1 час), так как данные за 5 мин - 3 часа берутся
SELECT create_hypertable('metric', 'time', chunk_time_interval => INTERVAL '1 hour');

-- Включаем сжатие для старых данных
ALTER TABLE metric SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'service_url',
    timescaledb.compress_orderby = 'time DESC'
);

-- Настраиваем политику сжатия: сжимаем данные старше 12 часов
SELECT add_compression_policy('metric', INTERVAL '12 hours');

-- Настраиваем политику удаления: удаляем данные старше 7 дней
SELECT add_drop_chunks_policy('metric', INTERVAL '7 days');

-- Составной индекс для оптимизации выборок метрик по сервису, названию, времени и флагу аномальности
CREATE INDEX idx_metric_composite ON metric (time DESC, service_url, metric_name, is_anomaly);
