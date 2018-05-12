import json

import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))


# Formation of the vector of locked balances that users must have for
# the execution of generated transactions.


def user_balances_gen(file_name_inlet):
    with open(file_name_inlet) as f:
        inlet = json.load(f)

    user_balances_mult = inlet['user_balances_mult']

    with open(inlet['amountmatr_file_name']) as f:
        amountmatr = json.load(f)['amountmatr']

    balances = dict()
    for i in range(len(amountmatr)):
        sum = 0.
        for j in range(len(amountmatr[i])):
            if len(amountmatr[i][j]) > 0:
                for k in range(len(amountmatr[i][j])):
                    sum += user_balances_mult * amountmatr[i][j][k]
        balances[str(i)] = round(sum)

    # write the vector of locked balances into a file

    with open(inlet['user_balances_file_name'], 'w') as f:
        json.dump({'balances': balances}, f, sort_keys=True, indent=4 * ' ')


if __name__ == '__main__':
    user_balances_gen(file_name_inlet='inlet/user_balances_inlet.json')
