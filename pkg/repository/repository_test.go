package repository

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/korjavin/graphqlTinyExample/pkg/models"
)

func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *Repository) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}

	repo := NewRepository(db)
	return db, mock, repo
}

func TestGetSeller(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	// Define test data
	expectedSeller := &models.Seller{
		ID:      1,
		Name:    "Test Seller",
		Address: "123 Test St",
	}

	// Setup expectations
	rows := sqlmock.NewRows([]string{"id", "name", "address"}).
		AddRow(expectedSeller.ID, expectedSeller.Name, expectedSeller.Address)

	mock.ExpectQuery("SELECT id, name, address FROM sellers WHERE id = \\$1").
		WithArgs(1).
		WillReturnRows(rows)

	// Execute the function
	seller, err := repo.GetSeller(1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify expectations
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}

	// Verify result
	if seller.ID != expectedSeller.ID {
		t.Errorf("Expected seller ID %d, got %d", expectedSeller.ID, seller.ID)
	}
	if seller.Name != expectedSeller.Name {
		t.Errorf("Expected seller Name %s, got %s", expectedSeller.Name, seller.Name)
	}
	if seller.Address != expectedSeller.Address {
		t.Errorf("Expected seller Address %s, got %s", expectedSeller.Address, seller.Address)
	}
}

func TestGetAllSellers(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	// Define test data
	expectedSellers := []*models.Seller{
		{ID: 1, Name: "Seller 1", Address: "Address 1"},
		{ID: 2, Name: "Seller 2", Address: "Address 2"},
	}

	// Setup expectations
	rows := sqlmock.NewRows([]string{"id", "name", "address"})
	for _, s := range expectedSellers {
		rows.AddRow(s.ID, s.Name, s.Address)
	}

	mock.ExpectQuery("SELECT id, name, address FROM sellers").
		WillReturnRows(rows)

	// Execute the function
	sellers, err := repo.GetAllSellers()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify expectations
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}

	// Verify result
	if len(sellers) != len(expectedSellers) {
		t.Errorf("Expected %d sellers, got %d", len(expectedSellers), len(sellers))
	}

	for i, seller := range sellers {
		if seller.ID != expectedSellers[i].ID {
			t.Errorf("Expected seller ID %d, got %d", expectedSellers[i].ID, seller.ID)
		}
		if seller.Name != expectedSellers[i].Name {
			t.Errorf("Expected seller Name %s, got %s", expectedSellers[i].Name, seller.Name)
		}
		if seller.Address != expectedSellers[i].Address {
			t.Errorf("Expected seller Address %s, got %s", expectedSellers[i].Address, seller.Address)
		}
	}
}

func TestGetListings(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	// Define test data
	sellerId := 1
	minPrice := 50.0
	maxPrice := 100.0
	title := "test"

	filter := &models.ListingFilter{
		SellerID: &sellerId,
		MinPrice: &minPrice,
		MaxPrice: &maxPrice,
		Title:    &title,
	}

	// Setup expectations
	rows := sqlmock.NewRows([]string{"id", "seller_id", "title", "description", "price"}).
		AddRow(1, sellerId, "Test Listing", "Description", 75.0)

	mock.ExpectQuery("SELECT id, seller_id, title, description, price FROM listings WHERE seller_id = \\$1 AND price >= \\$2 AND price <= \\$3 AND title ILIKE \\$4").
		WithArgs(sellerId, minPrice, maxPrice, "%"+title+"%").
		WillReturnRows(rows)

	// Execute the function
	listings, err := repo.GetListings(filter)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify expectations
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}

	// Verify result
	if len(listings) != 1 {
		t.Errorf("Expected 1 listing, got %d", len(listings))
	}
	if listings[0].SellerID != sellerId {
		t.Errorf("Expected seller ID %d, got %d", sellerId, listings[0].SellerID)
	}
	if listings[0].Price != 75.0 {
		t.Errorf("Expected price %.2f, got %.2f", 75.0, listings[0].Price)
	}
}

func TestGetDeliveries(t *testing.T) {
	db, mock, repo := setupMockDB(t)
	defer db.Close()

	// Define test data
	purchaseId := 1
	status := "delivered"
	now := time.Now()
	fromDate := now.Add(-24 * time.Hour)
	toDate := now

	filter := &models.DeliveryFilter{
		PurchaseID: &purchaseId,
		Status:     &status,
		FromDate:   &fromDate,
		ToDate:     &toDate,
	}

	// Setup expectations
	rows := sqlmock.NewRows([]string{"id", "purchase_id", "timestamp", "status"}).
		AddRow(1, purchaseId, now.Add(-12*time.Hour), status)

	mock.ExpectQuery("SELECT id, purchase_id, timestamp, status FROM deliveries WHERE purchase_id = \\$1 AND status = \\$2 AND timestamp >= \\$3 AND timestamp <= \\$4 ORDER BY timestamp DESC").
		WithArgs(purchaseId, status, fromDate, toDate).
		WillReturnRows(rows)

	// Execute the function
	deliveries, err := repo.GetDeliveries(filter)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify expectations
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}

	// Verify result
	if len(deliveries) != 1 {
		t.Errorf("Expected 1 delivery, got %d", len(deliveries))
	}
	if deliveries[0].PurchaseID != purchaseId {
		t.Errorf("Expected purchase ID %d, got %d", purchaseId, deliveries[0].PurchaseID)
	}
	if deliveries[0].Status != status {
		t.Errorf("Expected status %s, got %s", status, deliveries[0].Status)
	}
}
