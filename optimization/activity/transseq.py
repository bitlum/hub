import json

import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))


# Formation of a general time-ordered transaction stream. Transaction stream
# element structure: time, sender_id, receiver_id, trans.

def transseq_gen(file_name_inlet):
    with open(file_name_inlet) as f:
        inlet = json.load(f)

    with open(inlet['users_id_file_name']) as f:
        users_id = json.load(f)['users_id']

    with open(inlet['timematr_file_name']) as f:
        timematr = json.load(f)['timematr']

    with open(inlet['amountmatr_file_name']) as f:
        amountmatr = json.load(f)['amountmatr']

    transstream_unsort = []
    for i in range(len(timematr)):
        for j in range(len(timematr[i])):
            if len(timematr[i][j]) > 0:
                for k in range(len(timematr[i][j])):
                    transstream_unsort.append(
                        dict(time=1.E9 * timematr[i][j][k],
                             payment=dict(sender=users_id[str(i)],
                                          receiver=users_id[str(j)],
                                          amount=round(amountmatr[i][j][k]),
                                          earned=int(0))))

    def take_time(trans):
        return trans['time']

    transseq = sorted(transstream_unsort, key=take_time)

    # write a general time-ordered transaction stream into a file

    with open(inlet['transseq_file_name'], 'w') as f:
        json.dump({'transseq': transseq}, f, sort_keys=True,
                  indent=4 * ' ')


if __name__ == '__main__':
    transseq_gen(file_name_inlet='inlet/transseq_inlet.json')
