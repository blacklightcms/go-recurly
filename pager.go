package recurly

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// Pager paginates records.
type Pager interface {
	// Count returns a total count of the request. Calling this function
	// more than once will return the value from the first call.
	Count(ctx context.Context) (int, error)

	// Next returns true if there is a next result expected, or false if
	// there is no next result.
	Next() bool

	// Cursor returns the next cursor (if available).
	Cursor() string

	// Fetch fetches results of a single page and populates dst with the results.
	// For use in a for loop with Next().
	Fetch(ctx context.Context, dst interface{}) error

	// FetchAll fetches all of the pages recurly has available for the result
	// set and populates dst with the results.
	FetchAll(ctx context.Context, dst interface{}) error
}

var _ Pager = &pager{}

// pager paginates API calls.
type pager struct {
	client *Client

	method string
	path   string
	opts   *PagerOptions

	count  *int
	cursor string

	expectResults bool
}

// returns a new pager and initializes params if nil. It ensures no cursor
// is set.
func (c *Client) newPager(method, path string, opts *PagerOptions) *pager {
	if opts == nil {
		opts = &PagerOptions{}
	}
	return &pager{
		client: c,
		method: method,
		path:   path,
		opts:   opts,
		cursor: opts.Cursor,

		expectResults: true,
	}
}

func (p *pager) Count(ctx context.Context) (int, error) {
	if p.count != nil {
		return *p.count, nil
	}

	req, err := p.client.newPagerRequest("HEAD", p.path, p.opts, nil)
	if err != nil {
		return 0, err
	}

	resp, err := p.client.do(ctx, req, nil)
	if err != nil {
		return 0, err
	}

	if count := resp.Header.Get("X-Records"); count == "" {
		return 0, nil
	} else if i, err := strconv.Atoi(count); err != nil {
		return 0, err
	} else {
		p.count = &i
		return i, nil
	}
}

func (p *pager) Next() bool { return p.expectResults }

// fetch retrieves the results and populates dst, setting the next
// cursor.
func (p *pager) Fetch(ctx context.Context, dst interface{}) error {
	select {
	default:
	case <-ctx.Done():
		return ctx.Err()
	}

	if !p.expectResults {
		return errors.New("no more results")
	}
	p.opts.Cursor = p.cursor

	req, err := p.client.newPagerRequest(p.method, p.path, p.opts, nil)
	if err != nil {
		return err
	}

	var unmarshaler struct {
		XMLName         xml.Name
		Account         []Account         `xml:"account"`
		Adjustment      []Adjustment      `xml:"adjustment"`
		AddOn           []AddOn           `xml:"add_on"`
		Coupon          []Coupon          `xml:"coupon"`
		CreditPayment   []CreditPayment   `xml:"credit_payment"`
		GiftCard        []GiftCard        `xml:"gift_card"`
		ExportDate      []ExportDate      `xml:"export_date"`
		ExportFile      []ExportFile      `xml:"export_file"`
		Invoice         []Invoice         `xml:"invoice"`
		Note            []Note            `xml:"note"`
		Plan            []Plan            `xml:"plan"`
		Redemption      []Redemption      `xml:"redemption"`
		ShippingAddress []ShippingAddress `xml:"shipping_address"`
		ShippingMethod  []ShippingMethod  `xml:"shipping_method"`
		Subscription    []Subscription    `xml:"subscription"`
		Transaction     []Transaction     `xml:"transaction"`
		Item            []Item            `xml:"item"`
	}

	resp, err := p.client.do(ctx, req, &unmarshaler)
	if err != nil {
		p.expectResults = false
		return err
	} else if p.cursor = resp.cursor; p.cursor == "" {
		p.expectResults = false
	}

	// note: this may be a good candidate for a code generator.
	switch v := dst.(type) {
	case *[]Account:
		*v = unmarshaler.Account
	case *[]Adjustment:
		*v = unmarshaler.Adjustment
	case *[]AddOn:
		*v = unmarshaler.AddOn
	case *[]Coupon:
		*v = unmarshaler.Coupon
	case *[]CreditPayment:
		*v = unmarshaler.CreditPayment
	case *[]GiftCard:
		*v = unmarshaler.GiftCard
	case *[]ExportDate:
		*v = unmarshaler.ExportDate
	case *[]ExportFile:
		*v = unmarshaler.ExportFile
	case *[]Invoice:
		*v = unmarshaler.Invoice
	case *[]Note:
		*v = unmarshaler.Note
	case *[]Plan:
		*v = unmarshaler.Plan
	case *[]Redemption:
		*v = unmarshaler.Redemption
	case *[]ShippingAddress:
		*v = unmarshaler.ShippingAddress
	case *[]ShippingMethod:
		*v = unmarshaler.ShippingMethod
	case *[]Subscription:
		*v = unmarshaler.Subscription
	case *[]Transaction:
		*v = unmarshaler.Transaction
	case *[]Item:
		*v = unmarshaler.Item
	default:
		return fmt.Errorf("unknown type used for pagination: %T", dst)
	}

	return nil
}

func (p *pager) FetchAll(ctx context.Context, dst interface{}) error {
	// Reduce HTTP calls needed by setting pagination to Recurly's max of 200.
	p.opts.PerPage = 200

	// note: this may be a good candidate for a code generator.
	switch v := dst.(type) {
	case *[]Account:
		var all []Account
		for p.Next() {
			var dst []Account
			if err := p.Fetch(ctx, &dst); err != nil {
				return err
			}
			all = append(all, dst...)
		}
		*v = all
	case *[]Adjustment:
		var all []Adjustment
		for p.Next() {
			var dst []Adjustment
			if err := p.Fetch(ctx, &dst); err != nil {
				return err
			}
			all = append(all, dst...)
		}
		*v = all
	case *[]AddOn:
		var all []AddOn
		for p.Next() {
			var dst []AddOn
			if err := p.Fetch(ctx, &dst); err != nil {
				return err
			}
			all = append(all, dst...)
		}
		*v = all
	case *[]Coupon:
		var all []Coupon
		for p.Next() {
			var dst []Coupon
			if err := p.Fetch(ctx, &dst); err != nil {
				return err
			}
			all = append(all, dst...)
		}
		*v = all
	case *[]CreditPayment:
		var all []CreditPayment
		for p.Next() {
			var dst []CreditPayment
			if err := p.Fetch(ctx, &dst); err != nil {
				return err
			}
			all = append(all, dst...)
		}
		*v = all
	case *[]GiftCard:
		var all []GiftCard
		for p.Next() {
			var dst []GiftCard
			if err := p.Fetch(ctx, &dst); err != nil {
				return err
			}
			all = append(all, dst...)
		}
		*v = all
	case *[]Invoice:
		var all []Invoice
		for p.Next() {
			var dst []Invoice
			if err := p.Fetch(ctx, &dst); err != nil {
				return err
			}
			all = append(all, dst...)
		}
		*v = all
	case *[]Note:
		var all []Note
		for p.Next() {
			var dst []Note
			if err := p.Fetch(ctx, &dst); err != nil {
				return err
			}
			all = append(all, dst...)
		}
		*v = all
	case *[]Plan:
		var all []Plan
		for p.Next() {
			var dst []Plan
			if err := p.Fetch(ctx, &dst); err != nil {
				return err
			}
			all = append(all, dst...)
		}
		*v = all
	case *[]Redemption:
		var all []Redemption
		for p.Next() {
			var dst []Redemption
			if err := p.Fetch(ctx, &dst); err != nil {
				return err
			}
			all = append(all, dst...)
		}
		*v = all
	case *[]ShippingAddress:
		var all []ShippingAddress
		for p.Next() {
			var dst []ShippingAddress
			if err := p.Fetch(ctx, &dst); err != nil {
				return err
			}
			all = append(all, dst...)
		}
		*v = all
	case *[]ShippingMethod:
		var all []ShippingMethod
		for p.Next() {
			var dst []ShippingMethod
			if err := p.Fetch(ctx, &dst); err != nil {
				return err
			}
			all = append(all, dst...)
		}
		*v = all
	case *[]Subscription:
		var all []Subscription
		for p.Next() {
			var dst []Subscription
			if err := p.Fetch(ctx, &dst); err != nil {
				return err
			}
			all = append(all, dst...)
		}
		*v = all
	case *[]Transaction:
		var all []Transaction
		for p.Next() {
			var dst []Transaction
			if err := p.Fetch(ctx, &dst); err != nil {
				return err
			}
			all = append(all, dst...)
		}
		*v = all
	default:
		return fmt.Errorf("unknown type used for pagination: %T", dst)
	}

	return nil
}

func (p *pager) Cursor() string { return p.cursor }

// PagerOptions are used to send pagination parameters with paginated requests.
type PagerOptions struct {
	// Results per page. If not provided, Recurly defaults to 50.
	PerPage int

	// The field to sort by (e.g. created_at). See Recurly's documentation.
	Sort string

	// asc or desc
	Order string

	// Returns records greater than or equal to BeginTime.
	BeginTime NullTime

	// Returns records less than or equal to EndTime.
	EndTime NullTime

	// supported by some endpoints. Check Recurly's documentation.
	State string
	// supported by some endpoints. Check Recurly's documentation.
	Type string
	// supported by some endpoints. Check Recurly's documentation.
	GifterAccountCode string
	// supported by some endpoints. Check Recurly's documentation.
	RecipientAccountCode string

	// query is for any one-off URL params used by a specific endpoint.
	// Values sent as time.Time or recurly.NullTime will be automatically
	// converted to a valid datetime format for Recurly.
	query query

	// Cursor is set internally by the library. If you are paginating
	// records non-consecutively and obtained the next cursor, you can set it
	// as the starting cursor here to continue where you left off.
	//
	// Use Pager.Cursor() to obtain the next cursor.
	Cursor string
}

type query map[string]interface{}

func (q query) append(u *url.URL) {
	if len(q) == 0 {
		return
	}

	vals := u.Query()
	for key, val := range q {
		switch v := val.(type) {
		case string:
			if v != "" {
				vals.Add(key, v)
			}
		case time.Time:
			if !v.IsZero() {
				vals.Add(key, v.UTC().Format(DateTimeFormat))
			}
		case NullTime:
			if _, ok := v.Value(); ok {
				vals.Add(key, v.String())
			}
		case bool:
			vals.Add(key, fmt.Sprintf("%t", v))
		case int, int64, uint, uint64:
			vals.Add(key, fmt.Sprintf("%d", v))
		default:
			vals.Add(key, fmt.Sprintf("%v", val))
		}
	}
	u.RawQuery = vals.Encode()
}

// append appends params to a URL.
func (p PagerOptions) append(u *url.URL) {
	if p.query == nil {
		p.query = map[string]interface{}{}
	}
	if p.PerPage > 0 {
		p.query["per_page"] = p.PerPage
	}

	if len(p.GifterAccountCode) > 0 {
		p.query["gifter_account_code"] = p.GifterAccountCode
	}

	if len(p.RecipientAccountCode) > 0 {
		p.query["recipient_account_code"] = p.RecipientAccountCode
	}

	p.query["begin_time"] = p.BeginTime.String()
	p.query["end_time"] = p.EndTime.String()
	p.query["sort"] = p.Sort
	p.query["order"] = p.Order
	p.query["state"] = p.State
	p.query["type"] = p.Type
	p.query["cursor"] = p.Cursor
	p.query.append(u)
}
