package models

import "time"

// CheckoutRequest represents a request to create a Paddle checkout session
type CheckoutRequest struct {
	Tier string `json:"tier" binding:"omitempty"` // Always "pro", kept for API compatibility
}

// CheckoutResponse contains the checkout URL for the user to complete payment
type CheckoutResponse struct {
	CheckoutURL string `json:"checkout_url"`
	Mode        string `json:"mode"` // mock, sandbox, or production
}

// SubscriptionStatusResponse returns the user's current subscription details
type SubscriptionStatusResponse struct {
	Tier             string     `json:"tier"`        // "pro" or null
	Status           string     `json:"status"`      // active, cancelled, past_due, inactive
	ExpiresAt        *time.Time `json:"expires_at,omitempty"`
	PaddleCustomerID *string    `json:"paddle_customer_id,omitempty"`
	ManagementPortalURL string  `json:"management_portal_url,omitempty"`
}

// CustomerPortalResponse contains the URL to Paddle's customer portal
type CustomerPortalResponse struct {
	PortalURL string `json:"portal_url"`
}
