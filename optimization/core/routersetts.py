import sys
import os
import configparser

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))


class RouterSetts:

    def __init__(self):
        # Main settings:
        self.router_log = str()

        # Output settings:
        self.output_log = bool(True)
        self.serialized_log = str()
        self.output_stat = bool(True)
        self.stat_file = str()
        self.make_drawing = bool(True)

        # Time settings:
        self.average_period = float(1)
        self.mgt_period = float(1)
        self.output_period = float(1)
        self.plot_period = float(1)
        self.acceleration = float(1)

        # Terminal settings:
        self.show_log = bool(True)
        self.show_pretty_log = bool(True)

        # InitLocking settings:
        self.init_period = int()
        self.init_mult = float(1)

        # Optimization settings:
        self.income = bool(True)
        self.blch_fee = int(1)
        self.pmnt_fee_base = int(1)
        self.pmnt_fee_prop = int(0)
        self.blch_period = float(0)
        self.idle_mult = float(1)
        self.period_mult = float(1)
        self.prob_cut = float(0.5)

    def get_from_file(self, file_name):
        config = configparser.ConfigParser()
        config.read(file_name)

        self.router_log = str(config.get('Main', 'router_log'))

        self.output_log = config.getboolean('Output', 'output_log')
        self.serialized_log = str(config.get('Output', 'serialized_log'))
        self.output_stat = config.getboolean('Output', 'output_stat')
        self.stat_file = str(config.get('Output', 'stat_file'))
        self.make_drawing = config.getboolean('Output', 'make_drawing')

        self.average_period = float(config.get('Time', 'average_period'))
        self.mgt_period = float(config.get('Time', 'mgt_period'))
        self.output_period = float(config.get('Time', 'output_period'))
        self.plot_period = float(config.get('Time', 'plot_period'))
        self.acceleration = float(config.get('Time', 'acceleration'))

        self.show_log = config.getboolean('Terminal', 'show_log')
        self.show_pretty_log = config.getboolean('Terminal', 'show_pretty_log')

        self.init_period = int(config.get('InitLocking', 'init_period'))
        self.init_mult = float(config.get('InitLocking', 'init_mult'))

        self.income = config.getboolean('Optimization', 'income')
        self.blch_fee = int(config.get('Optimization', 'blch_fee'))
        self.pmnt_fee_base = int(config.get('Optimization', 'pmnt_fee_base'))
        self.pmnt_fee_prop = int(config.get('Optimization', 'pmnt_fee_prop'))
        self.blch_period = float(config.get('Optimization', 'blch_period'))
        self.idle_mult = float(config.get('Optimization', 'idle_mult'))
        self.period_mult = float(config.get('Optimization', 'period_mult'))
        self.prob_cut = float(config.get('Optimization', 'prob_cut'))

    def __str__(self):
        out_str = '\nRouter settings\n'

        out_str += '\nMain:\n'
        out_str += 'router_log' + ' = ' + str(self.router_log) + '\n'

        out_str += '\nOutput:\n'
        out_str += 'output_log' + ' = ' + str(self.output_log) + '\n'
        out_str += 'serialized_log' + ' = ' + str(self.serialized_log) + '\n'
        out_str += 'output_stat' + ' = ' + str(self.output_stat) + '\n'
        out_str += 'stat_file' + ' = ' + str(self.stat_file) + '\n'
        out_str += 'make_drawing' + ' = ' + str(self.make_drawing) + '\n'

        out_str += '\nTime:\n'
        out_str += 'average_period' + ' = ' + str(self.average_period) + '\n'
        out_str += 'mgt_period' + ' = ' + str(self.mgt_period) + '\n'
        out_str += 'output_period' + ' = ' + str(self.output_period) + '\n'
        out_str += 'plot_period' + ' = ' + str(self.plot_period) + '\n'
        out_str += 'acceleration' + ' = ' + str(self.acceleration) + '\n'

        out_str += '\nTerminal:\n'
        out_str += 'show_log' + ' = ' + str(self.show_log) + '\n'
        out_str += 'show_pretty_log' + ' = ' + str(self.show_pretty_log) + '\n'

        out_str += '\nInitLocking:\n'
        out_str += 'init_period' + ' = ' + str(self.init_period) + '\n'
        out_str += 'init_mult' + ' = ' + str(self.init_mult) + '\n'

        out_str += '\nOptimization:\n'
        out_str += 'income' + ' = ' + str(self.income) + '\n'
        out_str += 'blch_fee' + ' = ' + str(self.blch_fee) + '\n'
        out_str += 'pmnt_fee_base' + ' = ' + str(self.pmnt_fee_base) + '\n'
        out_str += 'pmnt_fee_prop' + ' = ' + str(self.pmnt_fee_prop) + '\n'
        out_str += 'blch_period' + ' = ' + str(self.blch_period) + '\n'
        out_str += 'idle_mult' + ' = ' + str(self.idle_mult) + '\n'
        out_str += 'period_mult' + ' = ' + str(self.period_mult) + '\n'
        out_str += 'prob_cut' + ' = ' + str(self.prob_cut) + '\n'

        return out_str


if __name__ == '__main__':
    router_setts = RouterSetts()
    router_setts.get_from_file(sys.argv[1])
    print(router_setts)
