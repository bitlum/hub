import json
import random
import copy

import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))


# Formation of matrixes of mean values (mean) and root-mean-square
# deviations (stdev) of transactions size using random gauss.


def sizematr_gen(file_name_inlet):
    with open(file_name_inlet) as f:
        inlet = json.load(f)

    with open(inlet['flowmatr_file_name']) as f:
        flowmatr = json.load(f)['flowmatr']

    with open(inlet['periodmatr_mean_file_name']) as f:
        periodmatr_mean = json.load(f)['periodmatr_mean']

    sizematr_mean = copy.deepcopy(flowmatr)
    sizematr_stdev = copy.deepcopy(flowmatr)

    for i in range(len(flowmatr)):
        for j in range(len(flowmatr[i])):
            if flowmatr[i][j] is not None:

                sizematr_mean[i][j] = flowmatr[i][j] * periodmatr_mean[i][j]

                sizematr_stdev[i][j] = random.gauss(inlet['mean_stdev_size'],
                                                    inlet['mean_stdev_size'] *
                                                    inlet['stdev_stdev_size'])
                if sizematr_stdev[i][j] < 0:
                    sizematr_stdev[i][j] = inlet['mean_stdev_size']

            else:

                sizematr_mean[i][j] = None
                sizematr_stdev[i][j] = None

    # write mean size matrix into a file

    with open(inlet['sizematr_mean_file_name'], 'w') as f:
        json.dump({'sizematr_mean': sizematr_mean}, f, sort_keys=True,
                  indent=4 * ' ')

    # write stdev size matrix into a file

    with open(inlet['sizematr_stdev_file_name'], 'w') as f:
        json.dump({'sizematr_stdev': sizematr_stdev}, f, sort_keys=True,
                  indent=4 * ' ')


if __name__ == '__main__':
    sizematr_gen(file_name_inlet='inlet/sizematr_inlet.json')
