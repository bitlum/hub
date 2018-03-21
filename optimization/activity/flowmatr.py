import json
import random


def flowmatr_gen(file_name_inlet):
    with open(file_name_inlet) as f:
        inlet = json.load(f)

    with open(inlet['flowvect_file_name']) as f:
        flowvect = json.load(f)['flowvect']

    with open(inlet['flowvect_file_name']) as f:
        receivers = json.load(f)['receivers']

    matrflow = [flowvect.copy() for _ in range(len(flowvect))]

    for i in range(len(matrflow)):
        matrflow[i][i] = 0

    for i in range(len(matrflow)):
        for j in range(len(matrflow[i])):
            matrflow[i][j] *= random.uniform(
                inlet['min_mult'], inlet['max_mult'])
            matrflow[i][j] *= receivers[j]

    with open(inlet['flowmatr_file_name'], 'w') as f:
        json.dump({'flowmatr': matrflow}, f, sort_keys=True,
                  indent=4 * ' ')


if __name__ == '__main__':
    flowmatr_gen(file_name_inlet='flowmatr_inlet.json')
