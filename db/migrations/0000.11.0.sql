ALTER TABLE workout
	ADD CONSTRAINT un_workout_user_start_end UNIQUE INDEX workout (user_id, `start`, `end`);

ALTER TABLE workout_details
	ADD COLUMN `acceleration` BLOB DEFAULT NULL
	COMMENT 'Acceleration data for this point as a repeating array:
 - relative timestamp in ms for point
 - x,y,z stored as signed Int16, scaled in g: 2048 = 1 g (9.80665 m/s²)';
        
ALTER TABLE workout_details
	ADD COLUMN `horizontal_accuracy` DECIMAL(6,2) DEFAULT NULL
	COMMENT 'Horizontal accuracy in meters';

CREATE TABLE external_api (
	`id`			 INT(10) NOT NULL PRIMARY KEY AUTO_INCREMENT,
	`user_id`   	 INT(10) NOT NULL
		COMMENT 'User for which this credential is applied',
	`type`			 VARCHAR(20) NOT NULL
		COMMENT 'Unique identification of the server type (e.g. strava)',
	`server_address` VARCHAR(70) NOT NULL
		COMMENT 'Optional (base) server address to connect to',
	`credentials` 	 VARCHAR(400) NOT NULL
		COMMENT 'Type specific credentials used to connect to the remote API server',
	
	
	CONSTRAINT `fk_external_api_user_id` FOREIGN KEY (`user_id`) REFERENCES `user`(id) ON DELETE CASCADE,
	CONSTRAINT `un_external_api_type` UNIQUE INDEX workout (user_id, `type`)
);

ALTER TABLE `workout`
	ADD COLUMN `pumpfoilorg_sync`	VARCHAR(60) DEFAULT NULL
		COMMENT 'Internal ID of the workout on the pumpfoil.org server';
ALTER TABLE `workout`
	ADD COLUMN `strava_sync`		VARCHAR(60) DEFAULT NULL
		COMMENT 'Internal ID of the workout on the strava server';