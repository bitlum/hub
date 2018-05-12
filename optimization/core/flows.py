import json

import copy

import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))


# Calculating incoming and outgoing user funds flows by means of
# the statistical characteristics of transaction sizes and time periods
# between transactions.

def flows_calc(file_name_inlet):
    with open(file_name_inlet) as f:
        inlet = json.load(f)

    with open(inlet['amountmatr_mean_calc_file_name']) as f:
        amountmatr_mean_calc = json.load(f)['amountmatr_mean_calc']

    with open(inlet['periodmatr_mean_calc_file_name']) as f:
        periodmatr_mean_calc = json.load(f)['periodmatr_mean_calc']

    flowmatr_calc = copy.deepcopy(amountmatr_mean_calc)

    for i in range(len(flowmatr_calc)):
        for j in range(len(flowmatr_calc[i])):
            if flowmatr_calc[i][j] is not None:
                flowmatr_calc[i][j] /= periodmatr_mean_calc[i][j]

    flowvect_out_calc = [0 for _ in range(len(flowmatr_calc))]
    flowvect_in_calc = [0 for _ in range(len(flowmatr_calc))]

    for i in range(len(flowmatr_calc)):
        for j in range(len(flowmatr_calc[i])):
            if flowmatr_calc[i][j] is not None:
                flowvect_out_calc[i] += flowmatr_calc[i][j]
                flowvect_in_calc[j] += flowmatr_calc[i][j]

    # write flow matrix into a file

    with open(inlet['flowmatr_calc_file_name'], 'w') as f:
        json.dump({'flowmatr_calc': flowmatr_calc}, f, sort_keys=True,
                  indent=4 * ' ')

    # write flow outlet vector into a file

    with open(inlet['flowvect_out_calc_file_name'], 'w') as f:
        json.dump({'flowvect_out_calc': flowvect_out_calc}, f, sort_keys=True,
                  indent=4 * ' ')

    # write flow inlet vector into a file

    with open(inlet['flowvect_in_calc_file_name'], 'w') as f:
        json.dump({'flowvect_in_calc': flowvect_in_calc}, f, sort_keys=True,
                  indent=4 * ' ')


if __name__ == '__main__':
    flows_calc(file_name_inlet='inlet/flows_inlet.json')
