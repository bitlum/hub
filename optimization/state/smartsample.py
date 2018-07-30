import statistics

import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))

import state.erfcut as erfcut

from state.samplegen import generate_sample


class SmartSample:

    def __init__(self, sample):
        self.sample = sample
        self.number = int()
        self.prob_cut = float()
        self.sum = float()
        self.mean = float()
        self.minimum = float()
        self.maximum = float()
        self.stdev = float()
        self.variance = float()
        self.cut = float()

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

    def calc_stat(self, prob_cut=0.5):
        self.prob_cut = prob_cut
        self.number = len(self.sample)
        if self.number == 0:
            self.sum = None
            self.mean = None
            self.minimum = None
            self.maximum = None
            self.stdev = None
            self.variance = None
            self.cut = None
        else:
            self.calc_sum()
            self.mean = statistics.mean(self.sample)
            self.minimum = min(self.sample)
            self.maximum = max(self.sample)
            self.cut = self.mean
            if self.number == 1:
                self.stdev = 0
                self.variance = 0
            else:
                self.stdev = statistics.stdev(self.sample)
                self.variance = statistics.variance(self.sample)
                if self.stdev > 0:
                    self.calc_erfcut()

    def calc_sum(self):
        self.sum = 0.
        for i in range(self.number):
            self.sum += self.sample[i]

    def calc_erfcut(self):
        self.cut = erfcut.erfcut_calc(self.prob_cut, self.mean, self.stdev,
                                      erfcut.Method.newton, 1.E-5)


if __name__ == '__main__':
    sample = generate_sample(100, 50, 30)
    smart_sample = SmartSample(sample)
    smart_sample.calc_stat()
    print(smart_sample)
