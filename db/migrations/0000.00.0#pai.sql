DELIMITER $$

CREATE OR REPLACE PROCEDURE PopulateYearDay(
	IN start_date DATE,
    IN end_date   DATE
)
BEGIN
    DECLARE cur_date DATE DEFAULT start_date;
    DECLARE counter INT DEFAULT 1;

	TRUNCATE TABLE year_day;

    WHILE cur_date <= end_date DO
        INSERT INTO year_day (`id`, `start`, `end`, `start_offset`, `end_offset`, `day_year`, `day_week`)
        VALUES (
            counter, 
            cur_date,
            CONCAT(cur_date, ' 23:59:59'),
            DATE_ADD(cur_date, INTERVAL -14 HOUR),
            DATE_ADD(CONCAT(cur_date, ' 23:59:59'), INTERVAL 12 HOUR),
            DAYOFYEAR(cur_date),
            WEEKDAY(cur_date)
        );

        SET cur_date = DATE_ADD(cur_date, INTERVAL 1 DAY);
        SET counter = counter + 1;
    END WHILE;

    COMMIT;
END$$

CREATE OR REPLACE VIEW v_user_timezone AS
	SELECT
		u.id,
		u.timezone,
		TIME_TO_SEC( TIMEDIFF(CONVERT_TZ(UTC_TIMESTAMP(), 'UTC', u.timezone), UTC_TIMESTAMP()) ) AS `offset`
	FROM `user` u
$$

CREATE OR REPLACE VIEW v_year_day_user_offset AS
	SELECT 
		yd.id,
		yd.start,
		yd.end,
		yd.start_offset,
		yd.end_offset,
		DATE_ADD(yd.start, INTERVAL ut.offset SECOND) AS user_start_offset,
		DATE_ADD(yd.end, INTERVAL ut.offset SECOND) AS user_end_offset,
		ut.id AS user_id
	FROM year_day yd 
	CROSS JOIN v_user_timezone ut
$$

CREATE OR REPLACE VIEW pai_daily AS
	SELECT 
		yd.id,
		NVL(s.steps, 0) AS steps_total,
		NVL(workout.steps, 0)  AS steps_workout,
		NVL(workout.pai, 0) AS workout_pai,
		(CASE
			WHEN NVL(s.steps, 0) - NVL(workout.steps, 0) >= 30000 THEN 10
			WHEN NVL(s.steps, 0) - NVL(workout.steps, 0) >= 20000 THEN 5
			WHEN NVL(s.steps, 0) - NVL(workout.steps, 0) >= 10000 THEN 2
			ELSE 0
		END) steps_pai,
		NVL(s.user_id, workout.user_id) AS user_id
	FROM year_day yd
	-- This isn't totally correct because a workout could not have steps tracked. But we can't relay
	-- on the start and end time of the workout because no steps in the pauses / when workout got merged
	-- are counted. Checking the workout details would be too slow so this is the only solution
	LEFT JOIN (
		SELECT
			yd.id,
			SUM(w.steps) AS steps,
			SUM(w.pai) AS pai,
			w.user_id
		FROM year_day yd
		INNER JOIN ( 
			SELECT 
				w.*, u.offset
		  	FROM workout w 
		  	INNER JOIN v_user_timezone u ON u.id = w.user_id
		) w ON w.start > yd.start_offset AND w.start < yd.end_offset AND w.`start` > DATE_ADD(yd.`start`, INTERVAL w.offset SECOND) AND w.`start` < DATE_ADD(yd.`end`, INTERVAL w.offset SECOND)
		GROUP BY yd.id, w.user_id
	) workout ON workout.id = yd.id
	LEFT JOIN (
		SELECT
			yd.id,
			SUM(s.count) AS steps,
			s.user_id
		FROM year_day yd
		INNER JOIN (
			SELECT s.*, u.offset
			FROM steps s
			INNER JOIN v_user_timezone u ON u.id = s.user_id
		) s ON s.start > yd.start_offset AND s.end < yd.end_offset AND s.`start` > DATE_ADD(yd.`start`, INTERVAL s.offset SECOND) AND s.`start` < DATE_ADD(yd.`end`, INTERVAL s.offset SECOND)
		-- Because of the previously mentioned problem, only use workouts with small pauses. But this operation is heavy!
		-- LEFT JOIN workout w ON s.`start` > w.`start` AND s.`start` < w.end AND w.user_id = s.user_id AND TIMESTAMPDIFF(SECOND, w.`start`, w.end) - w.duration < w.duration * 0. AND w.steps = 0
		-- WHERE w.id IS NULL
		GROUP BY yd.id, s.user_id
	) s ON s.id = yd.id AND (workout.user_id IS NULL OR workout.user_id = s.user_id);
$$

DELIMITER ;