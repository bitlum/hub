from __future__ import print_function

import grpc

import protobuf.hubrpc_pb2 as proto_hub
import protobuf.hubrpc_pb2_grpc as proto_rpc_hub

from test.test_router_rpc_client import ActivityGenerator


class OptimisationGenerator(ActivityGenerator):

    def __init__(self, users_number):
        super().__init__(users_number, 1, 2)

        self.channels = [proto_hub.Channel() for _ in
                         range(self.users_number)]

        for i in range(self.users_number):
            self.channels[i].user_id = self.users_id[i]
            self.channels[i].channel_id = self.users_id[i]
            self.channels[i].router_balance = 1

    def set_router_balances(self, ind_user, router_balance):
        self.channels[ind_user].router_balance = router_balance

    def set_state_request(self):
        proto_hub.Channel()
        request = proto_hub.SetStateRequest()

        request.time = 1

        for i in range(self.users_number):
            channels = request.channels.add()
            channels.user_id = self.channels[i].user_id
            channels.channel_id = self.channels[i].channel_id
            channels.router_balance = self.channels[i].router_balance

        print('state:', request, sep='\n')
        return request


if __name__ == '__main__':
    users_num = 3

    generator = OptimisationGenerator(users_num)

    channel = grpc.insecure_channel('localhost:8686')

    stub = proto_rpc_hub.ManagerStub(channel)

    generator.set_router_balances(ind_user=0, router_balance=1)
    generator.set_router_balances(ind_user=1, router_balance=1)
    generator.set_router_balances(ind_user=2, router_balance=1)

    stub.SetState(generator.set_state_request())
