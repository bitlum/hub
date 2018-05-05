import json

import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))


# Formation of the vector of locked balances that users must have for
# the execution of generated transactions.


def balances_gen(file_name_inlet):
    with open(file_name_inlet) as f:
        inlet = json.load(f)

    with open(inlet['transmatr_file_name']) as f:
        transmatr = json.load(f)['transmatr']

    balances = dict()
    for i in range(len(transmatr)):
        sum = 0.
        for j in range(len(transmatr[i])):
            if len(transmatr[i][j]) > 0:
                for k in range(len(transmatr[i][j])):
                    sum += transmatr[i][j][k]
        balances[str(i)] = round(sum)

    # write the vector of locked balances into a file

    with open(inlet['balances_file_name'], 'w') as f:
        json.dump({'balances': balances}, f, sort_keys=True, indent=4 * ' ')


if __name__ == '__main__':
    balances_gen(file_name_inlet='inlet/balances_inlet.json')
