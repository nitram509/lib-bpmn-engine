package tests

import (
	"context"

	"github.com/rqlite/rqlite/v8/command/proto"
)

type TestStorage struct {
}

func (s *TestStorage) Query(ctx context.Context, req *proto.QueryRequest) ([]*proto.QueryRows, error) {
	return nil, nil
}

func (s *TestStorage) Execute(ctx context.Context, req *proto.ExecuteRequest) ([]*proto.ExecuteQueryResponse, error) {
	return nil, nil
}

func (s *TestStorage) IsLeader(ctx context.Context) bool {
	return true
}
