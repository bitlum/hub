import json
import random
import copy

import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))


# Formation of matrixes of mean values (mean) and root-mean-square
# deviations (stdev) of transactions amount using random gauss.


def amountmatr_gen(file_name_inlet):
    with open(file_name_inlet) as f:
        inlet = json.load(f)

    with open(inlet['flowmatr_file_name']) as f:
        flowmatr = json.load(f)['flowmatr']

    with open(inlet['periodmatr_mean_file_name']) as f:
        periodmatr_mean = json.load(f)['periodmatr_mean']

    amountmatr_mean = copy.deepcopy(flowmatr)
    amountmatr_stdev = copy.deepcopy(flowmatr)

    for i in range(len(flowmatr)):
        for j in range(len(flowmatr[i])):
            if flowmatr[i][j] is not None:

                amountmatr_mean[i][j] = flowmatr[i][j] * periodmatr_mean[i][j]

                amountmatr_stdev[i][j] = random.gauss(inlet['mean_stdev_amount'],
                                                    inlet['mean_stdev_amount'] *
                                                    inlet['stdev_stdev_amount'])
                if amountmatr_stdev[i][j] < 0:
                    amountmatr_stdev[i][j] = inlet['mean_stdev_amount']

            else:

                amountmatr_mean[i][j] = None
                amountmatr_stdev[i][j] = None

    # write mean amount matrix into a file

    with open(inlet['amountmatr_mean_file_name'], 'w') as f:
        json.dump({'amountmatr_mean': amountmatr_mean}, f, sort_keys=True,
                  indent=4 * ' ')

    # write stdev amount matrix into a file

    with open(inlet['amountmatr_stdev_file_name'], 'w') as f:
        json.dump({'amountmatr_stdev': amountmatr_stdev}, f, sort_keys=True,
                  indent=4 * ' ')


if __name__ == '__main__':
    amountmatr_gen(file_name_inlet='inlet/amountmatr_inlet.json')
