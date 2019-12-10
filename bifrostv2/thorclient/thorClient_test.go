package thorclient

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	. "gopkg.in/check.v1"

	"github.com/cosmos/cosmos-sdk/client/keys"
	cKeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	retryablehttp "github.com/hashicorp/go-retryablehttp"

	"gitlab.com/thorchain/thornode/bifrostv2/config"
	"gitlab.com/thorchain/thornode/bifrostv2/metrics"
	"gitlab.com/thorchain/thornode/bifrostv2/thorclient/types"
	"gitlab.com/thorchain/thornode/x/thorchain"
)

func TestPackage(t *testing.T) { TestingT(t) }

type ThorClientSuite struct {
	server *httptest.Server
}

var _ = Suite(&ThorClientSuite{})

func (s *ThorClientSuite) SetUpSuite(c *C) {
	thorchain.SetupConfigForTest()
	s.server = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}))
}

func SetupStateChainForTest(c *C) (config.ThorChainConfiguration, cKeys.Info, func()) {
	thorcliDir := filepath.Join(os.TempDir(), ".thorcli")
	cfg := config.ThorChainConfiguration{
		ChainID:         "statechain",
		ChainHost:       "localhost",
		SignerName:      "bob",
		SignerPasswd:    "password",
		ChainHomeFolder: thorcliDir,
	}
	kb, err := keys.NewKeyBaseFromDir(thorcliDir)
	c.Assert(err, IsNil)
	info, _, err := kb.CreateMnemonic(cfg.SignerName, cKeys.English, cfg.SignerPasswd, cKeys.Secp256k1)
	c.Assert(err, IsNil)
	return cfg, info, func() {
		if err := os.RemoveAll(thorcliDir); nil != err {
			c.Error(err)
		}
	}
}

func (s *ThorClientSuite) TestGetThorChainUrl(c *C) {
	cfg, _, cleanup := SetupStateChainForTest(c)
	defer cleanup()
	cfg.ChainHost = s.server.URL
	client, err := NewClient(cfg, getMetricForTest(c))
	c.Assert(err, IsNil)
	c.Assert(client, NotNil)

	uri := client.getThorChainUrl("")
	c.Assert(uri, Equals, s.server.URL)
}

func (s *ThorClientSuite) TestGet(c *C) {
	cfg, _, cleanup := SetupStateChainForTest(c)
	defer cleanup()
	cfg.ChainHost = s.server.URL
	spew.Dump(cfg)
	client, err := NewClient(cfg, getMetricForTest(c))
	c.Assert(err, IsNil)
	c.Assert(client, NotNil)

	buf, err := client.get("")
	c.Assert(err, IsNil)
	c.Assert(buf, NotNil)
}

// func (s ThorClientSuite) TestSign(c *C) {
// 	cfg, info, cleanup := SetupStateChainForTest(c)
// 	defer cleanup()
// 	// Start a local HTTP server
// 	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
// 		// Test request parameters
// 		c.Check(req.URL.String(), Equals, fmt.Sprintf("/auth/accounts/%s", info.GetAddress()))
// 		// Send response to be tested
// 		_, err := rw.Write([]byte(`{
// "height":"78",
// "result":{
// 			  "type": "cosmos-sdk/Account",
// 			  "value": {
// 				"address": "thor1vx80hen38j5w0jn6gqh3crqvktj9stnhw56kn0",
// 				"coins": [
// 				  {
// 					"denom": "thor",
// 					"amount": "1000"
// 				  }
// 				],
// 				"public_key": {
//         "type": "tendermint/PubKeySecp256k1",
//         "value": "ArYQdiiY4s1MgIEKm+7LXYQsH+ptH09neh9OWqY5VHYr"
//       },
// 				"account_number": "0",
// 				"sequence": "14"
// 			  }
// 			}}`))
// 		c.Assert(err, IsNil)
// 	}))
// 	defer server.Close()
// 	u, err := url.Parse(server.URL)
// 	c.Assert(err, IsNil)
// 	cfg.ChainHost = u.Host
// 	observedAddress := stypes.GetRandomPubKey()
// 	c.Assert(err, IsNil)
// 	tx := stypes.NewTxInVoter(common.TxID("20D150DF19DAB33405D375982E479F48F607D0C9E4EE95B146F6C35FA2A09269"), []stypes.TxIn{
// 		stypes.NewTxIn(
// 			common.Coins{
// 				common.NewCoin(common.BNBAsset, sdk.NewUint(123400000)),
// 			},
// 			"This is my memo!",
// 			common.Address("bnb1ntqj0v0sv62ut0ehxt7jqh7lenfrd3hmfws0aq"),
// 			common.Address("bnb1ntqj0v0sv62ut0ehxt7jqh7lenfrd3hmfws0aq"),
// 			common.BNBGasFeeSingleton,
// 			sdk.NewUint(1),
// 			observedAddress,
// 		),
// 	})
//
// 	bridge, err := NewClient(cfg, getMetricForTest(c))
// 	c.Assert(err, IsNil)
// 	c.Assert(bridge, NotNil)
// 	err = bridge.Start()
// 	c.Assert(err, IsNil)
// 	signedMsg, err := bridge.Sign([]stypes.TxInVoter{tx})
// 	c.Log(err)
// 	c.Assert(signedMsg, NotNil)
// 	c.Assert(err, IsNil)
// }

func (s ThorClientSuite) TestSend(c *C) {
	cfg, _, cleanup := SetupStateChainForTest(c)
	defer cleanup()
	// Start a local HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Test request parameters
		c.Check(req.URL.String(), Equals, "/txs")
		// Send response to be tested
		_, err := rw.Write([]byte(`{"txhash":"E43FA2330C4317ECC084B0C6044DFE75AAE1FAB8F84A66107809E9739D02F80D", "height": "test_height"}`))
		c.Assert(err, IsNil)
	}))
	// Close the server when test finishes
	defer server.Close()

	u, err := url.Parse(server.URL)
	c.Assert(err, IsNil)
	cfg.ChainHost = u.Host

	bridge, err := NewClient(cfg, getMetricForTest(c))
	c.Assert(err, IsNil)
	c.Assert(bridge, NotNil)
	stdTx := authtypes.StdTx{}
	mode := types.TxSync
	txID, err := bridge.Send(stdTx, mode)
	c.Assert(err, IsNil)
	c.Check(
		txID.String(),
		Equals,
		"E43FA2330C4317ECC084B0C6044DFE75AAE1FAB8F84A66107809E9739D02F80D",
	)
}
func getMetricForTest(c *C) *metrics.Metrics {
	m, err := metrics.NewMetrics(config.MetricConfiguration{
		Enabled:      false,
		ListenPort:   9000,
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
	})
	c.Assert(m, NotNil)
	c.Assert(err, IsNil)
	return m
}

func (ThorClientSuite) TestNewStateChainBridge(c *C) {
	var testFunc = func(cfg config.ThorChainConfiguration, errChecker Checker, sbChecker Checker) {
		sb, err := NewClient(cfg, getMetricForTest(c))
		c.Assert(err, errChecker)
		c.Assert(sb, sbChecker)
	}
	testFunc(config.ThorChainConfiguration{
		ChainID:         "",
		ChainHost:       "localhost",
		ChainHomeFolder: "~/.thorcli",
		SignerName:      "signer",
		SignerPasswd:    "signerpassword",
	}, NotNil, IsNil)
	testFunc(config.ThorChainConfiguration{
		ChainID:         "chainid",
		ChainHost:       "",
		ChainHomeFolder: "~/.thorcli",
		SignerName:      "signer",
		SignerPasswd:    "signerpassword",
	}, NotNil, IsNil)
	testFunc(config.ThorChainConfiguration{
		ChainID:         "chainid",
		ChainHost:       "localhost",
		ChainHomeFolder: "~/.thorcli",
		SignerName:      "",
		SignerPasswd:    "signerpassword",
	}, NotNil, IsNil)
	testFunc(config.ThorChainConfiguration{
		ChainID:         "chainid",
		ChainHost:       "localhost",
		ChainHomeFolder: "~/.thorcli",
		SignerName:      "signer",
		SignerPasswd:    "",
	}, NotNil, IsNil)
	cfg, _, cleanup := SetupStateChainForTest(c)
	testFunc(cfg, IsNil, NotNil)
	defer cleanup()
}

func (ThorClientSuite) TestGetAccountNumberAndSequenceNumber(c *C) {
	testfunc := func(handleFunc http.HandlerFunc, expectedAccNum uint64, expectedSeq uint64, errChecker Checker) {
		cfg, keyInfo, cleanup := SetupStateChainForTest(c)
		defer cleanup()
		scb, err := NewClient(cfg, getMetricForTest(c))
		c.Assert(err, IsNil)
		c.Assert(scb, NotNil)
		client := retryablehttp.NewClient()
		client.Backoff = func(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
			return time.Millisecond * 100
		}
		client.RetryMax = 3
		client.RetryWaitMax = 3 * time.Second
		scb.WithRetryableHttpClient(client)
		_ = keyInfo
		if nil != handleFunc {
			s := httptest.NewServer(handleFunc)
			defer s.Close()
			scb.cfg.ChainHost = s.Listener.Addr().String()
		}

		requestUrl := scb.getAccountInfoUrl()
		c.Logf("requestUrl:%s", requestUrl)
		if scb.cfg.ChainHost == "localhost" {
			requestUrl = ""
		}
		c.Logf("requestUrl:%s", requestUrl)
		accountNumber, seqNo, err := scb.getAccountNumberAndSequenceNumber(requestUrl)
		c.Log("account Number:", accountNumber)
		c.Log("seqNo:", seqNo)
		c.Assert(accountNumber, Equals, expectedAccNum)
		c.Assert(seqNo, Equals, expectedSeq)
		c.Assert(err, errChecker)
	}
	testfunc(nil, 0, 0, NotNil)
	testfunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusInternalServerError)
	}, 0, 0, NotNil)
	testfunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusInternalServerError)
	}, 0, 0, NotNil)
	testfunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusAccepted)
	}, 0, 0, NotNil)
	testfunc(func(writer http.ResponseWriter, request *http.Request) {
		if _, err := writer.Write([]byte("whatever")); nil != err {
			c.Error(err)
		}
	}, 0, 0, NotNil)
	testfunc(func(writer http.ResponseWriter, request *http.Request) {
		if _, err := writer.Write([]byte("")); nil != err {
			c.Error(err)
		}
	}, 0, 0, NotNil)
	testfunc(func(writer http.ResponseWriter, request *http.Request) {
		if _, err := writer.Write([]byte(`{
"height":"78",
"result":{
"type": "cosmos-sdk/Account",
"value": {
"address": "",
"coins": [],
"public_key": null,
"account_number": "asdf",
"sequence": "0"
}
}}`)); nil != err {
			c.Error(err)
		}
	}, 0, 0, NotNil)
	testfunc(func(writer http.ResponseWriter, request *http.Request) {
		if _, err := writer.Write([]byte(`{
"height":"78",
"result":{
"type": "cosmos-sdk/Account",
"value": {
"address": "",
"coins": [],
"public_key": null,
"account_number": "0",
"sequence": "whatever"
}
}}`)); nil != err {
			c.Error(err)
		}
	}, 0, 0, NotNil)
	testfunc(func(writer http.ResponseWriter, request *http.Request) {
		if _, err := writer.Write([]byte(`{
"height":"78",
"result":{
"type": "cosmos-sdk/Account",
"value": {
"address": "",
"coins": [],
"public_key": null,
"account_number": "5",
"sequence": "6"
}
}}`)); nil != err {
			c.Error(err)
		}
	}, 5, 6, IsNil)
	testfunc(func(writer http.ResponseWriter, request *http.Request) {
		if _, err := writer.Write([]byte(`{
	"height":"78",
	"result":{
  "type": "cosmos-sdk/Account",
  "value": {
    "address": "thor1vx80hen38j5w0jn6gqh3crqvktj9stnhw56kn0",
    "coins": [
      {
        "denom": "bnb",
        "amount": "1000"
      },
      {
        "denom": "btc",
        "amount": "1000"
      },
      {
        "denom": "runed",
        "amount": "1000"
      }
    ],
    "public_key": {
        "type": "tendermint/PubKeySecp256k1",
        "value": "ArYQdiiY4s1MgIEKm+7LXYQsH+ptH09neh9OWqY5VHYr"
      },
    "account_number": "0",
    "sequence": "2"
  }
}}
`)); nil != err {
			c.Error(err)
		}
	}, 0, 2, IsNil)

}

// func (ThorClientSuite) TestSignEx(c *C) {
// 	testFunc := func(in []stypes.TxInVoter, handleFunc http.HandlerFunc, resultChecker Checker, errChecker Checker) {
// 		cfg, _, cleanup := SetupStateChainForTest(c)
// 		defer cleanup()
// 		if nil != handleFunc {
// 			s := httptest.NewServer(handleFunc)
// 			defer s.Close()
// 			cfg.ChainHost = s.Listener.Addr().String()
// 		}
// 		scb, err := NewClient(cfg, getMetricForTest(c))
// 		c.Assert(err, IsNil)
// 		c.Assert(scb, NotNil)
// 		err = scb.Start()
// 		c.Assert(err, IsNil)
// 		stx, err := scb.Sign(in)
// 		c.Assert(stx, resultChecker)
// 		c.Assert(err, errChecker)
// 	}
// 	testBNBAddress, err := common.NewAddress("tbnb1hv4rmzajm3rx5lvh54sxvg563mufklw0dzyaqa")
// 	if nil != err {
// 		c.Error(err)
// 	}
// 	testFunc([]stypes.TxInVoter{
// 		{
// 			TxID: "EBB78FA6FDFBB19EBD188316B5FF9E60799C3149214A263274D31F4F605B8FDE",
// 			Txs: []stypes.TxIn{
// 				{
// 					Status:    stypes.Incomplete,
// 					OutHashes: nil,
// 					Memo:      "",
// 					Coins:     nil,
// 					Sender:    testBNBAddress,
// 				},
// 			},
// 		},
// 	}, func(writer http.ResponseWriter, request *http.Request) {
// 		fmt.Printf("RequestURL:%s", request.RequestURI)
// 		if strings.HasPrefix(request.RequestURI, "/auth/accounts") {
// 			n, err := writer.Write([]byte(`{
// 				"height":"78",
// 					"result":{
// 					"type": "cosmos-sdk/Account",
// 						"value": {
// 						"address": "thor1vx80hen38j5w0jn6gqh3crqvktj9stnhw56kn0",
// 							"coins": [
// 						{
// 							"denom": "bnb",
// 							"amount": "1000"
// 						},
// 						{
// 							"denom": "btc",
// 							"amount": "1000"
// 						},
// 						{
// 							"denom": "runed",
// 							"amount": "1000"
// 						}
// 		],
// 			"public_key": {
//         "type": "tendermint/PubKeySecp256k1",
//         "value": "ArYQdiiY4s1MgIEKm+7LXYQsH+ptH09neh9OWqY5VHYr"
//       },
// 			"account_number": "0",
// 			"sequence": "2"
// 			}
// 		}}
// 			`))
// 			c.Assert(n > 0, Equals, true)
// 			c.Assert(err, IsNil)
// 			return
// 		}
// 		writer.WriteHeader(http.StatusInternalServerError)
// 	}, NotNil, IsNil)
// }

// func (ThorClientSuite) TestSendEx(c *C) {
// 	testFunc := func(in []stypes.TxInVoter, mode types.TxMode, handleFunc http.HandlerFunc, resultChecker Checker, errChecker Checker) {
// 		cfg, _, cleanup := SetupStateChainForTest(c)
// 		defer cleanup()
// 		if nil != handleFunc {
// 			s := httptest.NewServer(handleFunc)
// 			defer s.Close()
// 			cfg.ChainHost = s.Listener.Addr().String()
// 		}
// 		scb, err := NewClient(cfg, getMetricForTest(c))
// 		c.Assert(err, IsNil)
// 		c.Assert(scb, NotNil)
// 		client := retryablehttp.NewClient()
// 		client.Backoff = func(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
// 			return time.Millisecond * 100
// 		}
// 		client.RetryMax = 3
// 		client.RetryWaitMax = 3 * time.Second
// 		scb.WithRetryableHttpClient(client)
// 		err = scb.Start()
// 		c.Assert(err, IsNil)
// 		c.Assert(scb.seqNumber, Equals, uint64(6))
// 		c.Assert(scb.accountNumber, Equals, uint64(5))
// 		stx, err := scb.Sign(in)
// 		c.Assert(stx, NotNil)
// 		c.Assert(err, IsNil)
// 		_, err = scb.Send(*stx, mode)
// 		c.Assert(err, errChecker)
//
// 	}
// 	testBNBAddress, err := common.NewAddress("tbnb1hv4rmzajm3rx5lvh54sxvg563mufklw0dzyaqa")
// 	if nil != err {
// 		c.Error(err)
// 	}
// 	txIns := []stypes.TxIn{
// 		{
// 			Status:    stypes.Incomplete,
// 			OutHashes: nil,
// 			Memo:      "",
// 			Coins:     nil,
// 			Sender:    testBNBAddress,
// 		},
// 	}
// 	txInVoters := []stypes.TxInVoter{
// 		stypes.NewTxInVoter("EBB78FA6FDFBB19EBD188316B5FF9E60799C3149214A263274D31F4F605B8FDE", txIns),
// 	}
// 	testFunc(txInVoters, types.TxUnknown, func(writer http.ResponseWriter, request *http.Request) {
// 		if _, err := writer.Write([]byte(`{
// "height":"78",
// "result":{
// "type": "cosmos-sdk/Account",
// "value": {
// "address": "",
// "coins": [],
// "public_key": null,
// "account_number": "5",
// "sequence": "6"
// }
// }}`)); nil != err {
// 			c.Error(err)
// 		}
// 	}, IsNil, NotNil)
// 	testFunc(txInVoters, types.TxSync, func(writer http.ResponseWriter, request *http.Request) {
// 		if strings.HasPrefix(request.RequestURI, "/auth/accounts") {
// 			if _, err := writer.Write([]byte(`{
// "height":"78",
// "result":{
// "type": "cosmos-sdk/Account",
// "value": {
// "address": "",
// "coins": [],
// "public_key": null,
// "account_number": "5",
// "sequence": "6"
// }
// }}`)); nil != err {
// 				c.Error(err)
// 			}
// 			return
// 		}
// 		writer.WriteHeader(http.StatusInternalServerError)
// 	}, IsNil, NotNil)
// 	testFunc(txInVoters, types.TxSync, func(writer http.ResponseWriter, request *http.Request) {
// 		if strings.HasPrefix(request.RequestURI, "/auth/accounts") {
// 			if _, err := writer.Write([]byte(`{
// "height":"78",
// "result":{
// "type": "cosmos-sdk/Account",
// "value": {
// "address": "",
// "coins": [],
// "public_key": null,
// "account_number": "5",
// "sequence": "6"
// }
// }}`)); nil != err {
// 				c.Error(err)
// 			}
// 			return
// 		}
//
// 		if _, err := writer.Write([]byte(`
// whatever`)); nil != err {
// 			c.Error(err)
// 		}
//
// 	}, IsNil, NotNil)
//
// }