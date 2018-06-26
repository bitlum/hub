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
        self.single_trans = set()

    def calc_users_data(self):

        self.count_single_trans()

        users_id_list = list()
        for i in range(len(self.transseq)):
            if i not in self.single_trans:
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

    def count_single_trans(self):

        users_mention = dict()
        for i in range(len(self.transseq)):
            sender = self.transseq[i]['payment']['sender']
            if sender in users_mention:
                users_mention[sender] += 1
            else:
                users_mention[sender] = 1

            receiver = self.transseq[i]['payment']['receiver']
            if receiver in users_mention:
                users_mention[receiver] += 1
            else:
                users_mention[receiver] = 1

        self.single_trans.clear()
        for i in range(len(self.transseq)):
            sender = self.transseq[i]['payment']['sender']
            receiver = self.transseq[i]['payment']['receiver']
            if users_mention[sender] == 1 or users_mention[receiver] == 1:
                self.single_trans.add(i)


if __name__ == '__main__':
    with open('../activity/outlet/transseq.json') as f:
        transseq = json.load(f)['transseq']

    trans_stat = UsersSample(transseq)
    trans_stat.calc_users_data()

    print('users_number ', trans_stat.users_number)
    print('users_id ', trans_stat.users_id)
    print('users_ind ', trans_stat.users_ind)
    print()
