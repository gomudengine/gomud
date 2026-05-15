package usercommands

import (
	"github.com/GoMudEngine/GoMud/internal/characters"
	"github.com/GoMudEngine/GoMud/internal/mobs"
	"github.com/GoMudEngine/GoMud/internal/rooms"
	"github.com/GoMudEngine/GoMud/internal/users"
	"github.com/GoMudEngine/GoMud/internal/util"
)

// ShopListRequest is the value threaded through OnShopList and
// OnShopListRendered handlers. Stock is the effective inventory for this
// seller; handlers may add, remove, or modify entries (including Price)
// before the list or buy command uses it. Exactly one of SellerMob and
// SellerUser is non-nil. Buyer and Room are always non-nil.
// IsBuy is true when the request originates from the buy command, false
// when it originates from the list command.
type ShopListRequest struct {
	Stock      characters.Shop
	Buyer      *users.UserRecord
	SellerMob  *mobs.Mob
	SellerUser *users.UserRecord
	Room       *rooms.Room
	IsBuy      bool
}

// OnShopList is fired once per seller before the stock list is rendered
// (list command) or matched against a purchase request (buy command).
// Modules register handlers here to dynamically modify shop inventory.
//
// Handlers may add items to Stock, remove items to block purchase, or
// modify ShopItem fields (including Price) to override pricing.
//
// Example registration from a module:
//
//	usercommands.OnShopList.Register(func(r usercommands.ShopListRequest) usercommands.ShopListRequest {
//	    if r.SellerMob != nil && r.SellerMob.MobId == mySpecialMobId {
//	        r.Stock = append(r.Stock, characters.ShopItem{ItemId: 42, Price: 100})
//	    }
//	    return r
//	})
var OnShopList util.Hook[ShopListRequest]

// InsufficientFundsRequest is passed to OnInsufficientFunds handlers when a
// buyer cannot afford a purchase. Gold is the buyer's current gold. Price is
// the full price that was required (may be tax-inclusive). Handlers may send
// a custom message to the buyer and set Handled to true to suppress the
// default "You don't have enough gold" message from the buy command.
type InsufficientFundsRequest struct {
	Buyer      *users.UserRecord
	SellerMob  *mobs.Mob
	SellerUser *users.UserRecord
	Room       *rooms.Room
	Gold       int
	Price      int
	Handled    bool
}

// OnInsufficientFunds is fired when a buyer's gold is less than the purchase
// price. Handlers may inspect the request, send a custom message to the buyer,
// and set Handled to true to prevent the default insufficient-funds message.
var OnInsufficientFunds util.Hook[InsufficientFundsRequest]

// OnShopListRendered is fired once per seller after all shop tables for
// that seller have been sent to the buyer (list command only; not fired
// during buy). Modules register handlers here to append additional text
// after the listing, such as zone tax notices.
//
// Example registration from a module:
//
//	usercommands.OnShopListRendered.Register(func(r usercommands.ShopListRequest) usercommands.ShopListRequest {
//	    r.Buyer.SendText("Some extra info about this shop.")
//	    return r
//	})
var OnShopListRendered util.Hook[ShopListRequest]
