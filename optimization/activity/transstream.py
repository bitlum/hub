import json

import sys

sys.path.append('../')


def transstream_gen(file_name_inlet):
    with open(file_name_inlet) as f:
        inlet = json.load(f)

    with open(inlet['users_id_file_name']) as f:
        users_id = json.load(f)['users_id']

    with open(inlet['timematr_file_name']) as f:
        timematr = json.load(f)['timematr']

    with open(inlet['transmatr_file_name']) as f:
        transmatr = json.load(f)['transmatr']

    transstream_unsort = []
    for i in range(len(timematr)):
        for j in range(len(timematr[i])):
            if len(timematr[i][j]) > 1:
                for k in range(1, len(timematr[i][j])):
                    transstream_unsort.append(
                        dict(time=timematr[i][j][k], sender_id=users_id[i],
                             receiver_id=users_id[j],
                             trans=transmatr[i][j][k]))

    def take_time(trans):
        return trans['time']

    transstream = sorted(transstream_unsort, key=take_time)

    with open(inlet['transstream_file_name'], 'w') as f:
        json.dump({'transstream': transstream}, f, sort_keys=True,
                  indent=4 * ' ')


if __name__ == '__main__':
    transstream_gen(file_name_inlet='inlet/transstream_inlet.json')
