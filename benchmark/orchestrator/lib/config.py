# Class Config defines a class to set up configuration for benchmarking scenarios
#from yaml import load, dump, YAMLError
import sys
from re import split
from copy import deepcopy
#from ruamel import yaml
import yaml
class Benchmark:

    def __init__(self):
        # parameters represent list of all varying parameters
        #self.PARAMETERS = {'int': {'block_size', 'ValueSize', 'clients'},
        #            'bool': {'encrypt', 'compress'},
        #            'string': {'method'}}
        self.PARAMETERS = {'block_size', 'key_size', 'ValueSize', 'clients', 'encrypt', 'compress', 'method'}

        self.prime = {'id':'', 'range':[]}
        self.second = {'id':'', 'range':[]}

    
    def valid(self):
        if self.prime['id'] not in self.PARAMETERS:
            return False    
        if len(self.prime['range']) == 0:
            return False
        if self.second['id'] not in self.PARAMETERS:
            return False    
        if len(self.second['range']) == 0:
            return False        
        return True

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

        if self.template == None:
            sys.exit('no template config given')

        # extract benchmarking parameters
        benchmarks = config.pop('benchmarks', None)

        self.benchmarks = []
        for bench in benchmarks:
            self.benchmarks.append(bench)        

    # pops next benchmark from self.benchmarks
    def pop(self):
        benchmark = Benchmark()
        benchmark_next = self.benchmarks.pop()

        benchmark.prime = benchmark_next.pop('prime_parameter', None)
        benchmark.prime['range'] = split("\W+", benchmark.prime['range'])

        benchmark.second = benchmark_next.pop('second_parameter', None)
        benchmark.second['range'] = split("\W+", benchmark.second['range'])

      
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

