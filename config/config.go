package config

import (
	"errors"
	"fmt"
	"os"

	"lingolift/pkg/log"

	"github.com/toolkits/net"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

var (
	// G Global configuration instance
	G *LingoLiftConfig

	// AppLogger is used for recording the logs of the application.
	// including (startup logs, exception logs, job logs).
	AppLogger *zap.Logger

	// AccessLogger HTTP access logs
	AccessLogger *zap.Logger

	ServerNodeIP = ""
)

// LingoLiftConfig
type LingoLiftConfig struct {
	Filename string     `yaml:"filename"`
	App      *AppConfig `yaml:"app_conf"`

	Speech TencentCloudSpeechConfig `yaml:"tencent_speech_conf"`
}

// NewConfig
func NewConfig() *LingoLiftConfig {
	return &LingoLiftConfig{}
}

// LoadFile
func (c *LingoLiftConfig) LoadFile(filename string) error {
	c.Filename = filename

	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	if err = yaml.UnmarshalStrict(content, c); err != nil {
		return err
	}

	if err = c.check(); err != nil {
		return err
	}

	c.fillDefault()

	G = c

	fmt.Println(G.Speech)

	return nil
}

// check
func (c *LingoLiftConfig) check() error {
	//检查服务基本配置是否正确
	if err := c.App.Check(); err != nil {
		return err
	}

	return nil
}

// fillDefault
func (c *LingoLiftConfig) fillDefault() {
	ServerNodeIP = c.App.ServerIP
}

// AppConfig
type AppConfig struct {
	// Server listening IP, auto get intranet IP when empty
	ServerIP string `yaml:"server_ip"`

	// Region identifier for multi-region deployment
	Region string `yaml:"region"`

	// HTTP server configuration (timeouts, address etc.)
	HTTP *HTTPServerConfig `yaml:"http_conf"`

	// Log configuration options (output path, log level etc.)
	Log *log.Options `yaml:"log_conf"`

	// Enable Prometheus exporter metrics collection
	EnableExporterMetrics bool `yaml:"enable_exporter_metrics"`

	// Metrics exposure path, default /metrics
	MetricsPath string `yaml:"metrics_path"`

	// Enable pprof performance analysis endpoints
	EnablePProf bool `yaml:"enable_pprof"`
}

// check 检查基础配置
func (c *AppConfig) Check() error {
	if len(c.ServerIP) <= 0 {
		if err := c.ParseLocalServeIntranetIP(); err != nil {
			return err
		}
	}

	c.HTTP.fillDefault()

	return nil
}

// ParseLocalServeIntranetIP
func (c *AppConfig) ParseLocalServeIntranetIP() error {
	ips, err := net.IntranetIP()
	if err != nil {
		return fmt.Errorf("failed to get intranet IPs: %w", err)
	}

	if len(ips) == 0 {
		return errors.New("not resolve to the intranet IP.")
	}

	c.ServerIP = ips[0]

	return nil
}

// HTTPServerConfig
type HTTPServerConfig struct {
	Address        string `yaml:"address"`
	IdleTimeout    int    `yaml:"idle_timeout"`
	ReadTimeout    int    `yaml:"read_timeout"`
	WriteTimeout   int    `yaml:"write_timeout"`
	MaxHeaderBytes int    `yaml:"max_header_bytes"`
}

// checkout
func (c *HTTPServerConfig) fillDefault() {
	if c.IdleTimeout <= 0 {
		c.IdleTimeout = 60
	}

	if c.ReadTimeout <= 0 {
		c.ReadTimeout = 15
	}

	if c.WriteTimeout <= 0 {
		c.WriteTimeout = 60
	}
}

// TencentCloudSpeechConfig
type TencentCloudSpeechConfig struct {
	AppID     string `yaml:"app_id"`
	SecretID  string `yaml:"secret_id"`
	SecretKey string `yaml:"secret_key"`
	Token     string `yaml:"token"`
	SliceSize int    `yaml:"slice_size"`
}
