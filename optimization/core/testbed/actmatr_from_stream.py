import json
import copy

import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../../'))


# Calculation of user id list and transaction size sequence matrix and
# transaction period (and time) sequence matrix by means of a general
# time-ordered transaction stream.  Transaction stream element structure:
# time, sender_id, receiver_id, trans.

def actmatr_calc(file_name_inlet):
    with open(file_name_inlet) as f:
        inlet = json.load(f)

    with open(inlet['transstream_file_name']) as f:
        transstream = json.load(f)['transstream']

    users_id_rep_calc = []
    for i in range(len(transstream)):
        users_id_rep_calc.append(transstream[i]["payment"]['sender'])
        users_id_rep_calc.append(transstream[i]["payment"]['receiver'])
    users_id_list_calc = sorted(list(set(users_id_rep_calc)))

    user_id_calc = {i: users_id_list_calc[i] for i in
                    range(len(users_id_list_calc))}

    users_ind_calc = {users_id_list_calc[i]: i for i in
                      range(len(users_id_list_calc))}

    timematr_calc = [[[] for _ in range(len(users_id_list_calc))] for _ in
                     range(len(users_id_list_calc))]

    amountmatr_calc = [[[] for _ in range(len(users_id_list_calc))] for _ in
                      range(len(users_id_list_calc))]

    for i in range(len(transstream)):
        timematr_calc[users_ind_calc[transstream[i]["payment"]['sender']]][
            users_ind_calc[transstream[i]["payment"]['receiver']]].append(
            1.E-9 * (transstream[i]['time'] - transstream[0]['time']))

        amountmatr_calc[users_ind_calc[transstream[i]["payment"]['sender']]][
            users_ind_calc[transstream[i]["payment"]['receiver']]].append(
            transstream[i]["payment"]['amount'])

    periodmatr_calc = copy.deepcopy(timematr_calc)

    for i in range(len(periodmatr_calc)):
        for j in range(len(periodmatr_calc[i])):
            if len(periodmatr_calc[i][j]) > 0:
                for k in range(len(periodmatr_calc[i][j]) - 1, 0, -1):
                    periodmatr_calc[i][j][k] -= periodmatr_calc[i][j][k - 1]

    # write user id dict into a file

    with open(inlet['users_id_calc_file_name'], 'w') as f:
        json.dump({'users_id_calc': user_id_calc}, f, sort_keys=True,
                  indent=4 * ' ')

    # write transaction time sequence matrix into a file

    with open(inlet['timematr_calc_file_name'], 'w') as f:
        json.dump({'timematr_calc': timematr_calc}, f, sort_keys=True,
                  indent=4 * ' ')

    # write transaction period sequence matrix into a file

    with open(inlet['periodmatr_calc_file_name'], 'w') as f:
        json.dump({'periodmatr_calc': periodmatr_calc}, f, sort_keys=True,
                  indent=4 * ' ')

    # write transaction size sequence matrix into a file

    with open(inlet['amountmatr_calc_file_name'], 'w') as f:
        json.dump({'amountmatr_calc': amountmatr_calc}, f, sort_keys=True,
                  indent=4 * ' ')


if __name__ == '__main__':
    actmatr_calc(file_name_inlet='inlet/actmatr_from_stream_inlet.json')
