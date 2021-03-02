package mock

import (
	"github.com/blacklightcms/recurly"
)

// Client is a test wrapper for recurly.Client holding mocks for
// all of the services.
type Client struct {
	*recurly.Client

	Accounts          AccountsService
	AddOns            AddOnsService
	Adjustments       AdjustmentsService
	Billing           BillingService
	Coupons           CouponsService
	CreditPayments    CreditPaymentsService
	GiftCards         GiftCardsService
	Redemptions       RedemptionsService
	Invoices          InvoicesService
	Plans             PlansService
	Purchases         PurchasesService
	ShippingAddresses ShippingAddressesService
	ShippingMethods   ShippingMethodsService
	Subscriptions     SubscriptionsService
	Transactions      TransactionsService
}

// NewClient returns a new instance of *Client with the
// services assigned to mocks.
func NewClient(subdomain, apiKey string) *Client {
	c := &Client{Client: recurly.NewClient(subdomain, apiKey)}

	// Attach mock implementations.
	c.Client.Accounts = &c.Accounts
	c.Client.AddOns = &c.AddOns
	c.Client.Adjustments = &c.Adjustments
	c.Client.Billing = &c.Billing
	c.Client.Coupons = &c.Coupons
	c.Client.CreditPayments = &c.CreditPayments
	c.Client.GiftCards = &c.GiftCards
	c.Client.Redemptions = &c.Redemptions
	c.Client.Invoices = &c.Invoices
	c.Client.Plans = &c.Plans
	c.Client.Purchases = &c.Purchases
	c.Client.ShippingAddresses = &c.ShippingAddresses
	c.Client.ShippingMethods = &c.ShippingMethods
	c.Client.Subscriptions = &c.Subscriptions
	c.Client.Transactions = &c.Transactions
	return c
}
