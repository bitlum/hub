import sys
import os
import json

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))

from core.transstat import TransStat
from core.routersetts import RouterSetts


class FlowStat(TransStat):

    def __init__(self, transseq, setts):
        super().__init__(transseq, setts)
        self.amountmatr_mean = list()
        self.amountmatr_number = list()
        self.flowmatr = list()
        self.flowvect_out = list()
        self.flowvect_in = list()
        self.flowvect_in_eff = list()
        self.flowvect_gain = list()
        self.period_eff_gain = list()

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

        self.flowvect_gain.clear()
        self.flowvect_gain = [float() for _ in range(self.users_number)]

        self.period_eff_gain.clear()
        self.period_eff_gain = [float() for _ in range(self.users_number)]

        for i in range(self.users_number):
            for j in range(self.users_number):
                value = self.flowmatr[i][j]
                if value is not None:
                    self.flowvect_out[i] += value
                    self.flowvect_in[j] += value
                    if self.users_id[i] != '0' and self.users_id[j] != '0':
                        self.flowvect_gain[j] += value

                        weight_period = value * self.smart_period[i][j].mean
                        self.period_eff_gain[j] += weight_period

        for i in range(self.users_number):
            if self.period_eff_gain[i] != 0:
                self.period_eff_gain[i] /= self.flowvect_gain[i]
            else:
                self.period_eff_gain[i] = None

        self.flowvect_in_eff.clear()
        self.flowvect_in_eff = [flow for flow in self.flowvect_in]
        for i in range(self.users_number):
            self.flowvect_in_eff[i] -= self.flowvect_out[i]

        self.calc_amount_mean()


if __name__ == '__main__':
    router_setts = RouterSetts()
    router_setts.get_from_file('../optimizer/routersetts.ini')

    with open('../activity/outlet/transseq.json') as f:
        transseq = json.load(f)['transseq']

    prob_cut = 0.5

    flow_stat = FlowStat(transseq, router_setts)
    flow_stat.accelerate_transseq()
    flow_stat.calc_flow(prob_cut)

    print('flowmatr:')
    print(flow_stat.flowmatr)
    print()

    print('flowvect_out:')
    print(flow_stat.flowvect_out)
    print()

    print('flowvect_gain:')
    print(flow_stat.flowvect_gain)
    print()

    print('period_eff_gain:')
    print(flow_stat.period_eff_gain)
    print()

    print('flowvect_in:')
    print(flow_stat.flowvect_in)
    print()
