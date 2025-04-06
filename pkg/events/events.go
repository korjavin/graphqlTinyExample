package events

import (
	"log"
	"strconv"
	"sync"

	"github.com/korjavin/graphqlTinyExample/pkg/models"
)

// DeliveryEvent represents a delivery status update event
type DeliveryEvent struct {
	Delivery *models.Delivery
}

// EventBus manages subscription events
type EventBus struct {
	mu          sync.RWMutex
	subscribers map[string]map[chan DeliveryEvent]bool
	nextID      int
}

// NewEventBus creates a new event bus
func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string]map[chan DeliveryEvent]bool),
	}
}

// Subscribe registers a channel to receive delivery events for a specific purchase ID
// If purchaseID is empty, subscribe to all delivery events
func (b *EventBus) SubscribeToDeliveries(purchaseID string) chan DeliveryEvent {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan DeliveryEvent, 1) // Buffered channel to prevent blocking

	// Initialize map for this purchaseID if it doesn't exist
	if _, ok := b.subscribers[purchaseID]; !ok {
		b.subscribers[purchaseID] = make(map[chan DeliveryEvent]bool)
	}

	// Add this subscriber
	b.subscribers[purchaseID][ch] = true
	log.Printf("[EventBus] New subscriber for purchaseID=%s, total subscribers: %d",
		purchaseID, len(b.subscribers[purchaseID]))

	return ch
}

// Unsubscribe removes a channel from receiving events
func (b *EventBus) Unsubscribe(purchaseID string, ch chan DeliveryEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.subscribers[purchaseID]; ok {
		delete(b.subscribers[purchaseID], ch)
		log.Printf("[EventBus] Unsubscribed from purchaseID=%s, remaining subscribers: %d",
			purchaseID, len(b.subscribers[purchaseID]))

		if len(b.subscribers[purchaseID]) == 0 {
			delete(b.subscribers, purchaseID)
		}
	}
}

// PublishDelivery publishes a delivery event to all relevant subscribers
func (b *EventBus) PublishDelivery(delivery *models.Delivery) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	event := DeliveryEvent{Delivery: delivery}

	// Send to subscribers for the specific purchase ID
	purchaseID := strconv.Itoa(delivery.PurchaseID)
	if subscribers, ok := b.subscribers[purchaseID]; ok {
		for ch := range subscribers {
			// Use non-blocking send to prevent deadlocks
			select {
			case ch <- event:
				log.Printf("[EventBus] Delivered event to subscriber for purchaseID=%s", purchaseID)
			default:
				log.Printf("[EventBus] Subscriber channel for purchaseID=%s is full or closed, skipping", purchaseID)
			}
		}
	}

	// Also send to subscribers interested in all deliveries
	if subscribers, ok := b.subscribers[""]; ok {
		for ch := range subscribers {
			select {
			case ch <- event:
				log.Printf("[EventBus] Delivered event to global subscriber")
			default:
				log.Printf("[EventBus] Global subscriber channel is full or closed, skipping")
			}
		}
	}
}
