import sys
import os
import json

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))


class UsersSample:
    def __init__(self, transseq):
        self.transseq = transseq
        self.users_id = dict()
        self.users_ind = dict()
        self.users_number = int()
        self.receiver_activity = dict()

    def calc_users_data(self):
        users_id_list = list()
        for i in range(len(self.transseq)):
            users_id_list.append(self.transseq[i]["payment"]['sender'])
            users_id_list.append(
                self.transseq[i]["payment"]['receiver'])
        users_id_list = sorted(list(set(users_id_list)))

        self.users_number = len(users_id_list)

        self.users_id.clear()
        self.users_id = {i: users_id_list[i] for i in
                         range(self.users_number)}

        self.users_ind.clear()
        self.users_ind = {users_id_list[i]: i for i in
                          range(self.users_number)}

        self.receiver_activity.clear()
        for _, user in self.users_id.items():
            self.receiver_activity[user] = int(0)

        for trans in self.transseq:
            self.receiver_activity[trans['payment']['receiver']] += 1


if __name__ == '__main__':
    with open('../activity/outlet/transseq.json') as f:
        transseq = json.load(f)['transseq']

    trans_stat = UsersSample(transseq)
    trans_stat.calc_users_data()

    print('users_number ', trans_stat.users_number)
    print('users_id ', trans_stat.users_id)
    print('users_ind ', trans_stat.users_ind)
    print()
