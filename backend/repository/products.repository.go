package repository

import (
	"database/sql"
	"flash_drop/types"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
)

type ProductRepository struct {
	db *sql.DB
}

func InitProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{
		db: db,
	}
}

func (pr *ProductRepository) CreateProduct(c *gin.Context, productInput types.InputProductType) (sql.Result, *types.RepositoryErrorResponseType) {
	if len(productInput.ImageUrl) == 0 {
		return nil, &types.RepositoryErrorResponseType{
			HttpCode:     400,
			ErrorMessage: "Product image is required",
		}
	}

	values := make([]string, 0, len(productInput.ImageUrl))
	args := make([]any, 0, len(productInput.ImageUrl)*7)

	for _, image := range productInput.ImageUrl {
		values = append(values, "(?, ?, ?, ?, ?, ?, ?)")
		args = append(args, productInput.CategoryId, productInput.Name, productInput.Slug, productInput.Description, productInput.Price, productInput.Stock, image.NewFileName)
	}

	query := fmt.Sprintf(`INSERT INTO products ( category_id, name, slug, description, price, stock, image_url) VALUES %s`, strings.Join(values, ", "))
	result, err := pr.db.ExecContext(c, query, args...)

	if err != nil {
		me, ok := err.(*mysql.MySQLError)
		if !ok {
			return nil, &types.RepositoryErrorResponseType{
				HttpCode:     500,
				ErrorMessage: err.Error(),
			}
		}
		if me.Number == 1062 {
			return nil, &types.RepositoryErrorResponseType{
				HttpCode:     409,
				ErrorMessage: me.Message,
			}
		}
		return nil, &types.RepositoryErrorResponseType{
			HttpCode:     500,
			ErrorMessage: me.Message,
		}
	}

	return result, nil
}

func (pr *ProductRepository) GetAllProduct(c *gin.Context) (*sql.Rows, error) {
	query := "SELECT * FROM products;"
	rows, err := pr.db.QueryContext(c, query)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (pr *ProductRepository) GetProductById(c *gin.Context, id int) *sql.Row {
	query := "SELECT id, category_id, name, slug, description, price, stock, image_url, is_active, created_at, updated_at FROM products WHERE id = ?;"
	row := pr.db.QueryRowContext(c, query, id)
	return row
}

func (pr *ProductRepository) EditProduct(c *gin.Context, inputProduct types.InputProductType, productId int) (sql.Result, *types.RepositoryErrorResponseType) {
	if len(inputProduct.ImageUrl) == 0 {
		return nil, &types.RepositoryErrorResponseType{
			HttpCode:     400,
			ErrorMessage: "Product image is required",
		}
	}
	image := inputProduct.ImageUrl[0]

	query := `
		UPDATE products SET
			category_id = ?,
			name = ?,
			slug = ?,
			description = ?,
			price = ?,
			stock = ?,
			image_url = ?,
			is_active = ?,
			updated_at = ?
		WHERE id = ?
	`

	result, err := pr.db.ExecContext(
		c,
		query,
		inputProduct.CategoryId,
		inputProduct.Name,
		inputProduct.Slug,
		inputProduct.Description,
		inputProduct.Price,
		inputProduct.Stock,
		image.NewFileName,
		inputProduct.IsActive,
		time.Now(),
		productId,
	)

	if err != nil {
		me, ok := err.(*mysql.MySQLError)

		if !ok {
			return nil, &types.RepositoryErrorResponseType{
				HttpCode:     500,
				ErrorMessage: err.Error(),
			}
		}

		if me.Number == 1062 {
			return nil, &types.RepositoryErrorResponseType{
				HttpCode:     409,
				ErrorMessage: me.Message,
			}
		}

		return nil, &types.RepositoryErrorResponseType{
			HttpCode:     500,
			ErrorMessage: me.Message,
		}
	}

	return result, nil
}

func (pr *ProductRepository) DeleteProduct(c *gin.Context, productId int) (sql.Result, *types.RepositoryErrorResponseType) {
	query := "DELETE FROM products WHERE id = ?;"
	result, err := pr.db.ExecContext(c, query, productId)

	if err != nil {
		return nil, &types.RepositoryErrorResponseType{
			HttpCode:     500,
			ErrorMessage: err.Error(),
		}
	}

	return result, nil
}
