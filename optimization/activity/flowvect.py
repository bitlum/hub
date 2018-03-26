import json
import random

import sys

sys.path.append('../')

from samplegen import generate_sample


def flowvect_gen(file_name_inlet):
    with open(file_name_inlet) as f:
        inlet = json.load(f)

    flowvect = generate_sample(inlet['users_number'], inlet['flow_mean'],
                               inlet['flow_mean'] * inlet['flow_stdev'])

    for i in range(len(flowvect)):
        if flowvect[i] < 0:
            flowvect[i] = 0.

    receivers = [0 for _ in range(inlet['users_number'])]

    receivers_number = int(inlet['users_number'] * inlet['receivers_frac'])

    receivers_ind = [i for i in range(inlet['users_number'])]

    receivers_ind = random.sample(range(len(receivers_ind)), receivers_number)

    for i in range(len(receivers_ind)):
        receivers[receivers_ind[i]] = 1

    with open(inlet['flowvect_file_name'], 'w') as f:
        json.dump({'flowvect': flowvect, 'receivers': receivers}, f,
                  sort_keys=True, indent=4 * ' ')


if __name__ == '__main__':
    flowvect_gen(file_name_inlet='flowvect_inlet.json')
