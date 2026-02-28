DROP TRIGGER IF EXISTS movie_rating_stats_ins ON ratings;
DROP TRIGGER IF EXISTS movie_rating_stats_upd ON ratings;
DROP TRIGGER IF EXISTS movie_rating_stats_del ON ratings;

DROP FUNCTION IF EXISTS trg_movie_rating_stats_apply();

DROP TABLE IF EXISTS movie_rating_stats;
