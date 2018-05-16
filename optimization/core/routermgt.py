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
        self.balances = dict()
        self.idle_lim = list()
        self.periods_eff = list()
        self.period_lim = list()
        self.freqs_out = list()
        self.freqs_in = list()
        self.total_lim = list()
        self.freqs = dict()
        self.wanes = dict()
        self.bounds = dict()

    def calc_parameters(self):
        self.calc_flow(self.setts.prob_cut)
        self.calc_extremum()
        self.account_idle()
        self.calc_periods_eff()
        # self.account_period()
        # self.calc_freqs_out()
        # self.calc_freqs_in()
        # self.calc_total_lim()
        # self.calc_closure()
        # self.calc_closure()

    def calc_extremum(self):
        self.balances.clear()
        self.balances = {
            self.users_id[i]: self.setts.penalty / self.setts.commission
            for i in range(self.users_number)}
        if self.setts.income:
            for i in range(self.users_number):
                self.balances[self.users_id[i]] *= 2

    def account_idle(self):
        self.idle_lim.clear()

        self.idle_lim = [
            self.flowvect_out[i] * self.setts.time_p / self.setts.alpha_p
            for i in range(self.users_number)]

        for i in range(self.users_number):
            lim = self.idle_lim[i]
            user_id = self.users_id[i]
            if self.balances[user_id] < lim:
                self.balances[user_id] = lim

    def calc_periods_eff(self):
        self.periods_eff.clear()

        self.periods_eff = [None for _ in range(self.users_number)]
        for i in range(self.users_number):
            for j in range(self.users_number):
                period = self.smart_period[i][j].mean
                flow = self.flowmatr[i][j]
                if period is not None:
                    amount = period * flow
                    if self.periods_eff[j] is None:
                        self.periods_eff[j] = amount / self.flowvect_in[j]
                    else:
                        self.periods_eff[j] += amount / self.flowvect_in[j]

    def account_period(self):
        self.period_lim.clear()
        self.period_lim = [
            self.setts.alpha_T * self.flowvect_out[i] * self.periods_eff[i]
            for i in range(self.users_number)]

        for i in range(self.users_number):
            lim = self.period_lim[i]
            user_id = self.users_id[i]
            if self.balances[user_id] < lim:
                self.balances[user_id] = lim

    def calc_freqs_out(self):
        self.freqs_out.clear()
        self.freqs_out = [
            self.flowvect_out[i] / self.balances[self.users_id[i]]
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

        self.wanes.clear()
        self.wanes = {self.users_id[i]: bool(False)
                      for i in range(self.users_number)}
        self.freqs.clear()
        self.freqs = {self.users_id[i]: self.freqs_out[i]
                      for i in range(self.users_number)}
        self.balances.clear()
        self.balances = {self.users_id[i]: self.total_lim[i]
                         for i in range(self.users_number)}

        for i in range(self.users_number):
            balance = self.flowvect_in_eff[i] / self.freqs_in[i]
            user_id = self.users_id[i]

            if balance > self.total_lim[i]:
                self.balances[user_id] = balance

            if self.flowvect_in_eff[i] >= 0:
                self.wanes[user_id] = True
                self.freqs[user_id] = self.freqs_in[i]

        self.bounds.clear()
        self.bounds = copy.deepcopy(self.balances)

        for i in range(self.users_number):
            user_id = self.users_id[i]
            flow = self.flowvect_in_eff[i]
            self.bounds[user_id] -= flow / self.freqs[user_id]

        for i in range(self.users_number):
            user_id = self.users_id[i]
            if self.wanes[user_id]:
                self.bounds[user_id] = self.total_lim[i]

        for i in range(self.users_number):
            user_id = self.users_id[i]
            self.balances[user_id] = round(self.balances[user_id])

        # TODO remove this:
        for i in range(self.users_number):
            user_id = self.users_id[i]

            self.freqs_out[i] = round(self.freqs_out[i], 2)
            self.freqs_in[i] = round(self.freqs_in[i], 2)
            self.freqs[user_id] = round(self.freqs[user_id], 2)

            self.total_lim[i] = round(self.total_lim[i])
            self.balances[user_id] = round(self.balances[user_id])
            self.bounds[user_id] = round(self.bounds[user_id])

            self.flowvect_out[i] = round(self.flowvect_out[i], 2)
            self.flowvect_in[i] = round(self.flowvect_in[i], 2)
            self.flowvect_in_eff[i] = round(self.flowvect_in_eff[i], 2)


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

    # print('balances ', router_mgt.balances)
    # print('freqs ', router_mgt.freqs)
    # print('wanes ', router_mgt.wanes)
