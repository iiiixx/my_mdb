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