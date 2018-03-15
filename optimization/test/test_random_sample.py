import random


def calc(number, mean, minimum, maximum, stdev):

    meannormal = (mean - minimum) / (maximum - minimum)
    stdevnormal = stdev / (maximum - minimum)

    value = []
    for _ in range(number):
        value.append(random.gauss(meannormal, stdevnormal))
        value[-1] *= (maximum - minimum)
        value[-1] += minimum

    return value[:]
