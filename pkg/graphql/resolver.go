package graphql

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/korjavin/graphqlTinyExample/pkg/models"
	"github.com/korjavin/graphqlTinyExample/pkg/repository"
)

// Resolver is the root resolver for all GraphQL queries
type Resolver struct {
	repo *repository.Repository
}

// NewResolver creates a new resolver with the given repository
func NewResolver(repo *repository.Repository) *Resolver {
	return &Resolver{repo: repo}
}

// Schema loads the GraphQL schema from the schema.graphql file
func GetSchema(resolver *Resolver) (*graphql.Schema, error) {
	schemaString := Schema
	schema, err := graphql.ParseSchema(schemaString, resolver)
	if err != nil {
		return nil, err
	}
	return schema, nil
}

// Seller resolver
type SellerResolver struct {
	seller *models.Seller
	repo   *repository.Repository
}

func (r *SellerResolver) ID() graphql.ID {
	return graphql.ID(strconv.Itoa(r.seller.ID))
}

func (r *SellerResolver) Name() string {
	return r.seller.Name
}

func (r *SellerResolver) Address() string {
	return r.seller.Address
}

func (r *SellerResolver) Listings() ([]*ListingResolver, error) {
	log.Printf("[GraphQL] Fetching listings for seller ID: %d", r.seller.ID)
	
	sellerID := r.seller.ID
	filter := &models.ListingFilter{
		SellerID: &sellerID,
	}
	
	listings, err := r.repo.GetListings(filter)
	if err != nil {
		log.Printf("[GraphQL] Error fetching listings: %v", err)
		return nil, err
	}
	
	var resolvers []*ListingResolver
	for _, listing := range listings {
		resolvers = append(resolvers, &ListingResolver{listing: listing, repo: r.repo})
	}
	
	return resolvers, nil
}

// Listing resolver
type ListingResolver struct {
	listing *models.Listing
	repo    *repository.Repository
}

func (r *ListingResolver) ID() graphql.ID {
	return graphql.ID(strconv.Itoa(r.listing.ID))
}

func (r *ListingResolver) Seller() (*SellerResolver, error) {
	log.Printf("[GraphQL] Fetching seller for listing ID: %d", r.listing.ID)
	
	seller, err := r.repo.GetSeller(r.listing.SellerID)
	if err != nil {
		log.Printf("[GraphQL] Error fetching seller: %v", err)
		return nil, err
	}
	
	return &SellerResolver{seller: seller, repo: r.repo}, nil
}

func (r *ListingResolver) Title() string {
	return r.listing.Title
}

func (r *ListingResolver) Description() string {
	return r.listing.Description
}

func (r *ListingResolver) Price() float64 {
	return r.listing.Price
}

func (r *ListingResolver) Purchases() ([]*PurchaseResolver, error) {
	log.Printf("[GraphQL] Fetching purchases for listing ID: %d", r.listing.ID)
	
	listingID := r.listing.ID
	filter := &models.PurchaseFilter{
		ListingID: &listingID,
	}
	
	purchases, err := r.repo.GetPurchases(filter)
	if err != nil {
		log.Printf("[GraphQL] Error fetching purchases: %v", err)
		return nil, err
	}
	
	var resolvers []*PurchaseResolver
	for _, purchase := range purchases {
		resolvers = append(resolvers, &PurchaseResolver{purchase: purchase, repo: r.repo})
	}
	
	return resolvers, nil
}

// Purchase resolver
type PurchaseResolver struct {
	purchase *models.Purchase
	repo     *repository.Repository
}

func (r *PurchaseResolver) ID() graphql.ID {
	return graphql.ID(strconv.Itoa(r.purchase.ID))
}

func (r *PurchaseResolver) Listing() (*ListingResolver, error) {
	log.Printf("[GraphQL] Fetching listing for purchase ID: %d", r.purchase.ID)
	
	listing, err := r.repo.GetListing(r.purchase.ListingID)
	if err != nil {
		log.Printf("[GraphQL] Error fetching listing: %v", err)
		return nil, err
	}
	
	return &ListingResolver{listing: listing, repo: r.repo}, nil
}

func (r *PurchaseResolver) Price() float64 {
	return r.purchase.Price
}

func (r *PurchaseResolver) BankTxId() string {
	return r.purchase.BankTxID
}

func (r *PurchaseResolver) DeliveryAddress() string {
	return r.purchase.DeliveryAddress
}

func (r *PurchaseResolver) CreatedAt() string {
	return r.purchase.CreatedAt.Format(time.RFC3339)
}

func (r *PurchaseResolver) Deliveries() ([]*DeliveryResolver, error) {
	log.Printf("[GraphQL] Fetching deliveries for purchase ID: %d", r.purchase.ID)
	
	deliveries, err := r.repo.GetDeliveriesByPurchaseID(r.purchase.ID)
	if err != nil {
		log.Printf("[GraphQL] Error fetching deliveries: %v", err)
		return nil, err
	}
	
	var resolvers []*DeliveryResolver
	for _, delivery := range deliveries {
		resolvers = append(resolvers, &DeliveryResolver{delivery: delivery, repo: r.repo})
	}
	
	return resolvers, nil
}

// Delivery resolver
type DeliveryResolver struct {
	delivery *models.Delivery
	repo     *repository.Repository
}

func (r *DeliveryResolver) ID() graphql.ID {
	return graphql.ID(strconv.Itoa(r.delivery.ID))
}

func (r *DeliveryResolver) Purchase() (*PurchaseResolver, error) {
	log.Printf("[GraphQL] Fetching purchase for delivery ID: %d", r.delivery.ID)
	
	purchase, err := r.repo.GetPurchase(r.delivery.PurchaseID)
	if err != nil {
		log.Printf("[GraphQL] Error fetching purchase: %v", err)
		return nil, err
	}
	
	return &PurchaseResolver{purchase: purchase, repo: r.repo}, nil
}

func (r *DeliveryResolver) Timestamp() string {
	return r.delivery.Timestamp.Format(time.RFC3339)
}

func (r *DeliveryResolver) Status() string {
	// Convert status to uppercase to match the GraphQL enum
	switch r.delivery.Status {
	case "packed":
		return "PACKED"
	case "out_for_delivery":
		return "OUT_FOR_DELIVERY"
	case "delivered":
		return "DELIVERED"
	case "rescheduled":
		return "RESCHEDULED"
	case "canceled":
		return "CANCELED"
	default:
		return "UNKNOWN"
	}
}

// Input type resolvers
type ListingFilterInput struct {
	SellerID *graphql.ID
	MinPrice *float64
	MaxPrice *float64
	Title    *string
}

func (r *Resolver) resolveListingFilter(filter *ListingFilterInput) *models.ListingFilter {
	if filter == nil {
		return nil
	}
	
	result := &models.ListingFilter{}
	
	if filter.SellerID != nil {
		id, _ := strconv.Atoi(string(*filter.SellerID))
		result.SellerID = &id
	}
	
	result.MinPrice = filter.MinPrice
	result.MaxPrice = filter.MaxPrice
	result.Title = filter.Title
	
	return result
}

type PurchaseFilterInput struct {
	ListingID *graphql.ID
	BankTxID  *string
	FromDate  *string
	ToDate    *string
}

func (r *Resolver) resolvePurchaseFilter(filter *PurchaseFilterInput) *models.PurchaseFilter {
	if filter == nil {
		return nil
	}
	
	result := &models.PurchaseFilter{}
	
	if filter.ListingID != nil {
		id, _ := strconv.Atoi(string(*filter.ListingID))
		result.ListingID = &id
	}
	
	result.BankTxID = filter.BankTxID
	
	if filter.FromDate != nil {
		fromDate, err := time.Parse(time.RFC3339, *filter.FromDate)
		if err == nil {
			result.FromDate = &fromDate
		}
	}
	
	if filter.ToDate != nil {
		toDate, err := time.Parse(time.RFC3339, *filter.ToDate)
		if err == nil {
			result.ToDate = &toDate
		}
	}
	
	return result
}

type DeliveryFilterInput struct {
	PurchaseID *graphql.ID
	Status     *string
	FromDate   *string
	ToDate     *string
}

func (r *Resolver) resolveDeliveryFilter(filter *DeliveryFilterInput) *models.DeliveryFilter {
	if filter == nil {
		return nil
	}
	
	result := &models.DeliveryFilter{}
	
	if filter.PurchaseID != nil {
		id, _ := strconv.Atoi(string(*filter.PurchaseID))
		result.PurchaseID = &id
	}
	
	if filter.Status != nil {
		var status string
		// Convert GraphQL enum to database enum
		switch *filter.Status {
		case "PACKED":
			status = "packed"
		case "OUT_FOR_DELIVERY":
			status = "out_for_delivery"
		case "DELIVERED":
			status = "delivered"
		case "RESCHEDULED":
			status = "rescheduled"
		case "CANCELED":
			status = "canceled"
		}
		result.Status = &status
	}
	
	if filter.FromDate != nil {
		fromDate, err := time.Parse(time.RFC3339, *filter.FromDate)
		if err == nil {
			result.FromDate = &fromDate
		}
	}
	
	if filter.ToDate != nil {
		toDate, err := time.Parse(time.RFC3339, *filter.ToDate)
		if err == nil {
			result.ToDate = &toDate
		}
	}
	
	return result
}

// Root Query resolvers
func (r *Resolver) Seller(ctx context.Context, args struct{ ID graphql.ID }) (*SellerResolver, error) {
	log.Printf("[GraphQL] Seller query with ID: %s", args.ID)
	
	id, err := strconv.Atoi(string(args.ID))
	if err != nil {
		log.Printf("[GraphQL] Invalid seller ID format: %v", err)
		return nil, fmt.Errorf("invalid seller ID format: %v", err)
	}
	
	seller, err := r.repo.GetSeller(id)
	if err != nil {
		log.Printf("[GraphQL] Error fetching seller: %v", err)
		return nil, err
	}
	
	return &SellerResolver{seller: seller, repo: r.repo}, nil
}

func (r *Resolver) Sellers(ctx context.Context) ([]*SellerResolver, error) {
	log.Printf("[GraphQL] Sellers query")
	
	sellers, err := r.repo.GetAllSellers()
	if err != nil {
		log.Printf("[GraphQL] Error fetching sellers: %v", err)
		return nil, err
	}
	
	var resolvers []*SellerResolver
	for _, seller := range sellers {
		resolvers = append(resolvers, &SellerResolver{seller: seller, repo: r.repo})
	}
	
	return resolvers, nil
}

func (r *Resolver) Listing(ctx context.Context, args struct{ ID graphql.ID }) (*ListingResolver, error) {
	log.Printf("[GraphQL] Listing query with ID: %s", args.ID)
	
	id, err := strconv.Atoi(string(args.ID))
	if err != nil {
		log.Printf("[GraphQL] Invalid listing ID format: %v", err)
		return nil, fmt.Errorf("invalid listing ID format: %v", err)
	}
	
	listing, err := r.repo.GetListing(id)
	if err != nil {
		log.Printf("[GraphQL] Error fetching listing: %v", err)
		return nil, err
	}
	
	return &ListingResolver{listing: listing, repo: r.repo}, nil
}

func (r *Resolver) Listings(ctx context.Context, args struct{ Filter *ListingFilterInput }) ([]*ListingResolver, error) {
	log.Printf("[GraphQL] Listings query with filter")
	
	filter := r.resolveListingFilter(args.Filter)
	listings, err := r.repo.GetListings(filter)
	if err != nil {
		log.Printf("[GraphQL] Error fetching listings: %v", err)
		return nil, err
	}
	
	var resolvers []*ListingResolver
	for _, listing := range listings {
		resolvers = append(resolvers, &ListingResolver{listing: listing, repo: r.repo})
	}
	
	return resolvers, nil
}

func (r *Resolver) Purchase(ctx context.Context, args struct{ ID graphql.ID }) (*PurchaseResolver, error) {
	log.Printf("[GraphQL] Purchase query with ID: %s", args.ID)
	
	id, err := strconv.Atoi(string(args.ID))
	if err != nil {
		log.Printf("[GraphQL] Invalid purchase ID format: %v", err)
		return nil, fmt.Errorf("invalid purchase ID format: %v", err)
	}
	
	purchase, err := r.repo.GetPurchase(id)
	if err != nil {
		log.Printf("[GraphQL] Error fetching purchase: %v", err)
		return nil, err
	}
	
	return &PurchaseResolver{purchase: purchase, repo: r.repo}, nil
}

func (r *Resolver) Purchases(ctx context.Context, args struct{ Filter *PurchaseFilterInput }) ([]*PurchaseResolver, error) {
	log.Printf("[GraphQL] Purchases query with filter")
	
	filter := r.resolvePurchaseFilter(args.Filter)
	purchases, err := r.repo.GetPurchases(filter)
	if err != nil {
		log.Printf("[GraphQL] Error fetching purchases: %v", err)
		return nil, err
	}
	
	var resolvers []*PurchaseResolver
	for _, purchase := range purchases {
		resolvers = append(resolvers, &PurchaseResolver{purchase: purchase, repo: r.repo})
	}
	
	return resolvers, nil
}

func (r *Resolver) Delivery(ctx context.Context, args struct{ ID graphql.ID }) (*DeliveryResolver, error) {
	log.Printf("[GraphQL] Delivery query with ID: %s", args.ID)
	
	id, err := strconv.Atoi(string(args.ID))
	if err != nil {
		log.Printf("[GraphQL] Invalid delivery ID format: %v", err)
		return nil, fmt.Errorf("invalid delivery ID format: %v", err)
	}
	
	delivery, err := r.repo.GetDelivery(id)
	if err != nil {
		log.Printf("[GraphQL] Error fetching delivery: %v", err)
		return nil, err
	}
	
	return &DeliveryResolver{delivery: delivery, repo: r.repo}, nil
}

func (r *Resolver) Deliveries(ctx context.Context, args struct{ Filter *DeliveryFilterInput }) ([]*DeliveryResolver, error) {
	log.Printf("[GraphQL] Deliveries query with filter")
	
	filter := r.resolveDeliveryFilter(args.Filter)
	deliveries, err := r.repo.GetDeliveries(filter)
	if err != nil {
		log.Printf("[GraphQL] Error fetching deliveries: %v", err)
		return nil, err
	}
	
	var resolvers []*DeliveryResolver
	for _, delivery := range deliveries {
		resolvers = append(resolvers, &DeliveryResolver{delivery: delivery, repo: r.repo})
	}
	
	return resolvers, nil
}