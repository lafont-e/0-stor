# Class Output defines a class to format the output report of the benchmarking program

from yaml import load, YAMLError
from sys import exit
import matplotlib.pyplot as plt

# Time units
timeUnits = {'per_second': 1, 'per_minute': 60, 'per_hour': 3600}

class Output:

    def __init__(self, benchFile):
        self.benchFile = benchFile
        with open(self.benchFile, 'r') as stream:
            try:
                self.scenarios = load(stream)['scenarios']
            except YAMLError as exc:
                exit(exc) 

    # plotScenarios creates plot of number of operations vs time
    def plotScenarios(self): 
        for scName in self.scenarios:
            scenario = self.scenarios[scName]
            
            # check if results are given
            if len(scenario['results'])==0:
                exit('no results')

            # TEMPORARY: only first set of results is considered
            results = scenario['results'][0]

            # duration of the benchmarking
            try:
                duration = float(results['duration'])
            except:
                exit('duration format is not valid')        

            # count represents the total number op operations
            count = results['count']

            # timeUnitLiteral represents the time unit for aggregation of the results
            timeUnitLiteral = scenario['scenarioconf']['bench_conf']['result_output']
            timeUnit = timeUnits[timeUnitLiteral]

            # perInterval represents number of opperations per time unit
            perInterval = results['perinterval']

            # plot throughput vs time only if perInterval is not empty
            if len(perInterval)>0:
                # create time samples for every time unit
                timeLine = [i for i in range(1, int(duration))]

                plt.figure(1, figsize=(6, 6))
                plt.plot(timeLine, perInterval[:len(timeLine)])
                plt.xlabel('time, '+timeUnitLiteral[4:])
                plt.ylabel('number of operations')
                name = scName + '.png'
                plt.savefig(name)
                plt.close(1)
        
