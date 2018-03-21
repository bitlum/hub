import random


def generate_sample(number, mean, stdev):
    value = []
    for _ in range(number):
        value.append(random.gauss(mean, stdev))
        if value[-1] < 0:
            value[-1] = 0
    return value[:]
