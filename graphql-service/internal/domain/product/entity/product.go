package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Product represents a product in the system
type Product struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	Price       float64            `bson:"price" json:"price"`
	Category    string             `bson:"category" json:"category"`
	Stock       int                `bson:"stock" json:"stock"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// NewProduct creates a new product entity
func NewProduct(name, description string, price float64, category string, stock int) *Product {
	now := time.Now()
	return &Product{
		Name:        name,
		Description: description,
		Price:       price,
		Category:    category,
		Stock:       stock,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// GetID returns the product ID as string
func (p *Product) GetID() string {
	return p.ID.Hex()
}

// Update updates product fields
func (p *Product) Update(name, description string, price float64, category string, stock int) {
	p.Name = name
	p.Description = description
	p.Price = price
	p.Category = category
	p.Stock = stock
	p.UpdatedAt = time.Now()
}

// UpdateStock updates the product stock
func (p *Product) UpdateStock(stock int) {
	p.Stock = stock
	p.UpdatedAt = time.Now()
}

// IsInStock checks if the product is in stock
func (p *Product) IsInStock() bool {
	return p.Stock > 0
}

// CanFulfillOrder checks if the product can fulfill the requested quantity
func (p *Product) CanFulfillOrder(quantity int) bool {
	return p.Stock >= quantity
}
