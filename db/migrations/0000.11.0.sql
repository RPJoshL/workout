ALTER TABLE workout
	ADD CONSTRAINT un_workout_user_start_end UNIQUE INDEX workout (user_id, `start`, `end`);