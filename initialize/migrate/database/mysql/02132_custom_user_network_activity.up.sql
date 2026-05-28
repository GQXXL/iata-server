-- custom migration: create user_network_activity table
CREATE TABLE IF NOT EXISTS `user_network_activity` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `server_id` BIGINT NOT NULL DEFAULT 0,
  `user_id` BIGINT NOT NULL,
  `subscribe_id` BIGINT NOT NULL DEFAULT 0,
  `user_subscribe_id` BIGINT NOT NULL DEFAULT 0,
  `domain` VARCHAR(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `client_ip` VARCHAR(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `user_agent` VARCHAR(1024) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `upload` BIGINT NOT NULL DEFAULT 0,
  `download` BIGINT NOT NULL DEFAULT 0,
  `timestamp` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_subscribe_id` (`subscribe_id`),
  KEY `idx_domain` (`domain`),
  KEY `idx_timestamp` (`timestamp`),
  KEY `idx_user_subscribe_id` (`user_subscribe_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

