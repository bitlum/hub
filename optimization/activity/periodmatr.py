import json
import random
import copy

import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))


# Formation of matrixes of mean values (mean) and root-mean-square
# deviations (stdev) of time periods between transactions using random gauss.

def periodmatr_gen(file_name_inlet):
    with open(file_name_inlet) as f:
        inlet = json.load(f)

    with open(inlet['flowmatr_file_name']) as f:
        flowmatr = json.load(f)['flowmatr']

    periodmatr_mean = copy.deepcopy(flowmatr)
    periodmatr_stdev = copy.deepcopy(flowmatr)

    for i in range(len(flowmatr)):
        for j in range(len(flowmatr[i])):
            if flowmatr[i][j] is not None:

                periodmatr_mean[i][j] = random.gauss(
                    inlet['mean_mean_period'],
                    inlet['mean_mean_period'] * inlet['stdev_mean_period'])

                if periodmatr_mean[i][j] < 0:
                    periodmatr_mean[i][j] = inlet['mean_mean_period']

                periodmatr_stdev[i][j] = random.gauss(
                    inlet['mean_stdev_period'],
                    inlet['mean_stdev_period'] * inlet['stdev_stdev_period'])

                if periodmatr_stdev[i][j] < 0:
                    periodmatr_stdev[i][j] = inlet['mean_stdev_period']

            else:
                periodmatr_mean[i][j] = None
                periodmatr_stdev[i][j] = None

    # write mean period matrix into a file

    with open(inlet['periodmatr_mean_file_name'], 'w') as f:
        json.dump({'periodmatr_mean': periodmatr_mean}, f, sort_keys=True,
                  indent=4 * ' ')

    # write stdev period matrix into a file

    with open(inlet['periodmatr_stdev_file_name'], 'w') as f:
        json.dump({'periodmatr_stdev': periodmatr_stdev}, f, sort_keys=True,
                  indent=4 * ' ')


if __name__ == '__main__':
    periodmatr_gen(file_name_inlet='inlet/periodmatr_inlet.json')
