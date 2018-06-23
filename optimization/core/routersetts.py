import sys
import os
import json

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))


class RouterSetts:

    def __init__(self):
        self.income = bool(True)
        self.blch_fee = float(1)
        self.pmnt_fee_base = int(1)
        self.pmnt_fee_prop = int(0)
        self.blch_period = float(0)
        self.idle_mult = float(1)
        self.period_mult = float(1)
        self.prob_cut = float(0.5)
        self.init_period = int()
        self.mgt_period = float(1)
        self.average_period = float(1)
        self.init_mult = float(1)
        self.make_drawing = bool(True)
        self.output_stat = bool(True)
        self.output_log = bool(True)
        self.output_period = float(1)
        self.show_log = bool(True)

    def set_from_file(self, file_name):
        with open(file_name) as f:
            inlet = json.load(f)
        self.income = inlet['income']
        self.blch_fee = inlet['blch_fee']
        self.pmnt_fee_base = inlet['pmnt_fee_base']
        self.pmnt_fee_prop = inlet['pmnt_fee_prop']
        self.blch_period = inlet['blch_period']
        self.idle_mult = inlet['idle_mult']
        self.period_mult = inlet['period_mult']
        self.prob_cut = inlet['prob_cut']
        self.init_period = inlet['init_period']
        self.mgt_period = inlet['mgt_period']
        self.average_period = inlet['average_period']
        self.init_mult = inlet['init_mult']
        self.make_drawing = inlet['make_drawing']
        self.output_stat = inlet['output_stat']
        self.output_log = inlet['output_log']
        self.output_period = inlet['output_period']
        self.show_log = inlet['show_log']

    def __str__(self):
        out_str = 'router settings:\n'
        out_str += 'income ' + str(self.income) + '\n'
        out_str += 'blch_fee ' + str(self.blch_fee) + '\n'
        out_str += 'pmnt_fee_prop ' + str(self.pmnt_fee_prop) + '\n'
        out_str += 'pmnt_fee_base ' + str(self.pmnt_fee_base) + '\n'
        out_str += 'blch_period ' + str(self.blch_period) + '\n'
        out_str += 'idle_mult ' + str(self.idle_mult) + '\n'
        out_str += 'period_mult ' + str(self.period_mult) + '\n'
        out_str += 'prob_cut ' + str(self.prob_cut) + '\n'
        out_str += 'init_period ' + str(self.init_period) + '\n'
        out_str += 'mgt_period ' + str(self.mgt_period) + '\n'
        out_str += 'average_period ' + str(self.average_period) + '\n'
        out_str += 'init_mult ' + str(self.init_mult) + '\n'
        out_str += 'make_drawing ' + str(self.make_drawing) + '\n'
        out_str += 'output_stat ' + str(self.output_stat) + '\n'
        out_str += 'show_log ' + str(self.show_log)
        return out_str
