import json

import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))


# Formation of a general time-ordered transaction stream. Transaction stream
# element structure: time, sender_id, receiver_id, trans.

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
            if len(timematr[i][j]) > 0:
                for k in range(len(timematr[i][j])):
                    transstream_unsort.append(
                        dict(time=timematr[i][j][k],
                             sender_id=users_id[str(i)],
                             receiver_id=users_id[str(j)],
                             trans=transmatr[i][j][k]))

    def take_time(trans):
        return trans['time']

    transstream = sorted(transstream_unsort, key=take_time)

    # write a general time-ordered transaction stream into a file

    with open(inlet['transstream_file_name'], 'w') as f:
        json.dump({'transstream': transstream}, f, sort_keys=True,
                  indent=4 * ' ')


if __name__ == '__main__':
    transstream_gen(file_name_inlet='inlet/transstream_inlet.json')
