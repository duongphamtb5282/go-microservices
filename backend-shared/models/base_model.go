package models

import (
	"backend-shared/audit"
	"time"

	"github.com/google/uuid"
)

// BaseModel represents a base model with audit fields
type BaseModel struct {
	ID string `json:"id" db:"id" validate:"required"`
	audit.AuditEntity
}

// NewBaseModel creates a new base model
func NewBaseModel(createdBy string) BaseModel {
	return BaseModel{
		ID:          uuid.New().String(),
		AuditEntity: audit.NewAuditEntity(createdBy),
	}
}

// GetID returns the model ID
func (m *BaseModel) GetID() string {
	return m.ID
}

// SetID sets the model ID
func (m *BaseModel) SetID(id string) {
	m.ID = id
}

// GetAuditEntity returns the audit entity
func (m *BaseModel) GetAuditEntity() *audit.AuditEntity {
	return &m.AuditEntity
}

// GetEntityID returns the entity ID
func (m *BaseModel) GetEntityID() string {
	return m.ID
}

// GetEntityType returns the entity type
func (m *BaseModel) GetEntityType() string {
	return "base_model"
}

// UpdateAudit updates the audit fields
func (m *BaseModel) UpdateAudit(modifiedBy string) {
	m.AuditEntity.UpdateAudit(modifiedBy)
}

// GetAuditInfo returns audit information
func (m *BaseModel) GetAuditInfo() *audit.AuditInfo {
	return audit.NewAuditInfo(m.ID, m.GetEntityType(), "update", m.ModifiedBy)
}

// IsNew checks if the model is new
func (m *BaseModel) IsNew() bool {
	return m.AuditEntity.IsNew()
}

// GetAge returns the age of the model
func (m *BaseModel) GetAge() time.Duration {
	return m.AuditEntity.GetAge()
}

// GetLastModifiedAge returns the age since last modification
func (m *BaseModel) GetLastModifiedAge() time.Duration {
	return m.AuditEntity.GetLastModifiedAge()
}

// UserModel represents a user model with audit fields
type UserModel struct {
	BaseModel
	Username     string `json:"username" db:"username" validate:"required,min=3,max=20"`
	Email        string `json:"email" db:"email" validate:"required,email"`
	PasswordHash string `json:"-" db:"password_hash" validate:"required"`
	FirstName    string `json:"first_name" db:"first_name"`
	LastName     string `json:"last_name" db:"last_name"`
	IsActive     bool   `json:"is_active" db:"is_active"`
	IsVerified   bool   `json:"is_verified" db:"is_verified"`
}

// NewUserModel creates a new user model
func NewUserModel(username, email, passwordHash, createdBy string) UserModel {
	return UserModel{
		BaseModel:    NewBaseModel(createdBy),
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		IsActive:     true,
		IsVerified:   false,
	}
}

// GetEntityType returns the entity type
func (u *UserModel) GetEntityType() string {
	return "user"
}

// GetFullName returns the full name
func (u *UserModel) GetFullName() string {
	if u.FirstName != "" && u.LastName != "" {
		return u.FirstName + " " + u.LastName
	}
	return u.Username
}

// SetPasswordHash sets the password hash
func (u *UserModel) SetPasswordHash(passwordHash string) {
	u.PasswordHash = passwordHash
}

// Activate activates the user
func (u *UserModel) Activate() {
	u.IsActive = true
}

// Deactivate deactivates the user
func (u *UserModel) Deactivate() {
	u.IsActive = false
}

// Verify verifies the user
func (u *UserModel) Verify() {
	u.IsVerified = true
}

// Unverify unverifies the user
func (u *UserModel) Unverify() {
	u.IsVerified = false
}

// ProductModel represents a product model with audit fields
type ProductModel struct {
	BaseModel
	Name        string  `json:"name" db:"name" validate:"required,min=1,max=100"`
	Description string  `json:"description" db:"description"`
	Price       float64 `json:"price" db:"price" validate:"required,min=0"`
	Category    string  `json:"category" db:"category"`
	SKU         string  `json:"sku" db:"sku" validate:"required"`
	IsActive    bool    `json:"is_active" db:"is_active"`
	Stock       int     `json:"stock" db:"stock" validate:"min=0"`
}

// NewProductModel creates a new product model
func NewProductModel(name, description, category, sku string, price float64, stock int, createdBy string) ProductModel {
	return ProductModel{
		BaseModel:   NewBaseModel(createdBy),
		Name:        name,
		Description: description,
		Price:       price,
		Category:    category,
		SKU:         sku,
		IsActive:    true,
		Stock:       stock,
	}
}

// GetEntityType returns the entity type
func (p *ProductModel) GetEntityType() string {
	return "product"
}

// Activate activates the product
func (p *ProductModel) Activate() {
	p.IsActive = true
}

// Deactivate deactivates the product
func (p *ProductModel) Deactivate() {
	p.IsActive = false
}

// UpdateStock updates the stock
func (p *ProductModel) UpdateStock(stock int) {
	p.Stock = stock
}

// ReduceStock reduces the stock
func (p *ProductModel) ReduceStock(quantity int) {
	p.Stock -= quantity
}

// IncreaseStock increases the stock
func (p *ProductModel) IncreaseStock(quantity int) {
	p.Stock += quantity
}

// IsInStock checks if the product is in stock
func (p *ProductModel) IsInStock() bool {
	return p.Stock > 0
}

// OrderModel represents an order model with audit fields
type OrderModel struct {
	BaseModel
	UserID      string           `json:"user_id" db:"user_id" validate:"required"`
	TotalAmount float64          `json:"total_amount" db:"total_amount" validate:"required,min=0"`
	Status      string           `json:"status" db:"status" validate:"required"`
	Items       []OrderItemModel `json:"items,omitempty"`
}

// NewOrderModel creates a new order model
func NewOrderModel(userID string, totalAmount float64, status string, createdBy string) OrderModel {
	return OrderModel{
		BaseModel:   NewBaseModel(createdBy),
		UserID:      userID,
		TotalAmount: totalAmount,
		Status:      status,
		Items:       make([]OrderItemModel, 0),
	}
}

// GetEntityType returns the entity type
func (o *OrderModel) GetEntityType() string {
	return "order"
}

// AddItem adds an item to the order
func (o *OrderModel) AddItem(item OrderItemModel) {
	o.Items = append(o.Items, item)
}

// UpdateStatus updates the order status
func (o *OrderModel) UpdateStatus(status string) {
	o.Status = status
}

// UpdateTotalAmount updates the total amount
func (o *OrderModel) UpdateTotalAmount(amount float64) {
	o.TotalAmount = amount
}

// OrderItemModel represents an order item model
type OrderItemModel struct {
	ProductID string  `json:"product_id" db:"product_id" validate:"required"`
	Quantity  int     `json:"quantity" db:"quantity" validate:"required,min=1"`
	Price     float64 `json:"price" db:"price" validate:"required,min=0"`
	Total     float64 `json:"total" db:"total" validate:"required,min=0"`
}

// NewOrderItemModel creates a new order item model
func NewOrderItemModel(productID string, quantity int, price float64) OrderItemModel {
	return OrderItemModel{
		ProductID: productID,
		Quantity:  quantity,
		Price:     price,
		Total:     float64(quantity) * price,
	}
}

// CalculateTotal calculates the total for the item
func (i *OrderItemModel) CalculateTotal() {
	i.Total = float64(i.Quantity) * i.Price
}
