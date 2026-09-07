package library

import (
	"errors"
	"net/http"

	"gotickets/internal/domain/library/dto"
	"gotickets/internal/httpresponse"
	"gotickets/internal/querybuilder"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// FindAll godoc
// @Summary      Get all library items
// @Description  Retrieves a paginated list of library items. Supports search, sort, category filtering, and pagination via query params.
// @Tags         Library
// @Accept       json
// @Produce      json
// @Param        search    query  string  false  "Search by title, category, or description"
// @Param        category  query  string  false  "Filter by category"
// @Param        page      query  int     false  "Page number (default: 1)"
// @Param        limit     query  int     false  "Items per page (default: 10)"
// @Param        sort      query  string  false  "Sort column; prefix with - for DESC (e.g. -created_at)"
// @Success      200  {object}  dto.PaginatedLibraryResponse
// @Router       /api/v1/library [get]
func (h *Handler) FindAll(c *echo.Context) error {
	params := make(querybuilder.Params)
	for key, vals := range c.Request().URL.Query() {
		if len(vals) > 0 {
			params[key] = vals[0]
		}
	}

	res, err := h.svc.FindAll(params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, httpresponse.NewError(http.StatusInternalServerError, "Failed to retrieve library items", err.Error()))
	}
	return c.JSON(http.StatusOK, res)
}

// GetDetails godoc
// @Summary      Get library item details
// @Description  Retrieves details of a specific library item along with related items.
// @Tags         Library
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Library Item ID"
// @Success      200  {object}  dto.LibraryDetailsResponse
// @Failure      400  {object}  httpresponse.Error
// @Failure      404  {object}  httpresponse.Error
// @Router       /api/v1/library/{id} [get]
func (h *Handler) GetDetails(c *echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.NewError(http.StatusBadRequest, "Invalid ID format", err.Error()))
	}

	res, err := h.svc.GetDetails(id)
	if err != nil {
		if errors.Is(err, ErrLibraryNotFound) {
			return c.JSON(http.StatusNotFound, httpresponse.NewError(http.StatusNotFound, "Library item not found", ""))
		}
		return c.JSON(http.StatusInternalServerError, httpresponse.NewError(http.StatusInternalServerError, "Failed to retrieve library details", err.Error()))
	}
	return c.JSON(http.StatusOK, res)
}

// Create godoc
// @Summary      Create a library item
// @Description  Admin endpoint to create a new library item.
// @Tags         Library
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      dto.CreateLibraryReq  true  "Library payload"
// @Success      201      {object}  dto.LibraryResponse
// @Failure      400      {object}  httpresponse.Error
// @Router       /api/v1/admin/library [post]
func (h *Handler) Create(c *echo.Context) error {
	var req dto.CreateLibraryReq
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.NewError(http.StatusBadRequest, "Invalid request body", err.Error()))
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.NewError(http.StatusBadRequest, "Validation failed", err.Error()))
	}

	res, err := h.svc.Create(req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, httpresponse.NewError(http.StatusInternalServerError, "Failed to create library item", err.Error()))
	}
	return c.JSON(http.StatusCreated, res)
}

// Update godoc
// @Summary      Update a library item
// @Description  Admin endpoint to update an existing library item.
// @Tags         Library
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      string                   true  "Library Item ID"
// @Param        request  body      dto.UpdateLibraryReq  true  "Library payload"
// @Success      200      {object}  dto.LibraryResponse
// @Failure      400      {object}  httpresponse.Error
// @Failure      404      {object}  httpresponse.Error
// @Router       /api/v1/admin/library/{id} [put]
func (h *Handler) Update(c *echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.NewError(http.StatusBadRequest, "Invalid ID format", err.Error()))
	}

	var req dto.UpdateLibraryReq
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.NewError(http.StatusBadRequest, "Invalid request body", err.Error()))
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.NewError(http.StatusBadRequest, "Validation failed", err.Error()))
	}

	res, err := h.svc.Update(id, req)
	if err != nil {
		if errors.Is(err, ErrLibraryNotFound) {
			return c.JSON(http.StatusNotFound, httpresponse.NewError(http.StatusNotFound, "Library item not found", ""))
		}
		return c.JSON(http.StatusInternalServerError, httpresponse.NewError(http.StatusInternalServerError, "Failed to update library item", err.Error()))
	}
	return c.JSON(http.StatusOK, res)
}

// Delete godoc
// @Summary      Delete a library item
// @Description  Admin endpoint to delete an existing library item.
// @Tags         Library
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Library Item ID"
// @Success      204  "No Content"
// @Failure      400  {object}  httpresponse.Error
// @Router       /api/v1/admin/library/{id} [delete]
func (h *Handler) Delete(c *echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.NewError(http.StatusBadRequest, "Invalid ID format", err.Error()))
	}

	if err := h.svc.Delete(id); err != nil {
		return c.JSON(http.StatusInternalServerError, httpresponse.NewError(http.StatusInternalServerError, "Failed to delete library item", err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}

// FindAllCategories godoc
// @Summary      Get all library categories
// @Description  Retrieves all library categories.
// @Tags         Library Categories
// @Produce      json
// @Success      200  {array}   dto.CategoryResponse
// @Router       /api/v1/library-categories [get]
func (h *Handler) FindAllCategories(c *echo.Context) error {
	res, err := h.svc.FindAllCategories()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, httpresponse.NewError(http.StatusInternalServerError, "Failed to retrieve categories", err.Error()))
	}
	return c.JSON(http.StatusOK, res)
}

// CreateCategory godoc
// @Summary      Create a library category
// @Description  Admin endpoint to create a new category.
// @Tags         Library Categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      dto.CreateCategoryReq  true  "Category payload"
// @Success      201      {object}  dto.CategoryResponse
// @Failure      400      {object}  httpresponse.Error
// @Router       /api/v1/admin/library-categories [post]
func (h *Handler) CreateCategory(c *echo.Context) error {
	var req dto.CreateCategoryReq
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.NewError(http.StatusBadRequest, "Invalid request body", err.Error()))
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.NewError(http.StatusBadRequest, "Validation failed", err.Error()))
	}

	res, err := h.svc.CreateCategory(req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, httpresponse.NewError(http.StatusInternalServerError, "Failed to create category", err.Error()))
	}
	return c.JSON(http.StatusCreated, res)
}

// UpdateCategory godoc
// @Summary      Update a library category
// @Description  Admin endpoint to update a category.
// @Tags         Library Categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      string                 true  "Category ID"
// @Param        request  body      dto.UpdateCategoryReq  true  "Category payload"
// @Success      200      {object}  dto.CategoryResponse
// @Failure      400      {object}  httpresponse.Error
// @Router       /api/v1/admin/library-categories/{id} [put]
func (h *Handler) UpdateCategory(c *echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.NewError(http.StatusBadRequest, "Invalid ID format", err.Error()))
	}

	var req dto.UpdateCategoryReq
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.NewError(http.StatusBadRequest, "Invalid request body", err.Error()))
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.NewError(http.StatusBadRequest, "Validation failed", err.Error()))
	}

	res, err := h.svc.UpdateCategory(id, req)
	if err != nil {
		if errors.Is(err, ErrCategoryNotFound) {
			return c.JSON(http.StatusNotFound, httpresponse.NewError(http.StatusNotFound, "Category not found", ""))
		}
		return c.JSON(http.StatusInternalServerError, httpresponse.NewError(http.StatusInternalServerError, "Failed to update category", err.Error()))
	}
	return c.JSON(http.StatusOK, res)
}

// DeleteCategory godoc
// @Summary      Delete a library category
// @Description  Admin endpoint to delete a category.
// @Tags         Library Categories
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Category ID"
// @Success      204  "No Content"
// @Router       /api/v1/admin/library-categories/{id} [delete]
func (h *Handler) DeleteCategory(c *echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.NewError(http.StatusBadRequest, "Invalid ID format", err.Error()))
	}

	if err := h.svc.DeleteCategory(id); err != nil {
		return c.JSON(http.StatusInternalServerError, httpresponse.NewError(http.StatusInternalServerError, "Failed to delete category", err.Error()))
	}
	return c.NoContent(http.StatusNoContent)
}
