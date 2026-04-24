-- 002_create_purchases.sql — Migration kedua: buat table untuk track pembelian user.
-- Table ni simpan rekod setiap kali user beli item — penting untuk:
--   1. Semak limited item (dancing-zombie hanya boleh beli sekali)
--   2. Frontend tahu item mana yang user dah owned

-- Buat table 'purchases' kalau belum ada
CREATE TABLE IF NOT EXISTS purchases (
    id         SERIAL PRIMARY KEY,                           -- ID auto-increment
    user_id    VARCHAR(255) NOT NULL,                        -- ID user dari Xsolla (field "sub" dalam JWT)
    item_sku   VARCHAR(50)  NOT NULL,                        -- SKU item yang dibeli (contoh: "pvz-wall-nut")
    created_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP, -- Bila pembelian berlaku
    UNIQUE (user_id, item_sku)                               -- Satu user tak boleh ada duplicate purchase untuk item yang sama
);
