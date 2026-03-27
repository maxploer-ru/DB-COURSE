package service

import (
	"ZVideo/internal/domain/comment/entity"
	"ZVideo/internal/domain/comment/repository"
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrCommentNotFound = errors.New("comment not found")
	ErrEmptyContent    = errors.New("comment content cannot be empty")
	ErrTooLongContent  = errors.New("comment content too long")
	ErrVideoNotFound   = errors.New("video not found")
	ErrNotAuthorized   = errors.New("not authorized")
	ErrRatingNotFound  = errors.New("rating not found")
)

type VideoChecker interface {
	Exists(ctx context.Context, videoID int) (bool, error)
}

type CommentService interface {
	CreateComment(ctx context.Context, userID, videoID int, content string) (*entity.Comment, error)
	GetComment(ctx context.Context, commentID int) (*entity.Comment, error)
	GetCommentsByVideo(ctx context.Context, videoID int, limit, offset int) ([]*entity.Comment, error)
	UpdateComment(ctx context.Context, commentID, userID int, content string) (*entity.Comment, error)
	DeleteComment(ctx context.Context, commentID, userID int) error
	RateComment(ctx context.Context, userID, commentID int, liked bool) error
	GetCommentRatingStats(ctx context.Context, commentID int) (likes, dislikes int, err error)
}

type commentService struct {
	commentRepo  repository.CommentRepository
	ratingRepo   repository.CommentRatingRepository
	videoChecker VideoChecker
}

func NewCommentService(
	commentRepo repository.CommentRepository,
	ratingRepo repository.CommentRatingRepository,
	videoChecker VideoChecker,
) CommentService {
	return &commentService{
		commentRepo:  commentRepo,
		ratingRepo:   ratingRepo,
		videoChecker: videoChecker,
	}
}

func (s *commentService) CreateComment(ctx context.Context, userID, videoID int, content string) (*entity.Comment, error) {
	if content == "" {
		return nil, ErrEmptyContent
	}
	if len(content) > 1000 {
		return nil, ErrTooLongContent
	}
	exists, err := s.videoChecker.Exists(ctx, videoID)
	if err != nil {
		return nil, fmt.Errorf("check video: %w", err)
	}
	if !exists {
		return nil, ErrVideoNotFound
	}

	comment := &entity.Comment{
		UserID:    userID,
		VideoID:   videoID,
		Content:   content,
		CreatedAt: time.Now(),
	}
	if err := s.commentRepo.Create(ctx, comment); err != nil {
		return nil, fmt.Errorf("create comment: %w", err)
	}
	return comment, nil
}

func (s *commentService) GetComment(ctx context.Context, commentID int) (*entity.Comment, error) {
	comment, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		return nil, fmt.Errorf("get comment: %w", err)
	}
	if comment == nil {
		return nil, ErrCommentNotFound
	}
	return comment, nil
}

func (s *commentService) GetCommentsByVideo(ctx context.Context, videoID int, limit, offset int) ([]*entity.Comment, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return s.commentRepo.GetByVideoID(ctx, videoID, limit, offset)
}

func (s *commentService) UpdateComment(ctx context.Context, commentID, userID int, content string) (*entity.Comment, error) {
	comment, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		return nil, fmt.Errorf("get comment: %w", err)
	}
	if comment == nil {
		return nil, ErrCommentNotFound
	}
	if comment.UserID != userID {
		return nil, ErrNotAuthorized
	}
	if content == "" {
		return nil, ErrEmptyContent
	}
	if len(content) > 1000 {
		return nil, ErrTooLongContent
	}
	comment.Content = content
	if err := s.commentRepo.Update(ctx, comment); err != nil {
		return nil, fmt.Errorf("update comment: %w", err)
	}
	return comment, nil
}

func (s *commentService) DeleteComment(ctx context.Context, commentID, userID int) error {
	comment, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		return fmt.Errorf("get comment: %w", err)
	}
	if comment == nil {
		return ErrCommentNotFound
	}
	if comment.UserID != userID {
		return ErrNotAuthorized
	}
	if err := s.commentRepo.Delete(ctx, commentID); err != nil {
		return fmt.Errorf("delete comment: %w", err)
	}
	return nil
}

func (s *commentService) RateComment(ctx context.Context, userID, commentID int, liked bool) error {
	comment, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		return fmt.Errorf("get comment: %w", err)
	}
	if comment == nil {
		return ErrCommentNotFound
	}

	existing, err := s.ratingRepo.GetByUserAndComment(ctx, userID, commentID)
	if err != nil && !errors.Is(err, ErrRatingNotFound) {
		return fmt.Errorf("get rating: %w", err)
	}

	if existing == nil {
		rating := &entity.CommentRating{
			UserID:    userID,
			CommentID: commentID,
			Liked:     liked,
			RatedAt:   time.Now(),
		}
		return s.ratingRepo.Create(ctx, rating)
	}

	if existing.Liked == liked {
		return nil
	}
	existing.Liked = liked
	existing.RatedAt = time.Now()
	return s.ratingRepo.Update(ctx, existing)
}

func (s *commentService) GetCommentRatingStats(ctx context.Context, commentID int) (likes, dislikes int, err error) {
	return s.ratingRepo.GetCommentRatingStats(ctx, commentID)
}
