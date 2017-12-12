package main

import (
	"errors"
	"log"
	"os"
	"sync"

	"github.com/spf13/cobra"
	"github.com/zero-os/0-stor/benchmark/client/benchers"

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
	//rootCmd creates flags
	rootCmd = &cobra.Command{
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
		Run: func(cmd *cobra.Command, args []string) {
			root(cmd)
		},
	}
)

func init() {
	rootCmd.Flags().StringVarP(&BenchmarkFlags.confFile, "conf", "C", "config.yaml", "path to a config file")
	rootCmd.Flags().StringVar(&BenchmarkFlags.benchmarkOutPath, "out-benchmark", "benchmark.yaml", "path and filename where benchmarking results are written")
	rootCmd.Flags().StringVar(&BenchmarkFlags.profileOutPath, "out-profile", "profile", "path where profiling files are written")
	rootCmd.Flags().StringVar(&BenchmarkFlags.profileMode, "profile-mode", "", "enable profiling mode, one of [cpu, mem, trace, block]")
}

func main() {
	rootCmd.Execute()
}

func root(cmd *cobra.Command) {
	// open a config file
	yamlFile, err := os.Open(BenchmarkFlags.confFile)
	if err != nil {
		log.Fatal(err)
	}

	// parse the config file to clientConf structure
	clientConf, err := config.FromReader(yamlFile)
	if err != nil {
		log.Fatal(err)
	}

	// close config file
	err = yamlFile.Close()
	if err != nil {
		log.Fatal(err)
	}

	// Start profiling if profiling flag is given
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

	output := NewOutputFormat()

	//Run benchmarking for provided scenarios
	for scID, sc := range clientConf.Scenarios {
		var b benchers.Method
		var err error
		var clients []benchers.Method
		var cc int // client count
		var wg sync.WaitGroup
		var results []*benchers.Result

		// define the type of bencher for the method given in scenario
		benchConstructor, ok := benchers.Methods[sc.BenchConf.Method]
		if !ok {
			err = errors.New("benchmark method not found")
			goto WriteResult
		}

		// get concurrent clients
		cc = sc.BenchConf.Clients
		if cc < 1 {
			cc = 1
		}
		clients = make([]benchers.Method, cc)
		results = make([]*benchers.Result, cc)

		// init clients concurrently
		for i := range clients {
			wg.Add(1)
			go func(i int) {
				b, err = benchConstructor(scID, &sc)
				clients[i] = b
				wg.Done()
			}(i)
		}
		wg.Wait()
		if err != nil {
			goto WriteResult
		}

		// run benchmarks concurrently
		for i := range clients {
			wg.Add(1)
			go func(m benchers.Method, i int) {
				var result *benchers.Result
				result, err = b.RunBenchmark()
				results[i] = result
				wg.Done()
			}(clients[i], i)
		}
		wg.Wait()

		// collect results of the benchmarking cycle
	WriteResult:
		scBuf := sc
		output.Scenarios[scID] = *FormatOutput(results, &scBuf, err)
	}

	// write results to file
	err = writeOutput(BenchmarkFlags.benchmarkOutPath, output)
	if err != nil {
		log.Fatal(err)
	}
}
