-- 000001_create_metric.down.sql

-- Удаляем политику хранения данных (если существует)
SELECT remove_retention_policy('metric', if_exists => true);

-- Удаляем политику сжатия (если существует)
SELECT remove_compression_policy('metric', if_exists => true);

-- Отключаем сжатие и сбрасываем настройки
ALTER TABLE metric RESET (
    timescaledb.compress,
    timescaledb.compress_segmentby,
    timescaledb.compress_orderby
);

-- Удаляем уникальный индекс
DROP INDEX IF EXISTS idx_metric_unique_composite;

-- Удаляем таблицу (включая гипертаблицу и чанки)
DROP TABLE IF EXISTS metric;
