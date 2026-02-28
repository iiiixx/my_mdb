ALTER TABLE movie_rating_stats
ADD COLUMN weighted_score DOUBLE PRECISION;


WITH global_avg AS (
    SELECT AVG(avg_rating) AS c
    FROM movie_rating_stats
    WHERE votes > 0
)


UPDATE movie_rating_stats m
SET weighted_score =
    (
        (m.votes::float / (m.votes + 250)) * m.avg_rating +
        (250::float / (m.votes + 250)) * (SELECT c FROM global_avg)
    )
WHERE m.votes > 0;


CREATE INDEX idx_movie_rating_stats_weighted ON movie_rating_stats (weighted_score DESC);