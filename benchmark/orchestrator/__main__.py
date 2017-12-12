from yaml import dump
from subprocess import run
from lib import Config
from lib import Output

def main():
    # path to template yaml file
    templateConfFile = "templateConf.yaml"

    # path to the file containing benchmarking results
    #fileBenchResult = "~/go/src/github.com/zero-os/0-stor/benchmark/client/benchmark.yaml"
    fileBenchResult = "benchmark.yaml"

    # path where config for scenarios is written
    #confScenarios = "~/go/src/github.com/zero-os/0-stor/benchmark/client/confOrc.yaml"
    confScenarios = "confOrc.yaml"

    print('********************')
    print('****Benchmarking****')
    print('********************')

    # define an object of class Config
    config = Config(templateConfFile)

    # create config for scenarios
    scenarios = config.create_scenario()

    # write scenarios to a yaml file 
    with open(confScenarios, 'w+') as outfile:
        dump(scenarios, outfile, default_flow_style=False, default_style='')    
    
    # TODO: trigger benchmarking program
 

    # parse output of the benchmarking
    output = Output(fileBenchResult)

    # plot and save figures
    output.plotScenarios()

if __name__ == '__main__':
    main()    