package main

import (
	"io/ioutil"

	"github.com/zero-os/0-stor/benchmark/client/benchers"
	"github.com/zero-os/0-stor/benchmark/client/config"
	yaml "gopkg.in/yaml.v2"
)

// NewOutputFormat returns a new OutputFormat
func NewOutputFormat() OutputFormat {
	var o OutputFormat
	o.Scenarios = make(map[string]ScenarioOutputFormat)
	return o
}

// OutputFormat represents the output format of a full benchmark
type OutputFormat struct {
	Scenarios map[string]ScenarioOutputFormat
}

//ScenarioOutputFormat represents a scenario result for outputting
type ScenarioOutputFormat struct {
	Result       *benchers.Result
	ScenarioConf *config.Scenario
	Error        string `yaml:"error"`
}

//FormatOutput formats the output of the benchmarking program
func FormatOutput(result *benchers.Result, scenarioConfig *config.Scenario, err error) *ScenarioOutputFormat {
	output := new(ScenarioOutputFormat)
	if err != nil {
		output.Error = err.Error()
		return output
	}
	output.Result = result
	output.ScenarioConf = scenarioConfig
	return output
}

// writeOutput writes OutputFormat to provided file
func writeOutput(filePath string, output OutputFormat) error {
	yamlBytes, err := yaml.Marshal(output)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filePath, yamlBytes, 0644)
}
