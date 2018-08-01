import sys
import os
import json

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))

from core.flowstat import FlowStat
from core.routersetts import RouterSetts


class RouterMgt(FlowStat):

    def __init__(self, transseq, setts):
        super().__init__(transseq, setts)
        self.indexes = dict()
        self.pmnt_fee_prop_eff = 1.E-6 * self.setts.pmnt_fee_prop
        self.pmnt_fee_base_eff = 1.E-3 * self.setts.pmnt_fee_base
        self.lim_extrem = dict()
        self.balances = dict()
        self.lim_idle = dict()
        self.periods_in_eff = dict()
        self.periods_out_eff = dict()
        self.lim_period_in = dict()
        self.freqs_out = dict()
        self.freqs_in = dict()
        self.freqs = dict()
        self.wanes = dict()
        self.bounds = dict()
        self.balances_eff = dict()
        self.gain_eff = dict()

    def calc_parameters(self):

        self.calc_flow(self.setts.prob_cut)

        self.calc_ind_list()

        self.calc_extremum()

        self.account_idle()

        self.calc_periods_in_eff()
        self.calc_periods_out_eff()
        self.account_periods_in_eff()

        self.calc_freqs_in()
        self.calc_freqs_out()

        self.calc_closure()

    def calc_ind_list(self):

        self.indexes.clear()
        router_id = '0'
        for user_id, ind in self.users_ind.items():
            if user_id != router_id:
                self.indexes[user_id] = ind

    def calc_extremum(self):

        self.lim_extrem.clear()
        for user_id, ind in self.indexes.items():
            value = self.setts.blch_fee / self.pmnt_fee_prop_eff
            self.lim_extrem[user_id] = value
            if self.setts.income:
                self.lim_extrem[user_id] *= 2

        self.balances.clear()
        for user_id, ind in self.lim_extrem.items():
            self.balances[user_id] = self.lim_extrem[user_id]

    def account_idle(self):

        self.lim_idle.clear()
        for user_id, ind in self.indexes.items():
            amount = self.flowvect_in[ind] * self.setts.blch_period
            amount /= self.setts.acceleration
            self.lim_idle[user_id] = amount / self.setts.idle_mult

        for user_id, ind in self.indexes.items():
            lim = self.lim_idle[user_id]
            if self.balances[user_id] < lim:
                self.balances[user_id] = lim

    def calc_periods_in_eff(self):

        self.periods_in_eff.clear()
        for user_id, ind in self.indexes.items():
            self.periods_in_eff[user_id] = None

        for _, i in self.users_ind.items():
            for user_id, j in self.indexes.items():
                period = self.smart_period[i][j].mean
                flow = self.flowmatr[i][j]
                if period is not None:
                    value = period * flow / self.flowvect_in[j]
                    if self.periods_in_eff[user_id] is None:
                        self.periods_in_eff[user_id] = value
                    else:
                        self.periods_in_eff[user_id] += value

    def calc_periods_out_eff(self):

        self.periods_out_eff.clear()
        for user_id, ind in self.indexes.items():
            self.periods_out_eff[user_id] = None

        for user_id, i in self.indexes.items():
            for _, j in self.users_ind.items():
                period = self.smart_period[i][j].mean
                flow = self.flowmatr[i][j]
                if period is not None:
                    value = period * flow / self.flowvect_out[i]
                    if self.periods_out_eff[user_id] is None:
                        self.periods_out_eff[user_id] = value
                    else:
                        self.periods_out_eff[user_id] += value

    def account_periods_in_eff(self):

        self.lim_period_in.clear()
        for user_id, ind in self.indexes.items():
            self.lim_period_in[user_id] = None

        for user_id, ind in self.indexes.items():
            period = self.periods_in_eff[user_id]
            if period is not None:
                self.lim_period_in[user_id] = self.flowvect_in[ind] * period
                self.lim_period_in[user_id] *= self.setts.period_mult

        for user_id, ind in self.indexes.items():
            lim = self.lim_period_in[user_id]
            if lim is None or lim > self.balances[user_id]:
                self.balances[user_id] = lim

    def calc_freqs_in(self):

        self.freqs_in.clear()
        for user_id, ind in self.indexes.items():
            self.freqs_in[user_id] = None

        for user_id, ind in self.indexes.items():
            balance = self.balances[user_id]
            if balance is not None:
                self.freqs_in[user_id] = self.flowvect_in[ind] / balance

    def calc_freqs_out(self):

        self.freqs_out.clear()
        for user_id, ind in self.indexes.items():
            self.freqs_out[user_id] = None

        for user_id_i, i in self.indexes.items():
            for user_id_j, j in self.indexes.items():
                if self.smart_period[j][i].mean is not None:
                    freq_in = self.freqs_in[user_id_i]
                    freq_out = self.freqs_out[user_id_j]
                    if freq_in is not None:
                        if freq_out is None or freq_out > freq_in:
                            self.freqs_out[user_id_j] = freq_in

    def calc_closure(self):

        self.wanes.clear()
        for user_id, ind in self.indexes.items():
            self.wanes[user_id] = None

        self.freqs.clear()
        for user_id, ind in self.indexes.items():
            self.freqs[user_id] = None

        for user_id, ind in self.indexes.items():
            if self.flowvect_in_eff[ind] >= 0:
                self.freqs[user_id] = self.freqs_in[user_id]
                self.wanes[user_id] = True
            else:
                self.freqs[user_id] = self.freqs_out[user_id]
                self.wanes[user_id] = False

        self.bounds.clear()
        for user_id, ind in self.indexes.items():
            self.bounds[user_id] = None

        for user_id, ind in self.indexes.items():
            if self.wanes[user_id]:
                self.bounds[user_id] = self.balances[user_id]
            else:
                self.bounds[user_id] = 0
                if self.freqs[user_id] is not None:
                    self.bounds[user_id] = -self.flowvect_in_eff[ind] / \
                                           self.freqs_out[user_id]
                if self.balances[user_id] is not None:
                    self.bounds[user_id] += self.balances[user_id]

        for user_id, ind in self.indexes.items():
            if self.wanes[user_id]:
                self.balances[user_id] += self.flowvect_in_eff[ind] / \
                                          self.freqs[user_id]
            if self.balances[user_id] is None:
                self.balances[user_id] = float(0)

        for user_id, ind in self.indexes.items():
            self.balances[user_id] = round(self.balances[user_id])

        self.balances_eff.clear()
        for user_id, ind in self.indexes.items():
            self.balances_eff[user_id] = 0.5 * (
                    self.balances[user_id] + self.bounds[user_id])

        self.gain_eff.clear()
        for user_id, ind in self.indexes.items():
            if self.period_eff_gain[ind] is not None:
                gain = self.flowvect_gain[ind] * self.pmnt_fee_prop_eff
                gain += self.pmnt_fee_base_eff / self.period_eff_gain[ind]
                gain -= self.setts.blch_fee * self.freqs[user_id]
                self.gain_eff[user_id] = float(gain)
            else:
                self.gain_eff[user_id] = float(0)


if __name__ == '__main__':
    router_setts = RouterSetts()
    router_setts.get_from_file('../optimizer/routersetts.ini')

    with open('../activity/outlet/transseq.json') as f:
        transseq = json.load(f)['transseq']

    router_mgt = RouterMgt(transseq, router_setts)
    router_mgt.accelerate_transseq()
    router_mgt.calc_parameters()

    print('lim_extrem', router_mgt.lim_extrem)
    print('lim_idle', router_mgt.lim_idle)
    print('lim_period_in', router_mgt.lim_period_in)
    print('balances', router_mgt.balances)
    print('bounds', router_mgt.bounds)
    print('flowvect_in', router_mgt.flowvect_in)
    print('flowvect_out', router_mgt.flowvect_out)
    print('flowvect_in_eff', router_mgt.flowvect_in_eff)
    print('freqs_in', router_mgt.freqs_in)
    print('freqs_out', router_mgt.freqs_out)
    print('freqs', router_mgt.freqs)
    print('wanes', router_mgt.wanes)
    print('balances_eff', router_mgt.balances_eff)
    print('gain_eff', router_mgt.gain_eff)
