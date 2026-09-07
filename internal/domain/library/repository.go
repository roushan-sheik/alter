package library

import (
	"gotickets/internal/querybuilder"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(item *LibraryItem) error
	FindAll(params querybuilder.Params) ([]*LibraryItem, *querybuilder.Meta, error)
	FindByID(id uuid.UUID) (*LibraryItem, error)
	FindRelated(excludeID uuid.UUID, limit int) ([]*LibraryItem, error)
	Update(item *LibraryItem) error
	Delete(id uuid.UUID) error

	CreateCategory(category *LibraryCategory) error
	FindAllCategories() ([]*LibraryCategory, error)
	FindCategoryByID(id uuid.UUID) (*LibraryCategory, error)
	UpdateCategory(category *LibraryCategory) error
	DeleteCategory(id uuid.UUID) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(item *LibraryItem) error {
	return r.db.Create(item).Error
}

func (r *repository) FindAll(params querybuilder.Params) ([]*LibraryItem, *querybuilder.Meta, error) {
	var items []*LibraryItem

	meta, err := querybuilder.New(r.db.Model(&LibraryItem{}), params).
		Search([]string{"title", "short_description", "category"}).
		Filter().
		Sort().
		Paginate().
		ExecuteWithMeta(&items)

	if err != nil {
		return nil, nil, err
	}
	return items, meta, nil
}

func (r *repository) FindByID(id uuid.UUID) (*LibraryItem, error) {
	var item LibraryItem
	err := r.db.Where("id = ?", id).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *repository) FindRelated(excludeID uuid.UUID, limit int) ([]*LibraryItem, error) {
	var items []*LibraryItem
	err := r.db.Where("id != ?", excludeID).Order("created_at desc").Limit(limit).Find(&items).Error
	return items, err
}

func (r *repository) Update(item *LibraryItem) error {
	return r.db.Save(item).Error
}

func (r *repository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&LibraryItem{}).Error
}

func (r *repository) CreateCategory(category *LibraryCategory) error {
	return r.db.Create(category).Error
}

func (r *repository) FindAllCategories() ([]*LibraryCategory, error) {
	var categories []*LibraryCategory
	err := r.db.Order("name asc").Find(&categories).Error
	return categories, err
}

func (r *repository) FindCategoryByID(id uuid.UUID) (*LibraryCategory, error) {
	var category LibraryCategory
	err := r.db.Where("id = ?", id).First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *repository) UpdateCategory(category *LibraryCategory) error {
	return r.db.Save(category).Error
}

func (r *repository) DeleteCategory(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&LibraryCategory{}).Error
}
