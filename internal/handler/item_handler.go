package handler

import (
	"errors"
	"fmt"
	"mime/multipart"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
	"temp_backend/internal/service"
)

// ItemHandler exposes HTTP endpoints for item management.
type ItemHandler struct {
	svc service.ItemService
}

// NewItemHandler creates a new ItemHandler.
func NewItemHandler(svc service.ItemService) *ItemHandler {
	return &ItemHandler{svc: svc}
}

// ListItems returns a paginated list of stored items.
func (h *ItemHandler) ListItems(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)

	items, err := h.svc.List(c.UserContext(), page, limit)
	if err != nil {
		return fmt.Errorf("list items: %w", err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data": items,
		"meta": fiber.Map{
			"page":  page,
			"limit": limit,
		},
	})
}

// CreateItem creates a new item and optionally uploads a file.
func (h *ItemHandler) CreateItem(c *fiber.Ctx) error {
	name := c.FormValue("name")
	description := c.FormValue("description")

	file, err := c.FormFile("file")
	var fh *multipart.FileHeader
	if err != nil {
		if !errors.Is(err, fasthttp.ErrMissingFile) {
			return fmt.Errorf("read file: %w", err)
		}
	} else {
		fh = file
	}

	item, err := h.svc.Create(c.UserContext(), name, description, fh)
	if err != nil {
		return fmt.Errorf("create item: %w", err)
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

// GetItem retrieves a single item by id.
func (h *ItemHandler) GetItem(c *fiber.Ctx) error {
	id := c.Params("id")
	item, err := h.svc.GetByID(c.UserContext(), id)
	if err != nil {
		return fmt.Errorf("get item: %w", err)
	}
	return c.Status(fiber.StatusOK).JSON(item)
}

// UpdateItem updates item metadata and/or replaces the attached file.
func (h *ItemHandler) UpdateItem(c *fiber.Ctx) error {
	id := c.Params("id")
	name := c.FormValue("name")
	description := c.FormValue("description")

	file, err := c.FormFile("file")
	var fh *multipart.FileHeader
	if err != nil {
		if !errors.Is(err, fasthttp.ErrMissingFile) {
			return fmt.Errorf("read file: %w", err)
		}
	} else {
		fh = file
	}

	item, err := h.svc.Update(c.UserContext(), id, name, description, fh)
	if err != nil {
		return fmt.Errorf("update item: %w", err)
	}
	return c.Status(fiber.StatusOK).JSON(item)
}

// DeleteItem removes an item and its stored file.
func (h *ItemHandler) DeleteItem(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.svc.Delete(c.UserContext(), id); err != nil {
		return fmt.Errorf("delete item: %w", err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// DownloadItem streams the attached file for an item.
func (h *ItemHandler) DownloadItem(c *fiber.Ctx) error {
	id := c.Params("id")
	reader, item, err := h.svc.Download(c.UserContext(), id)
	if err != nil {
		return fmt.Errorf("download item: %w", err)
	}
	defer reader.Close()

	c.Set("Content-Type", item.ContentType)
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", item.FileName))
	return c.Status(fiber.StatusOK).SendStream(reader)
}
