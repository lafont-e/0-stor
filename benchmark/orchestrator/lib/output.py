# Class Output defines a class to format the output report of the benchmarking program

from yaml import load, YAMLError
from sys import exit
import matplotlib.pyplot as plt
import os

# Time units
timeUnits = {'per_second': 1, 'per_minute': 60, 'per_hour': 3600}

class Output:

    def __init__(self, benchFile, directory):
        self.benchFile = benchFile
        self.directory = directory

        # create output directory if needed
        if not os.path.exists(self.directory):
            os.makedirs(self.directory)      

        # read input file
        with open(self.benchFile, 'r') as stream:
            try:
                self.scenarios = load(stream)['scenarios']
            except YAMLError as exc:
                exit(exc) 

    def plot_per_interval(self): 
        # plot_per_interval creates plot of number of operations vs time
        
        for sc_name in self.scenarios:
            # loop over results for all scenarios
            scenario = self.scenarios[sc_name]
            if 'error' in scenario:
                exit(scenario['error'])
            
            # check if results are given
            if len(scenario['results'])==0:
                exit('no results')

            # TODO: decide how to represent result for concurrent clients
            # TEMPORARY: only the first set of results is considered
            results = scenario['results'][0]

            # duration of the benchmarking
            try:
                duration = float(results['duration'])
                print("dur = ", duration)
            except:
                exit('duration format is not valid')        

            # time_unit_literal represents the time unit for aggregation of the results
            time_unit_literal = scenario['scenario']['bench_conf']['result_output']
            timeUnit = timeUnits[time_unit_literal]

            # per_interval represents number of opperations per time unit
            per_interval = results['perinterval']

            # plot number of operations vs time only if per_interval is not empty
            if len(per_interval)>0:
                # create time samples for every time unit
                timeLine = [i for i in range(timeUnit, int(duration+timeUnit))]

                plt.figure()
                plt.plot(timeLine, per_interval[:len(timeLine)],'ro', label=sc_name)
                plt.legend()
                plt.xlabel('time, '+time_unit_literal[4:])
                plt.ylabel('number of operations per '+time_unit_literal[4:])
                name = self.directory+'/'+sc_name + '.png'
                plt.savefig(name)
                plt.close()

    def plot_range(self):    
        # plotRange plots throughput over range of scenarios

        # at least two scenarios have to be given
        if len(self.scenarios)<2:
            print("at least two scenarios have to be given")
            return
        
        # throughput measures zstor performance    
        throughput = []

        # parameter_id has to be extracted from the results
        parameter_id = ""

        # parameter contains list of values that parameter takes
        parameter = []

        sc_prev = ""
        for idx, sc_name in enumerate(self.scenarios):
            # loop over results for all scenarios
            scenario = self.scenarios[sc_name]
            print("name", sc_name)
            print(idx)
            if 'error' in scenario:
                exit(scenario['error'])        

            # check if results are given
            if 'results' in scenario ==0:
                exit('no results')


            # compare current scenario with the previous scenario
            if idx > 0:
                # count differences between scenarios
                count_dif = 0
                
                # shortcats for the previous and current scenario configs
                sc_conf = scenario['scenario']
                pr_conf = self.scenarios[sc_prev]['scenario']

                for item in sc_conf:
                    # loop over scenario config
                    for key in sc_conf[item]:
                        if sc_conf[item][key] != pr_conf[item][key]:
                            count_dif+=1
                            diff_key = key
                            pr_value = pr_conf[item][diff_key]
                            value = sc_conf[item][diff_key]

                if count_dif == 0:
                    exit('cannot define range, no difference between scenarios')    

                if count_dif > 1:
                    exit('cannot define range, the scenarios are too different')  

                # if parameter_id is empty, assign the parameter
                if parameter_id == "":
                    parameter_id = diff_key
                    parameter.append(pr_value)
                else:
                    # changing parameter should be the same for all scenarios
                    if parameter_id != diff_key:
                        exit('cannot define range, ranging parameter should be consistent')  
                parameter.append(value)

            # save scenario name for the next loop                
            sc_prev = sc_name

            # TODO: decide how to represent result for concurrent clients
            # TEMPORARY: only the first set of results is considered
            results = scenario['results'][0]

            # duration of the benchmarking
            try:
                duration = float(results['duration'])
                print("dur = ", duration)
            except:
                exit('duration format is not valid')        

            # count represents the total number op operations
            count = results['count']

            throughput.append( count/duration )

            # time_unit_literal represents the time unit for aggregation of the results
            time_unit_literal = scenario['scenario']['bench_conf']['result_output']
            timeUnit = timeUnits[time_unit_literal]


            # per_interval represents number of opperations per time unit
            per_interval = results['perinterval']

        print(parameter)        
        print(throughput)            
        plt.figure()
        plt.plot(parameter, throughput,'o', label=parameter_id)
        plt.legend()
        plt.xlabel(parameter_id)
        plt.ylabel('throughput, operations per second')
        name = self.directory+'/range_'+parameter_id + '.png'
        plt.savefig(name)
        plt.close()
   

                
                      