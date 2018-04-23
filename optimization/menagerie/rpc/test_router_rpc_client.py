from __future__ import print_function
import time
import random

import grpc

import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../../'))

import protobuffer.rpc_pb2 as proto
import protobuffer.rpc_pb2_grpc as proto_rpc


class ActivityGenerator:

    def __init__(self, users_number, min_amount, max_amount):
        self.users_number = users_number
        self.users_id = [str(i + 1) for i in range(0, users_number)]
        # TODO such chans_id is tmp
        self.chans_id = [str(i + 1) for i in range(0, users_number)]
        self.min_amount = min_amount
        self.max_amount = max_amount
        self.ind_sender = 0
        self.ind_receiver = 0
        self.ind_user = 0
        self.ind_chan = 0

    def send_payment_request(self):
        self.ind_sender = random.randint(0, self.users_number - 1)
        shift = random.randint(1, self.users_number - 1)
        self.ind_receiver = (self.ind_sender + shift) % self.users_number

        request = proto.SendPaymentRequest()
        request.sender = self.users_id[self.ind_sender]
        request.receiver = self.users_id[self.ind_receiver]
        request.amount = random.randint(self.min_amount, self.max_amount)
        print('send:', request, sep='\n')
        return request

    def open_channel_request(self, ind_user):
        self.ind_user = ind_user
        request = proto.OpenChannelRequest()
        request.user_id = self.users_id[self.ind_user]
        request.locked_by_user = 7
        print('open:', request, sep='\n')
        return request

    def set_channel_id(self, ind_user, response_):
        self.chans_id[ind_user] = response_.channel_id
        print('channel_id for user_id ', self.users_id[ind_user], ' is set ',
              self.chans_id[ind_user], '\n')

    def close_channel_request(self, ind_user):
        self.ind_user = ind_user
        request = proto.CloseChannelRequest()
        request.chan_id = self.chans_id[self.ind_user]
        print('close:', request, sep='\n')
        return request

    @staticmethod
    def set_block_gen_duration_request(duration):
        request = proto.SetBlockGenDurationRequest()
        request.duration = int(duration * 1E+6)
        print('SetBlockGenDurationRequest:', request, sep='\n')
        return request


if __name__ == '__main__':
    users_num = 2
    trans_num = 1
    min_am = 1
    max_am = 2

    sleep_time = 1

    generator = ActivityGenerator(users_num, min_am, max_am)

    channel = grpc.insecure_channel('localhost:9393')

    stub = proto_rpc.EmulatorStub(channel)

    for user_ind in range(users_num):
        time.sleep(sleep_time)
        response = stub.OpenChannel(generator.open_channel_request(user_ind))
        generator.set_channel_id(user_ind, response)

    stub.SetBlockGenDuration(
        generator.set_block_gen_duration_request(duration=300))
    # 
    #
    # for _ in range(trans_num):
    #     time.sleep(sleep_time)
    #     stub.SendPayment(generator.send_payment_request())
    #
    # for ind in range(users_num):
    #     time.sleep(sleep_time)
    #     stub.CloseChannel(generator.close_channel_request(ind))
    #
    # for ind in range(users_num):
    #     time.sleep(sleep_time)
    #     response = stub.OpenChannel(generator.open_channel_request(ind))
    #     generator.set_channel_id(ind, response)
