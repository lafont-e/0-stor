"""
    Orchestrator controls benchmarking process and generates report.
"""

import pdb #pdb.set_trace()
import sys
from argparse import ArgumentParser
from yaml import dump
from subprocess import run
from lib import Config
from lib import Report
from lib import Aggregator

def main(argv):
    # default path to template yaml file
    input_config = "orchConfig.yaml"

    # default path where config for scenarios is written
    output_config = "scenariosConf.yaml"

    # default path to the benchmark results
    result_benchmark_file = "benchmarkResult.yaml"
    
    report_directory = "report"
    # TODO - handle flags with argparse
    #try:
    #    opts, args = getopt(argv,"hi:o:",["ifile=","ofile="])
    #    # check if output directories are given
    #    for opt, arg in opts:
    #        if opt == '-i':
    #            # set new file for input
    #            input_config = arg
    #        if opt == '-o':
    #            # set new file for output
    #            output_config = arg      
    #except:
    #    pass           

    print('********************')
    print('****Benchmarking****')
    print('********************')

    report = Report(report_directory)
    config = Config(input_config)

    # loop over all benchmark sonfigs
    try:
        while True:
            # switch to the next benchmark config
            benchmark = next(config.benchmark)

            # define a new data collection        
            report.init_aggregator(benchmark)

            # loop over range of the secondary parameter
            for val_second in benchmark.second.range:
                report.aggregator.new()

                # alter the template config if secondary parameter is given
                if not benchmark.second.empty():
                    config.alter_template(benchmark.second.id, val_second)

                # loop over the prime parameter
                for val_prime in benchmark.prime.range:    
                    # alter the template config if prime parameter is given
                    if not benchmark.prime.empty():
                        config.alter_template(benchmark.prime.id, val_prime)  

                    # update config file 
                    config.get_deployment_config()
                    config.save(output_config)

                    # deploy zstor
                    config.deploy_zstor() 
                    
                    # perform benchmarking                
                    run(["zstorbench", "-C", output_config, "--out-benchmark", result_benchmark_file])
                    
                    # stop zstor
                    #import ipdb; ipdb.set_trace()
                    config.stop_zstor()                    
                    #import ipdb; ipdb.set_trace()
                    # aggregate results
                    report.aggregate(result_benchmark_file)

                    # add timeplots
                    report.add_timeplot()
            
            # add results of the benchmarking to the report
            report.add_aggregation()
    except StopIteration:           # Note 4
        print("Benchmarking is done")
  

if __name__ == '__main__':
    main(sys.argv[1:])    