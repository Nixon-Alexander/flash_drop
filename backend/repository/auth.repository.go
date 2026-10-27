package repository

import (
	"database/sql"
	"flash_drop/types"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
)

type AuthRepository struct {
	db *sql.DB
}

func InitAuthRepository(db *sql.DB) *AuthRepository {
	return &AuthRepository{
		db: db,
	}
}

func (ar *AuthRepository) Create(c *gin.Context, userRegisterType types.UserRegisterType) (sql.Result, *types.RepositoryErrorResponseType) {
	query := "INSERT INTO users (name, email, password_hash, role) VALUES(?, ?, ?, ?);"
	result, err := ar.db.ExecContext(c, query, userRegisterType.Name, userRegisterType.Email, userRegisterType.Password, types.Customer)
	if err != nil {
		me, ok := err.(*mysql.MySQLError)
		if !ok {
			return nil, &types.RepositoryErrorResponseType{
				HttpCode:     401,
				ErrorMessage: err.Error(),
			}
		}
		if me.Number == 1062 {
			return nil, &types.RepositoryErrorResponseType{
				HttpCode:     401,
				ErrorMessage: "User already exists",
			}
		} else {
			return nil, &types.RepositoryErrorResponseType{
				HttpCode:     500,
				ErrorMessage: err.Error(),
			}
		}
	}
	return result, nil
}

func (ar *AuthRepository) GetUserByEmail(c *gin.Context, userLoginType types.UserLoginType) *sql.Row {
	sql := "SELECT * FROM users WHERE email = ?;"
	rows := ar.db.QueryRowContext(c, sql, userLoginType.Email)
	return rows
}
