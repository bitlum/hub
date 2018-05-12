import sys
import os
import json

from core.transsample import TransSample
from samples.smartsample import SmartSample

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))


class TransStat(TransSample):

    def __init__(self, transstream):
        super().__init__(transstream)

        self.smart_period = list()
        self.smart_amount = list()
        self.period_mean = list()
        self.amount_mean = list()

    def calc_stat(self, prob_cut=0.5):
        self.calc_data()

        self.smart_period = [[SmartSample(period) for period in periodvect] for
                             periodvect in self.periodmatr]
        self.smart_amount = [[SmartSample(amount) for amount in amountvect] for
                             amountvect in self.transmatr]

        for i in range(self.users_number):
            for j in range(self.users_number):
                self.smart_period[i][j].calc_stat(prob_cut)
                self.smart_amount[i][j].calc_stat(prob_cut)


if __name__ == '__main__':
    file_inlet = 'inlet/actmatr_from_stream_inlet.json'

    with open(file_inlet) as f:
        inlet = json.load(f)

    with open(inlet['transstream_file_name']) as f:
        transstream = json.load(f)['transstream']

    prob_cut = 0.5

    trans_stat = TransStat(transstream)
    trans_stat.calc_stat(prob_cut)

    print('period:')
    for period_vect in trans_stat.smart_period:
        for period in period_vect:
            print(period.cut, end='\t')
        print()
    print()

    print('amount:')
    for amount_vect in trans_stat.smart_amount:
        for amount in amount_vect:
            print(amount.cut, end='\t')
        print()
