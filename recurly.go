package recurly

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// apiVersion is the API version in use by this client.
// NOTE: v2.19:
//		- Parent/child accounts not yet implemented.
const apiVersion = "2.27"

// uaVersion is the userAgent version sent to Recurly so they can track usage
// of this library.
const uaVersion = "1.0.0"

// HTTPDoer is used for making HTTP requests. This implementation is generally
// a *http.Client.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client manages communication with the Recurly API.
type Client struct {
	// apiKey is your account's API key used for authentication.
	apiKey string

	// baseURL is the base url for requests.
	baseURL *url.URL

	// userAgent sets the User-Agent header for requests so Recurly can
	// track usage of the client.
	// See https://github.com/blacklightcms/recurly/issues/41
	userAgent string

	// Client is the HTTP Client used to communicate with the API.
	// By default this uses http.DefaultClient, so there are no timeouts
	// configured. It's recommended you set your own HTTP client with
	// reasonable timeouts for your application.
	Client HTTPDoer

	// Services used for talking with different parts of the Recurly API
	Accounts          AccountsService
	Adjustments       AdjustmentsService
	AddOns            AddOnsService
	AutomatedExports  AutomatedExportsService
	Billing           BillingService
	Coupons           CouponsService
	CreditPayments    CreditPaymentsService
	GiftCards         GiftCardsService
	Invoices          InvoicesService
	Plans             PlansService
	Purchases         PurchasesService
	Redemptions       RedemptionsService
	ShippingAddresses ShippingAddressesService
	ShippingMethods   ShippingMethodsService
	Subscriptions     SubscriptionsService
	Transactions      TransactionsService
	Items             ItemsService
}

type serviceImpl struct {
	client *Client
}

// NewClient returns a new instance of *Client.
// By default this uses http.DefaultClient, so there are no timeouts configured.
// It's recommended you set your own HTTP client with reasonable timeouts
// for your application.
func NewClient(subdomain, apiKey string) *Client {
	baseEndpoint, _ := url.Parse(fmt.Sprintf("https://%s.recurly.com/", subdomain))
	client := &Client{
		Client: http.DefaultClient,

		baseURL: baseEndpoint,
		apiKey:  base64.StdEncoding.EncodeToString([]byte(apiKey)),

		userAgent: fmt.Sprintf(
			"Blacklight/%s; Go (%s) [%s-%s]",
			uaVersion,
			runtime.Version(),
			runtime.GOARCH,
			runtime.GOOS,
		),
	}

	client.Accounts = &accountsImpl{client: client}
	client.Adjustments = &adjustmentsImpl{client: client}
	client.AddOns = &addOnsImpl{client: client}
	client.AutomatedExports = &automatedExportsImpl{client: client}
	client.Billing = &billingImpl{client: client}
	client.Coupons = &couponsImpl{client: client}
	client.CreditPayments = &creditInvoicesImpl{client: client}
	client.GiftCards = &giftCardsImpl{client: client}
	client.Invoices = &invoicesImpl{client: client}
	client.Plans = &plansImpl{client: client}
	client.Purchases = &purchasesImpl{client: client}
	client.Redemptions = &redemptionsImpl{client: client}
	client.ShippingAddresses = &shippingAddressesImpl{client: client}
	client.ShippingMethods = &shippingMethodsImpl{client: client}
	client.Subscriptions = &subscriptionsImpl{client: client}
	client.Transactions = &transactionsImpl{client: client}
	client.Items = &itemsImpl{client: client}
	return client
}

// newRequest creates an authenticated API request that is ready to send.
func (c *Client) newRequest(method string, path string, body interface{}) (*http.Request, error) {
	path = fmt.Sprintf("/v2/%s", strings.TrimPrefix(path, "/"))
	u, err := c.baseURL.Parse(path)
	if err != nil {
		return nil, err
	}

	// Request body
	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		if err := xml.NewEncoder(buf).Encode(body); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", c.apiKey))
	req.Header.Set("Accept", "application/xml")
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("X-Api-Version", apiVersion)
	if body != nil {
		req.Header.Set("Content-Type", "application/xml; charset=utf-8")
	}
	return req, err
}

// newPagerRequest is used for pagination.
func (c *Client) newPagerRequest(method string, path string, opts *PagerOptions, body interface{}) (*http.Request, error) {
	req, err := c.newRequest(method, path, body)
	if err != nil {
		return nil, err
	} else if opts != nil {
		opts.append(req.URL)
	}
	return req, nil
}

// newQueryRequest is used to create requests that require query strings.
func (c *Client) newQueryRequest(method string, path string, q query, body interface{}) (*http.Request, error) {
	req, err := c.newRequest(method, path, body)
	if err != nil {
		return nil, err
	} else if len(q) > 0 {
		q.append(req.URL)
	}
	return req, nil
}

// do takes a prepared API request and makes the API call to Recurly.
// It will decode the XML into a destination struct you provide as well
// as parse any validation errors that may have occurred.
// It returns a Response object that provides a wrapper around http.Response
// with some convenience methods.
func (c *Client) do(ctx context.Context, req *http.Request, v interface{}) (*response, error) {
	req = req.WithContext(ctx)

	resp, err := c.Client.Do(req)
	if err != nil {
		// If we got an error, and the context has been canceled,
		// the context's error is probably more useful.
		select {
		default:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		return nil, err
	}
	defer resp.Body.Close()

	response := newResponse(resp)
	if resp.StatusCode == http.StatusNoContent {
		return response, nil
	} else if resp.StatusCode == http.StatusTooManyRequests {
		return nil, &RateLimitError{
			Response: resp,
			Rate:     response.rate,
		}
	} else if v != nil && resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		if w, ok := v.(io.Writer); ok {
			io.Copy(w, resp.Body)
		} else if err := xml.NewDecoder(resp.Body).Decode(&v); err != nil && err != io.EOF {
			return response, err
		}
		return response, nil
	} else if resp.StatusCode >= 400 && resp.StatusCode <= 499 {
		return response, response.parseClientError(v)
	} else if resp.StatusCode >= 500 && resp.StatusCode <= 599 {
		return nil, &ServerError{Response: resp}
	}

	return response, nil
}

// response is a Recurly API response. This wraps the standard http.Response
// returned from Recurly and provides access to pagination cursors and rate
// limits.
type response struct {
	*http.Response

	// The next cursor (if available) when paginating results.
	cursor string

	// Rate limits.
	rate Rate
}

// NewResponse creates a new Response for the provided http.Response.
func newResponse(r *http.Response) *response {
	resp := &response{Response: r}
	resp.populatePageCursor()
	resp.populateRateLimit()
	return resp
}

func (r *response) populatePageCursor() {
	links, ok := r.Response.Header["Link"]
	if !ok || len(links) == 0 {
		return
	}

	for _, link := range strings.Split(links[0], ",") {
		segments := strings.Split(strings.TrimSpace(link), ";")

		if len(segments) < 2 { // link must at least have href and rel
			continue
		} else if !strings.HasPrefix(segments[0], "<") || !strings.HasSuffix(segments[0], ">") { // ensure href is properly formatted
			continue
		}

		// try to pull out cursor parameter
		url, err := url.Parse(segments[0][1 : len(segments[0])-1])
		if err != nil {
			continue
		}

		cursor := url.Query().Get("cursor")
		if cursor == "" {
			continue
		}

		for _, segment := range segments[1:] {
			switch strings.TrimSpace(segment) {
			case `rel="next"`:
				r.cursor = cursor
			}
		}
	}
}

// populates rate limits.
func (r *response) populateRateLimit() {
	if limit := r.Header.Get("X-RateLimit-Limit"); limit != "" {
		r.rate.Limit, _ = strconv.Atoi(limit)
	}
	if remaining := r.Header.Get("X-RateLimit-Remaining"); remaining != "" {
		r.rate.Remaining, _ = strconv.Atoi(remaining)
	}
	if reset := r.Header.Get("X-RateLimit-Reset"); reset != "" {
		if v, _ := strconv.ParseInt(reset, 10, 64); v != 0 {
			r.rate.Reset = time.Unix(v, 0)
		}
	}
}

// parses client errors.
func (r *response) parseClientError(v interface{}) error {
	// Immediately return a client error if there is no response body.
	if r.Header.Get("Content-Length") == "0" {
		return &ClientError{Response: r.Response}
	}

	// Read the full response body so we can conditionally process the
	// xml based on the top level tag that is returned.
	b, err := ioutil.ReadAll(r.Response.Body)
	if err != nil {
		return err
	} else if len(b) == 0 {
		// Exit here to avoid io.EOF errors.
		return &ClientError{Response: r.Response}
	}

	var name struct {
		XMLName xml.Name
	}
	if err := xml.Unmarshal(b, &name); err != nil {
		return err
	}

	switch name.XMLName.Local {
	case "error":
		var e xmlSingleError
		if err := xml.Unmarshal(b, &e); err != nil {
			return err
		}
		return &ClientError{
			Response: r.Response,
			ValidationErrors: []ValidationError{{
				Description: e.Description,
				Field:       e.Field,
				Symbol:      e.Symbol,
			}},
		}
	case "errors":
		var e xmlMultiErrors
		if err := xml.Unmarshal(b, &e); err != nil {
			return err
		}

		// Any transaction errors return TransactionFailedError.
		if e.Transaction != nil || e.TransactionError != nil {
			transFailedErr := &TransactionFailedError{
				Response:    r.Response,
				Transaction: e.Transaction,
			}

			if e.TransactionError != nil {
				transFailedErr.TransactionError = *e.TransactionError
			}
			return transFailedErr
		}

		clientErr := &ClientError{Response: r.Response}
		if len(e.Errors) > 0 {
			clientErr.ValidationErrors = make([]ValidationError, len(e.Errors))
			for i := range e.Errors {
				clientErr.ValidationErrors[i] = ValidationError{
					Description: e.Errors[i].Description,
					Field:       e.Errors[i].Field,
					Symbol:      e.Errors[i].Symbol,
				}
			}
		}
		return clientErr
	}

	// Unknown body.
	return &ClientError{Response: r.Response}
}

// Rate represents the rate limit for the current client.
type Rate struct {
	// The total request limit during the 5 minute window (e.g. requests/min * 5 min)
	Limit int

	// The number of requests remaining until your requests will be denied.
	Remaining int

	// The time when the current window will completely reset assuming no further API requests are made.
	Reset time.Time
}

// RateLimitError occurs when Recurly returns a 429 Too Many Requests error.
type RateLimitError struct {
	Response *http.Response

	Rate Rate // Rate specifies the last known rate limit for the client
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("API rate limit exceeded: %s %s: %d %v",
		e.Response.Request.Method,
		e.Response.Request.URL.Path,
		e.Response.StatusCode,
		e.Rate.Reset.Sub(time.Now()),
	)
}

// ClientError occurs when Recurly returns 400-499 status code.
// There are two known exceptions to this:
// 1) 429 Too Many Requests. See RateLimitError.
// 2) 422 Unprocessable Entity if a failed transaction occurred. See TransactionFailedError.
// All other 422 responses not related to failed transactions will return
// ClientError.
type ClientError struct {
	Response *http.Response

	// ValidationErrors holds an array of validation errors if any occurred.
	ValidationErrors []ValidationError
}

// Is returns true if one of the validation errors has a matching symbol.
func (e *ClientError) Is(symbol string) bool {
	for _, e := range e.ValidationErrors {
		if e.Symbol == symbol {
			return true
		}
	}
	return false
}

func (e *ClientError) Error() string {
	var b strings.Builder
	for i, err := range e.ValidationErrors {
		b.WriteString(err.Error())
		if i > 0 {
			b.WriteString(";")
		}
	}
	return fmt.Sprintf("client error: %s %s: %d %v",
		e.Response.Request.Method,
		e.Response.Request.URL.Path,
		e.Response.StatusCode,
		b.String(),
	)
}

// TransactionFailedError is returned when a transaction fails.
type TransactionFailedError struct {
	Response *http.Response

	// Transaction holds the failed transaction (if any).
	Transaction *Transaction

	// TransactionError holds the transaction error. This will always be
	// available.
	TransactionError TransactionError
}

func (e *TransactionFailedError) Error() string {
	return fmt.Sprintf("transaction failed: %s %s: %d [%s/%s/%s]",
		e.Response.Request.Method,
		e.Response.Request.URL.Path,
		e.Response.StatusCode,
		e.TransactionError.ErrorCode,
		e.TransactionError.ErrorCategory,
		e.TransactionError.CustomerMessage,
	)
}

// ServerError occurs when Recurly returns 500-599 status code.
type ServerError struct {
	Response *http.Response
}

func (e *ServerError) Error() string {
	return fmt.Sprintf("server error: %s %s: %d",
		e.Response.Request.Method,
		e.Response.Request.URL.Path,
		e.Response.StatusCode)
}

// ValidationError is an individual validation error.
type ValidationError struct {
	Description string
	Field       string
	Symbol      string
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		if e.Symbol == "" {
			return fmt.Sprintf("%s %s", e.Field, e.Description)
		}
		return fmt.Sprintf("%s %s (%s)", e.Field, e.Description, e.Symbol)
	} else if e.Symbol != "" {
		return fmt.Sprintf("%s (%s)", e.Description, e.Symbol)
	}
	return fmt.Sprintf("%s", e.Description)
}

// xmlSingleError is returned as a standalone error.
type xmlSingleError struct {
	XMLName     xml.Name `xml:"error"`
	Description string   `xml:"description"`
	Field       string   `xml:"field"`
	Symbol      string   `xml:"symbol"`
}

// xmlMultiErrors is a collection of various errors.
type xmlMultiErrors struct {
	XMLName xml.Name `xml:"errors"`
	Errors  []struct {
		Description string `xml:",innerxml"`
		Field       string `xml:"field,attr"`
		Symbol      string `xml:"symbol,attr"`
	} `xml:"error"`
	Transaction      *Transaction      `xml:"transaction"`
	TransactionError *TransactionError `xml:"transaction_error"`
}
