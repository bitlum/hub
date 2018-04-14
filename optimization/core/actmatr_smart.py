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

    periodmatr_smart = [
        [SmartSample(periodmatr_calc[i][j][1:],
                     inlet['period_cut_frac']) if len(
            periodmatr_calc[i][j][1:]) > 0 else None for j in
         range(len(periodmatr_calc[i]))] for i in range(len(periodmatr_calc))]

    transmatr_smart = [
        [SmartSample(transmatr_calc[i][j][1:],
                     inlet['trans_cut_frac']) if len(
            transmatr_calc[i][j][1:]) > 0 else None for j in
         range(len(transmatr_calc[i]))] for i in range(len(transmatr_calc))]

    periodmatr_mean = [
        [SmartSample(periodmatr_calc[i][j][1:],
                     inlet['period_cut_frac']).mean if len(
            periodmatr_calc[i][j][1:]) > 0 else None for j in
         range(len(periodmatr_calc[i]))] for i in range(len(periodmatr_calc))]

    transmatr_mean = [
        [SmartSample(transmatr_calc[i][j][1:],
                     inlet['trans_cut_frac']).mean if len(
            transmatr_calc[i][j][1:]) > 0 else None for j in
         range(len(transmatr_calc[i]))] for i in range(len(transmatr_calc))]

    # write the statistical characteristics matrices into files

    with open(inlet['periodmatr_mean_file_name'], 'w') as f:
        json.dump({'periodmatr_mean': periodmatr_mean}, f, sort_keys=True,
                  indent=4 * ' ')

    with open(inlet['transmatr_mean_file_name'], 'w') as f:
        json.dump({'transmatr_mean': transmatr_mean}, f, sort_keys=True,
                  indent=4 * ' ')


if __name__ == '__main__':
    actmatr_smart_gen(file_name_inlet='inlet/actmatr_smart_inlet.json')
