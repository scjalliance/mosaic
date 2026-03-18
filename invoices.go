package mosaic

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// Invoice represents an invoice in Mosaic.
// Field names correspond to the snake_case JSON fields used by the Mosaic API.
type Invoice struct {
	// ID is the unique identifier for the invoice.
	ID int `json:"id,omitempty"`

	// InvoiceNumber is the invoice number or code.
	InvoiceNumber string `json:"invoice_number,omitempty"`

	// InvoiceType is the type of invoice (required for creation).
	InvoiceType string `json:"invoice_type"`

	// ProjectID is the ID of the associated project.
	ProjectID int `json:"project_id"`

	// PhaseID is the ID of the associated phase.
	PhaseID int `json:"phase_id"`

	// Amount is the total invoice amount.
	Amount float64 `json:"amount,omitempty"`

	// Percentage is the invoice percentage.
	Percentage float64 `json:"percentage,omitempty"`

	// InvoiceDate is the date the invoice was issued.
	InvoiceDate *Date `json:"invoice_date,omitempty"`

	// PeriodStart is the start of the billing period.
	PeriodStart *Date `json:"period_start"`

	// PeriodEnd is the end of the billing period.
	PeriodEnd *Date `json:"period_end"`

	// BillingCategoryID is the billing category for the invoice.
	BillingCategoryID int `json:"billing_category_id,omitempty"`

	// IsEstimate indicates whether this is an estimate rather than an invoice.
	IsEstimate bool `json:"is_estimate,omitempty"`

	// Notes are any additional notes on the invoice.
	Notes string `json:"notes,omitempty"`
}

// InvoiceFilter contains filter parameters for listing invoices.
type InvoiceFilter struct {
	// ProjectID filters invoices by project.
	ProjectID string

	// IsEstimate filters to estimates only.
	IsEstimate *bool
}

// toQuery converts InvoiceFilter into URL query parameter values.
func (f InvoiceFilter) toQuery() url.Values {
	q := make(url.Values)
	if f.ProjectID != "" {
		q.Set("project_id", f.ProjectID)
	}
	if f.IsEstimate != nil {
		q.Set("is_estimate", strconv.FormatBool(*f.IsEstimate))
	}
	return q
}

// ListInvoices retrieves a list of invoices matching the given filter.
// Uses GET /api/{team_id}/invoice.
func (c *Client) ListInvoices(ctx context.Context, filter InvoiceFilter) ([]Invoice, error) {
	var resp []Invoice
	path := c.apiPath("invoice")
	if err := c.get(ctx, path, filter.toQuery(), &resp); err != nil {
		return nil, fmt.Errorf("listing invoices: %w", err)
	}
	return resp, nil
}

// CreateInvoice creates a new invoice and returns the created invoice.
// Uses POST /api/{team_id}/invoice.
func (c *Client) CreateInvoice(ctx context.Context, invoice *Invoice) (*Invoice, error) {
	var created Invoice
	path := c.apiPath("invoice")
	if err := c.post(ctx, path, invoice, &created); err != nil {
		return nil, fmt.Errorf("creating invoice: %w", err)
	}
	return &created, nil
}

// UpdateInvoice updates an existing invoice and returns the updated invoice.
// Uses PUT /api/{team_id}/invoice/{invoice_id}.
func (c *Client) UpdateInvoice(ctx context.Context, invoiceID int, invoice *Invoice) (*Invoice, error) {
	var updated Invoice
	path := c.apiPath("invoice", strconv.Itoa(invoiceID))
	if err := c.put(ctx, path, invoice, &updated); err != nil {
		return nil, fmt.Errorf("updating invoice %d: %w", invoiceID, err)
	}
	return &updated, nil
}
