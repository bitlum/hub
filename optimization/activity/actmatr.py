import json
import random

import sys

sys.path.append('../')


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

    timematr = [[[0.] for _ in range(len(periodmatr_mean))] for _ in
                range(len(periodmatr_mean))]
    transmatr = [[[0.] for _ in range(len(periodmatr_mean))] for _ in
                 range(len(periodmatr_mean))]

    for i in range(len(periodmatr_mean)):
        for j in range(len(periodmatr_mean[i])):
            if periodmatr_mean[i][j] != 0:
                for _ in range(
                        int(inlet['time_period'] / periodmatr_mean[i][j])):
                    timematr[i][j].append(timematr[i][j][-1] +
                                          random.gauss(
                                              periodmatr_mean[i][j],
                                              periodmatr_mean[i][j] *
                                              periodmatr_stdev[i][j]))

                    if timematr[i][j][-1] <= timematr[i][j][-2]:
                        timematr[i][j][-1] = timematr[i][j][-2] + \
                                             periodmatr_mean[i][j]

                    transmatr[i][j].append(random.gauss(sizematr_mean[i][j],
                                                        sizematr_mean[i][j] *
                                                        sizematr_stdev[i][j]))

                    if transmatr[i][j][-1] <= 0:
                        transmatr[i][j][-1] = sizematr_mean[i][j]

    with open(inlet['timematr_file_name'], 'w') as f:
        json.dump({'timematr': timematr}, f, sort_keys=True, indent=4 * ' ')

    with open(inlet['transmatr_file_name'], 'w') as f:
        json.dump({'transmatr': transmatr}, f, sort_keys=True, indent=4 * ' ')


if __name__ == '__main__':
    actmatr_gen(file_name_inlet='inlet/actmatr_inlet.json')
