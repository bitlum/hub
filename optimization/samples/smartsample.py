import statistics

import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))

import samples.erfcut as erfcut


class SmartSample:

    def __init__(self, value, prob_cut):
        self.value = value[:]
        self.number = 0
        self.sum = 0.
        self.mean = 0.
        self.minimum = 0.
        self.maximum = 0.
        self.stdev = 0.
        self.variance = 0.
        self.cut = 0.
        self.prob_cut = prob_cut
        self.calcstatistic()

    def __str__(self):
        outstr = ''
        outstr += 'number is ' + str('{:.0f}'.format(self.number)) + '\n'
        outstr += 'sum is ' + str('{:.3f}'.format(self.sum)) + '\n'
        outstr += 'mean is ' + str('{:.3f}'.format(self.mean)) + '\n'
        outstr += 'minimum is ' + str('{:.3f}'.format(self.minimum)) + '\n'
        outstr += 'maximum is ' + str('{:.3f}'.format(self.maximum)) + '\n'
        outstr += 'stdev is ' + str('{:.3f}'.format(self.stdev)) + '\n'
        outstr += 'variance is ' + str('{:.3f}'.format(self.variance)) + '\n'
        outstr += 'prob_cut is ' + str('{:.3f}'.format(self.prob_cut)) + '\n'
        outstr += 'cut is ' + str('{:.3f}'.format(self.cut))
        return outstr

    def calcstatistic(self):
        self.number = len(self.value)
        self.calcsum()
        self.mean = statistics.mean(self.value)
        self.minimum = min(self.value)
        self.maximum = max(self.value)
        if len(self.value) > 1:
            self.stdev = statistics.stdev(self.value)
            self.variance = statistics.variance(self.value)
        else:
            self.stdev = 0
            self.variance = 0
        self.erfcut()

    def calcsum(self):
        self.sum = 0.
        for i in range(self.number):
            self.sum += self.value[i]

    def erfcut(self):
        if self.stdev > 0:
            self.cut = erfcut.erf_cut_calc(self.prob_cut,
                                           self.mean, self.stdev,
                                           erfcut.Method.newton, 1.E-5)
        else:
            self.cut = self.mean
