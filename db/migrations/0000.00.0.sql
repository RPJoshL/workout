CREATE TABLE `user` (
	`id`			INT(10) NOT NULL PRIMARY KEY AUTO_INCREMENT
		COMMENT 'Unique ID of the user',
	`name`			VARCHAR(50) NOT NULL
		COMMENT 'Display name showed to other users',
	`mail`			VARCHAR(100) NOT NULL UNIQUE
		COMMENT 'Unique E-Mail address',
	`password`		VARCHAR(96) NOT NULL
		COMMENT 'Hashed password with the argon2 algorithm',
	`weight`		INT(3) NOT NULL
		COMMENT 'Body height in kg',
	`height`		INT(3) NOT NULL
		COMMENT 'Body height in cm',
	`birth_year`	INT(4) NOT NULL
		COMMENT 'Year the user was born in',
	`vo2_max`		INT(2) NOT NULL
		COMMENT 'VO2max value in mL/kg/min'
) ENGINE = InnoDB;

CREATE TABLE `workout_type` (
	`id`			INT(10) NOT NULL PRIMARY KEY
		COMMENT 'Unique ID of this workout type',
	`name`			VARCHAR(20) NOT NULL
		COMMENT 'Description name of the workout type',
	`tag_dark`		VARCHAR(10) NOT NULL
		COMMENT 'Color code (#f20102) of the tag for the dark mode',
	`tag_white`		VARCHAR(10) NOT NULL
		COMMENT 'Color code (#f20102) of the tag for the white mode'
) ENGINE = InnoDB;

CREATE TABLE `tag` (
	`id`			INT(10) NOT NULL PRIMARY KEY
		COMMENT 'Unique ID of this tag',
	`name`			VARCHAR(20) NOT NULL
		COMMENT 'Short description name of the tag',
	`tag_dark`		VARCHAR(10) NOT NULL
		COMMENT 'Color code (#f20102) of the tag for the dark mode',
	`tag_white`		VARCHAR(10) NOT NULL
		COMMENT 'Color code (#f20102) of the tag for the white mode'
) ENGINE = InnoDB;

CREATE TABLE `workout` (
	`id`			INT(10) NOT NULL PRIMARY KEY
		COMMENT 'Unique ID of the workout',
	`user_id`		INT(10) NOT NULL
		COMMENT 'ID of the user the workout belongs to',
	`type_id`		INT(10) NOT NULL
		COMMENT 'Workout type or categorie',
	`start`			DATETIME NOT NULL
		COMMENT 'Time and date the workout was started',
	`end`			DATETIME NOT NULL
		COMMENT 'Time and date the workout was completed',
	`country`		VARCHAR(2) NOT NULL
		COMMENT '2 letter country code where the workout was started',
	`city`			VARCHAR(50) NOT NULL
		COMMENT 'Name of the city where the workout was started',
	`city_id`		INT(15) NOT NULL
		COMMENT 'Unique ID for the city in the geonames database where the workout was started',
	`city_latitude`	POINT NOT NULL
		COMMENT 'Latitude of the city',
	`city_longitude`POINT NOT NULL
		COMMENT 'Longitude of the city',
	`duration`		INT(8) NOT NULL
		COMMENT 'Duration in seconds the workout lasted without any pauses',
	`calories`		INT(5) NOT NULL
		COMMENT 'Number of calories that were burned during the workouts "duration"',
	`calories_default` INT(5) NOT NULL
		COMMENT 'Number of calories that were by default burned during the workouts "duration"',
	`distance`		INT(5) NOT NULL
		COMMENT 'Distance in meters traveled during the workout',
	`speed_av`		INT(5) NOT NULL
		COMMENT 'Average traveling speed in sec/km',
	`elevation_up`	INT(5) NOT NULL
		COMMENT 'Attitude meters (up) made during the workout',
	`elevation_down`INT(5) NOT NULL
		COMMENT 'Attitude meters (down) made during the workout',
	`heart_rate_av`	INT(4)
		COMMENT 'Average heart rate during the workout',
	`heart_rate_max`	INT(4)
		COMMENT 'Maximum heart rate during the workout',

	CONSTRAINT `fk_workout_user_id` FOREIGN KEY (`user_id`) REFERENCES `user`(`id`),
	CONSTRAINT `fk_workout_type_id` FOREIGN KEY (`type_id`) REFERENCES `workout_type`(`id`)
) ENGINE = InnoDB;

CREATE TABLE `workout_details` (
	`id`			INT(12) NOT NULL PRIMARY KEY
		COMMENT 'Unique ID of the workout details',
	`workout_id`	INT(10) NOT NULl
		COMMENT 'Workout reference',
	`type`			INT(1) NOT NULL
		COMMENT 'There are two different types of workout details stored:\n0 = detailed and all workout points | 1 = downsampled points for an overview table',
	`longitude`		DECIMAL(11,7) NOT NULL
		COMMENT 'Longitude of the data point',
	`latitude`		DECIMAL(11,7) NOT NULL
		COMMENT 'Latitude of the data point',
	`elevation`		INT(4) NOT NULL
		COMMENT 'Elevation height of the data point. This can be 0 if elevation is not supported by the tracker',
	`speed`			INT(5) NOT NULL
		COMMENT 'Cummolated traveling speed in sec/km',
	`heart_rate`	INT(4)
		COMMENT 'Current heart rate',

	CONSTRAINT `fk_workout_details_workout_id` FOREIGN KEY (`workout_id`) REFERENCES `workout`(`id`)
) ENGINE = InnoDB;

CREATE TABLE `workout_tags` 
(
    `workout_id`	INT(10) NOT NULL
		COMMENT 'Reference to workout',
    `tag_id`		INT(10) NOT NULL
		COMMENT 'Reference to assigned tag',

    PRIMARY KEY (`workout_id`, `tag_id`),
    CONSTRAINT `fk_workout_tags_workout_id`   FOREIGN KEY (`workout_id`)  REFERENCES `workout`(`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_workout_tags_tag_id`       FOREIGN KEY (`tag_id`)      REFERENCES `tag`(`id`)     ON DELETE CASCADE
) ENGINE = InnoDB;
