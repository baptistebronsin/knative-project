package domain

import (
	"github.com/google/uuid"
)

func generateUniqueID() string {
	return uuid.New().String()
}

type CommentService interface {
	GetComments() ([]Comment, error)
	GetComment(id string) (Comment, error)
	CreateComment(content string, emotion string) (Comment, error)
}

type CommentServiceImpl struct {
	CommentRepository CommentRepository
}

func (s *CommentServiceImpl) GetComments() ([]Comment, error) {
	return s.CommentRepository.GetAll()
}

func (s *CommentServiceImpl) GetComment(id string) (Comment, error) {
	comment, err := s.CommentRepository.GetById(id)
	if err != nil {
		return Comment{}, err
	}

	return *comment, nil
}

func (s *CommentServiceImpl) CreateComment(content string, emotion string) (Comment, error) {
	comment := Comment{
		ID:      generateUniqueID(),
		Content: content,
		Emotion: emotion,
	}

	err := s.CommentRepository.Save(&comment)
	if err != nil {
		return Comment{}, err
	}

	return comment, nil
}
