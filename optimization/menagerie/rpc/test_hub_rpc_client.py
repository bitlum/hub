from __future__ import print_function
import time
import grpc
import datetime

import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../../'))

from menagerie.rpc.test_router_rpc_client import ActivityGenerator

import protobuffer.hubrpc_pb2 as proto_hub
import protobuffer.hubrpc_pb2_grpc as proto_rpc_hub


class OptimisationGenerator(ActivityGenerator):

    def __init__(self, users_number):
        super().__init__(users_number, 1, 2)
        self.chans_id = [str(i + 1) for i in range(self.users_number)]

    def set_update_link_request(self, ind_user, router_balance):
        request = proto_hub.UpdateLinkRequest()
        request.time = int(time.time() * 1E9)
        request.user_id = self.users_id[ind_user]
        request.router_balance = router_balance
        print('date: ', datetime.datetime.now())
        print('UpdateLinkRequest:', request, sep='\n')
        return request


if __name__ == '__main__':
    users_num = 2
    sleep_time = 1

    generator = OptimisationGenerator(users_num)

    channel = grpc.insecure_channel('localhost:8686')

    stub = proto_rpc_hub.ManagerStub(channel)

    for i in range(users_num):
        try:
            stub.UpdateLink(
                generator.set_update_link_request(ind_user=i,
                                                  router_balance=i + 3))
        except grpc._channel._Rendezvous as er:
            print(er, ' is skiped\n')
