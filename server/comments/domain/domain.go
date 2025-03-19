package domain

import "fmt"

type Comment struct {
	ID      string
	Content string
	Emotion string
}

type CommentRepository interface {
	GetAll() ([]Comment, error)
	GetById(id string) (*Comment, error)
	Save(comment *Comment) error
}

type InMemoryCommentRepository struct {
	Comment map[string](*Comment)
}

func (r *InMemoryCommentRepository) GetAll() ([]Comment, error) {
	comments := []Comment{}
	for _, student := range r.Comment {
		comments = append(comments, *student)
	}
	return comments, nil
}

func (r *InMemoryCommentRepository) GetById(id string) (*Comment, error) {
	comment, ok := r.Comment[id]
	if !ok {
		return nil, fmt.Errorf("There is no comment with '%s' ID", id)
	}
	return comment, nil
}

func (r *InMemoryCommentRepository) Save(comment *Comment) error {
	r.Comment[comment.ID] = comment
	return nil
}

func NewInMemoryCommentRepository() *InMemoryCommentRepository {
	return &InMemoryCommentRepository{
		Comment: make(map[string](*Comment)),
	}
}
