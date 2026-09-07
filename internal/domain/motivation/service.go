package motivation

import (
	"errors"

	"gotickets/internal/domain/motivation/dto"
	"gotickets/internal/querybuilder"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrMotivationNotFound = errors.New("motivation not found")
)

type Service interface {
	Create(req dto.CreateMotivationReq) (*dto.MotivationResponse, error)
	FindAll(params querybuilder.Params) (*dto.PaginatedMotivationResponse, error)
	FindByID(id uuid.UUID) (*dto.MotivationResponse, error)
	GetDetails(id uuid.UUID) (*dto.MotivationDetailsResponse, error)
	Update(id uuid.UUID, req dto.UpdateMotivationReq) (*dto.MotivationResponse, error)
	Delete(id uuid.UUID) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(req dto.CreateMotivationReq) (*dto.MotivationResponse, error) {
	m := &Motivation{
		Title:        req.Title,
		SpeakerName:  req.SpeakerName,
		Description:  req.Description,
		VideoURL:     req.VideoURL,
		ThumbnailURL: req.ThumbnailURL,
		Duration:     req.Duration,
	}

	if err := s.repo.Create(m); err != nil {
		return nil, err
	}

	return toDTO(m), nil
}

func (s *service) FindAll(params querybuilder.Params) (*dto.PaginatedMotivationResponse, error) {
	motivations, meta, err := s.repo.FindAll(params)
	if err != nil {
		return nil, err
	}

	data := make([]*dto.MotivationResponse, 0, len(motivations))
	for _, m := range motivations {
		data = append(data, toDTO(m))
	}
	return &dto.PaginatedMotivationResponse{Data: data, Meta: meta}, nil
}

func (s *service) FindByID(id uuid.UUID) (*dto.MotivationResponse, error) {
	m, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMotivationNotFound
		}
		return nil, err
	}
	return toDTO(m), nil
}

func (s *service) GetDetails(id uuid.UUID) (*dto.MotivationDetailsResponse, error) {
	m, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMotivationNotFound
		}
		return nil, err
	}

	related, err := s.repo.FindRelated(id, 5)
	if err != nil {
		return nil, err
	}

	relatedDTOs := make([]*dto.MotivationResponse, 0, len(related))
	for _, rm := range related {
		relatedDTOs = append(relatedDTOs, toDTO(rm))
	}

	return &dto.MotivationDetailsResponse{
		Motivation: toDTO(m),
		Related:    relatedDTOs,
	}, nil
}

func (s *service) Update(id uuid.UUID, req dto.UpdateMotivationReq) (*dto.MotivationResponse, error) {
	m, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMotivationNotFound
		}
		return nil, err
	}

	m.Title = req.Title
	m.SpeakerName = req.SpeakerName
	m.Description = req.Description
	m.VideoURL = req.VideoURL
	m.ThumbnailURL = req.ThumbnailURL
	m.Duration = req.Duration

	if err := s.repo.Update(m); err != nil {
		return nil, err
	}

	return toDTO(m), nil
}

func (s *service) Delete(id uuid.UUID) error {
	return s.repo.Delete(id)
}

func toDTO(m *Motivation) *dto.MotivationResponse {
	return &dto.MotivationResponse{
		ID:           m.ID,
		Title:        m.Title,
		SpeakerName:  m.SpeakerName,
		Description:  m.Description,
		VideoURL:     m.VideoURL,
		ThumbnailURL: m.ThumbnailURL,
		Duration:     m.Duration,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}
