import json
import random

import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))


# Formation of transaction size sequence matrix and transaction time sequence
# matrix using random gauss based on matrixes of mean values (mean) and
# root-mean-square deviations (stdev) of time periods between transactions and
# size transactions respectively.

def actmatr_gen(file_name_inlet):
    with open(file_name_inlet) as f:
        inlet = json.load(f)

    with open(inlet['periodmatr_mean_file_name']) as f:
        periodmatr_mean = json.load(f)['periodmatr_mean']

    with open(inlet['periodmatr_stdev_file_name']) as f:
        periodmatr_stdev = json.load(f)['periodmatr_stdev']

    with open(inlet['amountmatr_mean_file_name']) as f:
        amountmatr_mean = json.load(f)['amountmatr_mean']

    with open(inlet['amountmatr_stdev_file_name']) as f:
        amountmatr_stdev = json.load(f)['amountmatr_stdev']

    timematr = [[[] for _ in range(len(periodmatr_mean))] for _ in
                range(len(periodmatr_mean))]
    amountmatr = [[[] for _ in range(len(periodmatr_mean))] for _ in
                  range(len(periodmatr_mean))]

    for i in range(len(periodmatr_mean)):
        for j in range(len(periodmatr_mean[i])):
            if periodmatr_mean[i][j] is not None:

                for _ in range(
                        int(inlet['time_period'] / periodmatr_mean[i][j])):
                    timematr[i][j].append(random.gauss(periodmatr_mean[i][j],
                                                       periodmatr_mean[i][j] *
                                                       periodmatr_stdev[i][j]))

                    if timematr[i][j][-1] <= 0:
                        timematr[i][j][-1] = periodmatr_mean[i][j]

                    amountmatr[i][j].append(
                        random.gauss(amountmatr_mean[i][j],
                                     amountmatr_mean[i][j] *
                                     amountmatr_stdev[i][j]))

                    if amountmatr[i][j][-1] <= 0:
                        amountmatr[i][j][-1] = amountmatr_mean[i][j]

                for k in range(1, len(timematr[i][j])):
                    timematr[i][j][k] += timematr[i][j][k - 1]

    # write transaction size sequence matrix into a file

    with open(inlet['timematr_file_name'], 'w') as f:
        json.dump({'timematr': timematr}, f, sort_keys=True, indent=4 * ' ')

    # write transaction time sequence matrix into a file

    with open(inlet['amountmatr_file_name'], 'w') as f:
        json.dump({'amountmatr': amountmatr}, f, sort_keys=True,
                  indent=4 * ' ')


if __name__ == '__main__':
    actmatr_gen(file_name_inlet='inlet/actmatr_inlet.json')
