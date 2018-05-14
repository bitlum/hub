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
        self.time_p = float(0)
        self.alpha_p = float(1)
        self.alpha_T = float(1)
        self.prob_cut = float(0.5)

    def set_settings(self, income,
                     penalty, commission,
                     time_p, alpha_p,
                     alpha_T, prob_cut=0.5):
        self.set_income(income)
        self.set_penalty(penalty)
        self.set_commission(commission)
        self.set_time_p(time_p)
        self.set_alpha_p(alpha_p)
        self.set_alpha_T(alpha_T)
        self.set_prob_cut(prob_cut)

    def set_setts_from_file(self, file_name):
        with open(file_name) as f:
            inlet = json.load(f)
        self.set_income(inlet['income'])
        self.set_penalty(inlet['penalty'])
        self.set_commission(inlet['commission'])
        self.set_time_p(inlet['time_p'])
        self.set_alpha_p(inlet['alpha_p'])
        self.set_alpha_T(inlet['alpha_T'])
        self.set_prob_cut(inlet['prob_cut'])

    def set_income(self, income):
        self.income = income

    def set_penalty(self, penalty):
        self.penalty = penalty

    def set_commission(self, commission):
        self.commission = commission

    def set_time_p(self, time_p):
        self.time_p = time_p

    def set_alpha_p(self, alpha_p):
        self.alpha_p = alpha_p

    def set_alpha_T(self, alpha_T):
        self.alpha_T = alpha_T

    def set_prob_cut(self, prob_cut):
        self.prob_cut = prob_cut

    def __str__(self):
        out_str = 'router settings:\n'
        out_str += 'income ' + str(self.income) + '\n'
        out_str += 'penalty ' + str(self.penalty) + '\n'
        out_str += 'commission ' + str(self.commission) + '\n'
        out_str += 'time_p ' + str(self.time_p) + '\n'
        out_str += 'alpha_p ' + str(self.alpha_p) + '\n'
        out_str += 'alpha_T ' + str(self.alpha_T) + '\n'
        out_str += 'prob_cut ' + str(self.prob_cut)
        return out_str
