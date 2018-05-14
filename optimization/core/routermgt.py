import sys
import os
import json
import copy

from core.flowstat import FlowStat
from core.routersetts import RouterSetts

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))


class RouterMgt(FlowStat):

    def __init__(self, transseq, setts):
        super().__init__(transseq)
        self.setts = setts
        self.balances = list()
        self.idle_lim = list()
        self.periods_max = list()
        self.period_lim = list()
        self.freqs_out = list()
        self.freqs_in = list()
        self.total_lim = list()
        self.freqs = list()
        self.wanes = list()

    def calc_parameters(self):
        self.calc_flow(self.setts.prob_cut)
        self.calc_extremum()
        self.account_idle()
        self.calc_periods_max()
        self.account_period()
        self.calc_freqs_out()
        self.calc_freqs_in()
        self.calc_total_lim()
        self.calc_closure()
        self.calc_closure()

    def calc_extremum(self):
        self.balances.clear()
        self.balances = [self.setts.penalty / self.setts.commission
                         for _ in range(self.users_number)]
        if self.setts.income:
            for i in range(self.users_number):
                self.balances[i] *= 2

    def account_idle(self):
        self.idle_lim.clear()

        self.idle_lim = [
            self.flowvect_out[i] * self.setts.time_p / self.setts.alpha_p
            for i in range(self.users_number)]

        for i in range(self.users_number):
            lim = self.idle_lim[i]
            if self.balances[i] < lim:
                self.balances[i] = lim

    def calc_periods_max(self):
        self.periods_max.clear()

        self.periods_max = [float(0) for _ in range(self.users_number)]
        for i in range(self.users_number):
            for period in self.smart_period[i]:
                if period.mean is not None:
                    if self.periods_max[i] < period.mean:
                        self.periods_max[i] = period.mean

    def account_period(self):
        self.period_lim.clear()
        self.period_lim = [
            self.setts.alpha_T * self.flowvect_out[i] * self.periods_max[i]
            for i in range(self.users_number)]

        for i in range(self.users_number):
            lim = self.period_lim[i]
            if self.balances[i] < lim:
                self.balances[i] = lim

    def calc_freqs_out(self):
        self.freqs_out.clear()
        self.freqs_out = [self.flowvect_out[i] / self.balances[i]
                          for i in range(self.users_number)]

    def calc_freqs_in(self):
        self.freqs_in.clear()
        self.freqs_in = copy.deepcopy(self.freqs_out)

        for i in range(self.users_number):
            for j in range(self.users_number):
                freq = self.freqs_out[i]
                if self.freqs_in[j] < freq:
                    self.freqs_in[j] = freq

    def calc_total_lim(self):
        self.total_lim.clear()
        self.total_lim = [lim for lim in self.idle_lim]
        for i in range(self.users_number):
            lim = self.period_lim[i]
            if self.total_lim[i] < lim:
                self.total_lim[i] = lim

    def calc_closure(self):
        self.freqs.clear()
        self.wanes.clear()
        self.balances.clear()

        self.wanes = [bool(False) for _ in range(self.users_number)]
        self.freqs = copy.deepcopy(self.freqs_out)
        self.balances = copy.deepcopy(self.total_lim)

        for i in range(self.users_number):
            flow_delta = self.flowvect_in[i] - self.flowvect_out[i]
            balance = flow_delta / self.freqs_in[i]

            if balance > self.total_lim[i]:
                self.balances[i] = balance

            if flow_delta >= 0:
                self.wanes[i] = True
                self.freqs[i] = self.freqs_in[i]

        for i in range(self.users_number):
            self.balances[i] = round(self.balances[i])


if __name__ == '__main__':
    file_inlet = 'inlet/routermgt_inlet.json'

    with open(file_inlet) as f:
        inlet = json.load(f)

    router_setts = RouterSetts()

    router_setts.set_income(inlet['income'])
    router_setts.set_penalty(inlet['penalty'])
    router_setts.set_commission(inlet['commission'])
    router_setts.set_time_p(inlet['time_p'])
    router_setts.set_alpha_p(inlet['alpha_p'])
    router_setts.set_alpha_T(inlet['alpha_T'])

    with open('../activity/outlet/transseq.json') as f:
        transseq = json.load(f)['transseq']

    router_mgt = RouterMgt(transseq, router_setts)
    router_mgt.calc_parameters()

    print('balances ', router_mgt.balances)
    print('freqs ', router_mgt.freqs)
    print('wanes ', router_mgt.wanes)
