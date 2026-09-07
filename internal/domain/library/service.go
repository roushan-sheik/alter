package library

import (
	"errors"

	"gotickets/internal/domain/library/dto"
	"gotickets/internal/querybuilder"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrLibraryNotFound  = errors.New("library item not found")
	ErrCategoryNotFound = errors.New("library category not found")
)

type Service interface {
	Create(req dto.CreateLibraryReq) (*dto.LibraryResponse, error)
	FindAll(params querybuilder.Params) (*dto.PaginatedLibraryResponse, error)
	GetDetails(id uuid.UUID) (*dto.LibraryDetailsResponse, error)
	Update(id uuid.UUID, req dto.UpdateLibraryReq) (*dto.LibraryResponse, error)
	Delete(id uuid.UUID) error

	CreateCategory(req dto.CreateCategoryReq) (*dto.CategoryResponse, error)
	FindAllCategories() ([]*dto.CategoryResponse, error)
	UpdateCategory(id uuid.UUID, req dto.UpdateCategoryReq) (*dto.CategoryResponse, error)
	DeleteCategory(id uuid.UUID) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(req dto.CreateLibraryReq) (*dto.LibraryResponse, error) {
	item := &LibraryItem{
		Title:            req.Title,
		Category:         req.Category,
		ShortDescription: req.ShortDescription,
		ContentText:      req.ContentText,
		ThumbnailURL:     req.ThumbnailURL,
		MediaURL:         req.MediaURL,
	}

	if err := s.repo.Create(item); err != nil {
		return nil, err
	}

	return toDTO(item), nil
}

func (s *service) FindAll(params querybuilder.Params) (*dto.PaginatedLibraryResponse, error) {
	items, meta, err := s.repo.FindAll(params)
	if err != nil {
		return nil, err
	}

	data := make([]*dto.LibraryResponse, 0, len(items))
	for _, item := range items {
		data = append(data, toDTO(item))
	}
	return &dto.PaginatedLibraryResponse{Data: data, Meta: meta}, nil
}

func (s *service) GetDetails(id uuid.UUID) (*dto.LibraryDetailsResponse, error) {
	item, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrLibraryNotFound
		}
		return nil, err
	}

	related, err := s.repo.FindRelated(id, 5)
	if err != nil {
		return nil, err
	}

	relatedDTOs := make([]*dto.LibraryResponse, 0, len(related))
	for _, r := range related {
		relatedDTOs = append(relatedDTOs, toDTO(r))
	}

	return &dto.LibraryDetailsResponse{
		LibraryItem: toDTO(item),
		Related:     relatedDTOs,
	}, nil
}

func (s *service) Update(id uuid.UUID, req dto.UpdateLibraryReq) (*dto.LibraryResponse, error) {
	item, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrLibraryNotFound
		}
		return nil, err
	}

	item.Title = req.Title
	item.Category = req.Category
	item.ShortDescription = req.ShortDescription
	item.ContentText = req.ContentText
	item.ThumbnailURL = req.ThumbnailURL
	item.MediaURL = req.MediaURL

	if err := s.repo.Update(item); err != nil {
		return nil, err
	}

	return toDTO(item), nil
}

func (s *service) Delete(id uuid.UUID) error {
	return s.repo.Delete(id)
}

func toDTO(item *LibraryItem) *dto.LibraryResponse {
	return &dto.LibraryResponse{
		ID:               item.ID,
		Title:            item.Title,
		Category:         item.Category,
		ShortDescription: item.ShortDescription,
		ContentText:      item.ContentText,
		ThumbnailURL:     item.ThumbnailURL,
		MediaURL:         item.MediaURL,
		CreatedAt:        item.CreatedAt,
		UpdatedAt:        item.UpdatedAt,
	}
}

func (s *service) CreateCategory(req dto.CreateCategoryReq) (*dto.CategoryResponse, error) {
	cat := &LibraryCategory{
		Name: req.Name,
	}
	if err := s.repo.CreateCategory(cat); err != nil {
		return nil, err
	}
	return toCategoryDTO(cat), nil
}

func (s *service) FindAllCategories() ([]*dto.CategoryResponse, error) {
	cats, err := s.repo.FindAllCategories()
	if err != nil {
		return nil, err
	}
	res := make([]*dto.CategoryResponse, 0, len(cats))
	for _, c := range cats {
		res = append(res, toCategoryDTO(c))
	}
	return res, nil
}

func (s *service) UpdateCategory(id uuid.UUID, req dto.UpdateCategoryReq) (*dto.CategoryResponse, error) {
	cat, err := s.repo.FindCategoryByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}
	cat.Name = req.Name
	if err := s.repo.UpdateCategory(cat); err != nil {
		return nil, err
	}
	return toCategoryDTO(cat), nil
}

func (s *service) DeleteCategory(id uuid.UUID) error {
	return s.repo.DeleteCategory(id)
}

func toCategoryDTO(cat *LibraryCategory) *dto.CategoryResponse {
	return &dto.CategoryResponse{
		ID:   cat.ID,
		Name: cat.Name,
	}
}
