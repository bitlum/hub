import sys
import os
import json

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))

from core.flowstat import FlowStat
from core.routersetts import RouterSetts


class RouterMgt(FlowStat):

    def __init__(self, transseq, setts):
        super().__init__(transseq)
        self.setts = setts
        self.balances = dict()
        self.lim_idle = list()
        self.periods_in_eff = list()
        self.periods_out_eff = list()
        self.lim_period_in = list()
        self.freqs_out = list()
        self.freqs_in = list()
        self.freqs = dict()
        self.wanes = dict()
        self.bounds = dict()

    def calc_parameters(self):
        self.calc_flow(self.setts.prob_cut)
        self.calc_extremum()
        self.account_idle()
        self.calc_periods_in_eff()
        self.calc_periods_out_eff()
        self.account_periods_in_eff()
        self.calc_freqs_in()
        self.calc_freqs_out()
        self.calc_closure()

    def calc_extremum(self):
        self.balances.clear()
        for i in range(self.users_number):
            user = self.users_id[i]
            self.balances[user] = self.setts.penalty / self.setts.commission
            if self.setts.income:
                self.balances[user] *= 2

    def account_idle(self):
        self.lim_idle.clear()
        for i in range(self.users_number):
            amount = self.flowvect_out[i] * self.setts.time_p
            self.lim_idle.append(amount / self.setts.alpha_p)

        for i in range(self.users_number):
            lim = self.lim_idle[i]
            user_id = self.users_id[i]
            if self.balances[user_id] < lim:
                self.balances[user_id] = lim

    def calc_periods_in_eff(self):
        self.periods_in_eff.clear()
        for _ in range(self.users_number):
            self.periods_in_eff.append(None)

        for i in range(self.users_number):
            for j in range(self.users_number):
                period = self.smart_period[i][j].mean
                flow = self.flowmatr[i][j]
                if period is not None:
                    value = period * flow / self.flowvect_in[j]
                    if self.periods_in_eff[j] is None:
                        self.periods_in_eff[j] = value
                    else:
                        self.periods_in_eff[j] += value

    def calc_periods_out_eff(self):
        self.periods_out_eff.clear()
        for _ in range(self.users_number):
            self.periods_out_eff.append(None)

        for i in range(self.users_number):
            for j in range(self.users_number):
                period = self.smart_period[i][j].mean
                flow = self.flowmatr[i][j]
                if period is not None:
                    value = period * flow / self.flowvect_out[i]
                    if self.periods_out_eff[i] is None:
                        self.periods_out_eff[i] = value
                    else:
                        self.periods_out_eff[i] += value

    def account_periods_in_eff(self):

        self.lim_period_in.clear()
        for _ in range(self.users_number):
            self.lim_period_in.append(None)

        for i in range(self.users_number):
            periods_in_eff = self.periods_in_eff[i]
            if periods_in_eff is not None:
                self.lim_period_in[i] = self.flowvect_in[i] * periods_in_eff
                self.lim_period_in[i] *= self.setts.alpha_T

        for i in range(self.users_number):
            lim = self.lim_period_in[i]
            user_id = self.users_id[i]
            if lim is None or lim > self.balances[user_id]:
                self.balances[user_id] = lim

    def calc_freqs_in(self):

        self.freqs_in.clear()
        for _ in range(self.users_number):
            self.freqs_in.append(None)

        for i in range(self.users_number):
            balance = self.balances[self.users_id[i]]
            if balance is not None:
                self.freqs_in[i] = self.flowvect_in[i] / balance

    def calc_freqs_out(self):

        self.freqs_out.clear()
        for _ in range(self.users_number):
            self.freqs_out.append(None)

        for i in range(self.users_number):
            for j in range(self.users_number):
                if self.smart_period[j][i].mean is not None:
                    freq_in = self.freqs_in[i]
                    freq_out = self.freqs_out[j]
                    if freq_in is not None:
                        if freq_out is None or freq_out > freq_in:
                            self.freqs_out[j] = freq_in

    def calc_closure(self):

        self.wanes.clear()
        for i in range(self.users_number):
            user = self.users_id[i]
            self.wanes[user] = None

        self.freqs.clear()
        for i in range(self.users_number):
            user = self.users_id[i]
            self.freqs[user] = None

        for i in range(self.users_number):
            user_id = self.users_id[i]
            if self.flowvect_in_eff[i] >= 0:
                self.freqs[user_id] = self.freqs_in[i]
                self.wanes[user_id] = True
            else:
                self.freqs[user_id] = self.freqs_out[i]
                self.wanes[user_id] = False

        self.bounds.clear()
        for i in range(self.users_number):
            user = self.users_id[i]
            self.bounds[user] = None

        for i in range(self.users_number):
            user_id = self.users_id[i]

            if self.wanes[user_id]:
                self.bounds[user_id] = self.balances[user_id]
            else:
                balance = 0
                if self.freqs[user_id] is not None:
                    balance = -self.flowvect_in_eff[i] / self.freqs_out[i]
                self.bounds[user_id] = balance
                if self.balances[user_id] is not None:
                    self.bounds[user_id] += self.balances[user_id]

        for i in range(self.users_number):
            user_id = self.users_id[i]

            if self.wanes[user_id]:
                balance = self.flowvect_in_eff[i] / self.freqs[user_id]
                self.balances[user_id] += balance

            if self.balances[user_id] is None:
                self.balances[user_id] = float(0)

        for i in range(self.users_number):
            user_id = self.users_id[i]
            self.balances[user_id] = round(self.balances[user_id])

        # TODO remove this:
        for i in range(self.users_number):
            user_id = self.users_id[i]
            if self.freqs_out[i] is not None:
                self.freqs_out[i] = round(self.freqs_out[i], 2)
            if self.freqs_in[i] is not None:
                self.freqs_in[i] = round(self.freqs_in[i], 2)
            if self.freqs[user_id] is not None:
                self.freqs[user_id] = round(self.freqs[user_id], 2)

            if self.balances[user_id] is not None:
                self.balances[user_id] = round(self.balances[user_id])
            if self.bounds[user_id] is not None:
                self.bounds[user_id] = round(self.bounds[user_id])

            if self.flowvect_out[i] is not None:
                self.flowvect_out[i] = round(self.flowvect_out[i], 2)
            if self.flowvect_in[i] is not None:
                self.flowvect_in[i] = round(self.flowvect_in[i], 2)
            if self.flowvect_in_eff[i] is not None:
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

    print('balances', router_mgt.balances)
    print('bounds', router_mgt.bounds)
    print('flowvect_in', router_mgt.flowvect_in)
    print('flowvect_out', router_mgt.flowvect_out)
    print('flowvect_in_eff', router_mgt.flowvect_in_eff)
    print('freqs_in', router_mgt.freqs_in)
    print('freqs_out', router_mgt.freqs_out)
    print('freqs', router_mgt.freqs)
    print('wanes', router_mgt.wanes)
