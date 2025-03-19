package domain

import "fmt"

type Like struct {
	ID        string
	CommentId string
	Timestamp int
}

type LikeRepository interface {
	GetAll() ([]Like, error)
	GetById(id string) (*Like, error)
	Save(Like *Like) error
}

type InMemoryLikeRepository struct {
	Like map[string](*Like)
}

func (r *InMemoryLikeRepository) GetAll() ([]Like, error) {
	Likes := []Like{}
	for _, student := range r.Like {
		Likes = append(Likes, *student)
	}
	return Likes, nil
}

func (r *InMemoryLikeRepository) GetById(id string) (*Like, error) {
	Like, ok := r.Like[id]
	if !ok {
		return nil, fmt.Errorf("There is no like with '%s' ID", id)
	}
	return Like, nil
}

func (r *InMemoryLikeRepository) Save(Like *Like) error {
	r.Like[Like.ID] = Like
	return nil
}

func NewInMemoryLikeRepository() *InMemoryLikeRepository {
	return &InMemoryLikeRepository{
		Like: make(map[string](*Like)),
	}
}
