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
                print("timeLine",timeLine)
                plt.figure()
                plt.plot(timeLine, per_interval[:len(timeLine)],'ro', label=sc_name)
                plt.legend()
                plt.xlabel('time, '+time_unit_literal[4:])
                plt.ylabel('number of operations per '+time_unit_literal[4:])
                name = self.directory+'/'+sc_name + '.png'
                plt.savefig(name)
                plt.close()



                
                      