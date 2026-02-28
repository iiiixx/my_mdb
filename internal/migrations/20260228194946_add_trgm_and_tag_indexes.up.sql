CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX IF NOT EXISTS genome_tags_tag_trgm_idx ON genome_tags USING gin (tag gin_trgm_ops);
CREATE INDEX IF NOT EXISTS genome_scores_tag_id_idx ON genome_scores(tag_id);
