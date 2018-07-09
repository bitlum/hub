import json
import random
import copy

import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))


def get_true_frac(frac):
    return random.randrange(0, 100) < frac * 100


# Deploy flow vector in flow matrix using random factors and
# accounting gates and recipient.

def flowmatr_gen(file_name_inlet):
    with open(file_name_inlet) as f:
        inlet = json.load(f)

    with open(inlet['flowvect_file_name']) as f:
        flowvect = json.load(f)['flowvect']

    with open(inlet['flowvect_file_name']) as f:
        io_users = json.load(f)['io_users']

    with open(inlet['flowvect_file_name']) as f:
        receivers = json.load(f)['receivers']

    users_number = len(flowvect)

    flowmatr = [[None for _ in range(users_number)] for _ in
                range(users_number)]

    for i in range(len(flowmatr)):
        for j in range(len(flowmatr[i])):
            if receivers[j] and i != j:
                flowmatr[i][j] = flowvect[j] * random.uniform(
                    inlet['min_mult'],
                    inlet['max_mult'])
                if i == 0:
                    if io_users[j]:
                        flowmatr[i][j] *= inlet['io_mult']
                    else:
                        flowmatr[i][j] = None
                elif j == 0:
                    if io_users[i]:
                        flowmatr[i][j] *= inlet['io_mult']
                    else:
                        flowmatr[i][j] = None
                else:
                    if get_true_frac(inlet['sparse_frac']):
                        flowmatr[i][j] = None

    # write flow matrix into a file

    with open(inlet['flowmatr_file_name'], 'w') as f:
        json.dump({'flowmatr': flowmatr}, f, sort_keys=True,
                  indent=4 * ' ')


if __name__ == '__main__':
    flowmatr_gen(file_name_inlet='inlet/flowmatr_inlet.json')
