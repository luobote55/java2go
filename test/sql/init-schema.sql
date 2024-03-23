SET NAMES utf8mb4;
SET
FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for file_transfer_log
-- ----------------------------
DROP TABLE IF EXISTS `device_list`;
CREATE TABLE `device_list`
(
    `id`            bigint(0) NOT NULL AUTO_INCREMENT COMMENT 'id',
    `user_id`       bigint(0) NULL DEFAULT NULL COMMENT '设备id',
    `exec_status`   int(0) NULL DEFAULT NULL COMMENT '执行状态 1 未开始 2 执行中 3 执行成功 4 执行失败 5 执行终止',
    `machine_id`    bigint(0) NULL DEFAULT NULL COMMENT '机器id',
    `exit_code`     int(0) NULL DEFAULT NULL COMMENT '执行返回码',
    `file_token`    bigint(0) NULL DEFAULT NULL COMMENT '文件token',
    `cpu_range`     double(5, 2) NULL DEFAULT NULL COMMENT 'CPU占用率',
    `deleted` tinyint(0) NULL DEFAULT 1 COMMENT '是否删除 1未删除 2已删除',
    `create_time` datetime(4) NULL DEFAULT CURRENT_TIMESTAMP(4) COMMENT '创建时间',
    `update_time` datetime(4) NULL DEFAULT CURRENT_TIMESTAMP(4) ON UPDATE CURRENT_TIMESTAMP(4) COMMENT '修改时间',
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE INDEX `token_unidx`(`file_token`) USING BTREE,
    INDEX `user_machine_idx`(`user_id`, `machine_id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8 COLLATE = utf8_general_ci COMMENT = '设备表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for file_transfer_log
-- ----------------------------
DROP TABLE IF EXISTS `user_list`;
CREATE TABLE `user_list`
(
    `id`            bigint(0) NOT NULL AUTO_INCREMENT COMMENT 'id',
    `user_id`       bigint(0) NULL DEFAULT NULL COMMENT '设备id',
    `exec_status`   int(0) NULL DEFAULT NULL COMMENT '执行状态 1 未开始 2 执行中 3 执行成功 4 执行失败 5 执行终止',
    `machine_id`    bigint(0) NULL DEFAULT NULL COMMENT '机器id',
    `exit_code`     int(0) NULL DEFAULT NULL COMMENT '执行返回码',
    `file_token`    bigint(0) NULL DEFAULT NULL COMMENT '文件token',
    `cpu_range`     double(5, 2) NULL DEFAULT NULL COMMENT 'CPU占用率',
    `deleted` tinyint(0) NULL DEFAULT 1 COMMENT '是否删除 1未删除 2已删除',
    `create_time` datetime(4) NULL DEFAULT CURRENT_TIMESTAMP(4) COMMENT '创建时间',
    `update_time` datetime(4) NULL DEFAULT CURRENT_TIMESTAMP(4) ON UPDATE CURRENT_TIMESTAMP(4) COMMENT '修改时间',
    PRIMARY KEY (`id`) USING BTREE,
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8 COLLATE = utf8_general_ci COMMENT = '设备表' ROW_FORMAT = Dynamic;
