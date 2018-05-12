import sys
import os
import copy
import json

from core.userssample import UsersSample

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))


class TransSample(UsersSample):

    def __init__(self, transstream):
        super().__init__(transstream)
        self.timematr = list()
        self.transmatr = list()
        self.periodmatr = list()
        self.trans_number = int()

    def calc_matr_data(self):

        self.timematr.clear()
        self.transmatr.clear()
        self.periodmatr.clear()

        self.trans_number = len(self.transstream)

        self.timematr = [[list() for _ in range(self.users_number)] for _ in
                         range(self.users_number)]

        self.transmatr = [[list() for _ in range(self.users_number)] for _ in
                          range(self.users_number)]

        for i in range(len(self.transstream)):
            payment = self.transstream[i]["payment"]
            sender = self.users_ind[payment['sender']]
            receiver = self.users_ind[payment['receiver']]

            delta = self.transstream[i]['time'] - self.transstream[0]['time']
            self.timematr[sender][receiver].append(1.E-9 * delta)

            amount = payment['amount']
            self.transmatr[sender][receiver].append(amount)

        self.periodmatr = copy.deepcopy(self.timematr)

        for sender in self.periodmatr:
            for payments in sender:
                if payments is not None:
                    for i in range(len(payments) - 1, 0, -1):
                        payments[i] -= payments[i - 1]

    def calc_data(self):
        self.calc_users_data()
        self.calc_matr_data()


if __name__ == '__main__':
    file_inlet = 'inlet/actmatr_from_stream_inlet.json'

    with open(file_inlet) as f:
        inlet = json.load(f)

    with open(inlet['transstream_file_name']) as f:
        transstream = json.load(f)['transstream']

    trans_sample = TransSample(transstream)
    trans_sample.calc_data()

    print('users_number ', trans_sample.users_number)
    print('users_id ', trans_sample.users_id)
    print('users_ind ', trans_sample.users_ind)
    print()

    print('trans_number ', trans_sample.trans_number)
    print('transmatr ', trans_sample.transmatr)
    print('timematr ', trans_sample.timematr)
    print('periodmatr ', trans_sample.periodmatr)
    print()
