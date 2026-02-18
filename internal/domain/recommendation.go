package domain

type RecommendationItem struct {
	MovieID int     `json:"movie_id"`
	Score   float32 `json:"score"`
}

type SimilarItem struct {
	MovieID        int     `json:"movie_id"`
	SimilarMovieID int     `json:"similar_movie_id"`
	Score          float32 `json:"score"`
}
