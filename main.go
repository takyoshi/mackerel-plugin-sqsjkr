package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"

	"github.com/kayac/sqsjkr"
	mp "github.com/mackerelio/go-mackerel-plugin-helper"
)

// Default value
const (
	DefaultSocketPath = "/var/run/sqsjkr.sock"
)

// SQSJkrPlugin mackerel plugin
type SQSJkrPlugin struct {
	Tempfile string
	Socket   string
	Prefix   string
}

// GraphDefinition interface for mackerel plugin
func (sp SQSJkrPlugin) GraphDefinition() map[string](mp.Graphs) {
	return map[string]mp.Graphs{
		"stats.worker": {
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

func fakeDial(proto, addr string) (conn net.Conn, err error) {
	return net.Dial("unix", socket)
}

// FetchMetrics interface for mackerel plugin
func (sp SQSJkrPlugin) FetchMetrics() (map[string]interface{}, error) {
	// get sqsjkr stats
	var sta sqsjkr.StatsItem
	conn, err := net.Dial("unix", sp.Socket)
	if err != nil {
		return nil, err
	}

	tr := &http.Transport{
		Dial: fakeDial,
	}
	client := &http.Client{Transport: tr}
	res, err := http.Get("http:///dummy.local/stats/metrics")
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

var (
	socket  string
	tmpfile string
	prefix  string
)

func main() {
	flag.StringVar(&socket, "socket", DefaultSocketPath, "sqsjkr stats api socket path")
	flag.StringVar(&tmpfile, "tempfile", "", "tmpfile")
	flag.StringVar(&prefix, "metric-key", "sqsjkr", "tmpfile")
	flag.Parse()

	p := SQSJkrPlugin{
		Socket:   socket,
		Tempfile: tmpfile,
		Prefix:   prefix,
	}

	if p.Tempfile == "" {
		p.Tempfile = "/tmp/mackerel-plugin-sqsjkr"
	}

	helper := mp.NewMackerelPlugin(p)
	helper.Run()
}
