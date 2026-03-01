package grpc

import (
	"context"
	"fmt"

	"my_mdb/internal/domain"
	pb "my_mdb/protos/gen/recs"
)

func (x *Client) Recommend(ctx context.Context, userID int, limit int, excludeMovieIDs []int) ([]domain.RecommendationItem, error) {
	if userID <= 0 {
		return []domain.RecommendationItem{}, nil
	}
	if limit <= 0 {
		limit = 20
	}

	resp, err := x.c.Recommend(ctx, &pb.RecommendRequest{
		UserId:          int32(userID),
		Limit:           int32(limit),
		ExcludeMovieIds: toInt32s(excludeMovieIDs),
	})
	if err != nil {
		return nil, fmt.Errorf("recgrpc: recommend: %w", err)
	}

	out := make([]domain.RecommendationItem, 0, len(resp.Items))
	for _, it := range resp.Items {
		out = append(out, domain.RecommendationItem{
			MovieID: int(it.MovieId),
			Score:   float32(it.Score),
		})
	}
	return out, nil
}

func toInt32s(in []int) []int32 {
	if len(in) == 0 {
		return nil
	}
	out := make([]int32, 0, len(in))
	for _, v := range in {
		if v > 0 {
			out = append(out, int32(v))
		}
	}
	return out
}
