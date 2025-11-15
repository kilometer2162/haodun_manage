-- ============================================
-- 管理系统数据库初始化脚本
-- ============================================

-- 创建数据库
CREATE DATABASE IF NOT EXISTS `haodun` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE `haodun`;

-- 启用事件调度器（需要 SUPER 权限）
SET GLOBAL event_scheduler = ON;

-- ============================================
-- 自动维护订单表分区的事件
-- ============================================

DROP EVENT IF EXISTS `ev_manage_order_info_partitions`;
DELIMITER $$
CREATE EVENT `ev_manage_order_info_partitions`
  ON SCHEDULE EVERY 1 MONTH
  STARTS (TIMESTAMP(CURRENT_DATE) + INTERVAL 1 DAY)
  DO
  BEGIN
    DECLARE v_next_month_first DATE;
    DECLARE v_month_after_next DATE;
    DECLARE v_partition_name VARCHAR(32);
    DECLARE v_partition_exists INT DEFAULT 0;

    SET v_next_month_first = DATE_FORMAT(CURRENT_DATE + INTERVAL 1 MONTH, '%Y-%m-01');
    SET v_month_after_next = DATE_FORMAT(CURRENT_DATE + INTERVAL 2 MONTH, '%Y-%m-01');
    SET v_partition_name = CONCAT('p', DATE_FORMAT(v_next_month_first, '%Y_%m'));

    -- 检查下个月分区是否已经存在
    SELECT COUNT(*) INTO v_partition_exists
    FROM information_schema.partitions
    WHERE table_schema = DATABASE()
      AND table_name = 'order_info'
      AND partition_name = v_partition_name;

    IF v_partition_exists = 0 THEN
      SET @ddl = CONCAT(
        'ALTER TABLE `order_info` ADD PARTITION (PARTITION `', v_partition_name,
        '` VALUES LESS THAN (TO_DAYS(''', v_month_after_next, ''')))'
      );
      PREPARE stmt FROM @ddl;
      EXECUTE stmt;
      DEALLOCATE PREPARE stmt;
    END IF;
  END$$
DELIMITER ;


DROP TABLE IF EXISTS `departments`;
-- 部门表
CREATE TABLE IF NOT EXISTS `departments` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(128) NOT NULL COMMENT '部门名称',
  `parent_id` bigint unsigned DEFAULT NULL COMMENT '父部门ID',
  `description` varchar(255) DEFAULT NULL COMMENT '部门描述',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '状态: 1-启用, 0-禁用',
  `sort` int NOT NULL DEFAULT 0 COMMENT '排序',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_departments_name` (`name`),
  KEY `idx_departments_parent_id` (`parent_id`),
  KEY `idx_departments_status` (`status`),
  KEY `idx_departments_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='部门表';

-- ============================================
-- 创建表结构
-- ============================================

DROP TABLE IF EXISTS `resources`;
-- 资源表
CREATE TABLE IF NOT EXISTS `resources` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL COMMENT '资源名称',
  `path` varchar(255) NOT NULL COMMENT '资源路径',
  `method` varchar(255) NOT NULL COMMENT '请求方法: GET, POST, PUT, DELETE',
  `description` varchar(255) DEFAULT NULL COMMENT '描述',
  `type` varchar(255) DEFAULT NULL COMMENT '类型: api, menu, button',
  `parent_id` bigint unsigned DEFAULT NULL COMMENT '父资源ID',
  `sort` int DEFAULT '0' COMMENT '排序',
  `icon` varchar(255) DEFAULT NULL COMMENT '图标',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_resources_path` (`path`),
  KEY `idx_resources_deleted_at` (`deleted_at`),
  KEY `idx_resources_parent_id` (`parent_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='资源表';


DROP TABLE IF EXISTS `roles`;
-- 角色表
CREATE TABLE IF NOT EXISTS `roles` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL COMMENT '角色名称',
  `description` varchar(255) DEFAULT NULL COMMENT '角色描述',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_roles_name` (`name`),
  KEY `idx_roles_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色表';



DROP TABLE IF EXISTS `permissions`;
-- 权限表
CREATE TABLE IF NOT EXISTS `permissions` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL COMMENT '权限名称',
  `code` varchar(255) NOT NULL COMMENT '权限代码',
  `description` varchar(255) DEFAULT NULL COMMENT '权限描述',
  `resource_id` bigint unsigned NOT NULL COMMENT '关联资源ID',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_permissions_name` (`name`),
  UNIQUE KEY `idx_permissions_code` (`code`),
  KEY `idx_permissions_deleted_at` (`deleted_at`),
  KEY `idx_permissions_resource_id` (`resource_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='权限表';


DROP TABLE IF EXISTS `role_permissions`;

-- 角色权限关联表（多对多关系）
CREATE TABLE IF NOT EXISTS `role_permissions` (
  `role_id` bigint unsigned NOT NULL,
  `permission_id` bigint unsigned NOT NULL,
  PRIMARY KEY (`role_id`, `permission_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色权限关联表';


DROP TABLE IF EXISTS `users`;
-- 用户表
CREATE TABLE IF NOT EXISTS `users` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `username` varchar(255) NOT NULL COMMENT '用户名',
  `password` varchar(255) NOT NULL COMMENT '密码（bcrypt哈希）',
  `email` varchar(255) DEFAULT NULL COMMENT '邮箱',
  `status` int DEFAULT '1' COMMENT '状态: 1-正常, 0-禁用',
  `role_id` bigint unsigned NOT NULL COMMENT '角色ID',
  `department_id` bigint unsigned DEFAULT NULL COMMENT '部门ID',
  `employee_type` varchar(32) NOT NULL DEFAULT 'internal' COMMENT '员工类型: internal-内部, external-外部',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_users_username` (`username`),
  KEY `idx_users_deleted_at` (`deleted_at`),
  KEY `idx_users_role_id` (`role_id`),
  KEY `idx_users_department_id` (`department_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';



DROP TABLE IF EXISTS `logs`;
-- 日志表
CREATE TABLE IF NOT EXISTS `logs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned DEFAULT NULL COMMENT '用户ID',
  `username` varchar(255) DEFAULT NULL COMMENT '用户名',
  `action` varchar(255) DEFAULT NULL COMMENT '操作类型',
  `module` varchar(255) DEFAULT NULL COMMENT '模块',
  `content` text COMMENT '操作内容',
  `ip` varchar(255) DEFAULT NULL COMMENT 'IP地址',
  `user_agent` varchar(255) DEFAULT NULL COMMENT '用户代理',
  `status` int DEFAULT NULL COMMENT '状态: 1-成功, 0-失败',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_logs_deleted_at` (`deleted_at`),
  KEY `idx_logs_user_id` (`user_id`),
  KEY `idx_logs_username` (`username`),
  KEY `idx_logs_action` (`action`),
  KEY `idx_logs_module` (`module`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='日志表';



  DROP TABLE IF EXISTS `ip_accesses`;
-- IP访问统计表
CREATE TABLE IF NOT EXISTS `ip_accesses` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `ip` varchar(255) NOT NULL COMMENT 'IP地址',
  `date` date NOT NULL COMMENT '访问日期',
  `country` varchar(255) DEFAULT NULL COMMENT '国家',
  `province` varchar(255) DEFAULT NULL COMMENT '省份',
  `city` varchar(255) DEFAULT NULL COMMENT '城市',
  `isp` varchar(255) DEFAULT NULL COMMENT '运营商',
  `access_count` int DEFAULT '1' COMMENT '访问次数',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_ip_accesses_ip` (`ip`),
  KEY `idx_ip_accesses_date` (`date`),
  KEY `idx_ip_accesses_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='IP访问统计表';




  DROP TABLE IF EXISTS `configs`;
-- 系统参数表
CREATE TABLE IF NOT EXISTS `configs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `key` varchar(255) NOT NULL COMMENT '参数键',
  `value` text COMMENT '参数值',
  `label` varchar(255) NOT NULL COMMENT '参数标签',
  `type` varchar(50) NOT NULL DEFAULT 'text' COMMENT '参数类型',
  `group` varchar(100) NOT NULL DEFAULT 'system' COMMENT '参数分组',
  `description` text COMMENT '描述',
  `sort` int NOT NULL DEFAULT 0 COMMENT '排序',
  `status` int NOT NULL DEFAULT 1 COMMENT '状态: 1-启用, 0-禁用',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_configs_key` (`key`),
  KEY `idx_configs_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统参数表';

INSERT INTO `configs` VALUES 
(1,'site_name','豪盾后台管理','站点名称','text','system','系统展示名称',1,1,'2025-11-08 22:25:40.908','2025-11-10 15:04:03.329',NULL),
(2,'default_address','默认地址请在【系统参数】修改','默认地址','text','business','',0,1,'2025-11-09 12:41:09.213','2025-11-11 11:29:06.049',NULL),
(3,'storage_driver','local','存储驱动','select','storage','附件存储方式: local 或 cos',1,1,'2025-11-09 16:11:20.690','2025-11-10 15:03:43.306',NULL),
(4,'local_storage_path','./uploads','本地存储路径','text','storage','本地磁盘保存附件的路径',2,1,'2025-11-09 16:11:20.696','2025-11-09 16:11:20.696',NULL),
(5,'local_base_url','','本地访问URL前缀','text','storage','前端访问本地附件的URL前缀',3,1,'2025-11-09 16:11:20.698','2025-11-09 16:11:20.698',NULL),
(6,'cos_key_prefix','orders','COS对象前缀','text','storage','上传到COS的对象路径前缀',4,1,'2025-11-09 16:11:20.699','2025-11-09 16:11:20.699',NULL);



  DROP TABLE IF EXISTS `dict_types`;
-- 字典类型表
CREATE TABLE IF NOT EXISTS `dict_types` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `code` varchar(100) NOT NULL COMMENT '字典类型代码',
  `name` varchar(255) NOT NULL COMMENT '字典类型名称',
  `description` text COMMENT '描述',
  `status` int NOT NULL DEFAULT 1 COMMENT '状态: 1-启用, 0-禁用',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_dict_types_code` (`code`),
  KEY `idx_dict_types_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='字典类型表';

INSERT INTO `dict_types` VALUES (1,'shipping_warehouse','发货仓库','',1,'2025-11-09 21:52:57.818','2025-11-09 21:52:57.818',NULL),(2,'order_status','订单状态','',1,'2025-11-11 10:47:02.950','2025-11-11 10:47:02.950',NULL);




  DROP TABLE IF EXISTS `dict_items`;

-- 字典项表
CREATE TABLE IF NOT EXISTS `dict_items` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `type_id` bigint unsigned NOT NULL COMMENT '字典类型ID',
  `label` varchar(255) NOT NULL COMMENT '显示标签',
  `value` varchar(255) NOT NULL COMMENT '值',
  `sort` int NOT NULL DEFAULT 0 COMMENT '排序',
  `status` int NOT NULL DEFAULT 1 COMMENT '状态: 1-启用, 0-禁用',
  `description` text COMMENT '描述',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_dict_items_type_value` (`type_id`, `value`),
  KEY `idx_dict_items_type_id` (`type_id`),
  KEY `idx_dict_items_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='字典项表';

INSERT INTO `dict_items` VALUES (1,1,'YC.BY','YC.BY',0,1,'','2025-11-09 21:53:18.455','2025-11-09 21:53:18.455',NULL),(2,1,'YC.D','YC.D',1,1,'','2025-11-09 21:53:28.478','2025-11-09 21:53:52.865',NULL),(3,1,'YC.X','YC.X',2,1,'','2025-11-09 21:53:39.600','2025-11-09 21:54:04.471',NULL),(4,1,'YZG.CA','YZG.CA',3,1,'','2025-11-09 21:54:21.375','2025-11-09 21:54:21.375',NULL),(5,1,'YZG.D','YZG.D',4,1,'','2025-11-09 21:54:32.542','2025-11-09 21:54:32.542',NULL),(6,1,'YZG.X','YZG.X',5,1,'','2025-11-09 21:55:04.730','2025-11-09 21:55:04.730',NULL),(7,1,'QQH.NY','QQH.NY',6,1,'','2025-11-09 21:55:13.965','2025-11-09 21:55:13.965',NULL),(8,1,'QQH.WV','QQH.WV',7,1,'','2025-11-09 21:55:22.511','2025-11-09 21:55:22.511',NULL),(9,1,'TX.X','TX.X',8,1,'','2025-11-09 21:55:31.064','2025-11-09 21:55:35.837',NULL),(10,1,'TX.D','TX.D',9,1,'','2025-11-09 21:55:44.036','2025-11-09 21:55:44.036',NULL),(11,1,'HTC.X','HTC.X',9,1,'','2025-11-09 21:55:51.394','2025-11-09 21:56:03.938',NULL),(12,1,'HTX.D','HTX.D',10,1,'','2025-11-09 21:56:16.050','2025-11-09 21:56:16.050',NULL),(13,2,'完成','1',0,1,'','2025-11-11 10:47:22.427','2025-11-11 10:47:22.427',NULL),(14,2,'未完成','0',1,1,'','2025-11-11 10:47:31.902','2025-11-11 10:47:31.902',NULL);




  DROP TABLE IF EXISTS `notices`;
-- 消息公告表
CREATE TABLE IF NOT EXISTS `notices` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `title` varchar(255) NOT NULL COMMENT '标题',
  `content` text COMMENT '内容',
  `type` varchar(50) NOT NULL DEFAULT 'notice' COMMENT '类型: notice-公告, message-消息',
  `status` int NOT NULL DEFAULT 1 COMMENT '状态: 1-发布, 0-草稿',
  `priority` int NOT NULL DEFAULT 0 COMMENT '优先级: 0-普通,1-重要,2-紧急',
  `created_by` bigint unsigned DEFAULT NULL COMMENT '创建人ID',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_notices_type` (`type`),
  KEY `idx_notices_status` (`status`),
  KEY `idx_notices_deleted_at` (`deleted_at`),
  KEY `idx_notices_created_by` (`created_by`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='消息公告表';



  DROP TABLE IF EXISTS `notice_reads`;
-- 消息阅读记录表
CREATE TABLE IF NOT EXISTS `notice_reads` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `notice_id` bigint unsigned NOT NULL COMMENT '消息ID',
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `is_read` int NOT NULL DEFAULT 0 COMMENT '是否已读: 1-已读, 0-未读',
  `read_at` datetime DEFAULT NULL COMMENT '阅读时间',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_notice_reads_unique` (`notice_id`, `user_id`),
  KEY `idx_notice_reads_notice_id` (`notice_id`),
  KEY `idx_notice_reads_user_id` (`user_id`),
  KEY `idx_notice_reads_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='消息阅读记录表';

-- 插入默认部门
INSERT INTO `departments` (`id`, `created_at`, `updated_at`, `name`, `status`, `sort`, `description`)
VALUES (1, NOW(), NOW(), '运营部门', 1, 1, '运营部门'),(2, NOW(), NOW(), '采购部门', 1, 2, '采购部门')
ON DUPLICATE KEY UPDATE `name` = `name`;

-- ============================================
-- 初始化默认数据
-- ============================================

-- 插入默认管理员角色
INSERT INTO `roles` (`id`, `created_at`, `updated_at`, `name`, `description`) 
VALUES (1, NOW(), NOW(), 'admin', '系统管理员')
ON DUPLICATE KEY UPDATE `name` = `name`;

-- 插入默认管理员用户
-- 密码: admin123 (bcrypt哈希值)
INSERT INTO `users` (`id`, `created_at`, `updated_at`, `username`, `password`, `email`, `status`, `role_id`, `department_id`, `employee_type`)
VALUES (1, NOW(), NOW(), 'admin', '$2a$10$1QucJZZHosnm50D8T6ptOuSrvz0pdBCpl4uoVpLVAu75AxtdMaTpS', 'admin@example.com', 1, 1, 3, 'internal'),(2, NOW(), NOW(), 'operator', '$2a$10$vvPqMcoFbKpe1XcxDuptE.WH2QE8eyLt6I/Wd5eJjXl3ThpSqsUqK', 'operator@example.com', 1, 2, 1, 'internal'),(3, NOW(), NOW(), 'buyer', '$2a$10$cRXP4oDrpFb4MnsnTIZqAuOdyFnxGT0EmT8ZZPAHwnn1JQVEuUhay', 'buyer@example.com', 1, 3, 2, 'internal')
ON DUPLICATE KEY UPDATE `username` = `username`;


-- 初始化菜单资源
INSERT INTO `resources` (`id`, `created_at`, `updated_at`, `name`, `path`, `method`, `description`, `type`, `parent_id`, `sort`, `icon`)
VALUES
  (1, NOW(), NOW(), '仪表盘', '/dashboard', 'GET', '仪表盘菜单', 'menu', NULL, 10, 'House'),
  (2, NOW(), NOW(), '系统管理', '/system', 'GET', '系统管理菜单', 'menu', NULL, 20, 'Setting'),
  (3, NOW(), NOW(), '消息公告', '/notices', 'GET', '消息公告菜单', 'menu', NULL, 30, 'Bell'),
  (4, NOW(), NOW(), '日志管理', '/logs', 'GET', '日志管理菜单', 'menu', NULL, 40, 'List'),
  (5, NOW(), NOW(), 'IP统计', '/ip-statistics', 'GET', 'IP统计菜单', 'menu', NULL, 60, 'DataAnalysis'),
  (12, NOW(), NOW(), '系统监控', '/system-monitor', 'GET', '系统监控菜单', 'menu', NULL, 45, 'Monitor'),
  (13, NOW(), NOW(), '订单管理', '/orders', 'GET', '订单管理菜单', 'menu', NULL, 50, 'List')
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `description` = VALUES(`description`),
  `type` = VALUES(`type`),
  `parent_id` = VALUES(`parent_id`),
  `sort` = VALUES(`sort`),
  `icon` = VALUES(`icon`);

INSERT INTO `resources` (`id`, `created_at`, `updated_at`, `name`, `path`, `method`, `description`, `type`, `parent_id`, `sort`, `icon`)
VALUES
  (6, NOW(), NOW(), '用户管理', '/users', 'GET', '用户管理菜单', 'menu', 2, 10, 'User'),
  (14, NOW(), NOW(), '部门管理', '/departments', 'GET', '部门管理菜单', 'menu', 2, 15, 'OfficeBuilding'),
  (15, NOW(), NOW(), '素材图库', '/materials', 'GET', '素材图库菜单', 'menu', NULL, 60, 'Picture'),
  (7, NOW(), NOW(), '角色管理', '/roles', 'GET', '角色管理菜单', 'menu', 2, 20, 'UserFilled'),
  (8, NOW(), NOW(), '权限管理', '/permissions', 'GET', '权限管理菜单', 'menu', 2, 30, 'Lock'),
  (9, NOW(), NOW(), '资源管理', '/resources', 'GET', '资源管理菜单', 'menu', 2, 40, 'Document'),
  (10, NOW(), NOW(), '系统字典', '/dicts', 'GET', '系统字典菜单', 'menu', 2, 50, 'List'),
  (11, NOW(), NOW(), '系统参数', '/configs', 'GET', '系统参数菜单', 'menu', 2, 60, 'Tools')
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `description` = VALUES(`description`),
  `type` = VALUES(`type`),
  `parent_id` = VALUES(`parent_id`),
  `sort` = VALUES(`sort`),
  `icon` = VALUES(`icon`);

-- 初始化菜单权限
INSERT INTO `permissions` (`id`, `created_at`, `updated_at`, `name`, `code`, `description`, `resource_id`)
VALUES
  (1, NOW(), NOW(), '仪表盘菜单权限', 'menu:dashboard:view', '访问仪表盘菜单', 1),
  (2, NOW(), NOW(), '系统管理菜单权限', 'menu:system:view', '访问系统管理菜单', 2),
  (14, NOW(), NOW(), '部门管理菜单权限', 'menu:departments:view', '访问部门管理菜单', 14),
  (15, NOW(), NOW(), '素材图库菜单权限', 'menu:materials:view', '访问素材图库菜单', 15),
  (3, NOW(), NOW(), '消息公告菜单权限', 'menu:notices:view', '访问消息公告菜单', 3),
  (4, NOW(), NOW(), '日志管理菜单权限', 'menu:logs:view', '访问日志管理菜单', 4),
  (5, NOW(), NOW(), 'IP统计菜单权限', 'menu:ip-statistics:view', '访问IP统计菜单', 5),
  (6, NOW(), NOW(), '用户管理菜单权限', 'menu:users:view', '访问用户管理菜单', 6),
  (7, NOW(), NOW(), '角色管理菜单权限', 'menu:roles:view', '访问角色管理菜单', 7),
  (8, NOW(), NOW(), '权限管理菜单权限', 'menu:permissions:view', '访问权限管理菜单', 8),
  (9, NOW(), NOW(), '资源管理菜单权限', 'menu:resources:view', '访问资源管理菜单', 9),
  (10, NOW(), NOW(), '系统字典菜单权限', 'menu:dicts:view', '访问系统字典菜单', 10),
  (11, NOW(), NOW(), '系统参数菜单权限', 'menu:configs:view', '访问系统参数菜单', 11),
  (12, NOW(), NOW(), '系统监控菜单权限', 'menu:system-monitor:view', '访问系统监控菜单', 12),
  (13, NOW(), NOW(), '订单管理菜单权限', 'menu:orders:view', '访问订单管理菜单', 13)
ON DUPLICATE KEY UPDATE
  `name` = VALUES(`name`),
  `description` = VALUES(`description`),
  `resource_id` = VALUES(`resource_id`);

-- 管理员角色拥有全部菜单权限
INSERT INTO `role_permissions` (`role_id`, `permission_id`)
VALUES
  (1, 1), (1, 2), (1, 3), (1, 4), (1, 5), (1, 6), (1, 7), (1, 8), (1, 9), (1, 10), (1, 11), (1, 12), (1, 13), (1, 14), (1, 15)
ON DUPLICATE KEY UPDATE `role_id` = `role_id`;

-- 插入operator角色
INSERT INTO `roles` (`id`, `created_at`, `updated_at`, `name`, `description`) 
VALUES (1, NOW(), NOW(), 'operator', '操作员'),
       (2, '2025-11-10 14:47:08.264', '2025-11-10 16:49:00.642', 'operator', '运营'),
       (3, '2025-11-10 14:47:44.647', '2025-11-10 14:48:20.261', 'buyer', '采购')
ON DUPLICATE KEY UPDATE `name` = `name`;

-- operator角色权限: 消息公告(3)、系统管理(2)、系统字典(10)、订单管理(13)
INSERT INTO `role_permissions` VALUES (1,1),(1,2),(2,2),(1,3),(2,3),(3,3),(1,4),(1,5),(1,6),(1,7),(1,8),(1,9),(1,10),(2,10),(3,10),(1,11),(1,12),(1,13),(2,13),(3,13)
ON DUPLICATE KEY UPDATE `role_id` = `role_id`;

DROP TABLE IF EXISTS `order_info`;
CREATE TABLE `order_info` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
  `gsp_order_no` VARCHAR(32) NOT NULL COMMENT 'GSP订单号',
  `order_type` VARCHAR(32) NOT NULL DEFAULT 'platform' COMMENT '订单类型: platform-平台面单, factory-工厂物流',
  `order_created_at` DATETIME NOT NULL COMMENT '订单创建时间',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '订单状态: 0-待支付, 1-待处理, 2-生产中, 3-发货中, 4-已完成, 5-已取消',
  `payment_time` DATETIME DEFAULT NULL COMMENT '支付时间',
  `completed_at` DATETIME DEFAULT NULL COMMENT '完成时间',
  `shipping_warehouse_code` VARCHAR(20) NULL COMMENT '发货仓库',
  `required_sign_at` DATETIME NULL COMMENT '要求签收时间',
  `shop_code` VARCHAR(50) NULL COMMENT '店铺编号',
  `product_id` VARCHAR(50) NULL COMMENT '商品ID',
  `owner_name` VARCHAR(64) NULL COMMENT '负责人',
  `product_name` VARCHAR(200) NULL COMMENT '商品名称',
  `spec` VARCHAR(64) NULL COMMENT '规格（例如 8x12）',
  `item_no` VARCHAR(64) NULL COMMENT '货号',
  `seller_sku` VARCHAR(64) NULL COMMENT '卖家SKU',
  `platform_sku` VARCHAR(64) NULL COMMENT '平台SKU',
  `platform_skc` VARCHAR(64) NULL COMMENT '平台SKC',
  `platform_spu` VARCHAR(64) NULL COMMENT '平台SPU',
  `product_price` DECIMAL(10,2) NULL COMMENT '商品价格',
  `expected_revenue` DECIMAL(10,2) NULL COMMENT '商品预计收入',
  `special_product_note` VARCHAR(200) NULL COMMENT '特殊产品备注（如定制、木画、墓碑）',
  `currency_code` VARCHAR(16) NULL COMMENT '币种',
  `expected_fulfillment_qty` INT NULL COMMENT '应履约件数',
  `item_count` INT NOT NULL DEFAULT 1 COMMENT '件数',
  `postal_code` VARCHAR(20) NULL COMMENT '邮编',
  `country` VARCHAR(64) NULL COMMENT '国家',
  `province` VARCHAR(64) NULL COMMENT '省份',
  `city` VARCHAR(64) NULL COMMENT '城市',
  `district` VARCHAR(64) NULL COMMENT '区',
  `address_line1` VARCHAR(200) NULL COMMENT '用户地址1',
  `address_line2` VARCHAR(200) NULL COMMENT '用户地址2',
  `customer_full_name` VARCHAR(128) NULL COMMENT '用户全称',
  `customer_last_name` VARCHAR(64) NULL COMMENT '用户姓氏',
  `customer_first_name` VARCHAR(64) NULL COMMENT '用户名字',
  `phone_number` VARCHAR(32) NULL COMMENT '手机号',
  `email` VARCHAR(128) NULL COMMENT '用户邮箱',
  `tax_number` VARCHAR(64) NULL COMMENT '税号',
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录更新时间',
  `deleted_at` DATETIME DEFAULT NULL COMMENT '删除时间',
  `created_by` BIGINT UNSIGNED DEFAULT NULL COMMENT '创建人ID',
  `updated_by` BIGINT UNSIGNED DEFAULT NULL COMMENT '更新人ID',
  PRIMARY KEY (`id`, `order_created_at`),
  UNIQUE KEY `uk_id_created_at` (`id`, `order_created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单信息表'
PARTITION BY RANGE (TO_DAYS(`order_created_at`)) (
  PARTITION `p2025_11` VALUES LESS THAN (TO_DAYS('2025-12-01')),
  PARTITION `pmax` VALUES LESS THAN MAXVALUE
);


  DROP TABLE IF EXISTS `order_attachments`;

-- 订单附件表（素材图片、面单等文件）
CREATE TABLE IF NOT EXISTS `order_attachments` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `order_id` bigint unsigned NOT NULL COMMENT '订单ID',
  `file_type` varchar(32) NOT NULL COMMENT '附件类型: material_image/ shipping_label 等',
  `file_name` varchar(255) NOT NULL COMMENT '文件原始名称',
  `file_path` varchar(512) NOT NULL COMMENT '文件路径',
  `file_ext` varchar(32) DEFAULT NULL COMMENT '文件扩展名',
  `file_size` bigint DEFAULT NULL COMMENT '文件大小（字节）',
  `checksum` varchar(128) DEFAULT NULL COMMENT '文件校验和',
  `storage` varchar(16) NOT NULL DEFAULT 'local' COMMENT '存储类型: local/cos',
  `uploader_id` bigint unsigned DEFAULT NULL COMMENT '上传者ID',
  `material_id` bigint unsigned DEFAULT NULL COMMENT '引用的素材ID',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_order_attachments_order_id` (`order_id`),
  KEY `idx_order_attachments_file_type` (`file_type`),
  KEY `idx_order_attachments_uploader_id` (`uploader_id`),
  KEY `idx_order_attachments_material_id` (`material_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单附件表';

DROP TABLE IF EXISTS `material_folders`;
CREATE TABLE IF NOT EXISTS `material_folders` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(128) NOT NULL COMMENT '文件夹名称',
  `parent_id` bigint unsigned DEFAULT NULL COMMENT '父文件夹ID',
  `path` varchar(512) DEFAULT NULL COMMENT '完整路径',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_material_folder_path` (`path`),
  KEY `idx_material_folders_parent_id` (`parent_id`),
  KEY `idx_material_folders_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='素材文件夹';

INSERT INTO `material_folders` (`id`, `created_at`, `updated_at`, `name`, `parent_id`, `path`)
VALUES (1, NOW(), NOW(), '默认文件夹', NULL, '默认文件夹')
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`), `path` = VALUES(`path`);

DROP TABLE IF EXISTS `material_assets`;
CREATE TABLE IF NOT EXISTS `material_assets` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `code` varchar(64) NOT NULL COMMENT '素材编号',
  `file_name` varchar(255) NOT NULL COMMENT '文件名称',
  `title` varchar(255) DEFAULT NULL COMMENT '素材标题',
  `width` int DEFAULT 0 COMMENT '宽度',
  `height` int DEFAULT 0 COMMENT '高度',
  `dimensions` varchar(64) DEFAULT NULL COMMENT '尺寸文本',
  `format` varchar(32) DEFAULT NULL COMMENT '文件格式',
  `file_size` bigint DEFAULT 0 COMMENT '文件大小（字节）',
  `storage` varchar(16) NOT NULL DEFAULT 'local' COMMENT '存储方式',
  `file_path` varchar(512) DEFAULT NULL COMMENT '文件存储路径',
  `created_by` bigint unsigned DEFAULT NULL COMMENT '创建人',
  `updated_by` bigint unsigned DEFAULT NULL COMMENT '更新人',
  `order_count` int DEFAULT 0 COMMENT '关联订单数量',
  `folder_id` bigint unsigned DEFAULT NULL COMMENT '归属文件夹ID',
  `shape` varchar(32) DEFAULT NULL COMMENT '素材形状',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_material_assets_code` (`code`),
  KEY `idx_material_assets_folder_id` (`folder_id`),
  KEY `idx_material_assets_deleted_at` (`deleted_at`),
  KEY `idx_material_assets_created_by` (`created_by`),
  KEY `idx_material_assets_shape` (`shape`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='素材图库';

INSERT INTO `material_folders` (`id`, `created_at`, `updated_at`, `name`, `parent_id`, `path`)
VALUES (1, NOW(), NOW(), '默认文件夹', NULL, '默认文件夹')
ON DUPLICATE KEY UPDATE `name` = VALUES(`name`), `path` = VALUES(`path`);