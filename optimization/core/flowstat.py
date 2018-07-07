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
        self.amountvec_mean = dict()
        self.amount_mean_forw = float()
        self.amount_mean_io = float()
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

    def calc_amount_mean(self):
        self.amountmatr_mean.clear()
        self.amountmatr_mean = [[amount.mean for amount in amount_vect] for
                                amount_vect in self.smart_amount]
        self.amountmatr_number.clear()
        self.amountmatr_number = [[amount.number for amount in amount_vect] for
                                  amount_vect in self.smart_amount]

        self.amountvec_mean.clear()
        for i in range(self.users_number):
            val = 0
            num = 0
            for j in range(self.users_number):
                if self.amountmatr_mean[i][j] is not None:
                    val += self.amountmatr_mean[i][j] * \
                           self.amountmatr_number[i][j]
                    num += self.amountmatr_number[i][j]
            if val > 0:
                self.amountvec_mean[self.users_id[i]] = val / num
            else:
                self.amountvec_mean[self.users_id[i]] = None

        num_total_forw = 0
        num_total_io = 0
        for i in range(self.users_number):
            val_forw = 0
            val_io = 0
            num_forw = 0
            num_io = 0
            for j in range(self.users_number):
                if self.amountmatr_mean[i][j] is not None:
                    val = self.amountmatr_mean[i][j] * \
                          self.amountmatr_number[i][j]

                    if self.users_id[i] == '0' or self.users_id[j] == '0':
                        val_io += val
                        num_io += self.amountmatr_number[i][j]
                    else:
                        val_forw += val
                        num_forw += self.amountmatr_number[i][j]

            self.amount_mean_forw += val_forw
            num_total_forw += num_forw

            self.amount_mean_io += val_io
            num_total_io += num_io

        if num_total_forw > 0:
            self.amount_mean_forw /= num_total_forw

        if num_total_io > 0:
            self.amount_mean_io /= num_total_io


if __name__ == '__main__':
    router_setts = RouterSetts()
    router_setts.set_from_file('../optimizer/routermgt_inlet.json')

    with open('../activity/outlet/transseq.json') as f:
        transseq = json.load(f)['transseq']

    prob_cut = 0.5

    flow_stat = FlowStat(transseq, router_setts)
    flow_stat.accelerate_transseq()
    flow_stat.calc_flow(prob_cut)

    print('amountvec_mean:')
    print(flow_stat.amountvec_mean)
    print()

    print('amount_mean_forw:')
    print(flow_stat.amount_mean_forw)
    print()

    print('amount_mean_io:')
    print(flow_stat.amount_mean_io)
    print()

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
