module gitlab.com/thorchain/thornode

go 1.13

require (
	github.com/binance-chain/go-sdk v1.1.3
	github.com/blang/semver v3.5.1+incompatible
	github.com/btcsuite/btcd v0.0.0-20190115013929-ed77733ec07d
	github.com/btcsuite/btcutil v0.0.0-20190425235716-9e5f4b9a998d
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/cenkalti/backoff/v4 v4.0.0
	github.com/cosmos/cosmos-sdk v0.37.4
	github.com/davecgh/go-spew v1.1.1
	github.com/didip/tollbooth v4.0.2+incompatible
	github.com/ethereum/go-ethereum v1.9.8
	github.com/gorilla/mux v1.7.3
	github.com/hashicorp/go-retryablehttp v0.6.4
	github.com/logrusorgru/aurora v0.0.0-20191116043053-66b7ad493a23
	github.com/mitranim/gow v0.0.0-20181105081807-8128c81042bd // indirect
	github.com/openlyinc/pointy v1.1.2
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.2.1
	github.com/rs/zerolog v1.15.0
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.4.0
	github.com/syndtr/goleveldb v1.0.1-0.20190923125748-758128399b1d
	github.com/tendermint/btcd v0.1.1
	github.com/tendermint/go-amino v0.15.1
	github.com/tendermint/tendermint v0.32.7
	github.com/tendermint/tm-db v0.2.0
	golang.org/x/sys v0.0.0-20191210023423-ac6580df4449 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127
)

replace gitlab.com/thorchain/thornode => ../thornode

replace github.com/tendermint/go-amino => github.com/binance-chain/bnc-go-amino v0.14.1-binance.1
