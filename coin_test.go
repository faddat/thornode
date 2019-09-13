package common

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"
)

type CoinSuite struct{}

var _ = Suite(&CoinSuite{})

func (s CoinSuite) TestCoin(c *C) {
	coin := NewCoin("bnb", sdk.NewUint(230000000))
	c.Check(coin.Denom.Equals(Ticker("BNB")), Equals, true)
	c.Check(coin.Amount.Uint64(), Equals, uint64(230000000))

}
