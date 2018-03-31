import json
import copy
import sys

sys.path.append('../')


def actmatr_calc(file_name_inlet):
    with open(file_name_inlet) as f:
        inlet = json.load(f)

    with open(inlet['transstream_file_name']) as f:
        transstream = json.load(f)['transstream']

    users_id_rep_calc = []
    for i in range(len(transstream)):
        users_id_rep_calc.append(transstream[i]['sender_id'])
        users_id_rep_calc.append(transstream[i]['receiver_id'])
    users_id_list_calc = sorted(list(set(users_id_rep_calc)))

    user_id_calc = {i: users_id_list_calc[i] for i in
                    range(len(users_id_list_calc))}

    users_ind_calc = {users_id_list_calc[i]: i for i in
                      range(len(users_id_list_calc))}

    timematr_calc = [[[] for _ in range(len(users_id_list_calc))] for _ in
                     range(len(users_id_list_calc))]

    transmatr_calc = [[[] for _ in range(len(users_id_list_calc))] for _ in
                      range(len(users_id_list_calc))]

    for i in range(len(transstream)):
        timematr_calc[users_ind_calc[transstream[i]['sender_id']]][
            users_ind_calc[transstream[i]['receiver_id']]].append(
            transstream[i]['time'])

        transmatr_calc[users_ind_calc[transstream[i]['sender_id']]][
            users_ind_calc[transstream[i]['receiver_id']]].append(
            transstream[i]['trans'])

    periodmatr_calc = copy.deepcopy(timematr_calc)

    for i in range(len(periodmatr_calc)):
        for j in range(len(periodmatr_calc[i])):
            if len(periodmatr_calc[i][j]) > 0:
                for k in range(len(periodmatr_calc[i][j]) - 1, 0, -1):
                    periodmatr_calc[i][j][k] -= periodmatr_calc[i][j][k - 1]

    with open(inlet['users_id_calc_file_name'], 'w') as f:
        json.dump({'users_id_calc': user_id_calc}, f, sort_keys=True,
                  indent=4 * ' ')

    with open(inlet['timematr_calc_file_name'], 'w') as f:
        json.dump({'timematr_calc': timematr_calc}, f, sort_keys=True,
                  indent=4 * ' ')

    with open(inlet['periodmatr_calc_file_name'], 'w') as f:
        json.dump({'periodmatr_calc': periodmatr_calc}, f, sort_keys=True,
                  indent=4 * ' ')

    with open(inlet['transmatr_calc_file_name'], 'w') as f:
        json.dump({'transmatr_calc': transmatr_calc}, f, sort_keys=True,
                  indent=4 * ' ')


if __name__ == '__main__':
    actmatr_calc(file_name_inlet='inlet/actmatr_from_stream_inlet.json')
