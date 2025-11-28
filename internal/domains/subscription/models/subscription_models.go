package models

import "time"

// CheckoutRequest represents a request to create a Paddle checkout session
type CheckoutRequest struct {
	Tier string `json:"tier" binding:"required,oneof=starter professional business enterprise"`
}

// CheckoutResponse contains the checkout URL for the user to complete payment
type CheckoutResponse struct {
	CheckoutURL string `json:"checkout_url"`
	Mode        string `json:"mode"` // mock, sandbox, or production
}

// SubscriptionStatusResponse returns the user's current subscription details
type SubscriptionStatusResponse struct {
	Tier                string     `json:"tier"`
	Status              string     `json:"status"`
	ExpiresAt           *time.Time `json:"expires_at,omitempty"`
	PaddleCustomerID    *string    `json:"paddle_customer_id,omitempty"`
	CanUpgrade          bool       `json:"can_upgrade"`
	CanDowngrade        bool       `json:"can_downgrade"`
	ManagementPortalURL string     `json:"management_portal_url,omitempty"`
}

// CustomerPortalResponse contains the URL to Paddle's customer portal
type CustomerPortalResponse struct {
	PortalURL string `json:"portal_url"`
}
