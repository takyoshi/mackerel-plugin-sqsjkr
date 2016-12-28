package mpsqsjkr

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"

	"github.com/kayac/sqsjkr"
	mp "github.com/mackerelio/go-mackerel-plugin-helper"
)

// SQSJkrPlugin mackerel plugin
type SQSJkrPlugin struct {
	Tempfile string
	Port     int
	Host     string
	Prefix   string
}

// GraphDefinition interface for mackerel plugin
func (sp SQSJkrPlugin) GraphDefinition() map[string](mp.Graphs) {
	return map[string]mp.Graphs{
		"worker": {
			Label: "sqsjkr worker number",
			Unit:  "integer",
			Metrics: []mp.Metrics{
				mp.Metrics{Name: "idle", Label: "Idle Worker Number", Diff: false, Type: "uint32"},
				mp.Metrics{Name: "busy", Label: "Busy Worker Number", Diff: false, Type: "uint32"},
			},
		},
	}
}

// MetricKeyPrefix interface for mackerel plugin
func (sp SQSJkrPlugin) MetricKeyPrefix() string {
	return sp.Prefix
}

// FetchMetrics interface for mackerel plugin
func (sp SQSJkrPlugin) FetchMetrics() (map[string]interface{}, error) {
	// get sqsjkr stats
	var sta sqsjkr.StatsItem
	endpoint := url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", sp.Host, sp.Port),
		Path:   "/stats/metrics",
	}
	res, err := http.Get(endpoint.String())
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer res.Body.Close()

	// decode json
	dec := json.NewDecoder(res.Body)
	if err := dec.Decode(&sta); err != nil {
		fmt.Println(err)
		return nil, err
	}

	stats := map[string]interface{}{
		"idle": sta.IdleWorkerNum,
		"busy": sta.BusyWorkerNum,
	}
	return stats, nil
}

// Do sqsjkr plugin
func Do() {
	var (
		port    int
		host    string
		tmpfile string
		prefix  string
	)
	flag.StringVar(&host, "host", "localhost", "sqsjkr hostname")
	flag.IntVar(&port, "port", 8061, "sqsjkr hostname")
	flag.StringVar(&tmpfile, "tempfile", "", "tmpfile")
	flag.StringVar(&prefix, "metric-key", "sqsjkr", "tmpfile")
	flag.Parse()

	p := SQSJkrPlugin{
		Port:     port,
		Host:     host,
		Tempfile: tmpfile,
		Prefix:   prefix,
	}

	if p.Tempfile == "" {
		p.Tempfile = "/tmp/mackerel-plugin-sqsjkr"
	}

	helper := mp.NewMackerelPlugin(p)
	helper.Run()
}
