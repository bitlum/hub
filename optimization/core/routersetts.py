import sys
import os
import json

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))


class RouterSetts:

    def __init__(self):
        self.income = bool(True)
        self.penalty = float(1)
        self.commission = float(0)
        self.payment_fee_base = int(0)
        self.payment_fee_proportional = int(0)
        self.time_p = float(0)
        self.alpha_p = float(1)
        self.alpha_T = float(1)
        self.prob_cut = float(0.5)
        self.init_time_period = int()
        self.mgt_period = float(1)
        self.stat_period = float(1)
        self.init_mult = float(1)
        self.make_drawing = bool(True)
        self.output_statistics = bool(True)
        self.show_log = bool(True)

    def set_settings(self, income, penalty, commission,
                     payment_fee_base, payment_fee_proportional,
                     time_p, alpha_p, alpha_T, prob_cut,
                     init_time_period, mgt_period,
                     stat_period, init_mult, make_drawing,
                     output_statistics, show_log):
        self.set_income(income)
        self.set_penalty(penalty)
        self.set_commission(commission)
        self.set_payment_fee_base(payment_fee_base)
        self.set_payment_fee_proportional(payment_fee_proportional)
        self.set_time_p(time_p)
        self.set_alpha_p(alpha_p)
        self.set_alpha_T(alpha_T)
        self.set_prob_cut(prob_cut)
        self.set_init_time_period(init_time_period)
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
        self.set_penalty(inlet['penalty'])
        self.set_commission(inlet['commission'])
        self.set_payment_fee_base(inlet['payment_fee_base'])
        self.set_payment_fee_proportional(inlet['payment_fee_proportional'])
        self.set_time_p(inlet['time_p'])
        self.set_alpha_p(inlet['alpha_p'])
        self.set_alpha_T(inlet['alpha_T'])
        self.set_prob_cut(inlet['prob_cut'])
        self.set_init_time_period(inlet['init_time_period'])
        self.set_mgt_period(inlet['mgt_period'])
        self.set_stat_period(inlet['stat_period'])
        self.set_init_mult(inlet['init_mult'])
        self.set_make_drawing(inlet['make_drawing'])
        self.set_output_statistics(inlet['output_statistics'])
        self.set_show_log(inlet['show_log'])

    def set_income(self, income):
        self.income = income

    def set_penalty(self, penalty):
        self.penalty = penalty

    def set_commission(self, commission):
        self.commission = commission

    def set_payment_fee_proportional(self, payment_fee_proportional):
        self.payment_fee_proportional = payment_fee_proportional

    def set_payment_fee_base(self, payment_fee_base):
        self.payment_fee_base = payment_fee_base

    def set_time_p(self, time_p):
        self.time_p = time_p

    def set_alpha_p(self, alpha_p):
        self.alpha_p = alpha_p

    def set_alpha_T(self, alpha_T):
        self.alpha_T = alpha_T

    def set_prob_cut(self, prob_cut):
        self.prob_cut = prob_cut

    def set_init_time_period(self, init_time_period):
        self.init_time_period = init_time_period

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
        out_str += 'penalty ' + str(self.penalty) + '\n'
        out_str += 'commission ' + str(self.commission) + '\n'
        out_str += 'payment_fee_proportional ' + str(
            self.payment_fee_proportional) + '\n'
        out_str += 'payment_fee_base ' + str(self.payment_fee_base) + '\n'
        out_str += 'time_p ' + str(self.time_p) + '\n'
        out_str += 'alpha_p ' + str(self.alpha_p) + '\n'
        out_str += 'alpha_T ' + str(self.alpha_T) + '\n'
        out_str += 'prob_cut ' + str(self.prob_cut) + '\n'
        out_str += 'init_time_period ' + str(self.init_time_period) + '\n'
        out_str += 'mgt_period ' + str(self.mgt_period) + '\n'
        out_str += 'stat_period ' + str(self.stat_period) + '\n'
        out_str += 'init_mult ' + str(self.init_mult) + '\n'
        out_str += 'make_drawing ' + str(self.make_drawing) + '\n'
        out_str += 'output_statistics ' + str(self.output_statistics) + '\n'
        out_str += 'show_log ' + str(self.show_log)
        return out_str
