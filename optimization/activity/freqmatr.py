import json
import random
import copy

import sys

sys.path.append('../')


def freqmatr_gen(file_name_inlet):
    with open(file_name_inlet) as f:
        inlet = json.load(f)

    with open(inlet['flowmatr_file_name']) as f:
        flowmatr = json.load(f)['flowmatr']

    freqmatr_mean = copy.deepcopy(flowmatr)
    freqmatr_stdev = copy.deepcopy(flowmatr)

    for i in range(len(flowmatr)):
        for j in range(len(flowmatr[i])):
            if flowmatr[i][j] != 0:

                freqmatr_mean[i][j] = random.gauss(inlet['mean_mean_freq'],
                                                   inlet['mean_mean_freq'] *
                                                   inlet['stdev_mean_freq'])
                if freqmatr_mean[i][j] < 0:
                    freqmatr_mean[i][j] = 0.

                freqmatr_stdev[i][j] = random.gauss(inlet['mean_stdev_freq'],
                                                    inlet['mean_stdev_freq'] *
                                                    inlet['stdev_stdev_freq'])
                if freqmatr_stdev[i][j] < 0:
                    freqmatr_stdev[i][j] = 0.

    with open(inlet['freqmatr_mean_file_name'], 'w') as f:
        json.dump({'freqmatr_mean': freqmatr_mean}, f, sort_keys=True,
                  indent=4 * ' ')

    with open(inlet['freqmatr_stdev_file_name'], 'w') as f:
        json.dump({'freqmatr_stdev': freqmatr_stdev}, f, sort_keys=True,
                  indent=4 * ' ')


if __name__ == '__main__':
    freqmatr_gen(file_name_inlet='freqmatr_inlet.json')
