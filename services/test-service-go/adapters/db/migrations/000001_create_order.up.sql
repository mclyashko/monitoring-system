-- 000001_create_order.up.sql

CREATE TABLE "order" (
    id SERIAL PRIMARY KEY,
    product_id INT NOT NULL,
    quantity INT NOT NULL,
    user_id INT NOT NULL
);

-- Добавление индексов для ускорения поиска
CREATE INDEX idx_order_product_id ON "order" (product_id);
CREATE INDEX idx_order_user_id ON "order" (user_id);
