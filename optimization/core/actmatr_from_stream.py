import json

import sys

sys.path.append('../')


def actmatr_calc(file_name_inlet):
    with open(file_name_inlet) as f:
        inlet = json.load(f)

    with open(inlet['transstream_file_name']) as f:
        transstream = json.load(f)['transstream']

    users_ind_rep_calc = []
    for i in range(len(transstream)):
        users_ind_rep_calc.append(transstream[i]['sender_ind'])
        users_ind_rep_calc.append(transstream[i]['receiver_ind'])
    users_ind_calc = list(set(users_ind_rep_calc))

    timematr_calc = [[[0.] for _ in range(len(users_ind_calc))] for _ in
                     range(len(users_ind_calc))]
    transmatr_calc = [[[0.] for _ in range(len(users_ind_calc))] for _ in
                      range(len(users_ind_calc))]

    for i in range(len(transstream)):
        timematr_calc[transstream[i]['sender_ind']][
            transstream[i]['receiver_ind']].append(transstream[i]['time'])
        transmatr_calc[transstream[i]['sender_ind']][
            transstream[i]['receiver_ind']].append(transstream[i]['trans'])

    with open(inlet['users_ind_calc_file_name'], 'w') as f:
        json.dump({'users_ind_calc': users_ind_calc}, f, sort_keys=True,
                  indent=4 * ' ')

    with open(inlet['timematr_calc_file_name'], 'w') as f:
        json.dump({'timematr_calc': timematr_calc}, f, sort_keys=True,
                  indent=4 * ' ')

    with open(inlet['transmatr_calc_file_name'], 'w') as f:
        json.dump({'transmatr_calc': transmatr_calc}, f, sort_keys=True,
                  indent=4 * ' ')

    print(transmatr_calc)


if __name__ == '__main__':
    actmatr_calc(file_name_inlet='inlet/actmatr_from_stream_inlet.json')
