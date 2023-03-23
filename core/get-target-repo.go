package core

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/storage/memory"
)

var mono *git.Repository

func GetTargetRepoLoader(target string) *InMemoryRepository {
	repo := GetTargetRepo()
	return &InMemoryRepository{target, repo, nil}
}

func GetTargetRepo() *git.Repository {
	if mono != nil {
		return mono
	}
	repo, _ := NewRepo()
	mono = repo
	return repo
}

type InMemoryRepository struct {
	Target string

	Repo *git.Repository

	EndPoint *transport.Endpoint
}

func (m *InMemoryRepository) Load(ep *transport.Endpoint) (storer.Storer, error) {
	fmt.Println("getting store of repo:", ep.String(), ep.Path, ep == m.EndPoint)
	return m.Repo.Storer, nil
}

func NewRepo() (*git.Repository, error) {
	target := memory.NewStorage()
	repo, err := git.Init(target, nil)

	h := plumbing.NewSymbolicReference(plumbing.HEAD, "refs/heads/main")
	if err := target.SetReference(h); err != nil {
		return nil, err
	}

	return repo, err
}
