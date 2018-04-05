import random

import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))


def generate_sample(number, mean, stdev):
    value = []
    for _ in range(number):
        value.append(random.gauss(mean, stdev))
        if value[-1] < 0:
            value[-1] = 0
    return value[:]
