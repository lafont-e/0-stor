# Class Output defines a class to format the output report of the benchmarking program

import yaml
import time
import matplotlib.pyplot as plt

# Time units
timeUnits = {'per_second': 1, 'per_minute': 60, 'per_hour': 3600}

# Time Literals
timeLiterals = {'ms':1E-3,'s':1}  

class Output:

    def __init__(self, benchFile):
        self.benchFile = benchFile
        with open(self.benchFile, 'r') as stream:
            try:
                self.scenarios = yaml.load(stream)['scenarios']
            except yaml.YAMLError as exc:
                sys.exit(exc) 


    def plotScenarios(self): 
        for scName in self.scenarios:
            scenario = self.scenarios[scName]
            
            # duration of the benchmarking
            duration = duration2num(scenario['results']['duration'])

            # count represents the total number op operations
            count = scenario['results']['count']

            # timeUnitLiteral represents the time unit for aggregation of the results
            timeUnitLiteral = scenario['scenarioconf']['bench_conf']['result_output']
            timeUnit = timeUnits[timeUnitLiteral]

            # perInterval represents number of opperations per time unit
            perInterval = scenario['results']['perinterval']

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
        

# duration2num converts string containing duration to a float
def duration2num(str):
    for literal in timeLiterals:
        if literal in str:
            try:
                return float(str[:len(literal)])*timeLiterals[literal]
            except:
                sys.exit("Duration is not valid")