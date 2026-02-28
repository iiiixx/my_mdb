CREATE OR REPLACE FUNCTION trg_movie_rating_stats_apply()
RETURNS TRIGGER AS $$
DECLARE
    m CONSTANT double precision := 250.0;
    c CONSTANT double precision := 3.5;
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO movie_rating_stats(movie_id, votes, sum_rating, avg_rating, weighted_score, updated_at)
        VALUES (NEW.movie_id, 1, NEW.rating, NEW.rating,
                (1.0 / (1.0 + m)) * NEW.rating + (m / (1.0 + m)) * c,
                now())
        ON CONFLICT (movie_id)
        DO UPDATE SET
            votes      = movie_rating_stats.votes + 1,
            sum_rating = movie_rating_stats.sum_rating + NEW.rating,
            avg_rating = (movie_rating_stats.sum_rating + NEW.rating) / (movie_rating_stats.votes + 1),
            weighted_score =
                ((movie_rating_stats.votes + 1)::double precision /
                 ((movie_rating_stats.votes + 1)::double precision + m)) *
                ((movie_rating_stats.sum_rating + NEW.rating) /
                 (movie_rating_stats.votes + 1)) +
                (m / ((movie_rating_stats.votes + 1)::double precision + m)) * c,
            updated_at = now();

        RETURN NEW;

    ELSIF TG_OP = 'UPDATE' THEN
        UPDATE movie_rating_stats
        SET sum_rating = sum_rating + NEW.rating - OLD.rating,
            avg_rating = (sum_rating + NEW.rating - OLD.rating) / votes,
            weighted_score =
                (votes::double precision / (votes::double precision + m)) *
                ((sum_rating + NEW.rating - OLD.rating) / votes) +
                (m / (votes::double precision + m)) * c,
            updated_at = now()
        WHERE movie_id = NEW.movie_id;

        RETURN NEW;

    ELSIF TG_OP = 'DELETE' THEN
        UPDATE movie_rating_stats
        SET votes      = votes - 1,
            sum_rating = sum_rating - OLD.rating,
            avg_rating = CASE
                            WHEN votes - 1 > 0 THEN (sum_rating - OLD.rating) / (votes - 1)
                            ELSE NULL
                         END,
            weighted_score = CASE
                                WHEN votes - 1 > 0 THEN
                                    ((votes - 1)::double precision /
                                     ((votes - 1)::double precision + m)) *
                                    ((sum_rating - OLD.rating) / (votes - 1)) +
                                    (m / ((votes - 1)::double precision + m)) * c
                                ELSE NULL
                             END,
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