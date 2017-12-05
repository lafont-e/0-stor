package main

import (
	"github.com/zero-os/0-stor/benchmark/client/benchers"
	"github.com/zero-os/0-stor/benchmark/client/config"
)

//OutputFormat represents results of benchmarking and scenario config
type OutputFormat struct {
	Result       *benchers.Result
	ScenarioConf *config.Scenario
	Error        string `yaml:"error"`
}

//FormatOutput formats the output of the benchmarking program
func FormatOutput(result *benchers.Result, scenarioConfig *config.Scenario, err error) *OutputFormat {
	output := new(OutputFormat)
	if err != nil {
		output.Error = err.Error()
		return output
	}
	output.Result = result
	output.ScenarioConf = scenarioConfig
	return output
}
