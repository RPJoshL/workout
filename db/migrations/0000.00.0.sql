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
		COMMENT 'Body weight in kg',
	`height`		INT(3) NOT NULL
		COMMENT 'Body height in cm',
	`birth_year`	INT(4) NOT NULL
		COMMENT 'Year the user was born in',
	`vo2_max`		INT(2) NOT NULL
		COMMENT 'VO2max value in mL/kg/min',
	`gender`		INT(1) NOT NULL
		COMMENT 'Male (0) or Female (1)',
	`dark_theme`	BOOLEAN NOT NULL
		COMMENT 'Whether the user enabled the dark theme instead of the light one',
	`timezone`      VARCHAR(20) DEFAULT 'UTC' NOT NULL
		COMMENT 'Timezone the user specified in the last request'
) ENGINE = InnoDB;

CREATE TABLE `workout_type` (
	`id`			INT(10) NOT NULL PRIMARY KEY AUTO_INCREMENT
		COMMENT 'Unique ID of this workout type',
	`name_de`		VARCHAR(20) NOT NULL
		COMMENT 'Description name of the workout type (DE)',
	`name_en`		VARCHAR(20) NOT NULL
		COMMENT 'Description name of the workout type (EN)',
	`tag_dark`		VARCHAR(10) NOT NULL
		COMMENT 'Color code (#f20102) of the tag for the dark mode',
	`tag_white`		VARCHAR(10) NOT NULL
		COMMENT 'Color code (#f20102) of the tag for the white mode'
) ENGINE = InnoDB;

CREATE TABLE `tag` (
	`id`			INT(10) NOT NULL PRIMARY KEY AUTO_INCREMENT
		COMMENT 'Unique ID of this tag',
	`name`			VARCHAR(20) NOT NULL
		COMMENT 'Short description name of the tag',
	`tag_dark`		VARCHAR(10) NOT NULL
		COMMENT 'Color code (#f20102) of the tag for the dark mode',
	`tag_white`		VARCHAR(10) NOT NULL
		COMMENT 'Color code (#f20102) of the tag for the white mode'
) ENGINE = InnoDB;

CREATE TABLE `workout` (
	`id`			INT(10) NOT NULL PRIMARY KEY AUTO_INCREMENT
		COMMENT 'Unique ID of the workout',
	`name`			VARCHAR(40) NOT NULL
		COMMENT 'Name that describes this workout',
	`user_id`		INT(10) NOT NULL
		COMMENT 'ID of the user the workout belongs to',
	`type_id`		INT(10) NOT NULL
		COMMENT 'Workout type or categorie',
	`start`			DATETIME NOT NULL
		COMMENT 'Time and date the workout was started',
	`end`			DATETIME NOT NULL
		COMMENT 'Time and date the workout was completed',
	`country`		VARCHAR(2) NOT NULL
		COMMENT '2-letter country code where the workout was started',
	`city`			VARCHAR(50) NOT NULL
		COMMENT 'Name of the city where the workout was started',
	`city_id`		INT(15) NOT NULL
		COMMENT 'Unique ID for the city in the geonames database where the workout was started',
	`city_location`	POINT NOT NULL
		COMMENT 'Location point of the city',
	`duration`		INT(8) NOT NULL
		COMMENT 'Duration in seconds the workout lasted without any pauses',
	`calories`		INT(5) NOT NULL
		COMMENT 'Number of calories that were burned during the workouts "duration"',
	`calories_default` INT(5) NOT NULL
		COMMENT 'Number of calories that were by default burned during the workouts "duration"',
	`pai` INT(4) NOT NULL DEFAULT 0
		COMMENT 'Physical acvivity score based on heart rate',
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
	`heart_rate_max`INT(4)
		COMMENT 'Maximum heart rate during the workout',
	`note` 			TEXT(4000)
		COMMENT 'Text describing this workout in Markdown format',

	CONSTRAINT `fk_workout_user_id` FOREIGN KEY (`user_id`) REFERENCES `user`(`id`),
	CONSTRAINT `fk_workout_type_id` FOREIGN KEY (`type_id`) REFERENCES `workout_type`(`id`)
) ENGINE = InnoDB;

CREATE TABLE `workout_details` (
	`id`			INT(12) NOT NULL PRIMARY KEY AUTO_INCREMENT
		COMMENT 'Unique ID of the workout details',
	`workout_id`	INT(10) NOT NULl
		COMMENT 'Workout reference',
	`type`			INT(1) NOT NULL
		COMMENT 'There are two different types of workout details stored:\n0 = detailed and all workout points | 1 = downsampled points for an overview table',
	`duration`		INT(7) NOT NULL
		COMMENT 'Duration (without pauses) since the beginning of the workout in seconds',
	`time`			DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		COMMENT 'Date and time of this point',
	`distance`		INT(7) NOT NULl
		COMMENT 'Distance in meters traveled for this point from the beginning of the workout (without pauses)',
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

	CONSTRAINT `fk_workout_details_workout_id` FOREIGN KEY (`workout_id`) REFERENCES `workout`(`id`) ON DELETE CASCADE
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
 
-- Supported workout types
INSERT INTO workout_type (name_de, name_en, tag_dark, tag_white) VALUES ('Gehen', 'Hiking', '#fff', '#000');
INSERT INTO workout_type (name_de, name_en, tag_dark, tag_white) VALUES ('Joggen', 'Running', '#fff', '#000');
INSERT INTO workout_type (name_de, name_en, tag_dark, tag_white) VALUES ('Surfen', 'Surf', '#fff', '#000');
INSERT INTO workout_type (name_de, name_en, tag_dark, tag_white) VALUES ('Segeln', 'Sailing', '#fff', '#000');
INSERT INTO workout_type (name_de, name_en, tag_dark, tag_white) VALUES ('Snowboarden', 'Snowboarding', '#fff', '#000');
INSERT INTO workout_type (name_de, name_en, tag_dark, tag_white) VALUES ('Schwimmen', 'Swimming', '#fff', '#000');
INSERT INTO workout_type (name_de, name_en, tag_dark, tag_white) VALUES ('Radfahren', 'Cycling', '#fff', '#000');
INSERT INTO workout_type (name_de, name_en, tag_dark, tag_white) VALUES ('Skateboarden', 'Skateboarding', '#fff', '#000');
INSERT INTO workout_type (name_de, name_en, tag_dark, tag_white) VALUES ('Volleyball', 'Volleyball', '#fff', '#000');

-- Geoname database dump
CREATE TABLE `geonames` (
    `geonameid` 		INT(11) NOT NULL,
    `name` 				VARCHAR(200) NOT NULL,
    `alternatenames` 	VARCHAR(4000) DEFAULT NULL,
    `location` 			POINT NOT NULL,
    `country` 			VARCHAR(2) NOT NULL,
    `population` 		INT(11) NOT NULL,
	`adm1`				VARCHAR(20),
	`adm2`				VARCHAR(20),
	`adm3`				VARCHAR(20),
	`adm4`				VARCHAR(20),
    PRIMARY KEY (`geonameid`),
    INDEX `name` (`name`),
    INDEX `country` (`country`),
    INDEX `population` (`population`),
	SPATIAL INDEX(location)
) ENGINE = InnoDB;
CREATE TABLE `geonames_adm` (
	`geonameid`			INT(11) NOT NULL PRIMARY KEY,
	`typ`				VARCHAR(5) NOT NULL,
	`value`				VARCHAR(20) NOT NULL,
	`name`				VARCHAR(200) NOT NULL,
	`alternatenames`	VARCHAR(4000) DEFAULT NULL,
	`adm0`				VARCHAR(3) NOT NULL,
	`adm1`				VARCHAR(20) NOT NULL,
	`adm2`				VARCHAR(20),
	`adm3`				VARCHAR(20),
	`root`				INT(11),

	INDEX `typ` (`typ`),
	INDEX `value` (`value`),
	INDEX `adm0` (`adm0`),
	INDEX `adm1` (`adm1`),
	INDEX `root` (`root`)
) ENGINE = InnoDB;

CREATE OR REPLACE VIEW v_geonames_all AS
  SELECT 
    g.*,
    NVL(g.name, adm4.name) AS display_name,
    adm3.name AS adm3_name,
    adm2.name AS adm2_name,
    adm1.name AS adm1_name
  FROM geonames g
  LEFT JOIN geonames_adm adm1
	  ON adm1.typ = 'ADM1' AND g.adm1 = adm1.value AND adm1.adm0 = g.country
  LEFT JOIN geonames_adm adm2
	  ON adm2.typ = 'ADM2' AND g.adm2 = adm2.value AND adm2.root = adm1.geonameid
  LEFT JOIN geonames_adm adm3
	  ON adm3.typ = 'ADM3' AND g.adm3 = adm3.value AND adm3.root = adm2.geonameid
  LEFT JOIN geonames_adm adm4
	  ON adm4.typ = 'ADM4' AND g.adm4 = adm4.value AND adm4.root = adm3.geonameid;


CREATE TABLE version (
	`release` 		VARCHAR(10) NOT NULL PRIMARY KEY,
	update_time		DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE = InnoDB;
INSERT INTO version(`release`) VALUES ('0.0.0');