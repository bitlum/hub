import sys
import os
import json

from core.transstat import TransStat

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))


class FlowStat(TransStat):

    def __init__(self, transseq):
        super().__init__(transseq)
        self.flowmatr = list()
        self.flowvect_out = list()
        self.flowvect_in = list()

    def calc_flow(self, prob_cut=0.5):
        self.calc_stat(prob_cut)
        self.flowmatr.clear()

        self.flowmatr = [[amount.cut for amount in amount_vect] for amount_vect
                         in self.smart_amount]

        for i in range(self.users_number):
            for j in range(self.users_number):
                period = self.smart_period[i][j].cut
                if self.flowmatr[i][j] is not None and period > 0:
                    self.flowmatr[i][j] /= period

        self.flowvect_out.clear()
        self.flowvect_out = [float() for _ in range(self.users_number)]

        self.flowvect_in.clear()
        self.flowvect_in = [float() for _ in range(self.users_number)]

        for i in range(self.users_number):
            for j in range(self.users_number):
                value = self.flowmatr[i][j]
                if value is not None:
                    self.flowvect_out[i] += value
                    self.flowvect_in[j] += value


if __name__ == '__main__':
    with open('../activity/outlet/transseq.json') as f:
        transseq = json.load(f)['transseq']

    prob_cut = 0.5

    flow_stat = FlowStat(transseq)
    flow_stat.calc_flow(prob_cut)

    print('flowmatr:')
    print(flow_stat.flowmatr)
    print()

    print('flowvect_out:')
    print(flow_stat.flowvect_out)
    print()

    print('flowvect_in:')
    print(flow_stat.flowvect_in)
    print()
