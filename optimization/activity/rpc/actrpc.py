from __future__ import print_function
import time
import json
from threading import Thread

import grpc

import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../../'))

import protobuffer.rpc_pb2 as proto
import protobuffer.rpc_pb2_grpc as proto_rpc


def set_duration(duration, stub):
    duration_request = proto.SetBlockGenDurationRequest()
    duration_request.duration = duration
    stub.SetBlockGenDuration(duration_request)


def open_channels(users_id, balances, stub):
    channels_id = dict()
    open_request = proto.OpenChannelRequest()
    for key, user_id in users_id.items():
        open_request.user_id = user_id
        open_request.locked_by_user = balances[key]
        open_response = stub.OpenChannel(open_request)
        channels_id[key] = open_response.channel_id
    return channels_id


def create_stub():
    rpc_channel = grpc.insecure_channel('localhost:9393')
    return proto_rpc.EmulatorStub(rpc_channel)


class TransactionThread(Thread):
    def __init__(self, stub, init_time, transaction):
        Thread.__init__(self)
        self.stub = stub
        self.init_time = init_time
        self.transaction = transaction
        self.request = proto.SendPaymentRequest()

    def run(self):
        time.sleep(self.transaction['time'] - self.init_time)
        self.request.sender = self.transaction['sender_id']
        self.request.receiver = self.transaction['receiver_id']
        self.request.amount = round(self.transaction['trans'])
        self.stub.SendPayment(self.request)


def actrpc_gen(file_name_inlet):
    with open(file_name_inlet) as f:
        inlet = json.load(f)

    duration = inlet['duration']

    with open(inlet['users_id_file_name']) as f:
        users_id = json.load(f)['users_id']

    with open(inlet['balances_file_name']) as f:
        balances = json.load(f)['balances']

    with open(inlet['transstream_file_name']) as f:
        transstream = json.load(f)['transstream']

    stub = create_stub()

    set_duration(duration, stub)

    channels_id = open_channels(users_id, balances, stub)

    # TransactionThread(stub, 0, transstream[0]).start()

    with open(inlet['channels_id_file_name'], 'w') as f:
        json.dump({'channels_id': channels_id}, f,
                  sort_keys=True, indent=4 * ' ')


if __name__ == '__main__':
    actrpc_gen(file_name_inlet='actrpc_inlet.json')
