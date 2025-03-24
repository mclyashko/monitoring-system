-- 000001_create_order.down.sql

-- Удаляем индексы
DROP INDEX IF EXISTS idx_order_product_id;
DROP INDEX IF EXISTS idx_order_user_id;

-- Удаляем таблицу
DROP TABLE IF EXISTS "order";
