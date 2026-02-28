package service

import (
	"context"
	"encoding/binary"
	"hash/fnv"
	"my_mdb/internal/domain"
	"time"
)

func (s *HomeService) pickChangingBlockDaily(ctx context.Context, userID int, limit int) (domain.ChangingBlock, error) {
	if limit <= 0 {
		limit = 10
	}

	loc, err := time.LoadLocation("Europe/Amsterdam")
	if err != nil {
		loc = time.Local
	}

	now := time.Now().In(loc)
	day := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	rng := newSplitmix64(dailySeed(userID, day))

	// 2 типа блоков: years / tags
	if rng.intn(2) == 0 {
		return s.buildChangingByYearRangeRNG(ctx, limit, rng)
	}
	return s.buildChangingByTagCategoryRNG(ctx, limit, rng)
}

func (s *HomeService) buildChangingByYearRangeRNG(ctx context.Context, limit int, rng *splitmix64) (domain.ChangingBlock, error) {
	p := yearRangePicks[rng.intn(len(yearRangePicks))]

	movies, err := s.movies.MoviesByYearRange(ctx, int16(p.from), int16(p.to), limit)
	if err != nil {
		return domain.ChangingBlock{}, err
	}

	return domain.ChangingBlock{
		Kind:   p.kind,
		Title:  p.title,
		Movies: toCards(movies),
	}, nil
}

func (s *HomeService) buildChangingByTagCategoryRNG(ctx context.Context, limit int, rng *splitmix64) (domain.ChangingBlock, error) {
	if s.tags == nil || len(tagCategories) == 0 {
		return s.buildChangingByYearRangeRNG(ctx, limit, rng)
	}

	cat := tagCategories[rng.intn(len(tagCategories))]
	p := cat.picks[rng.intn(len(cat.picks))]

	ids, err := s.tags.TopMoviesByTagQuery(ctx, p.query, limit)
	if err != nil {
		return domain.ChangingBlock{}, err
	}

	if len(ids) == 0 {
		return s.buildChangingByYearRangeRNG(ctx, limit, rng)
	}

	movies, err := s.movies.GetMoviesByIDs(ctx, ids)
	if err != nil {
		return domain.ChangingBlock{}, err
	}

	return domain.ChangingBlock{
		Kind:   "tag_" + p.kind,
		Title:  p.title,
		Movies: toCards(movies),
	}, nil
}

type splitmix64 struct{ x uint64 }

func newSplitmix64(seed uint64) *splitmix64 { return &splitmix64{x: seed} }

func (s *splitmix64) next() uint64 {
	s.x += 0x9e3779b97f4a7c15
	z := s.x
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return z ^ (z >> 31)
}

func (s *splitmix64) intn(n int) int {
	if n <= 0 {
		return 0
	}
	return int(s.next() % uint64(n))
}

func dailySeed(userID int, day time.Time) uint64 {
	h := fnv.New64a()

	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], uint64(userID))
	_, _ = h.Write(buf[:])

	yyyymmdd := uint64(day.Year()*10000 + int(day.Month())*100 + day.Day())
	binary.LittleEndian.PutUint64(buf[:], yyyymmdd)
	_, _ = h.Write(buf[:])

	return h.Sum64()
}
