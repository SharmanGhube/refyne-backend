package models

import "time"

// PaddleWebhookEvent represents the incoming webhook from Paddle
type PaddleWebhookEvent struct {
	EventID    string                 `json:"event_id"`
	EventType  string                 `json:"event_type"`
	OccurredAt time.Time              `json:"occurred_at"`
	Data       map[string]interface{} `json:"data"`
}

// SubscriptionCreated represents subscription.created event data
type SubscriptionCreated struct {
	SubscriptionID       string                 `json:"id"`
	Status               string                 `json:"status"`
	CustomerID           string                 `json:"customer_id"`
	Items                []SubscriptionItem     `json:"items"`
	CustomData           map[string]interface{} `json:"custom_data"`
	CurrentBillingPeriod *BillingPeriod         `json:"current_billing_period"`
}

// SubscriptionUpdated represents subscription.updated event data
type SubscriptionUpdated struct {
	SubscriptionID       string                 `json:"id"`
	Status               string                 `json:"status"`
	CustomerID           string                 `json:"customer_id"`
	Items                []SubscriptionItem     `json:"items"`
	CustomData           map[string]interface{} `json:"custom_data"`
	CurrentBillingPeriod *BillingPeriod         `json:"current_billing_period"`
}

// SubscriptionCanceled represents subscription.canceled event data
type SubscriptionCanceled struct {
	SubscriptionID string                 `json:"id"`
	Status         string                 `json:"status"`
	CustomerID     string                 `json:"customer_id"`
	CanceledAt     time.Time              `json:"canceled_at"`
	CustomData     map[string]interface{} `json:"custom_data"`
}

// SubscriptionPastDue represents subscription.past_due event data
type SubscriptionPastDue struct {
	SubscriptionID string                 `json:"id"`
	Status         string                 `json:"status"`
	CustomerID     string                 `json:"customer_id"`
	CustomData     map[string]interface{} `json:"custom_data"`
}

// SubscriptionPaused represents subscription.paused event data
type SubscriptionPaused struct {
	SubscriptionID string                 `json:"id"`
	Status         string                 `json:"status"`
	CustomerID     string                 `json:"customer_id"`
	PausedAt       time.Time              `json:"paused_at"`
	CustomData     map[string]interface{} `json:"custom_data"`
}

// SubscriptionResumed represents subscription.resumed event data
type SubscriptionResumed struct {
	SubscriptionID string                 `json:"id"`
	Status         string                 `json:"status"`
	CustomerID     string                 `json:"customer_id"`
	ResumedAt      time.Time              `json:"resumed_at"`
	CustomData     map[string]interface{} `json:"custom_data"`
}

// TransactionCompleted represents transaction.completed event data
type TransactionCompleted struct {
	TransactionID  string                 `json:"id"`
	Status         string                 `json:"status"`
	CustomerID     string                 `json:"customer_id"`
	SubscriptionID *string                `json:"subscription_id,omitempty"`
	Items          []TransactionItem      `json:"items"`
	CustomData     map[string]interface{} `json:"custom_data"`
}

// TransactionPaymentFailed represents transaction.payment_failed event data
type TransactionPaymentFailed struct {
	TransactionID  string                 `json:"id"`
	Status         string                 `json:"status"`
	CustomerID     string                 `json:"customer_id"`
	SubscriptionID *string                `json:"subscription_id,omitempty"`
	CustomData     map[string]interface{} `json:"custom_data"`
}

// SubscriptionItem represents an item in a subscription
type SubscriptionItem struct {
	PriceID   string                 `json:"price_id"`
	ProductID string                 `json:"product_id"`
	Quantity  int                    `json:"quantity"`
	Price     map[string]interface{} `json:"price"`
}

// TransactionItem represents an item in a transaction
type TransactionItem struct {
	PriceID   string                 `json:"price_id"`
	ProductID string                 `json:"product_id"`
	Quantity  int                    `json:"quantity"`
	Price     map[string]interface{} `json:"price"`
}

// BillingPeriod represents the current billing period
type BillingPeriod struct {
	StartsAt *time.Time `json:"starts_at,omitempty"`
	EndsAt   *time.Time `json:"ends_at,omitempty"`
}
