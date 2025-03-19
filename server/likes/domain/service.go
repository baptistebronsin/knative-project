package domain

import (
	"time"

	"github.com/google/uuid"
)

func generateUniqueID() string {
	return uuid.New().String()
}

type LikeService interface {
	GetLikes() ([]Like, error)
	GetLike(id string) (Like, error)
	CreateLike(commentId string) (Like, error)
}

type LikeServiceImpl struct {
	LikeRepository LikeRepository
}

func (s *LikeServiceImpl) GetLikes() ([]Like, error) {
	return s.LikeRepository.GetAll()
}

func (s *LikeServiceImpl) GetLike(id string) (Like, error) {
	like, err := s.LikeRepository.GetById(id)
	if err != nil {
		return Like{}, err
	}

	return *like, nil
}

func (s *LikeServiceImpl) CreateLike(commentId string) (Like, error) {
	like := Like{
		ID:        generateUniqueID(),
		CommentId: commentId,
		Timestamp: int(time.Now().Unix()),
	}

	err := s.LikeRepository.Save(&like)
	if err != nil {
		return Like{}, err
	}

	return like, nil
}
