package repository

import (
	"database/sql"
	"flash_drop/types"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type CartRepository struct {
	db *sql.DB
}

func InitCartRepository(db *sql.DB) *CartRepository {
	return &CartRepository{db: db}
}

func internalErr(err error) *types.RepositoryErrorResponseType {
	return &types.RepositoryErrorResponseType{HttpCode: 500, ErrorMessage: err.Error()}
}

func (cr *CartRepository) GetItems(c *gin.Context, userId int) (*sql.Rows, error) {
	query := `
		SELECT ci.id AS cart_items_id, ci.cart_id, ci.product_id, ci.quantity, cat.name AS category_name, p.name, p.slug, p.price, p.image_url FROM carts c
		JOIN cart_items ci
		ON cart_id = c.id
		JOIN products p
		ON p.id = ci.product_id
		JOIN categories cat
		ON cat.id = p.category_id
		WHERE c.user_id = ?;
	`

	rows, err := cr.db.QueryContext(c, query, userId)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

func (cr *CartRepository) AddItems(c *gin.Context, userId int, items []types.UserInputItemType) *types.RepositoryErrorResponseType {
	tx, err := cr.db.BeginTx(c, nil)
	if err != nil {
		return internalErr(err)
	}
	defer tx.Rollback()

	now := time.Now()

	var cartId int64
	err = tx.QueryRowContext(c, `SELECT id FROM carts WHERE user_id = ? LIMIT 1 FOR UPDATE`, userId).Scan(&cartId)
	if err == sql.ErrNoRows {
		res, execErr := tx.ExecContext(c,
			`INSERT INTO carts (user_id, created_at, updated_at) VALUES (?, ?, ?)`,
			userId, now, now)
		if execErr != nil {
			return internalErr(execErr)
		}
		cartId, err = res.LastInsertId()
		if err != nil {
			return internalErr(err)
		}
	} else if err != nil {
		return internalErr(err)
	}

	placeholders := make([]string, 0, len(items))
	args := make([]any, 0, len(items)*5)
	for _, it := range items {
		placeholders = append(placeholders, "(?, ?, ?, ?, ?)")
		args = append(args, cartId, *it.ProductId, *it.Quantity, now, now)
	}

	query := fmt.Sprintf(`
		INSERT INTO cart_items (cart_id, product_id, quantity, created_at, updated_at)
		VALUES %s
		ON DUPLICATE KEY UPDATE
			quantity   = quantity + VALUES(quantity),
			updated_at = VALUES(updated_at)`,
		strings.Join(placeholders, ", "))

	if _, err = tx.ExecContext(c, query, args...); err != nil {
		return internalErr(err)
	}

	if err = tx.Commit(); err != nil {
		return internalErr(err)
	}
	return nil
}

func (cr *CartRepository) UpdateCart(c *gin.Context, userId int, userInput []types.UserInputItemType) *types.RepositoryErrorResponseType {
	tx, err := cr.db.BeginTx(c, nil)
	if err != nil {
		return internalErr(err)
	}
	defer tx.Rollback()

	var cartId int
	err = tx.QueryRowContext(c, `SELECT id FROM carts WHERE user_id = ? LIMIT 1`, userId).Scan(&cartId)
	if err == sql.ErrNoRows {
		return &types.RepositoryErrorResponseType{HttpCode: 404, ErrorMessage: "Cart not found"}
	}
	if err != nil {
		return internalErr(err)
	}

	for _, item := range userInput {
		if item.ProductId == nil || item.Quantity == nil {
			return &types.RepositoryErrorResponseType{HttpCode: 400, ErrorMessage: "product_id and quantity are required"}
		}
		if *item.Quantity < 0 {
			return &types.RepositoryErrorResponseType{HttpCode: 400, ErrorMessage: "quantity cannot be negative"}
		}

		if *item.Quantity == 0 {
			if _, err = tx.ExecContext(c,
				`DELETE FROM cart_items WHERE cart_id = ? AND product_id = ?`,
				cartId, *item.ProductId); err != nil {
				return internalErr(err)
			}
			continue
		}

		result, err := tx.ExecContext(c,
			`UPDATE cart_items SET quantity = ?, updated_at = ? WHERE cart_id = ? AND product_id = ?`,
			*item.Quantity, time.Now(), cartId, *item.ProductId)
		if err != nil {
			return internalErr(err)
		}

		affected, err := result.RowsAffected()
		if err != nil {
			return internalErr(err)
		}
		if affected == 0 {
			return &types.RepositoryErrorResponseType{HttpCode: 404, ErrorMessage: "Product is not in cart"}
		}
	}

	if err = tx.Commit(); err != nil {
		return internalErr(err)
	}
	return nil
}
