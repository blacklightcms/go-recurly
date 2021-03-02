package recurly

import (
	"context"
	"encoding/xml"
	"strconv"
)

type GiftCardsService interface {
	// List returns a list of all purchased gift cards on your site, across all accounts.
	//
	// https://developers.recurly.com/api-v2/v2.29/index.html#operation/listGiftCards
	List(opts *PagerOptions) Pager

	// Create purchases a gift card on the gifter's account.
	//
	// https://developers.recurly.com/api-v2/v2.29/index.html#operation/createGiftCard
	Create(ctx context.Context, g GiftCard) (*GiftCard, error)

	// Preview a gift card purchase.
	// Allows the gifter to confirm that the delivery details provided are correct.
	//
	// https://developers.recurly.com/api-v2/v2.29/index.html#operation/previewGiftCard
	Preview(ctx context.Context, g GiftCard) (*GiftCard, error)

	// Lookup gift card.
	//
	// https://developers.recurly.com/api-v2/v2.29/index.html#operation/lookupGiftCard
	Lookup(ctx context.Context, id int64) (*GiftCard, error)

	// Redeem a gift card on a recipient's account, outside of a subscription purchase.
	//
	// https://developers.recurly.com/api-v2/v2.29/index.html#operation/redeemGiftCardOnAccount
	Redeem(ctx context.Context, accountCode, redemptionCode string) (*GiftCard, error)
}

var _ GiftCardsService = &giftCardsImpl{}

type GiftCard struct {
	XMLName           xml.Name       `xml:"gift_card"`
	ID                int64          `xml:"id,omitempty"`
	RedemptionCode    string         `xml:"redemption_code,omitempty"`
	BalanceInCents    int            `xml:"balance_in_cents,omitempty"`
	ProductCode       string         `xml:"product_code,omitempty"`
	UnitAmountInCents int            `xml:"unit_amount_in_cents,omitempty"`
	Currency          string         `xml:"currency,omitempty"`
	Delivery          *Delivery      `xml:"delivery,omitempty"`
	GifterAccount     *GifterAccount `xml:"gifter_account,omitempty"`
	CreatedAt         NullTime       `xml:"created_at,omitempty"`
	UpdatedAt         NullTime       `xml:"updated_at,omitempty"`
	DeliveredAt       NullTime       `xml:"delivered_at,omitempty"`
	RedeemedAt        NullTime       `xml:"redeemed_at,omitempty"`
	CanceledAt        NullTime       `xml:"canceled_at,omitempty"`
}

type GifterAccount struct {
	XMLName xml.Name `xml:"gifter_account"`

	Account
}

type Delivery struct {
	XMLName         xml.Name `xml:"delivery"`
	Method          string   `xml:"method,omitempty"`
	EmailAddress    string   `xml:"email_address,omitempty"`
	DeliverAt       NullTime `xml:"deliver_at,omitempty"`
	FirstName       string   `xml:"first_name,omitempty"`
	LastName        string   `xml:"last_name,omitempty"`
	Address         *Address `xml:"address,omitempty"`
	GifterName      string   `xml:"gifter_name,omitempty"`
	PersonalMessage string   `xml:"personal_message,omitempty"`
}

type GiftCardRedemption struct {
	XMLName     xml.Name `xml:"recipient_account"`
	AccountCode string   `xml:"account_code"`
}

type giftCardsImpl serviceImpl

func (s *giftCardsImpl) List(opts *PagerOptions) Pager {
	return s.client.newPager("GET", "/gift_cards", opts)
}

func (s *giftCardsImpl) Create(ctx context.Context, g GiftCard) (*GiftCard, error) {
	req, err := s.client.newRequest("POST", "/gift_cards", g)
	if err != nil {
		return nil, err
	}

	var dst GiftCard
	if _, err := s.client.do(ctx, req, &dst); err != nil {
		return nil, err
	}

	return &dst, nil
}

func (s *giftCardsImpl) Preview(ctx context.Context, g GiftCard) (*GiftCard, error) {
	req, err := s.client.newRequest("POST", "/gift_cards/preview", g)
	if err != nil {
		return nil, err
	}

	var dst GiftCard
	if _, err := s.client.do(ctx, req, &dst); err != nil {
		return nil, err
	}

	return &dst, nil
}

func (s *giftCardsImpl) Lookup(ctx context.Context, id int64) (*GiftCard, error) {
	req, err := s.client.newRequest("GET", "/gift_cards/"+strconv.FormatInt(id, 10), nil)
	if err != nil {
		return nil, err
	}

	var dst GiftCard
	if _, err := s.client.do(ctx, req, &dst); err != nil {
		return nil, err
	}

	return &dst, nil
}

func (s *giftCardsImpl) Redeem(ctx context.Context, accountCode, redemptionCode string) (*GiftCard, error) {
	body := GiftCardRedemption{AccountCode: accountCode}
	req, err := s.client.newRequest("POST", "/gift_cards/"+redemptionCode+"/redeem", body)
	if err != nil {
		return nil, err
	}

	var dst GiftCard
	if _, err := s.client.do(ctx, req, &dst); err != nil {
		return nil, err
	}

	return &dst, nil
}
