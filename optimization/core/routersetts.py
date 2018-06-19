import sys
import os
import json

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))


class RouterSetts:

    def __init__(self):
        self.income = bool(True)
        self.blch_fee = float(1)
        self.pmnt_fee_base = int(0)
        self.pmnt_fee_prop = int(0)
        self.blockchain_period = float(0)
        self.idle_mult = float(1)
        self.period_mult = float(1)
        self.prob_cut = float(0.5)
        self.init_period = int()
        self.mgt_period = float(1)
        self.stat_period = float(1)
        self.init_mult = float(1)
        self.make_drawing = bool(True)
        self.output_statistics = bool(True)
        self.show_log = bool(True)

    def set_settings(self, income, blch_fee,
                     pmnt_fee_base, pmnt_fee_prop,
                     blockchain_period, idle_mult, period_mult, prob_cut,
                     init_period, mgt_period,
                     stat_period, init_mult, make_drawing,
                     output_statistics, show_log):
        self.set_income(income)
        self.set_blch_fee(blch_fee)
        self.set_pmnt_fee_base(pmnt_fee_base)
        self.set_pmnt_fee_prop(pmnt_fee_prop)
        self.set_blockchain_period(blockchain_period)
        self.set_idle_mult(idle_mult)
        self.set_period_mult(period_mult)
        self.set_prob_cut(prob_cut)
        self.set_init_period(init_period)
        self.set_mgt_period(mgt_period)
        self.set_stat_period(stat_period)
        self.set_init_mult(init_mult)
        self.set_make_drawing(make_drawing)
        self.set_output_statistics(output_statistics)
        self.set_show_log(show_log)

    def set_setts_from_file(self, file_name):
        with open(file_name) as f:
            inlet = json.load(f)
        self.set_income(inlet['income'])
        self.set_blch_fee(inlet['blch_fee'])
        self.set_pmnt_fee_base(inlet['pmnt_fee_base'])
        self.set_pmnt_fee_prop(inlet['pmnt_fee_prop'])
        self.set_blockchain_period(inlet['blockchain_period'])
        self.set_idle_mult(inlet['idle_mult'])
        self.set_period_mult(inlet['period_mult'])
        self.set_prob_cut(inlet['prob_cut'])
        self.set_init_period(inlet['init_period'])
        self.set_mgt_period(inlet['mgt_period'])
        self.set_stat_period(inlet['stat_period'])
        self.set_init_mult(inlet['init_mult'])
        self.set_make_drawing(inlet['make_drawing'])
        self.set_output_statistics(inlet['output_statistics'])
        self.set_show_log(inlet['show_log'])

    def set_income(self, income):
        self.income = income

    def set_blch_fee(self, blch_fee):
        self.blch_fee = blch_fee

    def set_pmnt_fee_prop(self, pmnt_fee_prop):
        self.pmnt_fee_prop = pmnt_fee_prop

    def set_pmnt_fee_base(self, pmnt_fee_base):
        self.pmnt_fee_base = pmnt_fee_base

    def set_blockchain_period(self, blockchain_period):
        self.blockchain_period = blockchain_period

    def set_idle_mult(self, idle_mult):
        self.idle_mult = idle_mult

    def set_period_mult(self, period_mult):
        self.period_mult = period_mult

    def set_prob_cut(self, prob_cut):
        self.prob_cut = prob_cut

    def set_init_period(self, init_period):
        self.init_period = init_period

    def set_mgt_period(self, mgt_period):
        self.mgt_period = mgt_period

    def set_stat_period(self, stat_period):
        self.stat_period = stat_period

    def set_init_mult(self, init_mult):
        self.init_mult = init_mult

    def set_make_drawing(self, make_drawing):
        self.make_drawing = make_drawing

    def set_output_statistics(self, output_statistics):
        self.output_statistics = output_statistics

    def set_show_log(self, show_log):
        self.show_log = show_log

    def __str__(self):
        out_str = 'router settings:\n'
        out_str += 'income ' + str(self.income) + '\n'
        out_str += 'blch_fee ' + str(self.blch_fee) + '\n'
        out_str += 'pmnt_fee_prop ' + str(
            self.pmnt_fee_prop) + '\n'
        out_str += 'pmnt_fee_base ' + str(self.pmnt_fee_base) + '\n'
        out_str += 'blockchain_period ' + str(self.blockchain_period) + '\n'
        out_str += 'idle_mult ' + str(self.idle_mult) + '\n'
        out_str += 'period_mult ' + str(self.period_mult) + '\n'
        out_str += 'prob_cut ' + str(self.prob_cut) + '\n'
        out_str += 'init_period ' + str(self.init_period) + '\n'
        out_str += 'mgt_period ' + str(self.mgt_period) + '\n'
        out_str += 'stat_period ' + str(self.stat_period) + '\n'
        out_str += 'init_mult ' + str(self.init_mult) + '\n'
        out_str += 'make_drawing ' + str(self.make_drawing) + '\n'
        out_str += 'output_statistics ' + str(self.output_statistics) + '\n'
        out_str += 'show_log ' + str(self.show_log)
        return out_str
