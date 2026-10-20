package repository

import (
	"database/sql"
	"flash_drop/types"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
)

type CategoryRepository struct {
	db *sql.DB
}

func InitCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{
		db: db,
	}
}

func (pr *CategoryRepository) CreateCategory(c *gin.Context, CategoryInput types.InputCategoryType) (sql.Result, *types.RepositoryErrorResponseType) {
	query := `INSERT INTO categories (name, slug) VALUES (?, ?);`
	result, err := pr.db.ExecContext(c, query, CategoryInput.Name, CategoryInput.Slug)

	if err != nil {
		me, ok := err.(*mysql.MySQLError)
		if !ok {
			return nil, &types.RepositoryErrorResponseType{
				HttpCode:     500,
				ErrorMessage: me.Message,
			}
		}
		if me.Number == 1062 {
			return nil, &types.RepositoryErrorResponseType{
				HttpCode:     409,
				ErrorMessage: me.Message,
			}
		}
	}

	return result, nil
}

func (pr *CategoryRepository) GetAllCategory(c *gin.Context) (*sql.Rows, *types.RepositoryErrorResponseType) {
	query := "SELECT * FROM categories;"
	rows, err := pr.db.QueryContext(c, query)
	if err != nil {
		return nil, &types.RepositoryErrorResponseType{
			ErrorMessage: err.Error(),
			HttpCode:     500,
		}
	}
	return rows, nil
}

func (pr *CategoryRepository) GetCategoryById(c *gin.Context, id int) *sql.Row {
	query := "SELECT * FROM categories WHERE id = ?;"
	row := pr.db.QueryRowContext(c, query, id)
	return row
}

func (pr *CategoryRepository) EditCategory(c *gin.Context, inputCategory types.InputCategoryType, categoryId int) (sql.Result, *types.RepositoryErrorResponseType) {
	query := `UPDATE categories SET name = ?, slug = ?, updated_at = ? WHERE id = ?; `
	result, err := pr.db.ExecContext(c, query, inputCategory.Name, inputCategory.Slug, time.Now(), categoryId)

	if err != nil {
		return nil, &types.RepositoryErrorResponseType{
			ErrorMessage: err.Error(),
			HttpCode:     500,
		}
	}

	return result, nil
}

func (pr *CategoryRepository) DeleteCategory(c *gin.Context, categoryId int) (sql.Result, *types.RepositoryErrorResponseType) {
	query := "DELETE FROM categories WHERE id = ?;"
	result, err := pr.db.ExecContext(c, query, categoryId)

	if err != nil {
		return nil, &types.RepositoryErrorResponseType{
			ErrorMessage: err.Error(),
			HttpCode:     500,
		}
	}

	return result, nil
}
