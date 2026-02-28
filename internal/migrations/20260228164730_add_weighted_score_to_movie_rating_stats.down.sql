DROP INDEX IF EXISTS idx_movie_rating_stats_weighted;

ALTER TABLE movie_rating_stats
DROP COLUMN IF EXISTS weighted_score;