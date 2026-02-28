CREATE TABLE IF NOT EXISTS movie_rating_stats (
    movie_id   INT PRIMARY KEY REFERENCES movies(movie_id) ON DELETE CASCADE,
    votes      INT NOT NULL DEFAULT 0 CHECK (votes >= 0),
    sum_rating DOUBLE PRECISION NOT NULL DEFAULT 0 CHECK (sum_rating >= 0),
    avg_rating DOUBLE PRECISION GENERATED ALWAYS AS (
        CASE WHEN votes > 0 THEN sum_rating / votes ELSE 0 END
    ) STORED,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);


INSERT INTO movie_rating_stats (movie_id, votes, sum_rating)
SELECT r.movie_id, COUNT(*)::int, SUM(r.rating)::float8
FROM ratings r
GROUP BY r.movie_id
ON CONFLICT (movie_id)
DO UPDATE SET
    votes      = EXCLUDED.votes,
    sum_rating = EXCLUDED.sum_rating,
    updated_at = now();


CREATE OR REPLACE FUNCTION trg_movie_rating_stats_apply()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO movie_rating_stats(movie_id, votes, sum_rating, updated_at)
        VALUES (NEW.movie_id, 1, NEW.rating, now())
        ON CONFLICT (movie_id)
        DO UPDATE SET
            votes      = movie_rating_stats.votes + 1,
            sum_rating = movie_rating_stats.sum_rating + NEW.rating,
            updated_at = now();
        RETURN NEW;

    ELSIF TG_OP = 'UPDATE' THEN
        UPDATE movie_rating_stats
        SET sum_rating = sum_rating + NEW.rating - OLD.rating,
            updated_at = now()
        WHERE movie_id = NEW.movie_id;
        RETURN NEW;

    ELSIF TG_OP = 'DELETE' THEN
        UPDATE movie_rating_stats
        SET votes      = votes - 1,
            sum_rating = sum_rating - OLD.rating,
            updated_at = now()
        WHERE movie_id = OLD.movie_id;

        DELETE FROM movie_rating_stats
        WHERE movie_id = OLD.movie_id AND votes <= 0;

        RETURN OLD;
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;


DROP TRIGGER IF EXISTS movie_rating_stats_ins ON ratings;
CREATE TRIGGER movie_rating_stats_ins
AFTER INSERT ON ratings
FOR EACH ROW
EXECUTE FUNCTION trg_movie_rating_stats_apply();

DROP TRIGGER IF EXISTS movie_rating_stats_upd ON ratings;
CREATE TRIGGER movie_rating_stats_upd
AFTER UPDATE OF rating ON ratings
FOR EACH ROW
EXECUTE FUNCTION trg_movie_rating_stats_apply();

DROP TRIGGER IF EXISTS movie_rating_stats_del ON ratings;
CREATE TRIGGER movie_rating_stats_del
AFTER DELETE ON ratings
FOR EACH ROW
EXECUTE FUNCTION trg_movie_rating_stats_apply();