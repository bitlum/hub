import sys
import os
import json

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))


class StatSetts:

    def __init__(self):
        self.port = str()
        self.average_period = float()
        self.velocity_period = float()
        self.output_log = bool(False)
        self.show_log = bool(False)
        self.show_pretty_log = bool(False)
        self.acceleration = float(1)

    def set_from_file(self, file_name):
        with open(file_name) as f:
            inlet = json.load(f)
        self.port = inlet['port']
        self.average_period = inlet['average_period']
        self.velocity_period = inlet['velocity_period']
        self.output_log = inlet['output_log']
        self.show_log = inlet['show_log']
        self.show_pretty_log = inlet['show_pretty_log']
        self.acceleration = inlet['acceleration']

    def __str__(self):
        out_str = 'statistics settings:\n'
        out_str += 'port ' + str(self.port) + '\n'
        out_str += 'average_period ' + str(self.average_period) + '\n'
        out_str += 'velocity_period ' + str(self.velocity_period) + '\n'
        out_str += 'output_log ' + str(self.output_log) + '\n'
        out_str += 'show_log ' + str(self.show_log) + '\n'
        out_str += 'show_pretty_log ' + str(self.show_pretty_log) + '\n'
        out_str += 'acceleration ' + str(self.acceleration) + '\n'
        return out_str


if __name__ == '__main__':
    stat_setts = StatSetts()
    stat_setts.set_from_file(sys.argv[1])
    print(stat_setts)
