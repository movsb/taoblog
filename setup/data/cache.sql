BEGIN;

CREATE TABLE IF NOT EXISTS `cache` (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    `hash` INTEGER NOT NULL,
    `created_at` INTEGER NOT NULL,
    `expiring_at` INTEGER NOT NULL,
    `key` BLOB NOT NULL,
    `data` BLOB NOT NULL
);

CREATE UNIQUE INDEX `hash__key` ON `cache` (`hash`,`key`);

COMMIT;
