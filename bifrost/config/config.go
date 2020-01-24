package config

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Configuration struct {
	Observer  ObserverConfiguration  `json:"observer" mapstructure:"observer"`
	Signer    SignerConfiguration    `json:"signer" mapstructure:"signer"`
	Thorchain ThorchainConfiguration `json:"thorchain" mapstructure:"thorchain"`
	Metric    MetricConfiguration    `json:"metric" mapstructure:"metric"`
	Binance   BinanceConfiguration   `json:"binance" mapstructure:"binance"`
	TSS       TSSConfiguration       `json:"tss" mapstructure:"tss"`
}

// ObserverConfiguration values
type ObserverConfiguration struct {
	ObserverDbPath string                    `json:"observer_db_path" mapstructure:"observer_db_path"`
	BlockScanner   BlockScannerConfiguration `json:"block_scanner" mapstructure:"block_scanner"`
	RetryInterval  time.Duration             `json:"retry_interval" mapstructure:"retry_interval"`
}

// SignerConfiguration all the configures need by signer
type SignerConfiguration struct {
	SignerDbPath  string                    `json:"signer_db_path" mapstructure:"signer_db_path"`
	BlockScanner  BlockScannerConfiguration `json:"block_scanner" mapstructure:"block_scanner"`
	RetryInterval time.Duration             `json:"retry_interval" mapstructure:"retry_interval"`
}

// BinanceConfiguration all the configurations for binance client
type BinanceConfiguration struct {
	RPCHost string `json:"rpc_host" mapstructure:"rpc_host"`
}

// TSSConfiguration
type TSSConfiguration struct {
	Scheme string `json:"scheme" mapstructure:"scheme"`
	Host   string `json:"host" mapstructure:"host"`
	Port   int    `json:"port" mapstructure:"port"`
}

// BlockScannerConfiguration settings for BlockScanner
type BlockScannerConfiguration struct {
	RPCHost                    string        `json:"rpc_host" mapstructure:"rpc_host"`
	StartBlockHeight           int64         `json:"-"`
	BlockScanProcessors        int           `json:"block_scan_processors" mapstructure:"block_scan_processors"`
	HttpRequestTimeout         time.Duration `json:"http_request_timeout" mapstructure:"http_request_timeout"`
	HttpRequestReadTimeout     time.Duration `json:"http_request_read_timeout" mapstructure:"http_request_read_timeout"`
	HttpRequestWriteTimeout    time.Duration `json:"http_request_write_timeout" mapstructure:"http_request_write_timeout"`
	MaxHttpRequestRetry        int           `json:"max_http_request_retry" mapstructure:"max_http_request_retry"`
	BlockHeightDiscoverBackoff time.Duration `json:"block_height_discover_back_off" mapstructure:"block_height_discover_back_off"`
	BlockRetryInterval         time.Duration `json:"block_retry_interval" mapstructure:"block_retry_interval"`
	EnforceBlockHeight         bool          `json:"enforce_block_height" mapstructure:"enforce_block_height"`
	ChainID                    string        `json:"chain_id" mapstructure:"chain_id"`
}

// ThorchainConfiguration
type ThorchainConfiguration struct {
	ChainID         string `json:"chain_id" mapstructure:"chain_id" `
	ChainHost       string `json:"chain_host" mapstructure:"chain_host"`
	ChainHomeFolder string `json:"chain_home_folder" mapstructure:"chain_home_folder"`
	SignerName      string `json:"signer_name" mapstructure:"signer_name"`
	SignerPasswd    string `json:"signer_passwd" mapstructure:"signer_passwd"`
}

type MetricConfiguration struct {
	Enabled      bool          `json:"enabled" mapstructure:"enabled"`
	ListenPort   int           `json:"listen_port" mapstructure:"listen_port"`
	ReadTimeout  time.Duration `json:"read_timeout" mapstructure:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout" mapstructure:"write_timeout"`
}

// LoadConfig
func LoadConfig(file string) (*Configuration, error) {
	applyDefaultConfig()
	var cfg Configuration
	viper.AddConfigPath(".")
	viper.AddConfigPath(filepath.Dir(file))
	viper.SetConfigName(strings.TrimRight(path.Base(file), ".json"))
	if err := viper.ReadInConfig(); nil != err {
		return nil, errors.Wrap(err, "fail to read from config file")
	}
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	if err := viper.Unmarshal(&cfg); nil != err {
		return nil, errors.Wrap(err, "fail to unmarshal")
	}
	return &cfg, nil
}

func applyDefaultConfig() {
	viper.SetDefault("metric.listen_port", "9000")
	viper.SetDefault("metric.read_timeout", "30s")
	viper.SetDefault("metric.write_timeout", "30s")
	viper.SetDefault("thorchain.chain_id", "thorchain")
	viper.SetDefault("thorchain.chain_host", "localhost:1317")
	applyDefaultObserverConfig()
	applyDefaultSignerConfig()
}

func applyBlockScannerDefault(path string) {
	viper.SetDefault(fmt.Sprintf("%s.block_scanner.start_block_height", path), "0")
	viper.SetDefault(fmt.Sprintf("%s.block_scanner.block_scan_processors", path), "2")
	viper.SetDefault(fmt.Sprintf("%s.block_scanner.http_request_timeout", path), "30s")
	viper.SetDefault(fmt.Sprintf("%s.block_scanner.http_request_read_timeout", path), "30s")
	viper.SetDefault(fmt.Sprintf("%s.block_scanner.http_request_write_timeout", path), "30s")
	viper.SetDefault(fmt.Sprintf("%s.block_scanner.max_http_request_retry", path), "10")
	viper.SetDefault(fmt.Sprintf("%s.block_scanner.block_height_discover_back_off", path), "1s")
	viper.SetDefault(fmt.Sprintf("%s.block_scanner.block_retry_interval", path), "1s")
}

func applyDefaultObserverConfig() {
	viper.SetDefault("observer.observer_db_path", "observer_data")
	viper.SetDefault("observer.retry_interval", "2s")
	applyBlockScannerDefault("observer")
	viper.SetDefault("observer.block_scanner.chain_id", "BNB")
}

func applyDefaultSignerConfig() {
	viper.SetDefault("signer.signer_db_path", "signer_db")
	applyBlockScannerDefault("signer")
	viper.SetDefault("signer.retry_interval", "2s")
	viper.SetDefault("signer.block_scanner.chain_id", "ThorChain")
}
