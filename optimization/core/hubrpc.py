import time
import grpc
import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))

import protobuffer.hubrpc_pb2 as proto_hub
import protobuffer.hubrpc_pb2_grpc as proto_rpc_hub


class HubRPC:
    def __init__(self, balances, update_set):
        self.balances = balances
        self.update_set = update_set
        self.stub = None
        self.create_stub()

    def create_stub(self):
        rpc_channel = grpc.insecure_channel('localhost:8686')
        self.stub = proto_rpc_hub.ManagerStub(rpc_channel)

    def update_link(self, user_id, balance):
        request_update = proto_hub.UpdateLinkRequest()
        request_update.time = int(time.time() * 1E9)
        request_update.user_id = user_id
        request_update.router_balance = balance
        self.stub.UpdateLink(request_update)

    def update(self):
        for user in self.update_set:
            try:
                self.update_link(user, int(self.balances[user]))
            except Exception as er:
                print(er, 'is skipped')
