"""
    Package report contains tools for collecting results of the benchmarking 
    and generating final report. 

    Output files show benchmarking results by means of tables and figures 
    allowing for visual representation. Report for each benchmark is added 
    to the output files as soon as the benchmark is finished.

    Two types of output files are avaliable: main report file and a timeplot
    collection. Main report file consists of performance measures and allows 
    for performance comparision among various scenarios.

    The performance metric, reflected in the report is throughput, 
    defined as an average data rate observed throughout the benchmark.

    Timeplot collection report soncists the scope of timeplots
    collected during the benchmarking. Timeplots show number of 
    operations per time unit, observes during the benchmark.
"""
import os
import matplotlib.pyplot as plt
import yaml
import sys

TimeUnits = {'per_second': 1, 
             'per_minute': 60, 
               'per_hour': 3600}

FilterKeys={'organization', 
               'namespace', 
              'iyo_app_id', 
              'iyo_app_id', 
          'iyo_app_secret',
             'data_shards', 
             'meta_shards', 
             'meta_shards',
             'encrypt_key'}

class Aggregator:
    """
    Aggregator aggregates average throughput over a set of benchmarks
    """
    def __init__(self, benchmark):
        self.benchmark = benchmark
        self.throughput= []

    def new(self):
        self.throughput.append([])

class Report:
    """
    Class Report is used to collect results of benchmarking
        and to create final report.
    """
    def __init__(self, directory='report', report='report.md', timeplots='timeplots.md'):
        self.directory = directory # set output directory for report files
        self.main_file =  "{0}/{1}".format(self.directory,report)
        self.timeplots_collection =  "{0}/{1}".format(self.directory, timeplots)

        if not os.path.exists(self.directory):
            os.makedirs(self.directory)  

        with open(self.main_file, 'w+') as outfile:
            outfile.write("# Benchmark report\n")     
            outfile.write("[Timeplot collection is here]({0})\n".format(timeplots))    

        with open(self.timeplots_collection, 'w+') as outfile:
            outfile.write("# Timeplot collection report\n")  
            outfile.write("[Main report in here]({0}) \n\n".format(report)) 

        self.reports_added = 0   # keep track of added reports
        self.timeplots_added = 0 # keep track of number of timeplots added 
        self.scenarios = {}

    def init_aggregator(self, benchmark):
        self.aggregator = Aggregator(benchmark)

    def get_scenario_config(self, input_file):
        """
        Fetch benchmark scenario config
        """
        with open(input_file, 'r') as stream:
            try:
                self.scenarios = yaml.load(stream)['scenarios']
            except yaml.YAMLError as exc:
                sys.exit(exc)
        err = self.scenarios['scenario'].pop('error', None)
        if err:
            sys.exit("last benchmark exited with error: %s"%err )
        self.filter_scenario_config()
        
    def filter_scenario_config(self):
        """
        Delete irrelevant keys of the scenario config 
        """
        for sc_name in self.scenarios:
            for key in FilterKeys:
                try:
                    self.scenarios[sc_name]['scenario']['zstor_config'].pop(key, None)
                except:
                    pass 

    def aggregate(self, input_file):
        self.get_scenario_config(input_file)
        th = self.__get_throughput__()
        self.aggregator.throughput[-1].append(th[0])
    
    def __get_throughput__(self):
        throughput = []
        for sc_name in self.scenarios:
            scenario = self.scenarios[sc_name]
            if 'error' in scenario:
                exit(scenario['error'])
            
            if 'results' not in scenario:
                sys.exit("no results are provided")

            throughput.append(0)
            for result in scenario['results']:    
                # duration of the benchmarking
                try:
                    duration = float(result['duration'])
                except:
                    exit('duration is not given, or format is not float')   

                # number of operations in the benchmarking
                try:
                    count = int(result['count'])
                except:
                    exit('count is not given, or format is not int')   

                # size of each value
                try:
                    value_size = int(scenario['scenario']['bench_conf']['ValueSize'])
                except:
                    exit('value size is not given, or format is not int')  
                     
                throughput[-1] += count*value_size/duration/len(scenario['results'])
            
            throughput[-1] = int(throughput[-1])
           
        return throughput

    def add_aggregation(self):
        self.reports_added += 1

        # name of the figure
        fig_name = 'fig' +str(self.reports_added) + '.png'

        # refer the figure in the report
        with open(self.main_file, 'a+') as outfile:
            outfile.write("\n # Report {0} \n".format(str(self.reports_added)))
            outfile.write('**Benchmark config:** \n')
            outfile.write('```yaml \n')
            yaml.dump(self.scenarios, outfile,default_flow_style=False)
            outfile.write('\n```')

        # check if more then one output was collected   
        if sum(map(len, self.aggregator.throughput)) > 1:
            # create a bar plot
            self.__bar_plot__( fig_name)

            # incerst bar plot to the report
            with open(self.main_file, 'a+') as outfile:
                outfile.write("\n![Fig: throughput vs parameter]({0})".format(fig_name))            

        # add the table of the data set
        self.__add_table__()


    def __bar_plot__(self, fig_name):
        # define range  from prime parameter
        ticks_labels = self.aggregator.benchmark.prime.range

        # af first results are plot vs counting number of samples
        rng = [i for i, tmp in enumerate(ticks_labels)]

        # number of data sets combined in the figure
        if len(self.aggregator.throughput) == 0:
            sys.exit("no results are included")

        n_plots = len(self.aggregator.throughput[0])

        """ figure settings """

        # number of samples for each data set
        n_samples = len(rng)

        # bar width
        width = rng[-1]/(n_samples*n_plots+1)

        # gap between bars
        gap = width/10

        # create figure
        fig, ax = plt.subplots()

        # limmit number of ticks to the number of samples
        plt.xticks(rng)

        # substitute tick labels
        ax.set_xticklabels(ticks_labels)
        
        # define color cycle 
        ax.set_color_cycle(['blue', 'red', 'green', 'yellow', 'black', 'brown'])

        ax.set_xlabel(self.aggregator.benchmark.prime.id)
        ax.set_ylabel('throughput, byte/s')

        # loop over data sets
        for i, th in enumerate(self.aggregator.throughput):
            # define plot label
            lb = " "
            if self.aggregator.benchmark.second.id:
                lb = "{0}={1}".format(self.aggregator.benchmark.second.id,
                                    self.aggregator.benchmark.second.range[i])

            # add bar plot to the figure
            ax.bar(rng, th, width, label=lb)
            lgd = ax.legend(loc='upper left', bbox_to_anchor=(1,1))
            #plt.tight_layout(pad=20)

            # add labels to bars
            for i, v in enumerate(th):
                ax.text(rng[i]-width/2, v , str(v), color='blue', fontweight='bold')

            # shift bars for the next plot
            rng = [x+gap+width for x in rng]

        # label axes
        plt.savefig(self.directory+"/"+fig_name, bbox_extra_artists=(lgd,), bbox_inches='tight')       
        plt.close()

    def __add_table__(self):         
        # add hidden table with data
        with open(self.main_file, 'a+') as outfile:
            # create a table
            outfile.write("""\n <h3> Throughput, byte/s: </h3>
            <head> 
                <style>
                    table, th, td {
                        border: 1px solid black;
                        border-collapse: collapse;
                    }
                    th, td {
                        text-align: left;    
                    }
                </style>
            </head>
            <table>  
                <tr> <th> """ + self.aggregator.benchmark.prime.id + "</th>")
            # add titles to the columns    
            for item in self.aggregator.benchmark.second.range:                
                if self.aggregator.benchmark.second.id:
                    outfile.write("<th> {0} = {1} </th>".format(self.aggregator.benchmark.second.id,item))
                else:
                    outfile.write("<th>  </th>")

                
            outfile.write(" </tr> ")

            # fill in the table
            for row, val in enumerate(self.aggregator.benchmark.prime.range):
                outfile.write("<tr> <th> {0} </th>".format(val))
                for col, tmp in enumerate(self.aggregator.benchmark.second.range):
                    outfile.write("<th> {0} </th>".format(str(self.aggregator.throughput[col][row])))
                outfile.write("</tr>")                    
                             


            outfile.write("\n </table></details>\n")


    def add_timeplot(self):
        """
        Add timeplots to the report
        """
        files = self.__plot_per_interval__()

        if len(files)>0:
            with open(self.timeplots_collection, 'a+') as outfile:
                outfile.write('\n**Config:**\n```yaml \n')
                yaml.dump(self.scenarios, outfile, default_flow_style=False)
                outfile.write('\n```')
                outfile.write("\n _____________ \n".format(str(self.reports_added)))
                for file in files:
                    outfile.write("\n![Fig](../{0}) \n".format(file))

    def __plot_per_interval__(self): 
        """
        Create timeplots
        """
        # file_names returns list of the output files
        file_names = []

        # plot_per_interval creates plot of number of operations vs time
        for sc_name in self.scenarios:
            # loop over results for all scenarios
            scenario = self.scenarios[sc_name]
            if 'error' in scenario:
                exit(scenario['error'])

            # check if results are given
            if len(scenario['results'])==0:
                exit('no results')

            # time_unit_literal represents the time unit for aggregation of the results
            time_unit_literal = scenario['scenario']['bench_conf']['result_output']
            timeUnit = TimeUnits[time_unit_literal]
           
            for idx, result in enumerate(scenario['results']):
                # duration of the benchmarking
                try:
                    duration = float(result['duration'])
                except:
                    exit('duration format is not valid')        

                # per_interval represents number of opperations per time unit
                try:
                    per_interval = result['perinterval']
                except:
                    per_interval = []
                # plot number of operations vs time only if per_interval is not empty
                if len(per_interval)>0:
                    # create time samples for every time unit
                    time_line = [i for i in range(timeUnit, int(duration+timeUnit))]

                    plt.figure()
                    plt.plot(time_line, per_interval[len(per_interval)-len(time_line):],'bo--', label=self.timeplots_added)
                    plt.xlabel('time, '+time_unit_literal[4:])
                    plt.ylabel('number of operations per '+time_unit_literal[4:])
                    file = '{0}/plot_per_interval_{1}_{2}.png'.format(self.directory, sc_name, str(self.timeplots_added))
                    plt.savefig(file)
                    plt.close()
                    file_names.append(file)
                    self.timeplots_added+=1
        return file_names


