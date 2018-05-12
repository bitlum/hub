import json

import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))

from samples.smartsample import SmartSample


# Calculating the statistical characteristics of transaction sizes and
# time periods between transactions using the previously created class
# SmartSample and previously computed transaction size sequence matrix and
# transaction time sequence matrix

def actmatr_smart_gen(file_name_inlet):
    with open(file_name_inlet) as f:
        inlet = json.load(f)

    with open(inlet['periodmatr_calc_file_name']) as f:
        periodmatr_calc = json.load(f)['periodmatr_calc']

    with open(inlet['transmatr_calc_file_name']) as f:
        transmatr_calc = json.load(f)['transmatr_calc']

    periodmatr_smart = [[SmartSample(period) for period in periodvect] for
                        periodvect in periodmatr_calc]

    transmatr_smart = [[SmartSample(amount) for amount in amountvect] for
                       amountvect in transmatr_calc]

    for period_vect in periodmatr_smart:
        for period in period_vect:
            period.calc_stat(inlet['period_cut_frac'])

    for amount_vect in transmatr_smart:
        for amount in amount_vect:
            amount.calc_stat(inlet['trans_cut_frac'])

    periodmatr_mean_calc = [[period.mean for period in periodvect] for
                            periodvect in periodmatr_smart]

    transmatr_mean_calc = [[amount.mean for amount in amountvect] for
                           amountvect in transmatr_smart]

    # write the statistical characteristics matrices into files

    with open(inlet['periodmatr_mean_calc_file_name'], 'w') as f:
        json.dump({'periodmatr_mean_calc': periodmatr_mean_calc}, f,
                  sort_keys=True, indent=4 * ' ')

    with open(inlet['transmatr_mean_calc_file_name'], 'w') as f:
        json.dump({'transmatr_mean_calc': transmatr_mean_calc}, f,
                  sort_keys=True, indent=4 * ' ')


if __name__ == '__main__':
    actmatr_smart_gen(file_name_inlet='inlet/actmatr_smart_inlet.json')
