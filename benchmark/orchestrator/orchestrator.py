import pdb #pdb.set_trace()
from yaml import dump
from subprocess import run
from lib import Config
from lib import Output
from lib import Report
import sys
from getopt import getopt

def main(argv):
    # default path to template yaml file
    input_config = "orchConfig.yaml"

    # default path where config for scenarios is written
    output_config = "scenariosConf.yaml"

    # default path to the benchmark results
    result_benchmark_file = "benchmarkResult.yaml"
    
    report_directory = "report"
    try:
        opts, args = getopt(argv,"hi:o:",["ifile=","ofile="])
        # check if output directories are given
        for opt, arg in opts:
            if opt == '-i':
                # set new file for input
                input_config = arg
            if opt == '-o':
                # set new file for output
                output_config = arg      
    except:
        pass           

    print('********************')
    print('****Benchmarking****')
    print('********************')

    report = Report(report_directory)
    
    # define an object of class Config
    config = Config(input_config)

    # loop over all benchmark sonfigs
    while len(config.benchmarks) > 0:
        # pop next benchmark config
        benchmark = config.pop()

        # predefine list of throughput
        throughput = []

        # loop over range of the secondary parameter
        for val_second in benchmark.second['range']:
            throughput.append([])

            # alter the template config if secondary parameter is given
            if val_second != '':
                config.alter_template(benchmark.second['id'], val_second)

            # loop over the prime parameter
            for val_prime in benchmark.prime['range']:    
                # alter the template config
                config.alter_template(benchmark.prime['id'], val_prime)

                # save new config
                config.save(output_config)

                # run benchmarking program
                run(["../client/client", "-C", output_config, "--out-benchmark", result_benchmark_file])

                # fetch results of benchmarking
                output = Output(result_benchmark_file)

                # append results
                throughput[-1].append(output.throughput()) 


        
        # add results of the benchmarking to the report
        report.add(config.template, benchmark, throughput)
  

if __name__ == '__main__':
    main(sys.argv[1:])    