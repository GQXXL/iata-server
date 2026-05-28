CREATE TABLE IF NOT EXISTS `node_latency_monitor_task` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(128) NOT NULL DEFAULT '' COMMENT 'task name',
  `monitor_type` VARCHAR(32) NOT NULL DEFAULT 'tcp' COMMENT 'monitor type',
  `target` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'target host:port or endpoint',
  `node_ids` TEXT NOT NULL COMMENT 'comma separated node ids',
  `interval_seconds` INT NOT NULL DEFAULT 60 COMMENT 'run interval seconds',
  `enabled` TINYINT(1) NOT NULL DEFAULT 1,
  `last_run_at` DATETIME(3) NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_enabled` (`enabled`),
  KEY `idx_updated_at` (`updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `node_latency_monitor_result` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `task_id` BIGINT NOT NULL,
  `node_id` BIGINT NOT NULL,
  `isp` VARCHAR(16) NOT NULL DEFAULT '',
  `latency_ms` BIGINT NOT NULL DEFAULT -1,
  `status` VARCHAR(16) NOT NULL DEFAULT 'offline',
  `error_msg` VARCHAR(255) NOT NULL DEFAULT '',
  `checked_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_task_node_isp` (`task_id`,`node_id`,`isp`),
  KEY `idx_node_isp_checked` (`node_id`,`isp`,`checked_at`),
  KEY `idx_checked_at` (`checked_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
