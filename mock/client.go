package mock

import (
	"net/http"

	recurly "github.com/blacklightcms/go-recurly"
)

// NewClient sets the unexported fields on the struct and returns a Client.
func NewClient(httpClient *http.Client) *recurly.Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	client := recurly.NewClient("a", "b", httpClient)

	// Services not implemented in mock package are nil so that they panic when used.
	client.Accounts = nil
	client.Adjustments = nil
	client.Billing = nil
	client.Coupons = nil
	client.Redemptions = nil
	client.Invoices = nil
	client.Plans = nil
	client.AddOns = nil
	client.Subscriptions = &SubscriptionsService{}
	client.Transactions = &TransactionsService{}

	return client
}
