package main

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	"github.com/zero-os/0-stor/benchmark/client/benchers"
	yaml "gopkg.in/yaml.v2"

	"github.com/pkg/profile"
	"github.com/zero-os/0-stor/benchmark/client/config"
)

//BenchmarkFlags defines flags
var BenchmarkFlags struct {
	confFile         string
	benchmarkOutPath string
	profileOutPath   string
	profileMode      string
}

var (
	//benchmarkCmd creates flags
	benchmarkCmd = &cobra.Command{
		Use:   "performance testing",
		Short: "runs benchmarking and profiling of a zstor client",
		Long: `
		
		Profiling and benchmarking of the zstor client is implemented.
		The result of benchmarking will be described in YAML format and written to file.
		
		Profiling mode is given using the --profile-mode flag, taking one of the following options:
			+ cpu
			+ mem
			+ trace 
			+ block
		In case --profile-mode is not given, no profiling will be performed.

		Output directory for profiling is given by --out-profile flag.

		Config file used to initialize the benchmarking is given by --conf flag. 
		Default config file is clientConf.yaml

		Output file for the benchmarking result can be given by --out-benchmark flag.
		Default output file is benchmark.yaml
	`,
		Run: func(cmd *cobra.Command, args []string) {},
	}
)

func init() {
	benchmarkCmd.Flags().StringVarP(&BenchmarkFlags.confFile, "conf", "C", "clientConf.yaml", "path to a config file")
	benchmarkCmd.Flags().StringVar(&BenchmarkFlags.benchmarkOutPath, "out-benchmark", "benchmark.yaml", "path and filename where benchmarking results are written")
	benchmarkCmd.Flags().StringVar(&BenchmarkFlags.profileOutPath, "out-profile", "", "path where profiling files are written")
	benchmarkCmd.Flags().StringVar(&BenchmarkFlags.profileMode, "profile-mode", "", "enable profiling mode, one of [cpu, mem, trace, block]")
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	// parse flags
	err := benchmarkCmd.Execute()
	panicOnError(err)

	// open a config file
	yamlFile, err := os.Open(BenchmarkFlags.confFile)
	panicOnError(err)

	// parse the config file to clientConf structure
	clientConf, err := config.FromReader(yamlFile)
	panicOnError(err)

	// Start profiling if profiling flag is given
	// TODO: change flags to cobra
	switch BenchmarkFlags.profileMode {
	case "cpu":
		defer profile.Start(profile.ProfilePath(BenchmarkFlags.profileOutPath), profile.CPUProfile).Stop()
	case "mem":
		defer profile.Start(profile.ProfilePath(BenchmarkFlags.profileOutPath), profile.MemProfile).Stop()
	case "trace":
		defer profile.Start(profile.ProfilePath(BenchmarkFlags.profileOutPath), profile.TraceProfile).Stop()
	case "block":
		defer profile.Start(profile.ProfilePath(BenchmarkFlags.profileOutPath), profile.BlockProfile).Stop()
	default:
	}

	//Collector of the results of benchmarking
	resultCollector := make(map[string]OutputFormat)

	//Run benchmarking for provided scenarios
	for scID, sc := range clientConf.Scenarios {
		result := new(benchers.Result)
		var b benchers.Method
		var err error

		// define the type of bencher for the method given in scenario
		benchConstructor, ok := benchers.Methods[sc.BenchConf.Method]
		if !ok {
			err = errors.New("benchmark method not found")
			goto WriteResult
		}

		// Initialize the benchmarker
		b, err = benchConstructor(scID, &sc)
		if err != nil {
			goto WriteResult
		}
		result, err = b.RunBenchmark()

		// collect results of the benchmarking cycle
	WriteResult:
		resultCollector[scID] = *FormatOutput(result, &sc, err)

	}
	yamlBytes, err := yaml.Marshal(resultCollector)
	panicOnError(err)

	err = ioutil.WriteFile(BenchmarkFlags.benchmarkOutPath, yamlBytes, 0644)
	panicOnError(err)

}
