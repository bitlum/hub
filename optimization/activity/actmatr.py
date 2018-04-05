import json
import random

import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))


def actmatr_gen(file_name_inlet):
    with open(file_name_inlet) as f:
        inlet = json.load(f)

    with open(inlet['periodmatr_mean_file_name']) as f:
        periodmatr_mean = json.load(f)['periodmatr_mean']

    with open(inlet['periodmatr_stdev_file_name']) as f:
        periodmatr_stdev = json.load(f)['periodmatr_stdev']

    with open(inlet['sizematr_mean_file_name']) as f:
        sizematr_mean = json.load(f)['sizematr_mean']

    with open(inlet['sizematr_stdev_file_name']) as f:
        sizematr_stdev = json.load(f)['sizematr_stdev']

    timematr = [[[] for _ in range(len(periodmatr_mean))] for _ in
                range(len(periodmatr_mean))]
    transmatr = [[[] for _ in range(len(periodmatr_mean))] for _ in
                 range(len(periodmatr_mean))]

    for i in range(len(periodmatr_mean)):
        for j in range(len(periodmatr_mean[i])):
            if periodmatr_mean[i][j] is not None:

                for _ in range(
                        int(inlet['time_period'] / periodmatr_mean[i][j])):
                    timematr[i][j].append(random.gauss(periodmatr_mean[i][j],
                                                       periodmatr_stdev[i][j]))

                    if timematr[i][j][-1] <= 0:
                        timematr[i][j][-1] = periodmatr_mean[i][j]

                    transmatr[i][j].append(random.gauss(sizematr_mean[i][j],
                                                        sizematr_stdev[i][j]))

                    if transmatr[i][j][-1] <= 0:
                        transmatr[i][j][-1] = sizematr_mean[i][j]

                for k in range(1, len(timematr[i][j])):
                    timematr[i][j][k] += timematr[i][j][k - 1]

    with open(inlet['timematr_file_name'], 'w') as f:
        json.dump({'timematr': timematr}, f, sort_keys=True, indent=4 * ' ')

    with open(inlet['transmatr_file_name'], 'w') as f:
        json.dump({'transmatr': transmatr}, f, sort_keys=True, indent=4 * ' ')


if __name__ == '__main__':
    actmatr_gen(file_name_inlet='inlet/actmatr_inlet.json')
