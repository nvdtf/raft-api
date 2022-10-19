package api

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	repoProcessesRequested = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "repo_request_total",
		Help: "The total number of request to process a GitHub repository.",
	}, []string{
		"repo", "network",
	})
)

func RegisterRepoProcessRequestMetrics(repo string, network string) {
	repoProcessesRequested.WithLabelValues(strings.ToLower(repo), strings.ToLower(network)).Inc()
}
