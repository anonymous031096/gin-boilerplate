package product

import (
	"strconv"

	httpresp "gin-boilerplate/internal/shared/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	db *pgxpool.Pool
}

func NewHandler(db *pgxpool.Pool) *Handler { return &Handler{db: db} }

// Create godoc
// @Summary Create product
// @Tags Products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateProductRequest true "Create product payload"
// @Success 201 {object} http.APIResponse
// @Failure 400 {object} http.APIResponse
// @Router /products [post]
func (h *Handler) Create(c *gin.Context) {
	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpresp.BadRequest(c, err.Error())
		return
	}
	var item ProductResponse
	err := h.db.QueryRow(c, `
		INSERT INTO products (name,description,price)
		VALUES ($1,$2,$3) RETURNING id,name,description,price
	`, req.Name, req.Description, req.Price).
		Scan(&item.ID, &item.Name, &item.Description, &item.Price)
	if err != nil {
		httpresp.BadRequest(c, err.Error())
		return
	}
	httpresp.Created(c, item)
}

// List godoc
// @Summary List products
// @Tags Products
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} http.APIResponse
// @Failure 500 {object} http.APIResponse
// @Router /products [get]
func (h *Handler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	rows, err := h.db.Query(c, `SELECT id,name,description,price FROM products ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		httpresp.Internal(c, err.Error())
		return
	}
	defer rows.Close()
	items := make([]ProductResponse, 0)
	for rows.Next() {
		var item ProductResponse
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Price); err != nil {
			httpresp.Internal(c, err.Error())
			return
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		httpresp.Internal(c, err.Error())
		return
	}
	httpresp.OK(c, items)
}

// GetByID godoc
// @Summary Get product by id
// @Tags Products
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Success 200 {object} http.APIResponse
// @Failure 400 {object} http.APIResponse
// @Failure 404 {object} http.APIResponse
// @Router /products/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpresp.BadRequest(c, "invalid id")
		return
	}
	var item ProductResponse
	err = h.db.QueryRow(c, `SELECT id,name,description,price FROM products WHERE id=$1`, id).
		Scan(&item.ID, &item.Name, &item.Description, &item.Price)
	if err != nil {
		httpresp.NotFound(c, "product not found")
		return
	}
	httpresp.OK(c, item)
}

// Update godoc
// @Summary Update product
// @Tags Products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Param request body UpdateProductRequest true "Update product payload"
// @Success 200 {object} http.APIResponse
// @Failure 400 {object} http.APIResponse
// @Router /products/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpresp.BadRequest(c, "invalid id")
		return
	}
	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpresp.BadRequest(c, err.Error())
		return
	}
	ct, err := h.db.Exec(c, `UPDATE products SET name=$2,description=$3,price=$4,updated_at=now() WHERE id=$1`, id, req.Name, req.Description, req.Price)
	if err != nil {
		httpresp.BadRequest(c, err.Error())
		return
	}
	if ct.RowsAffected() == 0 {
		httpresp.BadRequest(c, pgx.ErrNoRows.Error())
		return
	}
	httpresp.OK(c, gin.H{"id": id})
}

// Delete godoc
// @Summary Delete product
// @Tags Products
// @Produce json
// @Security BearerAuth
// @Param id path string true "Product ID"
// @Success 200 {object} http.APIResponse
// @Failure 400 {object} http.APIResponse
// @Router /products/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpresp.BadRequest(c, "invalid id")
		return
	}
	ct, err := h.db.Exec(c, `DELETE FROM products WHERE id=$1`, id)
	if err != nil {
		httpresp.BadRequest(c, err.Error())
		return
	}
	if ct.RowsAffected() == 0 {
		httpresp.BadRequest(c, pgx.ErrNoRows.Error())
		return
	}
	httpresp.OK(c, gin.H{"id": id})
}
