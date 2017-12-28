# Class Config defines a class to set up configuration for benchmarking scenarios
#from yaml import load, dump, YAMLError
import pdb
import sys
from re import split
from copy import deepcopy
import yaml
from lib.zstor_local_setup import SetupZstor

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
        
        self.deploy = SetupZstor()

    def benchmark_generator(self,benchmarks):       
        if benchmarks:
            for bench in benchmarks:
                yield BenchmarkPair(bench)        
        else:
            yield BenchmarkPair()
        
        
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


    def get_deployment_config(self):
        # fetch zstor server deployment config        
        self.data_shards_nr=self.template['zstor_config']['distribution_data']+ \
                        self.template['zstor_config']['distribution_parity']

        self.meta_shards_nr = self.template['zstor_config']['meta_shards_nr']

        data_shards = self.template['zstor_config']['data_shards']
        meta_shards = self.template['zstor_config']['meta_shards']
        self.port_data = port2int(split(':', data_shards[0])[-1])
        self.port_meta = port2int(split(':', meta_shards[0])[-1])
        self.template['zstor_config']['data_shards'] = self.fix_port_list(data_shards, self.data_shards_nr)
        self.template['zstor_config']['meta_shards'] = self.fix_port_list(meta_shards, self.meta_shards_nr)
        
    def deploy_zstor(self):
        self.deploy.run_zstordb_servers(servers=self.data_shards_nr,
                                    start_port=self.port_data,)
        self.deploy.run_etcd_servers(servers=self.meta_shards_nr,
                                    start_port=self.port_meta)   

    def stop_zstor(self):
        self.deploy.stop_etcd_servers()
        self.deploy.stop_zstordb_servers()
        self.deploy.cleanup()


    @staticmethod
    def fix_port_list(addrs, servers):
        """ 
        Correct list of addresses for the local server deployment.
        
        @addrs gives list of addresses.
            Preserve first given address delete others.
        @servers defines required number of servers.
        
        Return list of localhosts using ports starting 
        from the preserved @addrs and then +1 for each port.
        """
        if not addrs:
            return []
        try:
            [host, start_port] = split(':', addrs[0])
        except ValueError:
            host = "127.0.0.1"
            start_port = addrs[0]

        start_port = port2int(start_port)
        new_addrs=[]
        for port in range(start_port, start_port+servers):
            new_addrs.append("%s:%s"%(host,port))
        return new_addrs


def port2int(port):
    try:
        port = int(port)
    except ValueError:
        sys.exit("error config: wrong port format")        
    return port    
        
                                            