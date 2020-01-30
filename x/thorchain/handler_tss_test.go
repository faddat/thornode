package thorchain

import (
	"github.com/blang/semver"
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "gopkg.in/check.v1"

	"gitlab.com/thorchain/thornode/common"
)

type HandlerTssSuite struct{}

type TestTssValidKeepr struct {
	KVStoreDummy
	na NodeAccount
}

func (k *TestTssValidKeepr) GetNodeAccount(ctx sdk.Context, signer sdk.AccAddress) (NodeAccount, error) {
	return k.na, nil
}

var _ = Suite(&HandlerTssSuite{})

func (s *HandlerTssSuite) TestValidate(c *C) {
	ctx, _ := setupKeeperForTest(c)

	keeper := &TestTssValidKeepr{
		na: GetRandomNodeAccount(NodeActive),
	}
	versionedTxOutStore := NewVersionedTxOutStoreDummy()
	versionedVaultMgrDummy := NewVersionedVaultMgrDummy(versionedTxOutStore)
	handler := NewTssHandler(keeper, versionedVaultMgrDummy)
	// happy path
	ver := semver.MustParse("0.1.0")
	pk := GetRandomPubKey()
	pks := common.PubKeys{
		GetRandomPubKey(), GetRandomPubKey(), GetRandomPubKey(),
	}
	msg := NewMsgTssPool(pks, pk, 10, keeper.na.NodeAddress)
	err := handler.validate(ctx, msg, ver)
	c.Assert(err, IsNil)

	// invalid version
	err = handler.validate(ctx, msg, semver.Version{})
	c.Assert(err, Equals, errInvalidVersion)

	// inactive node account
	keeper.na = GetRandomNodeAccount(NodeStandby)
	msg = NewMsgTssPool(pks, pk, 10, keeper.na.NodeAddress)
	err = handler.validate(ctx, msg, ver)
	c.Assert(err, Equals, notAuthorized)

	// invalid msg
	msg = MsgTssPool{}
	err = handler.validate(ctx, msg, ver)
	c.Assert(err, NotNil)
}

type TestTssHandlerKeeper struct {
	KVStoreDummy
	active NodeAccounts
	tss    TssVoter
	chains common.Chains
}

func (s *TestTssHandlerKeeper) ListActiveNodeAccounts(_ sdk.Context) (NodeAccounts, error) {
	return s.active, nil
}

func (s *TestTssHandlerKeeper) GetTssVoter(_ sdk.Context, _ string) (TssVoter, error) {
	return s.tss, nil
}

func (s *TestTssHandlerKeeper) SetTssVoter(_ sdk.Context, voter TssVoter) {
	s.tss = voter
}

func (s *TestTssHandlerKeeper) GetChains(_ sdk.Context) (common.Chains, error) {
	return s.chains, nil
}

func (s *HandlerTssSuite) TestHandle(c *C) {
	ctx, _ := setupKeeperForTest(c)
	ctx = ctx.WithBlockHeight(12)
	ver := semver.MustParse("0.1.0")

	keeper := &TestTssHandlerKeeper{
		active: NodeAccounts{GetRandomNodeAccount(NodeActive)},
		chains: common.Chains{common.BNBChain},
		tss:    TssVoter{},
	}
	versionedVaultMgrDummy := NewVersionedVaultMgrDummy(NewVersionedTxOutStoreDummy())

	handler := NewTssHandler(keeper, versionedVaultMgrDummy)
	// happy path
	pk := GetRandomPubKey()
	pks := common.PubKeys{
		GetRandomPubKey(), GetRandomPubKey(), GetRandomPubKey(),
	}
	msg := NewMsgTssPool(pks, pk, 12, keeper.active[0].NodeAddress)
	result := handler.handle(ctx, msg, ver)
	c.Assert(result.IsOK(), Equals, true)
	c.Check(keeper.tss.Signers, HasLen, 1)
	c.Check(keeper.tss.BlockHeight, Equals, int64(12))
	c.Check(versionedVaultMgrDummy.vaultMgrDummy.vault.PubKey.Equals(pk), Equals, true, Commentf("%+v\n", versionedVaultMgrDummy.vaultMgrDummy.vault))

	// running again doesn't rotate the pool again
	ctx = ctx.WithBlockHeight(10)
	result = handler.handle(ctx, msg, ver)
	c.Assert(result.IsOK(), Equals, true)
	c.Check(keeper.tss.BlockHeight, Equals, int64(12))
}
