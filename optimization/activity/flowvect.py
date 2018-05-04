import json
import random

import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))

from samples.samplegen import generate_sample


# Formation of a vector of flows (each vector element corresponds to a
# specific user), which will subsequently be deployed into a matrix of flows.
# A certain proportion of users is a gates, the flows for such users is
# multiplied by a certain amount.  A logical vector of recipients (users who
# not only send but receive funds) is formed separately.

def flowvect_gen(file_name_inlet):
    with open(file_name_inlet) as f:
        inlet = json.load(f)

    # Initial flow vector

    users_id = {i: str(i + 1) for i in range(inlet['users_number'])}

    channels_id = {i: str(i + 1) for i in range(inlet['users_number'])}

    flowvect = generate_sample(inlet['users_number'], inlet['flow_mean'],
                               inlet['flow_mean'] * inlet['flow_stdev'])

    for i in range(len(flowvect)):
        if flowvect[i] < 0:
            flowvect[i] = 0.

    # Gates accounting in the flow vector

    gates_number = int(inlet['users_number'] * inlet['gates_frac'])

    gates_ind = [i for i in range(inlet['users_number'])]

    gates_ind = random.sample(range(len(gates_ind)), gates_number)

    for i in range(len(gates_ind)):
        flowvect[gates_ind[i]] *= random.uniform(inlet['gates_min_mult'],
                                                 inlet['gates_max_mult'])

    # Recipient accounting

    receivers = [False for _ in range(inlet['users_number'])]

    receivers_number = int(inlet['users_number'] * inlet['receivers_frac'])

    receivers_ind = [i for i in range(inlet['users_number'])]

    receivers_ind = random.sample(range(len(receivers_ind)), receivers_number)

    # write bool receiver vector into a file

    for i in range(len(receivers_ind)):
        receivers[receivers_ind[i]] = True

    # write user id dict into a file

    with open(inlet['users_id_file_name'], 'w') as f:
        json.dump({'users_id': users_id}, f,
                  sort_keys=True, indent=4 * ' ')

    # write channel id dict into a file

    with open(inlet['channels_id_file_name'], 'w') as f:
        json.dump({'channels_id': channels_id}, f,
                  sort_keys=True, indent=4 * ' ')

    # write flow vector into a file

    with open(inlet['flowvect_file_name'], 'w') as f:
        json.dump({'flowvect': flowvect, 'receivers': receivers}, f,
                  sort_keys=True, indent=4 * ' ')


if __name__ == '__main__':
    flowvect_gen(file_name_inlet='inlet/flowvect_inlet.json')
