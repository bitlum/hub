import sys
import os
import json

from core.transstat import TransStat

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))


class FlowStat(TransStat):

    def __init__(self, transstream):
        super().__init__(transstream)
        self.flowmatr = list()
        self.flowvect_out = list()
        self.flowvect_in = list()

    def calc_flow(self, prob_cut=0.5):
        self.calc_stat(prob_cut)
        self.flowmatr.clear()

        self.flowmatr = [[amount.cut for amount in amount_vect] for amount_vect
                         in self.smart_amount]

        for i in range(self.users_number):
            for j in range(self.users_number):
                if self.flowmatr[i][j] is not None:
                    self.flowmatr[i][j] /= self.smart_period[i][j].cut

        self.flowvect_out.clear()
        self.flowvect_out = [float() for _ in range(self.users_number)]

        self.flowvect_in.clear()
        self.flowvect_in = [float() for _ in range(self.users_number)]

        for i in range(self.users_number):
            for j in range(self.users_number):
                value = self.flowmatr[i][j]
                if value is not None:
                    self.flowvect_out[i] += value
                    self.flowvect_in[j] += value


if __name__ == '__main__':
    file_inlet = 'inlet/actmatr_from_stream_inlet.json'

    with open(file_inlet) as f:
        inlet = json.load(f)

    with open(inlet['transstream_file_name']) as f:
        transstream = json.load(f)['transstream']

    prob_cut = 0.5

    flow_stat = FlowStat(transstream)
    flow_stat.calc_flow(prob_cut)

    print('flowmatr:')
    print(flow_stat.flowmatr)
    print()

    print('flowvect_out:')
    print(flow_stat.flowvect_out)
    print()

    print('flowvect_in:')
    print(flow_stat.flowvect_in)
    print()
