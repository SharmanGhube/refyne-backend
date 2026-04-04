package monitoring


import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"go.uber.org/zap"
)

// GrafanaCloudPusher sends metrics to Grafana Cloud
type GrafanaCloudPusher struct {
	enabled        bool
	remoteWriteURL string
	username       string
	apiKey         string
	logger         *zap.Logger
	httpClient     *http.Client
	ticker         *time.Ticker
	stopChan       chan struct{}
}

// NewGrafanaCloudPusher creates a new Grafana Cloud pusher
func NewGrafanaCloudPusher(logger *zap.Logger) *GrafanaCloudPusher {
	enabled := os.Getenv("GRAFANA_CLOUD_ENABLED") == "true"
	remoteWriteURL := os.Getenv("GRAFANA_CLOUD_PROMETHEUS_URL")
	username := os.Getenv("GRAFANA_CLOUD_USERNAME")
	apiKey := os.Getenv("GRAFANA_CLOUD_API_KEY")

	return &GrafanaCloudPusher{
		enabled:        enabled,
		remoteWriteURL: remoteWriteURL,
		username:       username,
		apiKey:         apiKey,
		logger:         logger,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		stopChan: make(chan struct{}),
	}
}

// Start begins the periodic metrics push
func (g *GrafanaCloudPusher) Start(ctx context.Context) error {
	if !g.enabled {
		g.logger.Info("Grafana Cloud metrics push is disabled")
		return nil
	}

	// Validate configuration
	if g.remoteWriteURL == "" || g.username == "" || g.apiKey == "" {
		g.logger.Error("Grafana Cloud configuration incomplete",
			zap.Bool("has_url", g.remoteWriteURL != ""),
			zap.Bool("has_username", g.username != ""),
			zap.Bool("has_api_key", g.apiKey != ""),
		)
		return fmt.Errorf("incomplete Grafana Cloud configuration")
	}

	// Start pushing metrics every 15 seconds
	g.ticker = time.NewTicker(15 * time.Second)

	go func() {
		for {
			select {
			case <-g.ticker.C:
				g.pushMetrics()
			case <-g.stopChan:
				g.ticker.Stop()
				return
			case <-ctx.Done():
				g.ticker.Stop()
				return
			}
		}
	}()

	g.logger.Info("Grafana Cloud metrics pusher started",
		zap.String("remote_write_url", g.remoteWriteURL),
	)

	return nil
}

// pushMetrics collects and sends metrics to Grafana Cloud
func (g *GrafanaCloudPusher) pushMetrics() {
	// Get all registered metrics
	gatherers := prometheus.DefaultGatherer
	families, err := gatherers.Gather()
	if err != nil {
		g.logger.Error("Failed to gather metrics", zap.Error(err))
		return
	}

	if len(families) == 0 {
		return
	}

	// Encode metrics in Prometheus text format
	buf := &bytes.Buffer{}
	encoder := expfmt.NewEncoder(buf, expfmt.FmtText)

	for _, family := range families {
		if err := encoder.Encode(family); err != nil {
			g.logger.Error("Failed to encode metric family", zap.Error(err))
			return
		}
	}

	// Send to Grafana Cloud
	if err := g.sendMetricsHTTP(buf.Bytes()); err != nil {
		g.logger.Error("Failed to send metrics to Grafana Cloud", zap.Error(err))
	}
}

// sendMetricsHTTP sends metrics using HTTP POST
func (g *GrafanaCloudPusher) sendMetricsHTTP(data []byte) error {
	// Create POST request with metrics in text format
	req, err := http.NewRequest("POST", g.remoteWriteURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add basic auth header
	auth := base64.StdEncoding.EncodeToString([]byte(g.username + ":" + g.apiKey))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")
	req.Header.Set("User-Agent", "refyne-backend/1.0")

	// Send request
	resp, err := g.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send metrics: %w", err)
	}
	defer resp.Body.Close()

	// Read response body for debugging
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("grafana cloud returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Stop stops the metrics pusher
func (g *GrafanaCloudPusher) Stop() {
	if g.enabled && g.ticker != nil {
		close(g.stopChan)
	}
}
