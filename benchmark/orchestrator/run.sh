
# path to template yaml file
export bench_config="templateConf.yaml"

# path to the file containing benchmarking results
export bench_result="benchmark.yaml"

# path where config for scenarios is generated
export confScenarios="scenariosConf.yaml"

# path where report and plots are written
export report_path="report"

# setupconf.py creates config for benchmarking
python3 setupconf.py -i $bench_config -o $confScenarios

# benchmark client
#../client/client -C $confScenarios --out-benchmark $bench_result

# create report for the benchmarking results
python3 create_report.py -i $bench_result -o $report_path -c $bench_config