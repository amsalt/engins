package monitor

import (
	"fmt"
	"sync"
)

var metrics map[string]Metric
var mutex sync.RWMutex

func init() {
	metrics = make(map[string]Metric)
	RegisterMetric(&HelpMetric{})
}

// RegisterMetric registers new Metric to monitor.
func RegisterMetric(metric Metric) {
	mutex.Lock()
	defer mutex.Unlock()

	_, exist := metrics[metric.cmd()]
	if exist {
		panic(fmt.Errorf("metric with cmd name %+v has been registered", metric.cmd()))
	}
	metrics[metric.cmd()] = metric
}

func getMetric(cmd string) Metric {
	mutex.RLock()
	defer mutex.RUnlock()
	return metrics[cmd]
}

// Metric represents a kind of metric.
type Metric interface {
	// cmd the name of command
	cmd() string

	// desc the description of command
	desc() string

	// run the action of this command
	run() string
}

type HelpMetric struct {
}

func (h *HelpMetric) cmd() string {
	return "help"
}

func (h *HelpMetric) desc() string {
	return "type `help` for more information"
}

func (h *HelpMetric) run() string {
	result := "The commands are:\n\n"
	for n, m := range metrics {
		result += fmt.Sprintf("    %-15s %s\n", n, m.desc())
	}

	return result
}
