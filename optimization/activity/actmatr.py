import json
import random

import sys

sys.path.append('../')


def actmatr_gen(file_name_inlet):
    with open(file_name_inlet) as f:
        inlet = json.load(f)

    with open(inlet['freqmatr_mean_file_name']) as f:
        freqmatr_mean = json.load(f)['freqmatr_mean']

    with open(inlet['freqmatr_stdev_file_name']) as f:
        freqmatr_stdev = json.load(f)['freqmatr_stdev']

    with open(inlet['sizematr_mean_file_name']) as f:
        sizematr_mean = json.load(f)['sizematr_mean']

    with open(inlet['sizematr_stdev_file_name']) as f:
        sizematr_stdev = json.load(f)['sizematr_stdev']

    timematr = [[[0.] for _ in range(len(freqmatr_mean))] for _ in
                range(len(freqmatr_mean))]
    transmatr = [[[0.] for _ in range(len(freqmatr_mean))] for _ in
                 range(len(freqmatr_mean))]

    for i in range(len(freqmatr_mean)):
        for j in range(len(freqmatr_mean[i])):
            if freqmatr_mean[i][j] != 0:
                for _ in range(
                        int(freqmatr_mean[i][j] * inlet['time_period'])):
                    timematr[i][j].append(timematr[i][j][-1] +
                                          random.gauss(
                                              1. / freqmatr_mean[i][j],
                                              1. / freqmatr_mean[i][j] *
                                              freqmatr_stdev[i][j]))

                    if timematr[i][j][-1] <= timematr[i][j][-2]:
                        timematr[i][j][-1] = timematr[i][j][-1] + \
                                             freqmatr_mean[i][j]

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
    actmatr_gen(file_name_inlet='actmatr_inlet.json')
