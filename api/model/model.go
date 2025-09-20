package model

import (
	"context"
	"fmt"
	"log"
)

type StringFilterOp string

const (
	StringFilterOp_EQ    StringFilterOp = "eq"
	StringFilterOp_Like  StringFilterOp = "like"
	StringFilterOp_ILike StringFilterOp = "ilike"
)

type StringFilter struct {
	Op    StringFilterOp `json:"op"`
	Value *string        `json:"value"`
}

type SortDir string

const (
	SortDir_ASC  SortDir = "asc"
	SortDir_DESC SortDir = "desc"
)

type Repos interface {
	User() UserRepository
}

type ReposWithCtx func(ctx context.Context) (Repos, context.Context, error)

var reposMap = map[string]ReposWithCtx{}

func Register(id string, repos ReposWithCtx) {
	reposMap[id] = repos
}

func GetRepos(id string, ctx context.Context) (Repos, context.Context, error) {
	repos := reposMap[id]
	if repos == nil {
		log.Fatal(fmt.Errorf("Repos '%s' not initialized", id))
	}
	return repos(ctx)
}
