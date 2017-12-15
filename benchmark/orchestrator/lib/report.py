import pdb #pdb.set_trace()
import os
import matplotlib.pyplot as plt


class Report:

    def __init__(self, directory):
        # output directories for report file
        self.directory = directory
        self.report_name =  self.directory+"/report.md"

        # create output directory if needed
        if not os.path.exists(self.directory):
            os.makedirs(self.directory)  

        # create a report file
        with open(self.report_name, 'w+') as outfile:
            outfile.write("# Benchmark report\n")        
        
        # keep track of added reports
        self.reports_added = 0


    def add(self, config, benchmark, throughput):
        self.reports_added += 1
        self.benchmark = benchmark
        self.throughput = throughput

        # name of the figure
        fig_name = 'fig' +str(self.reports_added) + '.png'

        # refer the figure in the report
        with open(self.report_name, 'a+') as outfile:
            outfile.write("\n ## Report " + str(self.reports_added)+"\n")
            outfile.write("\n![Fig: throughput vs parameter]("+fig_name+")")
        
        # create a bar plot
        self.__bar_plot__( fig_name)

        # add the table of the data set
        self.__add_table__()



    def __bar_plot__(self, fig_name):
        # define range  from prime parameter
        ticks_labels = self.benchmark.prime['range']

        # af first results are plot vs counting number of samples
        rng = [i for i, tmp in enumerate(ticks_labels)]

        # number of data sets combined in the figure
        n_plots = len(self.throughput[0])

        # number of samples for each data set
        n_samples = len(rng)

        ### figure view ###
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

        ax.set_xlabel(self.benchmark.prime['id'])
        ax.set_ylabel('throughput, byte/s')

        # loop over data sets
        for i, th in enumerate(self.throughput):
            # define plot latel
            lb = self.benchmark.second['id']+"="+ self.benchmark.second['range'][i]

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
        with open(self.report_name, 'a+') as outfile:
            # create a table
            outfile.write("""\n <details> <summary> data </summary> \n
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
            </table>
            <table> 
                <tr> """ + self.benchmark.prime['id'] + "<\th> \n")

            for row, tmp in enumerate(self.benchmark.prime['range']):
                for col, tmp in enumerate(self.benchmark.second['range']):
                    outfile.write(str(self.throughput[col][row])+"\t")
                outfile.write("\n")                    
                             


            outfile.write("\n </table></details> \n")




