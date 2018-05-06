import json
import random
import copy

import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))


# Deploy flow vector in flow matrix using random factors and
# accounting gates and recipient.

def flowmatr_gen(file_name_inlet):
    with open(file_name_inlet) as f:
        inlet = json.load(f)

    with open(inlet['flowvect_file_name']) as f:
        flowvect = json.load(f)['flowvect']

    with open(inlet['flowvect_file_name']) as f:
        receivers = json.load(f)['receivers']

    flowmatr = [copy.deepcopy(flowvect) for _ in range(len(flowvect))]

    for i in range(len(flowmatr)):
        for j in range(len(flowmatr[i])):
            if receivers[j] and i != j:
                flowmatr[i][j] *= random.uniform(inlet['min_mult'],
                                                 inlet['max_mult'])
            else:
                flowmatr[i][j] = None

    # write flow matrix into a file

    with open(inlet['flowmatr_file_name'], 'w') as f:
        json.dump({'flowmatr': flowmatr}, f, sort_keys=True,
                  indent=4 * ' ')


if __name__ == '__main__':
    flowmatr_gen(file_name_inlet='inlet/flowmatr_inlet.json')
