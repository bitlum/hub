from __future__ import print_function
import time
import json
import grpc
import datetime

import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../../'))

from menagerie.rpc.test_router_rpc_client import ActivityGenerator

import protobuffer.hubrpc_pb2 as proto_hub
import protobuffer.hubrpc_pb2_grpc as proto_rpc_hub


def update_link(user_id, stub, router_balance):
    request_update = proto_hub.UpdateLinkRequest()
    request_update.time = int(time.time() * 1E9)
    request_update.user_id = user_id
    request_update.router_balance = router_balance
    stub.UpdateLink(request_update)


def acthubrpc_gen(file_name_inlet):
    with open(file_name_inlet) as f:
        inlet = json.load(f)

    with open(inlet['users_id_file_name']) as f:
        users_id = json.load(f)['users_id']

    with open(inlet['channels_id_file_name']) as f:
        channels_id = json.load(f)['channels_id']

    rpc_channel = grpc.insecure_channel('localhost:8686')

    stub = proto_rpc_hub.ManagerStub(rpc_channel)

    for _, user_id in users_id.items():
        update_link(user_id, stub, 1)


if __name__ == '__main__':
    acthubrpc_gen(file_name_inlet='acthubrpc_inlet.json')
