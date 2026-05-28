CREATE TABLE IF NOT EXISTS `probe_agent` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `server_id` BIGINT NOT NULL COMMENT 'bind to servers.id',
  `name` VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'agent name',
  `token_hash` VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'sha256(token)',
  `status` VARCHAR(16) NOT NULL DEFAULT 'offline' COMMENT 'online/offline',
  `version` VARCHAR(32) NOT NULL DEFAULT '',
  `last_seen_at` DATETIME(3) NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_server_id` (`server_id`),
  UNIQUE KEY `uk_token_hash` (`token_hash`),
  KEY `idx_status_last_seen` (`status`, `last_seen_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `probe_agent_target` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `server_id` BIGINT NOT NULL,
  `target_ct` VARCHAR(255) NOT NULL DEFAULT '',
  `target_cu` VARCHAR(255) NOT NULL DEFAULT '',
  `target_cm` VARCHAR(255) NOT NULL DEFAULT '',
  `enabled` TINYINT(1) NOT NULL DEFAULT 1,
  `interval_seconds` INT NOT NULL DEFAULT 30,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_server_target` (`server_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `probe_agent_result` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `server_id` BIGINT NOT NULL,
  `isp` VARCHAR(16) NOT NULL DEFAULT '' COMMENT 'ct/cu/cm',
  `latency_ms` BIGINT NOT NULL DEFAULT -1,
  `status` VARCHAR(16) NOT NULL DEFAULT 'offline',
  `error_msg` VARCHAR(255) NOT NULL DEFAULT '',
  `probe_mode` VARCHAR(32) NOT NULL DEFAULT 'tcp',
  `checked_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_server_isp` (`server_id`, `isp`),
  KEY `idx_server_checked` (`server_id`, `checked_at`),
  KEY `idx_isp_checked` (`isp`, `checked_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
