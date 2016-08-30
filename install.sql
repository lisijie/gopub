CREATE TABLE `t_action` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `action` varchar(20) NOT NULL DEFAULT '',
  `actor` varchar(20) NOT NULL DEFAULT '',
  `object_type` varchar(20) NOT NULL DEFAULT '',
  `object_id` int(11) NOT NULL DEFAULT '0',
  `extra` varchar(1000) NOT NULL DEFAULT '',
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `create_time` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `t_env` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `project_id` int(11) NOT NULL DEFAULT '0',
  `name` varchar(20) NOT NULL DEFAULT '',
  `ssh_user` varchar(20) NOT NULL DEFAULT '',
  `ssh_port` varchar(10) NOT NULL DEFAULT '',
  `ssh_key` varchar(100) NOT NULL DEFAULT '',
  `pub_dir` varchar(100) NOT NULL DEFAULT '',
  `before_shell` longtext NOT NULL,
  `after_shell` longtext NOT NULL,
  `server_count` int(11) NOT NULL DEFAULT '0',
  `send_mail` int(11) NOT NULL DEFAULT '0',
  `mail_tpl_id` int(11) NOT NULL DEFAULT '0',
  `mail_to` varchar(1000) NOT NULL DEFAULT '',
  `mail_cc` varchar(1000) NOT NULL DEFAULT '',
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `t_env_project_id` (`project_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `t_env_server` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `project_id` int(11) NOT NULL DEFAULT '0',
  `env_id` int(11) NOT NULL DEFAULT '0',
  `server_id` int(11) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`),
  KEY `t_env_server_env_id` (`env_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `t_mail_tpl` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_id` int(11) NOT NULL DEFAULT '0',
  `name` varchar(100) NOT NULL DEFAULT '',
  `subject` varchar(200) NOT NULL DEFAULT '',
  `content` longtext NOT NULL,
  `mail_to` varchar(1000) NOT NULL DEFAULT '',
  `mail_cc` varchar(1000) NOT NULL DEFAULT '',
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `t_perm` (
  `module` varchar(20) NOT NULL DEFAULT '' COMMENT '模块名',
  `action` varchar(20) NOT NULL DEFAULT '' COMMENT '操作名',
  UNIQUE KEY `module` (`module`,`action`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `t_project` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(100) NOT NULL DEFAULT '',
  `domain` varchar(100) NOT NULL DEFAULT '',
  `version` varchar(20) NOT NULL DEFAULT '',
  `version_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `repo_url` varchar(100) NOT NULL DEFAULT '',
  `status` int(11) NOT NULL DEFAULT '0',
  `error_msg` longtext NOT NULL,
  `agent_id` int(11) NOT NULL DEFAULT '0' COMMENT '跳板机ID',
  `ignore_list` longtext NOT NULL,
  `before_shell` longtext NOT NULL,
  `after_shell` longtext NOT NULL,
  `create_verfile` int(11) NOT NULL DEFAULT '0',
  `verfile_path` varchar(50) NOT NULL DEFAULT '',
  `task_review` tinyint(4) NOT NULL DEFAULT '0' COMMENT '发布是否需要审批',
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `t_role` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `role_name` varchar(20) NOT NULL DEFAULT '',
  `project_ids` varchar(1000) NOT NULL DEFAULT '',
  `description` varchar(200) NOT NULL DEFAULT '',
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `t_role_perm` (
  `role_id` int(11) unsigned NOT NULL,
  `perm` varchar(50) NOT NULL DEFAULT '',
  PRIMARY KEY (`role_id`,`perm`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `t_server` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `type_id` tinyint(4) NOT NULL DEFAULT '0' COMMENT '0:普通服务器, 1:跳板机',
  `ip` varchar(20) NOT NULL DEFAULT '' COMMENT '服务器IP',
  `area` varchar(20) NOT NULL DEFAULT '' COMMENT '机房',
  `description` varchar(200) NOT NULL DEFAULT '' COMMENT '描述',
  `ssh_port` int(11) NOT NULL COMMENT 'ssh端口',
  `ssh_user` varchar(50) NOT NULL DEFAULT '' COMMENT 'ssh帐号',
  `ssh_pwd` varchar(100) NOT NULL DEFAULT '' COMMENT 'ssh密码',
  `ssh_key` varchar(100) NOT NULL DEFAULT '' COMMENT 'sshkey路径',
  `work_dir` varchar(100) NOT NULL DEFAULT '' COMMENT '工作目录',
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `t_task` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `project_id` int(11) NOT NULL DEFAULT '0' COMMENT '项目ID',
  `start_ver` varchar(20) NOT NULL DEFAULT '' COMMENT '起始版本',
  `end_ver` varchar(20) NOT NULL DEFAULT '' COMMENT '结束版本',
  `message` longtext NOT NULL COMMENT '版本内容',
  `user_id` int(11) NOT NULL DEFAULT '0' COMMENT '提单人ID',
  `user_name` varchar(20) NOT NULL DEFAULT '' COMMENT '提单人',
  `build_status` tinyint(4) NOT NULL DEFAULT '0' COMMENT '构建状态',
  `change_logs` longtext NOT NULL COMMENT '修改日志',
  `change_files` longtext NOT NULL COMMENT '修改文件列表',
  `filepath` varchar(200) NOT NULL DEFAULT '' COMMENT '发布包路径',
  `pub_env_id` int(11) NOT NULL DEFAULT '0' COMMENT '发布环境ID',
  `pub_time` datetime DEFAULT NULL COMMENT '发布时间',
  `error_msg` longtext NOT NULL COMMENT '错误消息',
  `pub_log` longtext NOT NULL COMMENT '发布日志',
  `pub_status` tinyint(4) NOT NULL DEFAULT '0' COMMENT '发布状态',
  `review_status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '审批状态',
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `t_task_project_id` (`project_id`),
  KEY `t_task_user_id` (`user_id`),
  KEY `pub_time` (`pub_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `t_task_review` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `task_id` int(11) NOT NULL COMMENT '任务ID',
  `user_id` int(11) NOT NULL COMMENT '审批人ID',
  `user_name` varchar(20) NOT NULL DEFAULT '' COMMENT '审批人名称',
  `status` int(11) NOT NULL DEFAULT '0' COMMENT '审批结果(1:通过;0:不通过)',
  `message` varchar(1000) NOT NULL DEFAULT '' COMMENT '审批说明',
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '审批时间',
  PRIMARY KEY (`id`),
  KEY `task_id` (`task_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `t_user` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_name` varchar(20) NOT NULL DEFAULT '',
  `password` varchar(32) NOT NULL DEFAULT '',
  `salt` varchar(10) NOT NULL DEFAULT '',
  `sex` int(11) NOT NULL DEFAULT '0',
  `email` varchar(50) NOT NULL DEFAULT '',
  `last_login` datetime DEFAULT NULL,
  `last_ip` varchar(15) NOT NULL DEFAULT '',
  `status` int(11) NOT NULL DEFAULT '0',
  `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `user_name` (`user_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `t_user_role` (
  `user_id` int(11) unsigned NOT NULL,
  `role_id` int(11) unsigned NOT NULL,
  UNIQUE KEY `user_id` (`user_id`,`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


INSERT INTO `t_perm` (`module`, `action`)
VALUES
	('agent','add'),
	('agent','del'),
	('agent','edit'),
	('agent','list'),
	('agent','projects'),
	('env','add'),
	('env','del'),
	('env','edit'),
	('env','list'),
	('mailtpl','add'),
	('mailtpl','del'),
	('mailtpl','edit'),
	('mailtpl','list'),
	('project','add'),
	('project','del'),
	('project','edit'),
	('project','list'),
	('review','detail'),
	('review','list'),
	('review','review'),
	('role','add'),
	('role','del'),
	('role','edit'),
	('role','list'),
	('role','perm'),
	('server','add'),
	('server','del'),
	('server','edit'),
	('server','list'),
	('server','projects'),
	('task','create'),
	('task','del'),
	('task','detail'),
	('task','list'),
	('task','publish'),
	('user','add'),
	('user','del'),
	('user','edit'),
	('user','list');


INSERT INTO `t_role` (`id`, `role_name`, `project_ids`, `description`, `create_time`, `update_time`)
VALUES
	(1,'系统管理员','','',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
	(2,'发版人员','','',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
	(3,'审批人员','','',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP);


INSERT INTO `t_role_perm` (`role_id`, `perm`)
VALUES
	(1,'agent.add'),
	(1,'agent.del'),
	(1,'agent.edit'),
	(1,'agent.list'),
	(1,'agent.projects'),
	(1,'env.add'),
	(1,'env.del'),
	(1,'env.edit'),
	(1,'env.list'),
	(1,'mailtpl.add'),
	(1,'mailtpl.del'),
	(1,'mailtpl.edit'),
	(1,'mailtpl.list'),
	(1,'project.add'),
	(1,'project.del'),
	(1,'project.edit'),
	(1,'project.list'),
	(1,'review.detail'),
	(1,'review.list'),
	(1,'review.review'),
	(1,'role.add'),
	(1,'role.del'),
	(1,'role.edit'),
	(1,'role.list'),
	(1,'role.perm'),
	(1,'server.add'),
	(1,'server.del'),
	(1,'server.edit'),
	(1,'server.list'),
	(1,'server.projects'),
	(1,'task.create'),
	(1,'task.del'),
	(1,'task.detail'),
	(1,'task.list'),
	(1,'task.publish'),
	(1,'user.add'),
	(1,'user.del'),
	(1,'user.edit'),
	(1,'user.list'),
	(2,'task.create'),
	(2,'task.del'),
	(2,'task.detail'),
	(2,'task.list'),
	(2,'task.publish'),
	(3,'review.detail'),
	(3,'review.list'),
	(3,'review.review');

INSERT INTO `t_user` (`id`, `user_name`, `password`, `salt`, `sex`, `email`, `last_login`, `last_ip`, `status`, `create_time`, `update_time`)
VALUES
	(1,'admin','7fef6171469e80d32c0559f88b377245','',1,'admin@admin.com','2016-05-11 10:33:49','127.0.0.1',0,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP);