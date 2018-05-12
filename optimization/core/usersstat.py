import sys
import os
import json

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))


class UsersStat:
    def __init__(self, transstream):
        self.transstream = transstream
        self.users_id = dict()
        self.users_ind = dict()
        self.users_number = int()

    def calc_users_data(self):
        users_id_list = []
        for i in range(len(self.transstream)):
            users_id_list.append(self.transstream[i]["payment"]['sender'])
            users_id_list.append(
                self.transstream[i]["payment"]['receiver'])
        users_id_list = sorted(list(set(users_id_list)))

        self.users_number = len(users_id_list)

        self.users_id.clear()
        self.users_id = {i: users_id_list[i] for i in
                         range(self.users_number)}

        self.users_ind.clear()
        self.users_ind = {users_id_list[i]: i for i in
                          range(self.users_number)}


if __name__ == '__main__':
    file_inlet = 'inlet/actmatr_from_stream_inlet.json'

    with open(file_inlet) as f:
        inlet = json.load(f)

    with open(inlet['transstream_file_name']) as f:
        transstream = json.load(f)['transstream']

    trans_stat = UsersStat(transstream)
    trans_stat.calc_users_data()

    print('users_number ', trans_stat.users_number)
    print('users_id ', trans_stat.users_id)
    print('users_ind ', trans_stat.users_ind)
    print()


