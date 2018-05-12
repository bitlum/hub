import math
from enum import Enum

import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))


def gaussian(x, mean, stdev):
    return math.exp(
        -((x - mean) ** 2) / (2 * stdev ** 2)) / stdev / math.sqrt(
        2 * math.pi)


def gauss_distr(x, mean, stdev):
    return (1 + math.erf((x - mean) / math.sqrt(2 * stdev ** 2))) / 2


# gaussian() and gauss_distr() test
#
# x_ = - 0.5
# mean_ = 0
# stdev_ = math.sqrt(0.2)
#
# print('gaussian = ', '{:.3f}'.format(gaussian(x_, mean_, stdev_)))
# print('gauss_distr = ', '{:.3f}'.format(gauss_distr(x_, mean_, stdev_)))


def newton_method(func, func_der, y, x_ini, accuracy):
    x_cur = x_ini
    while True:
        x_tmp = x_cur
        x_cur = x_tmp - (func(x_tmp) - y) / func_der(x_tmp)
        if (True if x_tmp + x_cur == 0 else math.fabs(
                2 * (x_cur - x_tmp) / (x_cur + x_tmp)) <= accuracy):
            break
    return x_cur


def shooting_method(func, y, x_ini, step_ini, accuracy):
    if func(x_ini) == y:
        return x_ini
    else:
        x_cur = x_ini
        step_cur = step_ini
        while math.fabs(step_cur / step_ini) >= accuracy:
            while True:
                dist_cur = math.fabs(func(x_cur) - y)
                x_cur += step_cur
                if dist_cur <= math.fabs(func(x_cur) - y):
                    break
            step_cur *= -0.5
    return x_cur


# newton_method and shooting_method test
#
# def f(x):
#     return x ** 2
#
#
# def f_der(x):
#     return 2 * x
#
#
# value = 4.5
# x_init = 1
# shooting_step = 1
# accur = 1.E-5
#
# print('shooting_method: ', '{:.3f}'.format(
#     shooting_method(f, value, x_init, shooting_step, accur)))
# print('newton_method: ',
#       '{:.3f}'.format(newton_method(f, f_der, value, x_init, accur)))


class GaussWrap:

    def __init__(self, mean, stdev):
        self.mean = mean
        self.stdev = stdev

    def calc(self, x):
        return gaussian(x, self.mean, self.stdev)


class GaussDistrWrap:

    def __init__(self, mean, stdev):
        self.mean = mean
        self.stdev = stdev

    def calc(self, x):
        return gauss_distr(x, self.mean, self.stdev)


# Handle calculation cutting test
#
# mean_ = 0
# stdev_ = math.sqrt(5)
# probability = 0.2
# x_init_ = mean_
# shooting_step_ = stdev_ * 0.1
# accur_ = 1.E-5
#
# print('shooting_method: ',
#       '{:.3f}'.format(
#           shooting_method(GaussDistrWrap(mean_, stdev_).calc, probability,
#                           x_init_,
#                           shooting_step_, accur_)))
#
# print('newton_method: ',
#       '{:.3f}'.format(
#           newton_method(GaussDistrWrap(mean_, stdev_).calc,
#                         GaussWrap(mean_, stdev_).calc,
#                         probability, x_init_,
#                         accur_)))


class Method(Enum):
    shooting = 1
    newton = 2


def erfcut_calc(probability, mean, stdev, method, accuracy):
    x_init = mean
    shooting_step = stdev * 0.1
    if method == Method.shooting:
        return shooting_method(GaussDistrWrap(mean, stdev).calc, probability,
                               x_init, shooting_step, accuracy)
    elif method == Method.newton:
        return newton_method(GaussDistrWrap(mean, stdev).calc,
                             GaussWrap(mean, stdev).calc, probability, x_init,
                             accuracy)


if __name__ == '__main__':
    # Final calculation cutting test
    #
    mean_ = 0
    stdev_ = math.sqrt(5)
    prob_ = 0.2
    accur_ = 1.E-5

    print('shooting_method: ',
          '{:.3f}'.format(
              erfcut_calc(prob_, mean_, stdev_, Method.shooting, accur_)))

    print('newton_method: ',
          '{:.3f}'.format(
              erfcut_calc(prob_, mean_, stdev_, Method.newton, accur_)))
