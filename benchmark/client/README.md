# Benchmark zstor client

Provides tools for benchmarking and profiling zstor client for various scenarios.

Several scenarios can be bechmarked by the benchmarking client.
Configuration for all requested scenarios can be given from a config file in YAML format (see example of a config file bellow).
Benchmarking program outputs results for all provided scenarios to an output file in YAML format (see example of an output file bellow).




## Getting started

Build benchmark client
```
go build
```

The following optional flags are defined
```
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

Flags:
      --conf string            path to a config file (default "clientConf.yaml")
  -h, --help                   help for this command
      --out-benchmark string   path to the output file for benchmarking (default "benchmark.yaml")
      --out-profile string     path to the output directory for profiling
      --profile-mode string    enable profiling mode, one of [cpu, mem, trace, block]

```

For benchmarking with default input/output files call
``` 
./client.go
```

For benchmarking with optional input/output files call
``` 
./client.go --conf "clientFonfig.yaml" --out-benchmark string "dataset01.yaml"
```

all main function to start benchmarking and profiling
``` 
./client.go --out-profile "outputProfileInfo" --profile-mode cpu
```

## Example of a YAML config file
The following example of a config file represents two benchmarking scenarios.

``` yaml
scenarios:
  bench1:                                   # name of the scenario
    zstor_config:                           # zstor config
      organization: "<IYO organization>"
      namespace: <IYO namespace>
      iyo_app_id: "<an IYO app ID>"
      iyo_app_secret: "<an IYO app secret>"
      data_shards:
        - 127.0.0.1:12345
        - 127.0.0.1:12346
        - 127.0.0.1:12347
      meta_shards:
        - 127.0.0.1:2379
        - 127.0.0.1:22379
      block_size: 2048
      replication_nr: 2
      replication_max_size: 4096
      distribution_data: 2
      distribution_parity: 1
      compress: true
      encrypt: false
      encrypt_key: ab345678901234567890123456789012
    bench_conf:                                 # config for benchmarking 
      method: write                             # name of a benchmarking method
      result_output: per_second                 # time interval of data collection
      duration: 10                              # duration of the benchmarking in seconds
      operations: 0
      key_size: 48
      ValueSize: 128
  bench2:
    zstor_config:
      organization: "<IYO organization>"
      namespace: <IYO namespace>
      iyo_app_id: "<an IYO app ID>"
      iyo_app_secret: "<an IYO app secret>"
      data_shards:
        - 127.0.0.1:12345
        - 127.0.0.1:12346
        - 127.0.0.1:12347
      meta_shards:
        - 127.0.0.1:12345
      block_size: 2048
      replication_nr: 2
      replication_max_size: 4096
      distribution_data: 2
      distribution_parity: 1
      compress: true
      encrypt: false
      encrypt_key: ab345678901234567890123456789012
    bench_conf:
      method: write
      result_output: per_minute
      duration: 70
      operations: 20000
      key_size: 48
      ValueSize: 128
```

## Example of an output file bellow

``` yaml
bench1:
  result:
    count: 845
    duration: 10.004846224s
    perinterval:
    - 43
    - 14
    - 35
    - 107
    - 109
    - 108
    - 108
    - 109
    - 104
    - 107
    - 1
  scenarioconf:
    zstor_config:
      organization: ""
      namespace: <IYO namespace>
      iyo_app_id: ""
      iyo_app_secret: ""
      data_shards:
      - 127.0.0.1:12345
      - 127.0.0.1:12346
      - 127.0.0.1:12347
      meta_shards:
      - 127.0.0.1:12345
      block_size: 2048
      replication_nr: 2
      replication_max_size: 4096
      distribution_data: 2
      distribution_parity: 1
      compress: true
      encrypt: false
      encrypt_key: ab345678901234567890123456789012
    bench_conf:
      method: write
      result_output: per_minute
      duration: 70
      operations: 20000
      key_size: 48
      ValueSize: 128
  error: ""
bench2:
  result:
    count: 9746
    duration: 1m10.001785879s
    perinterval:
    - 8334
    - 1412
  scenarioconf:
    zstor_config:
      organization: ""
      namespace: <IYO namespace>
      iyo_app_id: ""
      iyo_app_secret: ""
      data_shards:
      - 127.0.0.1:12345
      - 127.0.0.1:12346
      - 127.0.0.1:12347
      meta_shards:
      - 127.0.0.1:12345
      block_size: 2048
      replication_nr: 2
      replication_max_size: 4096
      distribution_data: 2
      distribution_parity: 1
      compress: true
      encrypt: false
      encrypt_key: ab345678901234567890123456789012
    bench_conf:
      method: write
      result_output: per_minute
      duration: 70
      operations: 20000
      key_size: 48
      ValueSize: 128
  error: ""

```