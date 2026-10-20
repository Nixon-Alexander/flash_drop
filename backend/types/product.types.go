package types

import (
	"time"
)

type ProductType struct {
	Id          int       `json:"id"`
	CategoryId  int       `json:"category_id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	Price       float32   `json:"price"`
	Stock       int       `json:"stock"`
	ImageUrl    string    `json:"image_url"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type InputProductType struct {
	Name        string  `form:"name"`
	Description string  `form:"description"`
	CategoryId  int     `form:"category_id"`
	Price       float32 `form:"price"`
	Stock       int     `form:"stock"`
	IsActive    bool    `form:"is_active"`
	Slug        string
	ImageUrl    []File // variable name at form -> image_url
}

type CategoryType struct {
	Id        int       `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type InputCategoryType struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type ProductDetailForCalculationType struct {
	ProductId int
	Price     float32
	Name      string
	Category  string
}
