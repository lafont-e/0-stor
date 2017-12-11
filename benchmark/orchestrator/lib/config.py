# Class Config defines a class to set up configuration for benchmarking scenarios
import yaml
import re
import copy

class Config:

    def __init__(self, config):
        self.config = config
        self.templateConfig = {}

        # read template yaml file
        with open(self.config, 'r') as stream:
            try:
                self.templateConfig = yaml.load(stream)
            except yaml.YAMLError as exc:
                sys.exit(exc)

        # first scenario in the config file is treated as a template scenario          
        self.templateScenarioID = self.templateConfig["scenarios"].keys()[0]
        self.templateScenario = self.templateConfig["scenarios"][self.templateScenarioID]     

        # extract benchmarking parameter
        self.benchParameter = self.templateScenario.pop("parameter", None)

        # check if parameter is specified
        if self.is_multiscenario():
            self.mode = "multi"
        else:
            self.mode = "mono"


    def is_multiscenario(self):
        if self.benchParameter == None:
            return False

        # check if range for the parameter is given
        if 'par_range' not in self.benchParameter:
            return False

        # extract range
        self.options = re.split("\W+", self.benchParameter["par_range"])

        # check if any options are given
        if len(self.options) == 0:
            return False

        # check if parameter id is given
        if 'par_id' not in self.benchParameter:
            return False        

        # extract name of the parameter
        self.parameterID = self.benchParameter["par_id"]

        # define which configID among scenario configs the parameter belongs to
        self.configID = ""
        for conf in self.templateScenario:
            if self.parameterID in self.templateScenario[conf]:
                self.configID = conf

        # check if any key was found
        if self.configID == "":
            return False       
    
        return True
    
    def create_scenario(self):
        # initialize structure for created scenarios
        self.scenarios = {"scenarios":{},}
        
        if self.mode == 'multi':
            # loop over options
            for idx, opt in enumerate(self.options):
                # create new scenario using template
                self.scenarioID = self.parameterID+"_"+opt

                # name the scenario
                self.scenarios['scenarios'][self.scenarioID] = copy.deepcopy(self.templateScenario)

                # define type of the parameter
                parameterType = type(self.scenarios['scenarios'][self.scenarioID][self.configID][self.parameterID])               

                # set parameter of the scenario                
                try:
                    self.scenarios['scenarios'][self.scenarioID][self.configID][self.parameterID] = parameterType(opt)
                except:
                    sys.exit("cannot convert {} to {}".format(opt,parameterType))

                print(self.scenarios['scenarios'][self.scenarioID][self.configID][self.parameterID])     
        if self.mode == 'mono':
            self.scenarios['scenarios'][self.scenarioID] = copy.deepcopy(self.templateScenario)
        
        return self.scenarios
            