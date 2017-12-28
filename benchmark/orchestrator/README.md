

## Orchestrator config file

``` yaml
benchmarks:
- prime_parameter:
    id: block_size
    range: 128, 256, 512, 1024, 2048
  second_parameter:
    id: encrypt
    range: true, false 
template:
  zstor_config:
    data_shards: 
      - 127.0.0.1:12345
    meta_shards:
      - 127.0.0.1:13000
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
    duration: 2
    operations: 0
    key_size: 48
    ValueSize: 128

```
Port of the local host give n in `data shards` is used by the orchestrator as a starting port for zstor servers deployment. Each next server uses the port +1.
Number of servers deployed is `distribution_data`+`distribution_parity`.

Port of the local host give n in `meta shards` is used by the orchestrator as a starting port for etcd servers deployment. Each next server uses the port +1.


