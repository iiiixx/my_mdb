package grpc

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "my_mdb/protos/gen/recs"
)

type DialOptions struct {
	DialTimeout time.Duration
	DialOptions []grpc.DialOption
}

func defaultDialOptions() DialOptions {
	return DialOptions{
		DialTimeout: 5 * time.Second,
		DialOptions: []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
		},
	}
}

type Client struct {
	cc *grpc.ClientConn
	c  pb.RecommenderClient
}

func Dial(addr string, opt ...func(*DialOptions)) (*Client, error) {
	if addr == "" {
		return nil, fmt.Errorf("recgrpc: addr is empty")
	}

	o := defaultDialOptions()
	for _, apply := range opt {
		apply(&o)
	}

	ctx, cancel := context.WithTimeout(context.Background(), o.DialTimeout)
	defer cancel()

	cc, err := grpc.DialContext(ctx, addr, o.DialOptions...)
	if err != nil {
		return nil, fmt.Errorf("recgrpc: dial: %w", err)
	}

	return &Client{
		cc: cc,
		c:  pb.NewRecommenderClient(cc),
	}, nil
}

func (x *Client) Close() error {
	if x.cc != nil {
		return x.cc.Close()
	}
	return nil
}

func WithDialTimeout(d time.Duration) func(*DialOptions) {
	return func(o *DialOptions) { o.DialTimeout = d }
}
