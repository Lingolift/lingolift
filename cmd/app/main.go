package main

import (
	"fmt"
	"os"

	"lingolift/config"
	"lingolift/job"
	"lingolift/pkg/log"
	"lingolift/server"

	"github.com/alecthomas/kingpin"
	"github.com/prometheus/common/version"
	"go.uber.org/zap"
)

// initLibraries 初始化库
func initLibraries(cfg *config.LingoLiftConfig, logger *zap.Logger) (err error) {
	go job.HealthCheck()
	return
}

func main() {
	var (
		listenAddress = kingpin.Flag(
			"web.listen-address",
			"Address on which to expose metrics and web interface.",
		).Default("").String()
		metricsPath = kingpin.Flag(
			"web.telemetry-path",
			"Path under which to expose metrics.",
		).Default("/metrics").String()
		enableExporterMetrics = kingpin.Flag(
			"web.enable-exporter-metrics",
			"Include metrics about the exporter itself (http_*, process_*, go_*).",
		).Default("true").Bool()
		configFile = kingpin.Flag(
			"config.file", "Kingsoft cloud monitor openapi configuration file.",
		).Default("cfg.yml").String()
	)

	kingpin.Version(version.Print("monitor-openapi"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	openAPIConf := config.NewConfig()
	if err := openAPIConf.LoadFile(*configFile); err != nil {
		fmt.Fprintln(os.Stderr, "Unable to load configuration file:", configFile, err)
		os.Exit(1)
	}

	openAPIConf.App.EnableExporterMetrics = *enableExporterMetrics
	openAPIConf.App.MetricsPath = *metricsPath

	// Initialize the service application log instance.
	config.AppLogger = log.NewLogger(func(option *log.Options) {
		option.LogFileDir = openAPIConf.App.Log.LogFileDir
		option.AppName = "app"
	})

	config.AppLogger.Info(`Load configuration file successfully.`, zap.String(`service`, `monitor-openapi`))

	if err := initLibraries(openAPIConf, config.AppLogger); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to init library:", *configFile, err)
		os.Exit(1)
	}

	server.NewHTTPServerWithConfig(openAPIConf.App, config.AppLogger, *listenAddress)

	os.Exit(0)
}
