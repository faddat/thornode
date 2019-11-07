package thorchain

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/bepswap/thornode/common"
)

type PoolAddressManagerSuite struct{}

var _ = Suite(&PoolAddressManagerSuite{})

func (ps *PoolAddressManagerSuite) SetUpSuite(c *C) {
	SetupConfigForTest()
}

func (PoolAddressManagerSuite) TestPoolAddressManager(c *C) {
	w := getHandlerTestWrapper(c, 1, true, false)
	c.Assert(w.poolAddrMgr.currentPoolAddresses.IsEmpty(), Equals, false)
	c.Assert(w.poolAddrMgr.GetCurrentPoolAddresses().IsEmpty(), Equals, false)

	rotateWindowOpenHeight := w.poolAddrMgr.currentPoolAddresses.RotateWindowOpenAt
	w.ctx = w.ctx.WithBlockHeight(rotateWindowOpenHeight)
	w.poolAddrMgr.BeginBlock(w.ctx)
	w.txOutStore.NewBlock(uint64(rotateWindowOpenHeight))
	c.Assert(w.poolAddrMgr.IsRotateWindowOpen, Equals, true)

	w.poolAddrMgr.currentPoolAddresses.Next = common.PoolPubKeys{
		common.NewPoolPubKey(common.BNBChain, 0, GetRandomPubKey()),
	}
	w.poolAddrMgr.EndBlock(w.ctx, w.txOutStore)
	// no asset get moved , because we just opened window
	c.Assert(w.txOutStore.blockOut.IsEmpty(), Equals, true)
	poolBNB := createTempNewPoolForTest(w.ctx, w.keeper, "BNB.BNB", c)
	poolTCan := createTempNewPoolForTest(w.ctx, w.keeper, "BNB.TCAN-014", c)
	poolLoki := createTempNewPoolForTest(w.ctx, w.keeper, "BNB.LOK-3C0", c)
	rotatePoolHeight := w.poolAddrMgr.currentPoolAddresses.RotateAt
	w.ctx = w.ctx.WithBlockHeight(rotatePoolHeight)
	w.txOutStore.NewBlock(uint64(rotatePoolHeight))
	w.poolAddrMgr.BeginBlock(w.ctx)
	w.poolAddrMgr.EndBlock(w.ctx, w.txOutStore)
	windowOpen := w.keeper.GetAdminConfigValidatorsChangeWindow(w.ctx, sdk.AccAddress{})
	rotatePerBlockHeight := w.keeper.GetAdminConfigRotatePerBlockHeight(w.ctx, sdk.AccAddress{})
	c.Assert(w.poolAddrMgr.currentPoolAddresses.RotateAt, Equals, 100+rotatePerBlockHeight)
	c.Assert(w.poolAddrMgr.currentPoolAddresses.RotateWindowOpenAt, Equals, 100+rotatePerBlockHeight-windowOpen)
	c.Assert(len(w.txOutStore.blockOut.TxArray) > 0, Equals, true)
	c.Assert(w.txOutStore.blockOut.Valid(), IsNil)
	totalBond := sdk.ZeroUint()
	nodeAccounts, err := w.keeper.ListNodeAccounts(w.ctx)
	c.Assert(err, IsNil)
	for _, item := range nodeAccounts {
		totalBond = totalBond.Add(item.Bond)
	}
	defaultPoolGas := PoolRefundGasKey.Default()
	poolGas, err := strconv.Atoi(defaultPoolGas)

	c.Assert(err, IsNil)
	for _, item := range w.txOutStore.blockOut.TxArray {
		c.Assert(item.Valid(), IsNil)
		// make sure the fund is sending from previous pool address to current
		c.Assert(len(item.Coins) > 0, Equals, true)
		chain := item.Coins[0].Asset.Chain
		newChainPoolAddr := w.poolAddrMgr.currentPoolAddresses.Current.GetByChain(chain)
		c.Assert(newChainPoolAddr, NotNil)
		newPoolAddr, err := newChainPoolAddr.GetAddress()
		c.Assert(err, IsNil)
		c.Assert(item.ToAddress.String(), Equals, newPoolAddr.String())
		// given we on
		if item.Coins[0].Asset.Equals(poolBNB.Asset) {
			// there are four coins , BNB,TCAN-014,LOK-3C0 and RUNE
			c.Assert(item.Coins[0].Amount.Uint64(), Equals, poolBNB.BalanceAsset.Uint64()-batchTransactionFee*4-uint64(poolGas))
		}
		if item.Coins[0].Asset.Equals(poolTCan.Asset) {
			c.Assert(item.Coins[0].Amount.Uint64(), Equals, poolTCan.BalanceAsset.Uint64())
		}
		if item.Coins[0].Asset.Equals(poolLoki.Asset) {
			c.Check(item.Coins[0].Amount.Uint64(), Equals, poolLoki.BalanceAsset.Uint64())
		}
		if common.IsRuneAsset(item.Coins[0].Asset) {
			totalRune := poolBNB.BalanceRune.Add(poolLoki.BalanceRune).Add(poolTCan.BalanceRune).Add(totalBond)
			c.Assert(item.Coins[0].Amount.String(), Equals, totalRune.String())
		}
	}
	w.txOutStore.CommitBlock(w.ctx)
}

func createTempNewPoolForTest(ctx sdk.Context, k Keeper, input string, c *C) *Pool {
	p := NewPool()
	asset, err := common.NewAsset(input)
	c.Assert(err, IsNil)
	p.Asset = asset
	// limiting balance to 59 bits, because the math done with floats looses
	// precision if the number is greater than 59 bits.
	// https://stackoverflow.com/questions/30897208/how-to-change-a-float64-number-to-uint64-in-a-right-way
	// https://github.com/golang/go/issues/29463
	p.BalanceRune = sdk.NewUint(1535169738538008)
	p.BalanceAsset = sdk.NewUint(1535169738538008)
	k.SetPool(ctx, p)
	k.AddChain(ctx, asset.Chain)
	return &p
}