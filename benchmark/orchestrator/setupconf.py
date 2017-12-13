from yaml import dump
from subprocess import run
from lib import Config
import sys
from getopt import getopt

def main(argv):
    # default path to template yaml file
    input_config = "templateConf.yaml"

    # default path where config for scenarios is written
    output_config = "scenariosConf.yaml"
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

    # define an object of class Config
    config = Config(input_config)

    # create config for scenarios
    scenarios = config.create_scenarios()

    # write scenarios to a yaml file 
    with open(output_config, 'w+') as outfile:
        dump(scenarios, outfile, default_flow_style=False, default_style='')    

if __name__ == '__main__':
    main(sys.argv[1:])    