package grpc

import (
	"context"
	"fmt"

	"my_mdb/internal/domain"
	pb "my_mdb/protos/gen/recs"
)

func (x *Client) SimilarMovies(ctx context.Context, movieID int, limit int) ([]domain.SimilarItem, error) {
	if movieID <= 0 {
		return []domain.SimilarItem{}, nil
	}
	if limit <= 0 {
		limit = 20
	}

	resp, err := x.c.SimilarMovies(ctx, &pb.SimilarMoviesRequest{
		MovieId: int32(movieID),
		Limit:   int32(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("recgrpc: similar_movies: %w", err)
	}

	out := make([]domain.SimilarItem, 0, len(resp.Items))
	for _, it := range resp.Items {
		out = append(out, domain.SimilarItem{
			MovieID:        movieID,
			SimilarMovieID: int(it.MovieId),
			Score:          float32(it.Similarity),
		})
	}

	return out, nil
}
