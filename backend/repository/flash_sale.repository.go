package repository

import (
	"database/sql"
	"flash_drop/types"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

type FlashSaleRepository struct {
	db *sql.DB
}

func InitFlashSaleRepository(db *sql.DB) *FlashSaleRepository {
	return &FlashSaleRepository{
		db: db,
	}
}

func (fr *FlashSaleRepository) CreateEvent(c *gin.Context, event types.UserInputFlashSalesType) error {
	tx, err := fr.db.BeginTx(c, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(c,
		`INSERT INTO flash_sales (name, description, start_time, end_time) VALUES (?, ?, ?, ?)`,
		event.Name, event.Description, event.StartTime, event.EndTime)
	if err != nil {
		return err
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		return err
	}

	valArgs := make([]any, 0, len(event.Items)*5)
	valString := make([]string, 0, len(event.Items))
	for _, e := range event.Items {
		valString = append(valString, "(?, ?, ?, ?, ?)")
		valArgs = append(valArgs, lastId, e.ProductId, e.FlashPrice, e.Stock, 0)
	}

	query := fmt.Sprintf(`INSERT INTO flash_sale_items (flash_sale_id, product_id, flash_price, stock, sold_quantity) VALUES %s`,
		strings.Join(valString, ", "))
	if _, err = tx.ExecContext(c, query, valArgs...); err != nil {
		return err
	}

	return tx.Commit()
}

func (fr *FlashSaleRepository) DeleteEvent(c *gin.Context, eventId int) error {
	tx, err := fr.db.BeginTx(c, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err = tx.ExecContext(c, `DELETE FROM flash_sale_items WHERE flash_sale_id = ?`, eventId); err != nil {
		return err
	}

	res, err := tx.ExecContext(c, `DELETE FROM flash_sales WHERE id = ?`, eventId)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}

	return tx.Commit()
}

func (fr *FlashSaleRepository) UpdateStatus(c *gin.Context, eventId int, status types.FlashSaleStatus) error {
	res, err := fr.db.ExecContext(c, `UPDATE flash_sales SET status = ? WHERE id = ?`, status, eventId)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

const selectJoin = `
	SELECT fs.id, fs.name, fs.description, fs.start_time, fs.end_time, fs.status, fs.created_at, fs.updated_at,
	       fsi.id, fsi.flash_sale_id, fsi.product_id, fsi.flash_price, fsi.stock, fsi.sold_quantity, fsi.created_at, fsi.updated_at
	FROM flash_sales fs
	LEFT JOIN flash_sale_items fsi ON fs.id = fsi.flash_sale_id`

func (fr *FlashSaleRepository) queryEvents(c *gin.Context, query string, args ...any) ([]types.FlashSaleWithItems, error) {
	rows, err := fr.db.QueryContext(c, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []types.FlashSaleWithItems{}
	index := map[int]int{}

	for rows.Next() {
		var fs types.FlashSalesType
		var itemId, itemSaleId, productId, stock, sold sql.NullInt64
		var price sql.NullFloat64
		var itemCreated, itemUpdated sql.NullTime

		err := rows.Scan(
			&fs.Id, &fs.Name, &fs.Description, &fs.StartTime, &fs.EndTime, &fs.Status, &fs.CreatedAt, &fs.UpdatedAt,
			&itemId, &itemSaleId, &productId, &price, &stock, &sold, &itemCreated, &itemUpdated,
		)
		if err != nil {
			return nil, err
		}

		pos, ok := index[fs.Id]
		if !ok {
			result = append(result, types.FlashSaleWithItems{FlashSalesType: fs, Items: []types.FlashSalesItems{}})
			pos = len(result) - 1
			index[fs.Id] = pos
		}

		if itemId.Valid {
			result[pos].Items = append(result[pos].Items, types.FlashSalesItems{
				ProductId:  int(productId.Int64),
				FlashPrice: float32(price.Float64),
				Stock:      int(stock.Int64),
			})
		}
	}

	return result, rows.Err()
}

func (fr *FlashSaleRepository) GetAllFlashSale(c *gin.Context) ([]types.FlashSaleWithItems, error) {
	return fr.queryEvents(c, selectJoin+` ORDER BY fs.id DESC`)
}

func (fr *FlashSaleRepository) GetFlashSaleById(c *gin.Context, eventId int) (*types.FlashSaleWithItems, error) {
	events, err := fr.queryEvents(c, selectJoin+` WHERE fs.id = ?`, eventId)
	if err != nil {
		return nil, err
	}
	if len(events) == 0 {
		return nil, sql.ErrNoRows
	}
	return &events[0], nil
}
