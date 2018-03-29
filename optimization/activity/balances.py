import json

import sys

sys.path.append('../')


def balances_gen(file_name_inlet):
    with open(file_name_inlet) as f:
        inlet = json.load(f)

    with open(inlet['transmatr_file_name']) as f:
        transmatr = json.load(f)['transmatr']

    balances = []
    for i in range(len(transmatr)):
        balances.append(dict(user_ind=i, balance=0.))
        for j in range(len(transmatr[i])):
            for k in range(1, len(transmatr[i][j])):
                balances[-1]['balance'] += transmatr[i][j][k]

    with open(inlet['balances_file_name'], 'w') as f:
        json.dump({'balances': balances}, f, sort_keys=True, indent=4 * ' ')


if __name__ == '__main__':
    balances_gen(file_name_inlet='inlet/balances_inlet.json')
