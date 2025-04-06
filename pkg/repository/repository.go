package repository

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/korjavin/graphqlTinyExample/pkg/models"
	_ "github.com/lib/pq"
)

// Repository handles all database operations
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new repository with the given database connection
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetSeller fetches a seller by ID
func (r *Repository) GetSeller(id int) (*models.Seller, error) {
	log.Printf("[DB] Fetching seller with ID: %d", id)

	var seller models.Seller
	err := r.db.QueryRow("SELECT id, name, address FROM sellers WHERE id = $1", id).
		Scan(&seller.ID, &seller.Name, &seller.Address)
	if err != nil {
		log.Printf("[DB] Error fetching seller: %v", err)
		return nil, err
	}

	return &seller, nil
}

// GetAllSellers fetches all sellers
func (r *Repository) GetAllSellers() ([]*models.Seller, error) {
	log.Printf("[DB] Fetching all sellers")

	rows, err := r.db.Query("SELECT id, name, address FROM sellers")
	if err != nil {
		log.Printf("[DB] Error fetching sellers: %v", err)
		return nil, err
	}
	defer rows.Close()

	var sellers []*models.Seller
	for rows.Next() {
		var seller models.Seller
		err := rows.Scan(&seller.ID, &seller.Name, &seller.Address)
		if err != nil {
			log.Printf("[DB] Error scanning seller row: %v", err)
			return nil, err
		}
		sellers = append(sellers, &seller)
	}

	if err = rows.Err(); err != nil {
		log.Printf("[DB] Error iterating seller rows: %v", err)
		return nil, err
	}

	log.Printf("[DB] Found %d sellers", len(sellers))
	return sellers, nil
}

// GetListing fetches a listing by ID
func (r *Repository) GetListing(id int) (*models.Listing, error) {
	log.Printf("[DB] Fetching listing with ID: %d", id)

	var listing models.Listing
	err := r.db.QueryRow("SELECT id, seller_id, title, description, price FROM listings WHERE id = $1", id).
		Scan(&listing.ID, &listing.SellerID, &listing.Title, &listing.Description, &listing.Price)
	if err != nil {
		log.Printf("[DB] Error fetching listing: %v", err)
		return nil, err
	}

	return &listing, nil
}

// GetListings fetches listings with optional filtering
func (r *Repository) GetListings(filter *models.ListingFilter) ([]*models.Listing, error) {
	log.Printf("[DB] Fetching listings with filter")

	query := "SELECT id, seller_id, title, description, price FROM listings"

	// Build WHERE clause based on filter
	var conditions []string
	var args []interface{}
	argCount := 1

	if filter != nil {
		if filter.SellerID != nil {
			conditions = append(conditions, fmt.Sprintf("seller_id = $%d", argCount))
			args = append(args, *filter.SellerID)
			argCount++
		}

		if filter.MinPrice != nil {
			conditions = append(conditions, fmt.Sprintf("price >= $%d", argCount))
			args = append(args, *filter.MinPrice)
			argCount++
		}

		if filter.MaxPrice != nil {
			conditions = append(conditions, fmt.Sprintf("price <= $%d", argCount))
			args = append(args, *filter.MaxPrice)
			argCount++
		}

		if filter.Title != nil {
			conditions = append(conditions, fmt.Sprintf("title ILIKE $%d", argCount))
			args = append(args, "%"+*filter.Title+"%")
			argCount++
		}
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	log.Printf("[DB] Executing query: %s with %d args", query, len(args))

	rows, err := r.db.Query(query, args...)
	if err != nil {
		log.Printf("[DB] Error fetching listings: %v", err)
		return nil, err
	}
	defer rows.Close()

	var listings []*models.Listing
	for rows.Next() {
		var listing models.Listing
		err := rows.Scan(&listing.ID, &listing.SellerID, &listing.Title, &listing.Description, &listing.Price)
		if err != nil {
			log.Printf("[DB] Error scanning listing row: %v", err)
			return nil, err
		}
		listings = append(listings, &listing)
	}

	if err = rows.Err(); err != nil {
		log.Printf("[DB] Error iterating listing rows: %v", err)
		return nil, err
	}

	log.Printf("[DB] Found %d listings", len(listings))
	return listings, nil
}

// CreateListing inserts a new listing into the database
func (r *Repository) CreateListing(sellerId int, title, description string, price float64) (*models.Listing, error) {
	log.Printf("[DB] Creating new listing with title: %s, price: %.2f", title, price)

	var id int
	err := r.db.QueryRow(
		`INSERT INTO listings (seller_id, title, description, price) 
		VALUES ($1, $2, $3, $4) RETURNING id`,
		sellerId, title, description, price).Scan(&id)

	if err != nil {
		log.Printf("[DB] Error creating listing: %v", err)
		return nil, err
	}

	// Return the newly created listing
	listing := &models.Listing{
		ID:          id,
		SellerID:    sellerId,
		Title:       title,
		Description: description,
		Price:       price,
	}

	log.Printf("[DB] Created new listing with ID: %d", id)
	return listing, nil
}

// GetPurchase fetches a purchase by ID
func (r *Repository) GetPurchase(id int) (*models.Purchase, error) {
	log.Printf("[DB] Fetching purchase with ID: %d", id)

	var purchase models.Purchase
	err := r.db.QueryRow(
		`SELECT id, listing_id, price, bank_tx_id, delivery_address, created_at 
		FROM purchases WHERE id = $1`, id).
		Scan(&purchase.ID, &purchase.ListingID, &purchase.Price,
			&purchase.BankTxID, &purchase.DeliveryAddress, &purchase.CreatedAt)
	if err != nil {
		log.Printf("[DB] Error fetching purchase: %v", err)
		return nil, err
	}

	return &purchase, nil
}

// GetPurchases fetches purchases with optional filtering
func (r *Repository) GetPurchases(filter *models.PurchaseFilter) ([]*models.Purchase, error) {
	log.Printf("[DB] Fetching purchases with filter")

	query := `SELECT id, listing_id, price, bank_tx_id, delivery_address, created_at 
			FROM purchases`

	// Build WHERE clause based on filter
	var conditions []string
	var args []interface{}
	argCount := 1

	if filter != nil {
		if filter.ListingID != nil {
			conditions = append(conditions, fmt.Sprintf("listing_id = $%d", argCount))
			args = append(args, *filter.ListingID)
			argCount++
		}

		if filter.BankTxID != nil {
			conditions = append(conditions, fmt.Sprintf("bank_tx_id = $%d", argCount))
			args = append(args, *filter.BankTxID)
			argCount++
		}

		if filter.FromDate != nil {
			conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argCount))
			args = append(args, *filter.FromDate)
			argCount++
		}

		if filter.ToDate != nil {
			conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argCount))
			args = append(args, *filter.ToDate)
			argCount++
		}
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	log.Printf("[DB] Executing query: %s with %d args", query, len(args))

	rows, err := r.db.Query(query, args...)
	if err != nil {
		log.Printf("[DB] Error fetching purchases: %v", err)
		return nil, err
	}
	defer rows.Close()

	var purchases []*models.Purchase
	for rows.Next() {
		var purchase models.Purchase
		err := rows.Scan(&purchase.ID, &purchase.ListingID, &purchase.Price,
			&purchase.BankTxID, &purchase.DeliveryAddress, &purchase.CreatedAt)
		if err != nil {
			log.Printf("[DB] Error scanning purchase row: %v", err)
			return nil, err
		}
		purchases = append(purchases, &purchase)
	}

	if err = rows.Err(); err != nil {
		log.Printf("[DB] Error iterating purchase rows: %v", err)
		return nil, err
	}

	log.Printf("[DB] Found %d purchases", len(purchases))
	return purchases, nil
}

// CreatePurchase inserts a new purchase into the database
func (r *Repository) CreatePurchase(listingId int, price float64, bankTxId, deliveryAddress string) (*models.Purchase, error) {
	log.Printf("[DB] Creating new purchase for listing ID: %d, price: %.2f", listingId, price)

	var id int
	var createdAt time.Time

	err := r.db.QueryRow(
		`INSERT INTO purchases (listing_id, price, bank_tx_id, delivery_address, created_at) 
		VALUES ($1, $2, $3, $4, NOW()) RETURNING id, created_at`,
		listingId, price, bankTxId, deliveryAddress).Scan(&id, &createdAt)

	if err != nil {
		log.Printf("[DB] Error creating purchase: %v", err)
		return nil, err
	}

	// Return the newly created purchase
	purchase := &models.Purchase{
		ID:              id,
		ListingID:       listingId,
		Price:           price,
		BankTxID:        bankTxId,
		DeliveryAddress: deliveryAddress,
		CreatedAt:       createdAt,
	}

	log.Printf("[DB] Created new purchase with ID: %d", id)
	return purchase, nil
}

// GetDelivery fetches a delivery by ID
func (r *Repository) GetDelivery(id int) (*models.Delivery, error) {
	log.Printf("[DB] Fetching delivery with ID: %d", id)

	var delivery models.Delivery
	err := r.db.QueryRow(
		"SELECT id, purchase_id, timestamp, status FROM deliveries WHERE id = $1", id).
		Scan(&delivery.ID, &delivery.PurchaseID, &delivery.Timestamp, &delivery.Status)
	if err != nil {
		log.Printf("[DB] Error fetching delivery: %v", err)
		return nil, err
	}

	return &delivery, nil
}

// GetDeliveries fetches deliveries with optional filtering
func (r *Repository) GetDeliveries(filter *models.DeliveryFilter) ([]*models.Delivery, error) {
	log.Printf("[DB] Fetching deliveries with filter")

	query := "SELECT id, purchase_id, timestamp, status FROM deliveries"

	// Build WHERE clause based on filter
	var conditions []string
	var args []interface{}
	argCount := 1

	if filter != nil {
		if filter.PurchaseID != nil {
			conditions = append(conditions, fmt.Sprintf("purchase_id = $%d", argCount))
			args = append(args, *filter.PurchaseID)
			argCount++
		}

		if filter.Status != nil {
			conditions = append(conditions, fmt.Sprintf("status = $%d", argCount))
			args = append(args, *filter.Status)
			argCount++
		}

		if filter.FromDate != nil {
			conditions = append(conditions, fmt.Sprintf("timestamp >= $%d", argCount))
			args = append(args, *filter.FromDate)
			argCount++
		}

		if filter.ToDate != nil {
			conditions = append(conditions, fmt.Sprintf("timestamp <= $%d", argCount))
			args = append(args, *filter.ToDate)
			argCount++
		}
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Add order by timestamp
	query += " ORDER BY timestamp DESC"

	log.Printf("[DB] Executing query: %s with %d args", query, len(args))

	rows, err := r.db.Query(query, args...)
	if err != nil {
		log.Printf("[DB] Error fetching deliveries: %v", err)
		return nil, err
	}
	defer rows.Close()

	var deliveries []*models.Delivery
	for rows.Next() {
		var delivery models.Delivery
		err := rows.Scan(&delivery.ID, &delivery.PurchaseID, &delivery.Timestamp, &delivery.Status)
		if err != nil {
			log.Printf("[DB] Error scanning delivery row: %v", err)
			return nil, err
		}
		deliveries = append(deliveries, &delivery)
	}

	if err = rows.Err(); err != nil {
		log.Printf("[DB] Error iterating delivery rows: %v", err)
		return nil, err
	}

	log.Printf("[DB] Found %d deliveries", len(deliveries))
	return deliveries, nil
}

// GetDeliveriesByPurchaseID fetches all deliveries for a specific purchase
func (r *Repository) GetDeliveriesByPurchaseID(purchaseID int) ([]*models.Delivery, error) {
	log.Printf("[DB] Fetching deliveries for purchase ID: %d", purchaseID)

	rows, err := r.db.Query(
		"SELECT id, purchase_id, timestamp, status FROM deliveries WHERE purchase_id = $1 ORDER BY timestamp DESC",
		purchaseID)
	if err != nil {
		log.Printf("[DB] Error fetching deliveries: %v", err)
		return nil, err
	}
	defer rows.Close()

	var deliveries []*models.Delivery
	for rows.Next() {
		var delivery models.Delivery
		err := rows.Scan(&delivery.ID, &delivery.PurchaseID, &delivery.Timestamp, &delivery.Status)
		if err != nil {
			log.Printf("[DB] Error scanning delivery row: %v", err)
			return nil, err
		}
		deliveries = append(deliveries, &delivery)
	}

	if err = rows.Err(); err != nil {
		log.Printf("[DB] Error iterating delivery rows: %v", err)
		return nil, err
	}

	log.Printf("[DB] Found %d deliveries for purchase ID %d", len(deliveries), purchaseID)
	return deliveries, nil
}

// CreateDelivery inserts a new delivery status update
func (r *Repository) CreateDelivery(purchaseID int, status string) (*models.Delivery, error) {
	log.Printf("[DB] Creating new delivery for purchase ID: %d with status: %s", purchaseID, status)

	var id int
	var timestamp time.Time

	err := r.db.QueryRow(
		`INSERT INTO deliveries (purchase_id, timestamp, status) 
		VALUES ($1, NOW(), $2) RETURNING id, timestamp`,
		purchaseID, status).Scan(&id, &timestamp)

	if err != nil {
		log.Printf("[DB] Error creating delivery: %v", err)
		return nil, err
	}

	// Return the newly created delivery
	delivery := &models.Delivery{
		ID:         id,
		PurchaseID: purchaseID,
		Timestamp:  timestamp,
		Status:     status,
	}

	log.Printf("[DB] Created new delivery with ID: %d", id)
	return delivery, nil
}
