import sys
import os
import configparser

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))


class StatSetts:

    def __init__(self):
        # Main settings:
        self.port = str()
        self.router_log = str()

        # Output settings:
        self.output_log = bool(False)
        self.serialized_log = str()

        # Time settings:
        self.average_period = float()
        self.velocity_period = float()
        self.output_period = float()
        self.acceleration = float(1)

        # Terminal settings:
        self.show_log = bool(False)
        self.show_pretty_log = bool(False)

    def get_from_file(self, file_name):
        config = configparser.ConfigParser()
        config.read(file_name)

        self.port = str(config.get('Main', 'port'))
        self.router_log = str(config.get('Main', 'router_log'))

        self.output_log = bool(config.get('Output', 'output_log'))
        self.serialized_log = str(config.get('Output', 'serialized_log'))

        self.average_period = float(config.get('Time', 'average_period'))
        self.velocity_period = float(config.get('Time', 'velocity_period'))
        self.output_period = float(config.get('Time', 'output_period'))
        self.acceleration = float(config.get('Time', 'acceleration'))

        self.show_log = bool(config.get('Terminal', 'show_log'))
        self.show_pretty_log = bool(config.get('Terminal', 'show_pretty_log'))

    def __str__(self):
        out_str = '\nRPC statistics settings\n'
        out_str += '\nMain:\n'
        out_str += 'port' + ' = ' + str(self.port) + '\n'
        out_str += 'router_log' + ' = ' + str(self.router_log) + '\n'

        out_str += '\nOutput:\n'
        out_str += 'output_log' + ' = ' + str(self.output_log) + '\n'
        out_str += 'serialized_log' + ' = ' + str(self.serialized_log) + '\n'

        out_str += '\nTime:\n'
        out_str += 'average_period' + ' = ' + str(self.average_period) + '\n'
        out_str += 'velocity_period' + ' = ' + str(self.velocity_period) + '\n'
        out_str += 'output_period' + ' = ' + str(self.output_period) + '\n'
        out_str += 'acceleration' + ' = ' + str(self.acceleration) + '\n'

        out_str += '\nTerminal:\n'
        out_str += 'show_log' + ' = ' + str(self.show_log) + '\n'
        out_str += 'show_pretty_log' + ' = ' + str(self.show_pretty_log) + '\n'

        return out_str


if __name__ == '__main__':
    stat_setts = StatSetts()
    stat_setts.get_from_file(sys.argv[1])
    print(stat_setts)
