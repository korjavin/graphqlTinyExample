package models

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

// Database connection string and pool
func NewDB(host, port, user, password, dbname string) (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}
	
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	
	log.Println("Successfully connected to the database")
	return db, nil
}

// Seller represents a seller entity
type Seller struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
}

// Listing represents a product listing
type Listing struct {
	ID          int     `json:"id"`
	SellerID    int     `json:"sellerId"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Seller      *Seller `json:"seller,omitempty"`
}

// Purchase represents a purchase transaction
type Purchase struct {
	ID              int       `json:"id"`
	ListingID       int       `json:"listingId"`
	Price           float64   `json:"price"`
	BankTxID        string    `json:"bankTxId"`
	DeliveryAddress string    `json:"deliveryAddress"`
	CreatedAt       time.Time `json:"createdAt"`
	Listing         *Listing  `json:"listing,omitempty"`
}

// Delivery represents a delivery status update
type Delivery struct {
	ID         int       `json:"id"`
	PurchaseID int       `json:"purchaseId"`
	Timestamp  time.Time `json:"timestamp"`
	Status     string    `json:"status"`
	Purchase   *Purchase `json:"purchase,omitempty"`
}

// Filter options for GraphQL queries
type ListingFilter struct {
	SellerID *int
	MinPrice *float64
	MaxPrice *float64
	Title    *string
}

type PurchaseFilter struct {
	ListingID *int
	BankTxID  *string
	FromDate  *time.Time
	ToDate    *time.Time
}

type DeliveryFilter struct {
	PurchaseID *int
	Status     *string
	FromDate   *time.Time
	ToDate     *time.Time
}