package types

import "time"

type CartsType struct {
	Id        int       `json:"id"`
	UserId    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CartItemsType struct {
	Id        int       `json:"id"`
	CartId    int       `json:"cart_id"`
	ProductId int       `json:"product_id"`
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserInputItemType struct {
	ProductId *int `json:"product_id"`
	Quantity  *int `json:"quantity"`
}

type UserInputUpdateCartItemsType struct {
	CartId int
	Item   []UserInputItemType `json:"item"`
}

type UserCartType struct {
	CartItemsId  int    `json:"cart_items_id"`
	CartId       int    `json:"cart_id"`
	ProuctId     int    `json:"product_id"`
	Quantity     int    `json:"quantity"`
	CategoryName string `json:"category_name"`
	Slug         string `json:"slug"`
	Name         string `json:"name"`
	Price        int    `json:"price"`
	ImageUrl     string `json:"image_url"`
}
