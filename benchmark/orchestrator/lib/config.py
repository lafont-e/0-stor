# Class Config defines a class to set up configuration for benchmarking scenarios
#from yaml import load, dump, YAMLError
import pdb
import sys
from re import split
from copy import deepcopy
import yaml

# list of supported parameters
Parameters = {'block_size', 
                'key_size', 
                'ValueSize', 
                'clients', 
                'encrypt', 
                'compress', 
                'method',
                'replication_max_size'}

class Benchmark():
# Benchmark defines and validates benchmark parameter   
    def __init__(self, parameter={}):
        if parameter:
            try:
                self.id = parameter['id']              
            except:
                print("parameter", parameter)
                sys.exit("invalid benchmark: parameter id field is missing")
            if not self.id:
                sys.exit("Invalid benchmark: parameter id is empty")             
            if self.id not in Parameters:
                sys.exit("invalid benchmark: {0} is not supported".format(self.id))
            try:
                self.range = split("\W+", parameter['range'])
            except:
                sys.exit("invalid benchmark: parameter range field is missing")
            if not range:
                sys.exit("invalid benchmark: no range is given for {0}".format(self.id))
        else:
            # return empty Benchmark
            self.range = [' ']
            self.id = ''

    def empty(self):
        if (len(self.range) == 1) and not self.id:
            return True
        return False


class BenchmarkPair():
# BenchmarkPair defines prime and secondary parameter for benchmarking
    def __init__(self, bench_pair={}):
        if bench_pair:
            # extract parameters from a dictionary
            self.prime = Benchmark(bench_pair.pop('prime_parameter', None))
            self.second = Benchmark(bench_pair.pop('second_parameter', None))

            if not self.prime.empty and self.prime.id == self.second.id:
                sys.exit("error: primary and secondary parameters should be different")
            
            if self.prime.empty() and not self.second.empty():
                sys.exit("error: if secondary parameter is given, primary parameter has to be given")
        else:
            # define empty benchmark
            self.prime = Benchmark()
            self.second = Benchmark()            

class Config:
    
    def __init__(self, config_file):
        # read config yaml file
        with open(config_file, 'r') as stream:
            try:
                config = yaml.load(stream)
            except yaml.YAMLError as exc:
                sys.exit(exc)

        # fetch template config for benchmarking
        self.template = config.pop('template', None)

        if not self.template:
            sys.exit('no template config given')

        # extract benchmarking parameters
        self.benchmark = iter(self.benchmark_generator(config.pop('benchmarks', None)))
        #pdb.set_trace()
    
    def benchmark_generator(self,benchmarks):       
        if benchmarks:
            for bench in benchmarks:
                yield BenchmarkPair(bench)        
        else:
            yield BenchmarkPair()
        

    # pops next benchmark from self.benchmarks
    def pop(self):
        benchmark = Benchmark()
        benchmark_next = self.benchmarks.pop()

        benchmark.prime = benchmark_next.pop('prime_parameter', {'id':'','range':[0]})
        try:        
            benchmark.prime['range'] = split("\W+", benchmark.prime['range'])
        except:
            pass
        #pdb.set_trace()
        benchmark.second = benchmark_next.pop('second_parameter', {'id':'','range':[0]})
        try:
            benchmark.second['range'] = split("\W+", benchmark.second['range'])
        except:
            pass

      
        if benchmark.valid() == False:
            sys.exit("benchmark parameteres are incorrect")

        return benchmark
        
    def alter_template(self, id, val):        
        for item in self.template:
            # loop over scenario config
            for key in self.template[item]:
                if  key == id:                   
                    # define type of the parameter
                    parameter_type = type(self.template[item][key])               

                    # update parameter of the scenario                
                    try:
                        self.template[item][key] = parameter_type(val)
                    except:
                        print(self.template[item][key])
                        print("val ", val)
                        print("id =", id)
                        sys.exit("cannot convert {} to {}".format(val,parameter_type))
                    
    def save(self, file_name):
        # prepare config for output
        output = {'scenarios': {'scenario': self.template}}

        # write scenarios to a yaml file 
        with open(file_name, 'w+') as outfile:
            yaml.dump(output, outfile, default_flow_style=False, default_style='')     

