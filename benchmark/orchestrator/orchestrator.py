#----------------------------------------------------------
# orchestrator.py
#
#
# Implements orchestration of benchmarking zstor client and server
#----------------------------------------------------------

import yaml
import time
import matplotlib.pyplot as plt
import sys
import copy
from subprocess import call

# Time units
timeUnits = {'per_second': 1, 'per_minute': 60, 'per_hour': 3600}

# Time Literals
timeLiterals = {'ms':1E-3,'s':1}

# Path to template yaml file
templateConfFile = "templateConf.yaml"

# Path to output yaml file
yamlFile = "../client/benchmark.yaml"

# Output yaml file with scenarios
confClient = "../client/confOrc.yaml"

# Dictionary of prefixes for scenarioID
prefixScenarioIDs = {"range"}

# Dictionary of parameters for benchmarking
parameterDictionary = {"block_size",
                      "replication_nr",
                      "replication_max_size"
                      "distribution_data",
                      "compress",
                      "encrypt",
                      "key_size",
                      "ValueSize" }

# main routine
def main():  
    # create scenarios from template yaml file
    scenarios = createScenarios(templateConfFile)

    # write scenarios to a yaml file 
    with open(confClient, 'w') as outfile:
        yaml.dump(scenarios, outfile, default_flow_style=False)

    # TODO: trigger benchmarking tests 

    # TEST: this is a test call  
    call(["../client/client", "-C confOrc.yaml"])

    # TODO: parse the output file

    # TODO: create output MD document 


#createScenarios creates structure representing scenarios
def createScenarios(templateConfFile):
    # read template yaml file
    with open(templateConfFile, 'r') as stream:
        try:
            templateConfig = yaml.load(stream)
        except yaml.YAMLError as exc:
            print(exc) 
    # first scenario in the template file is treated as a template scenario          
    templateScenarioID = templateConfig['scenarios'].keys()[0]
    templateScenario = templateConfig['scenarios'][templateScenarioID]

    # initialize structure for created scenarios
    Scenarios = {"scenarios":{},}

    # check if scenario starts with preffix "range"
    for pref in prefixScenarioIDs:
        if templateScenarioID.startswith(pref):
            # extract name of the parameter
            parameterID = templateScenarioID.split(pref+'_',1)[1]

            # locate given parameter and set key to configStructure
            configStructure = ""
            for configID in templateScenario:
                print configID
                if parameterID in templateScenario[configID]:
                    configStructure = configID
            
            # check if any key was found
            if configStructure == "":
                sys.exit("parameter is not found")

            # extract options for different scenarios
            options = templateScenario[pref].split(",")
            print options

            # loop over options to create scenarios
            for idx, opt in enumerate(options):
                # create new scenario using template
                scenarioID = parameterID+"_"+opt
                print scenarioID
                # name the scenario
                Scenarios['scenarios'][scenarioID] = copy.deepcopy(templateScenario)
                # set parameter of the scenario
                Scenarios['scenarios'][scenarioID][configStructure][parameterID] = opt
 #               print Scenarios['scenarios'][scenarioID]
 #           print Scenarios['scenarios']
    return Scenarios


#plotScenario plots graphs
def plotScenario():
    with open(yamlFile, 'r') as stream:
        try:
            scenarios = yaml.load(stream)
        except yaml.YAMLError as exc:
            print(exc)

    for scName in scenarios:
        print "Scenario: ", scName        
        print scenarios[scName]['result']['duration']
        duration = duration2num(scenarios[scName]['result']['duration'])

        count = scenarios[scName]['result']['count']
        perInterval = scenarios[scName]['result']['perinterval']

        timeUnitLiteral = scenarios[scName]['scenarioconf']['bench_conf']['result_output']
        timeUnit = timeUnits[timeUnitLiteral]

        # plot throughput vs time only if perInterval is not empty
        if len(perInterval)>0:
            print
            # create time samples for every time unit
            timeLine = [i for i in range(1, int(duration))]

            plt.figure(1, figsize=(6, 6))
            plt.plot(timeLine, perInterval[:len(timeLine)])
            plt.xlabel('time, '+timeUnitLiteral[4:])
            plt.ylabel('number of operations')
            plt.savefig('foo.png')
    

# duration2num converts string containing duration to a float
def duration2num(str):
    for literal in timeLiterals:
        if literal in str:
            try:
                return float(str[:len(literal)])*timeLiterals[literal]
            except:
                print "Duration is not valid"

main()        