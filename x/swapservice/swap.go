package swapservice

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"github.com/jpthor/cosmos-swap/config"
)

func isEmptyString(input string) bool {
	return strings.TrimSpace(input) == ""
}

// validateMessage is trying to validate the legitimacy of the incoming message and decide whether we can handle it
func validateMessage(ctx sdk.Context, keeper poolStorage, source, target Ticker, amount, requester, destination, requestTxHash, tradeSlipLimit string) error {
	if isEmptyString(requestTxHash) {
		return errors.New("request tx hash is empty")
	}
	if source.Empty() {
		return errors.New("source is empty")
	}
	if target.Empty() {
		return errors.New("target is empty")
	}
	if isEmptyString(amount) {
		return errors.New("amount is empty")
	}
	if isEmptyString(requester) {
		return errors.New("requester is empty")
	}
	if isEmptyString(destination) {
		return errors.New("destination is empty")
	}
	if isEmptyString(tradeSlipLimit) {
		return errors.New("trade slip limit is empty")
	}
	if !IsRune(source) {
		if !keeper.PoolExist(ctx, source) {
			return errors.New(fmt.Sprintf("%s doesn't exist", source))
		}
	}
	if !IsRune(target) {
		if !keeper.PoolExist(ctx, target) {
			return errors.New(fmt.Sprintf("%s doesn't exist", target))
		}
	}
	return nil
}

func swap(ctx sdk.Context, keeper poolStorage, setting *config.Settings, source, target Ticker, amount, requester, destination, requestTxHash, tradeSlipLimit string) (string, error) {
	if err := validateMessage(ctx, keeper, source, target, amount, requester, destination, requestTxHash, tradeSlipLimit); nil != err {
		ctx.Logger().Error(err.Error())
		return "0", err
	}
	isDoubleSwap := !IsRune(source) && !IsRune(target)
	swapRecord := SwapRecord{
		RequestTxHash:   requestTxHash,
		SourceTicker:    source,
		TargetTicker:    target,
		Requester:       requester,
		Destination:     destination,
		AmountRequested: amount,
	}
	if isDoubleSwap {
		runeAmount, err := swapOne(ctx, keeper, setting, source, RuneTicker, amount, requester, destination, tradeSlipLimit)
		if err != nil {
			return "0", errors.Wrapf(err, "fail to swap from %s to %s", source, RuneTicker)
		}
		tokenAmount, err := swapOne(ctx, keeper, setting, RuneTicker, target, runeAmount, requester, destination, tradeSlipLimit)
		swapRecord.AmountPaidBack = tokenAmount
		if err := keeper.SetSwapRecord(ctx, swapRecord); nil != err {
			ctx.Logger().Error("fail to save swap record", "error", err)
		}
		return tokenAmount, err
	}
	tokenAmount, err := swapOne(ctx, keeper, setting, source, target, amount, requester, destination, tradeSlipLimit)
	swapRecord.AmountPaidBack = tokenAmount
	if err := keeper.SetSwapRecord(ctx, swapRecord); nil != err {
		ctx.Logger().Error("fail to save swap record", "error", err)
	}
	return tokenAmount, err
}

func swapOne(ctx sdk.Context,
	keeper poolStorage,
	settings *config.Settings,
	source, target Ticker, amount, requester, destination, tradeSlipLimit string) (string, error) {
	ctx.Logger().Info(fmt.Sprintf("%s Swapping %s(%s) -> %s to %s", requester, source, amount, target, destination))
	ticker := source
	if IsRune(source) {
		ticker = target
	}
	if !keeper.PoolExist(ctx, ticker) {
		ctx.Logger().Debug(fmt.Sprintf("pool %s doesn't exist", ticker))
		return "0", errors.New(fmt.Sprintf("pool %s doesn't exist", ticker))
	}

	amt, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return "0", errors.Wrapf(err, "amount:%s is not valid", amount)
	}
	fslipLimit, err := strconv.ParseFloat(tradeSlipLimit, 64)
	if err != nil {
		return "0", errors.Wrapf(err, "trade slip limit %s is not valid", tradeSlipLimit)
	}

	pool := keeper.GetPoolStruct(ctx, ticker)
	if pool.Status != PoolEnabled {
		return "0", errors.Errorf("pool %s is in %s status, can't swap", ticker, pool.Status)
	}
	balanceRune, err := strconv.ParseFloat(pool.BalanceRune, 64)
	if err != nil {
		return "0", errors.Wrapf(err, "pool rune balance %s is invalid", pool.BalanceRune)
	}
	balanceToken, err := strconv.ParseFloat(pool.BalanceToken, 64)
	if err != nil {
		return "0", errors.Wrapf(err, "pool token balance %s is invalid", pool.BalanceToken)
	}
	poolSlip := calculatePoolSlip(source, balanceRune, balanceToken, amt)
	if poolSlip > settings.GlobalPoolSlip {
		return "0", errors.Errorf("pool slip:%f is over global pool slip limit :%f", poolSlip, settings.GlobalPoolSlip)
	}
	userPrice := calculateUserPrice(source, balanceRune, balanceToken, amt)
	if math.Abs(userPrice-fslipLimit)/fslipLimit > settings.GlobalTradeSlipLimit {
		return "0", errors.Errorf("user price %f is more than %.2f percent different than %f", userPrice, settings.GlobalTradeSlipLimit*100, fslipLimit)
	}
	// do we have enough balance to swap?
	if IsRune(source) {
		if balanceToken == 0 {
			return "0", errors.New("token :%s balance is 0, can't do swap")
		}
	} else {
		if balanceRune == 0 {
			return "0", errors.New(RuneTicker.String() + " balance is 0, can't swap ")
		}
	}
	ctx.Logger().Info(fmt.Sprintf("Pre-Pool: %sRune %sToken", pool.BalanceRune, pool.BalanceToken))
	newBalanceRune, newBalanceToken, returnAmt, err := calculateSwap(source, balanceRune, balanceToken, amt)
	if nil != err {
		return "0", errors.Wrap(err, "fail to swap")
	}
	pool.BalanceRune = strconv.FormatFloat(newBalanceRune, 'f', 8, 64)
	pool.BalanceToken = strconv.FormatFloat(newBalanceToken, 'f', 8, 64)
	returnTokenAmount := strconv.FormatFloat(returnAmt, 'f', 8, 64)
	keeper.SetPoolStruct(ctx, ticker, pool)
	ctx.Logger().Info(fmt.Sprintf("Post-swap: %sRune %sToken , user get:%s ", pool.BalanceRune, pool.BalanceToken, returnTokenAmount))
	return returnTokenAmount, nil
}

// calculateUserPrice return trade slip
func calculateUserPrice(source Ticker, balanceRune, balanceToken, amt float64) float64 {
	if IsRune(source) {
		return math.Pow(balanceRune+amt, 2.0) / (balanceRune * balanceToken)
	}
	return math.Pow(balanceToken+amt, 2.0) / (balanceRune * balanceToken)
}

// calculatePoolSlip the slip of total pool
func calculatePoolSlip(source Ticker, balanceRune, balanceToken, amt float64) float64 {
	if IsRune(source) {
		return amt * (2*balanceRune + amt) / math.Pow(balanceRune, 2.0)
	}
	return amt * (2*balanceToken + amt) / math.Pow(balanceToken, 2.0)
}

// calculateSwap how much rune, token and amount to emit
// return (Rune,Token,Amount)
func calculateSwap(source Ticker, balanceRune, balanceToken, amt float64) (float64, float64, float64, error) {
	if amt <= 0.0 {
		return balanceRune, balanceToken, 0.0, errors.New("amount is invalid")
	}
	if balanceRune <= 0 || balanceToken <= 0 {
		return balanceRune, balanceToken, amt, errors.New("invalid balance")
	}
	if IsRune(source) {
		balanceRune += amt
		tokenAmount := (amt * balanceToken) / balanceRune
		liquidityFee := math.Pow(amt, 2.0) * balanceToken / math.Pow(balanceRune, 2.0)
		tokenAmount -= liquidityFee
		balanceToken = balanceToken - tokenAmount
		return balanceRune, balanceToken, tokenAmount, nil
	} else {
		balanceToken += amt
		runeAmt := (balanceRune * amt) / balanceToken
		liquidityFee := (math.Pow(amt, 2.0) * balanceRune) / math.Pow(balanceToken, 2.0)
		runeAmt -= liquidityFee
		balanceRune = balanceRune - runeAmt
		return balanceRune, balanceToken, runeAmt, nil
	}
}

// swapComplete  mark a swap to be in complete state
func swapComplete(ctx sdk.Context, keeper poolStorage, requestTxHash, payTxHash string) error {
	if isEmptyString(requestTxHash) {
		return errors.New("request tx hash is empty")
	}
	if isEmptyString(payTxHash) {
		return errors.New("pay tx hash is empty")
	}
	return keeper.UpdateSwapRecordPayTxHash(ctx, requestTxHash, payTxHash)
}
