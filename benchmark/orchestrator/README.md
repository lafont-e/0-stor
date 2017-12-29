# Benchmark orchestrator

Benchmark orchestrator provides tools to analyse performance of `zstor`.

# Getting started
To start the benchmarking, zstor 

Run the benchmark orchestrator using optional parameters:
``` bash
python3 orchestrator --conf orchCong.yaml --out report
```

Here are the options of the orchestrator:
``` bash
optional arguments:
  -h, --help            help for orchestrator
  -C  --conf string     path to the config file (default rchConfig.yaml)
  --out string          directory where the benchmark report will be
                        written (default ./report)
```

## Orchestrator config file
Config file for the `orchestrator` consists of two parts:

  * `template` represents the template of the config file for the benchmark client. it consists of the two parts itself:

  * `benchmarks` contains set of the benchmark parameter.
Multiple benchmarks can be added to the final report. 

The config for each benchmark consists of `prime_parameter` and optional `second_parameter`. 
If no benchmark parameters are given, `template` will be directly used to run benchmarking and output the system throughput.
If only `prime_parameter` is given, `orchestrator` creates a plot of the system throughput versus values in `range` of the `prime parameter`.
If both `prime_parameter` and `second_parameter` are given, a few plots will be combined in the output figure, one for each value in `range` of `second_parameter`.

Here is an example of the config file:
``` yaml
benchmarks:
- prime_parameter:
    id: distribution_data
    range: 2,3,4,5
- prime_parameter:
    id: block_size
    range: 128, 256, 512, 1024, 2048
  second_parameter:
    id: encrypt
    range: true, false
template:
  zstor_config:
    data_shards:
      - 127.0.0.1:1200
    meta_shards:
      - 127.0.0.1:1300
    meta_shards_nr: 2
    block_size: 2048
    replication_nr: 2
    replication_max_size: 4096
    distribution_data: 2
    distribution_parity: 1
    compress: true
    encrypt: false
    encrypt_key: ab345678901234567890123456789012
  bench_conf:
    clients: 1
    method: write
    result_output: per_second
    duration: 10
    operations: 0
    key_size: 48
    ValueSize: 128
```
Port of the local host given in `data shards` is used by the orchestrator as a starting port for zstor servers deployment. Each next server uses the port +1.
Number of servers deployed is `distribution_data`+`distribution_parity`.

Port of the local host give n in `meta shards` is used by the orchestrator as a starting port for etcd servers deployment. Each next server uses the port +1.

 
