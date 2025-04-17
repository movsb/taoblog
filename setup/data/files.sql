BEGIN;

CREATE TABLE IF NOT EXISTS `files` (
    `id` INTEGER  PRIMARY KEY AUTOINCREMENT,
    `created_at` INTEGER NOT NULL,
    `updated_at` INTEGER NOT NULL,
    `post_id` INTEGER NOT NULL,
    `path` TEXT NOT NULL,
    `mode` INTEGER NOT NULL,
    `mod_time` INTEGER  NOT NULL,
    `size` INTEGER  NOT NULL,
    `meta` BLOB NOT NULL,
    `data` BLOB NOT NULL
);

CREATE UNIQUE INDEX `post_id__path` ON `files` (`post_id`,`path`);

COMMIT;
