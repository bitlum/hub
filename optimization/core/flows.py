import json

import copy

import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))


def flows_calc(file_name_inlet):
    with open(file_name_inlet) as f:
        inlet = json.load(f)

    with open(inlet['transmatr_mean_file_name']) as f:
        transmatr_mean = json.load(f)['transmatr_mean']

    with open(inlet['periodmatr_mean_file_name']) as f:
        periodmatr_mean = json.load(f)['periodmatr_mean']

    flowmatr_calc = copy.deepcopy(transmatr_mean)

    for i in range(len(flowmatr_calc)):
        for j in range(len(flowmatr_calc[i])):
            if flowmatr_calc[i][j] is not None:
                flowmatr_calc[i][j] /= periodmatr_mean[i][j]

    flowvect_calc_out = [0 for _ in range(len(flowmatr_calc))]
    flowvect_calc_in = [0 for _ in range(len(flowmatr_calc))]

    for i in range(len(flowmatr_calc)):
        for j in range(len(flowmatr_calc[i])):
            if flowmatr_calc[i][j] is not None:
                flowvect_calc_out[i] += flowmatr_calc[i][j]
                flowvect_calc_in[j] += flowmatr_calc[i][j]

    with open(inlet['flowmatr_calc_file_name'], 'w') as f:
        json.dump({'flowmatr_calc': flowmatr_calc}, f, sort_keys=True,
                  indent=4 * ' ')

    with open(inlet['flowvector_out_calc_file_name'], 'w') as f:
        json.dump({'flowvect_calc_out': flowvect_calc_out}, f, sort_keys=True,
                  indent=4 * ' ')

    with open(inlet['flowvector_in_calc_file_name'], 'w') as f:
        json.dump({'flowvect_calc_in': flowvect_calc_in}, f, sort_keys=True,
                  indent=4 * ' ')


if __name__ == '__main__':
    flows_calc(file_name_inlet='inlet/flows_calc_inlet.json')
