-- 001_create_items.sql — Migration pertama: buat table untuk simpan item dalam store.
-- Jalankan file ni sekali untuk setup database sebelum start server.

-- Buat table 'items' kalau belum ada
-- Setiap row = satu item yang boleh dibeli dalam game store
CREATE TABLE IF NOT EXISTS items (
    id          SERIAL PRIMARY KEY,        -- ID auto-increment (1, 2, 3, ...)
    name        VARCHAR(100) NOT NULL,     -- Nama item (contoh: "Wall-nut")
    description TEXT NOT NULL,            -- Penerangan panjang item
    price       INTEGER NOT NULL,         -- Harga dalam unit matawang (contoh: 5 = USD 5)
    currency    VARCHAR(10) NOT NULL,     -- Matawang (contoh: "USD")
    sku         VARCHAR(50) NOT NULL UNIQUE, -- SKU mesti unik — ID untuk Xsolla (contoh: "pvz-wall-nut")
    image_url   TEXT,                     -- URL gambar item (optional)
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP -- Tarikh/masa item ditambah
);

-- Insert data awal (seed data) — 4 item Plants vs Zombies
-- ON CONFLICT akan skip kalau SKU dah ada — selamat untuk run berkali-kali
INSERT INTO items (name, description, price, currency, sku, image_url)
VALUES
(
    'Wall-nut',
    'Wall-nuts have hard shells which you can use to protect your other plants. Toughness: high. "People wonder how I feel about getting constantly chewed on by zombies," says Wall-nut. "What they do not realize is that with my limited senses all I can feel is a kind of tingling, like a relaxing back rub."',
    5,
    'USD',
    'pvz-wall-nut',
    'https://static.wikia.nocookie.net/plantsvszombies/images/6/6b/WallNutPVZ.webp'
),
(
    'Torchwood',
    'Torchwoods turn peas that pass through them into fireballs that deal twice as much damage. Special: doubles the damage of peas that pass through it. Fireballs deal damage to nearby zombies on impact. Everybody likes and respects Torchwood. They like him for his integrity, for his steadfast friendship, for his ability to greatly maximize pea damage. But Torchwood has a secret: he cannot read.',
    8,
    'USD',
    'pvz-torchwood',
    'https://static.wikia.nocookie.net/plantsvszombies/images/7/72/Torchwood1.webp'
),
(
    'Snow Pea',
    'Snow Peas shoot frozen peas that damage and slow the enemy. Damage: normal, slows zombies. Folks often tell Snow Pea how "cool" he is, or exhort him to "chill out." They tell him to "stay frosty." Snow Pea just rolls his eyes. He has heard them all.',
    7,
    'USD',
    'pvz-snow-pea',
    'https://static.wikia.nocookie.net/plantsvszombies/images/9/96/Snowpea1.webp'
),
(
    'Cherry Bomb',
    'Cherry Bombs can blow up all zombies in an area. They have a short fuse so plant them near zombies. Damage: massive. Range: all zombies in a medium area. Usage: single use, instant. "I wanna explode," says Cherry #1. "No, lets detonate instead!" says his brother, Cherry #2. After intense consultation they agree to explodonate.',
    10,
    'USD',
    'pvz-cherry-bomb',
    'https://static.wikia.nocookie.net/plantsvszombies/images/8/81/Cherrybomb1.webp'
),
(
    'Dancing Zombie',
    'A rare limited-edition zombie dancer. Can only be owned once. This groovy zombie moonwalks through your defenses while summoning backup dancers. Limited to one per player — once he is gone, he is gone forever.',
    15,
    'USD',
    'dancing-zombie',
    'https://static.wikia.nocookie.net/plantsvszombies/images/8/8f/Dancing_Zombie.png'
);
