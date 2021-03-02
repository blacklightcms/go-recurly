package mock

import (
	"context"

	"github.com/blacklightcms/recurly"
)

var _ recurly.GiftCardsService = &GiftCardsService{}

type GiftCardsService struct {
	OnList      func(opts *recurly.PagerOptions) recurly.Pager
	ListInvoked bool

	OnCreate      func(ctx context.Context, g recurly.GiftCard) (*recurly.GiftCard, error)
	CreateInvoked bool

	OnPreview      func(ctx context.Context, g recurly.GiftCard) (*recurly.GiftCard, error)
	PreviewInvoked bool

	OnLookup      func(ctx context.Context, id int64) (*recurly.GiftCard, error)
	LookupInvoked bool

	OnRedeem      func(ctx context.Context, accountCode, redemptionCode string) (*recurly.GiftCard, error)
	RedeemInvoked bool
}

func (g *GiftCardsService) List(opts *recurly.PagerOptions) recurly.Pager {
	g.ListInvoked = true
	return g.OnList(opts)
}

func (g *GiftCardsService) Create(ctx context.Context, gc recurly.GiftCard) (*recurly.GiftCard, error) {
	g.CreateInvoked = true
	return g.OnCreate(ctx, gc)
}

func (g *GiftCardsService) Preview(ctx context.Context, gc recurly.GiftCard) (*recurly.GiftCard, error) {
	g.PreviewInvoked = true
	return g.OnPreview(ctx, gc)
}

func (g *GiftCardsService) Lookup(ctx context.Context, id int64) (*recurly.GiftCard, error) {
	g.ListInvoked = true
	return g.OnLookup(ctx, id)
}

func (g *GiftCardsService) Redeem(ctx context.Context, accountCode, redemptionCode string) (*recurly.GiftCard, error) {
	g.RedeemInvoked = true
	return g.OnRedeem(ctx, accountCode, redemptionCode)
}
