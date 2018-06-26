import sys
import os
import json

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))

from core.transsample import TransSample
from samples.smartsample import SmartSample
from core.routersetts import RouterSetts


class TransStat(TransSample):

    def __init__(self, transseq, setts):
        super().__init__(transseq, setts)

        self.smart_period = list()
        self.smart_amount = list()

    def calc_stat(self, prob_cut=0.5):
        self.calc_data()

        self.smart_period.clear()
        self.smart_period = [[SmartSample(period) for period in periodvect] for
                             periodvect in self.periodmatr]

        self.smart_amount.clear()
        self.smart_amount = [[SmartSample(amount) for amount in amountvect] for
                             amountvect in self.amountmatr]

        for i in range(self.users_number):
            for j in range(self.users_number):
                self.smart_period[i][j].calc_stat(prob_cut)
                self.smart_amount[i][j].calc_stat(prob_cut)


if __name__ == '__main__':
    router_setts = RouterSetts()
    router_setts.set_from_file('../optimizer/routermgt_inlet.json')

    with open('../activity/outlet/transseq.json') as f:
        transseq = json.load(f)['transseq']

    prob_cut = 0.5

    trans_stat = TransStat(transseq,router_setts)
    trans_stat.accelerate_transseq()
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
