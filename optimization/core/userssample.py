import sys
import os
import json

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))

from core.routersetts import RouterSetts


class UsersSample:
    def __init__(self, transseq, setts):
        self.transseq = transseq
        self.setts = setts
        self.users_id = dict()
        self.users_ind = dict()
        self.users_number = int()

    def calc_users_data(self):

        users_id_list = list()
        for i in range(len(self.transseq)):
            sender = self.transseq[i]['payment']['sender']
            receiver = self.transseq[i]['payment']['receiver']
            users_id_list.append(sender)
            users_id_list.append(receiver)
        users_id_list = sorted(list(set(users_id_list)))

        self.users_number = len(users_id_list)

        self.users_id.clear()
        self.users_id = {i: users_id_list[i] for i in
                         range(self.users_number)}

        self.users_ind.clear()
        self.users_ind = {users_id_list[i]: i for i in
                          range(self.users_number)}

    def accelerate_transseq(self):
        for i in range(len(self.transseq)):
            self.transseq[i]['time'] /= self.setts.acceleration


if __name__ == '__main__':
    router_setts = RouterSetts()
    router_setts.set_from_file('../optimizer/routermgt_inlet.json')

    with open('../activity/outlet/transseq.json') as f:
        transseq = json.load(f)['transseq']

    trans_stat = UsersSample(transseq, router_setts)
    trans_stat.accelerate_transseq()
    trans_stat.calc_users_data()

    print('users_number ', trans_stat.users_number)
    print('users_id ', trans_stat.users_id)
    print('users_ind ', trans_stat.users_ind)
    print()
