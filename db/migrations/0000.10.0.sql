CREATE TABLE `workout_metric` (
	`id`			INT(12) NOT NULL PRIMARY KEY AUTO_INCREMENT
		COMMENT 'Unique ID of the workout metric',
	`workout_id`	INT(10) NOT NULl
		COMMENT 'Workout reference',
	`type` 			VARCHAR(15) NOT NULL
		COMMENT 'Unique identification type of the workout metric',
	
	`int_val1`		INT(10) DEFAULT NULL,
	`int_val2`		INT(10) DEFAULT NULL,
	`int_val3` 		INT(10) DEFAULT NULL,

	CONSTRAINT `fk_workout_metric_workout_id` FOREIGN KEY (`workout_id`) REFERENCES `workout`(`id`) ON DELETE CASCADE
);

ALTER TABLE rule_tagging
	ADD COLUMN downsample_30 INT(1) NOT NULL DEFAULT 0
		COMMENT 'Downsample the workout points to 30 seconds';