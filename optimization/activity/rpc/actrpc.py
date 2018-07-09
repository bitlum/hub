import time
import json
from threading import Thread
import threading
import grpc
import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../../'))

import protobuffer.rpc_pb2 as proto
import protobuffer.rpc_pb2_grpc as proto_rpc
from core.routersetts import RouterSetts


def set_blch_period(blch_period, stub):
    request = proto.SetBlockGenDurationRequest()
    request.duration = blch_period
    stub.SetBlockGenDuration(request)


def open_channels(users_id, balances, stub):
    channels_id = dict()
    open_request = proto.OpenChannelRequest()
    for key, user_id in users_id.items():
        open_request.user_id = user_id
        open_request.locked_by_user = balances[key]
        open_response = stub.OpenChannel(open_request)
        channels_id[key] = open_response.channel_id
    return channels_id


def set_blch_fee(fee, stub):
    request = proto.SetBlockchainFeeRequest()
    request.fee = fee
    stub.SetBlockchainFee(request)


def create_stub():
    rpc_channel = grpc.insecure_channel('localhost:9393')
    return proto_rpc.EmulatorStub(rpc_channel)


class TransactionThread(Thread):
    def __init__(self, stub, time_shift, transaction, acceleration, id,
                 time_list, pmt_list):
        Thread.__init__(self)
        self.time_list = time_list
        self.pmt_list = pmt_list

        self.stub = stub
        self.time_shift = time_shift
        self.transaction = transaction
        self.acceleration = acceleration
        self.id = id
        self.request = proto.SendPaymentRequest()

    def run(self):
        amount = round(self.transaction['payment']['amount'])
        if amount > 0:
            period = 1.E-9 * self.transaction['time'] - self.time_shift
            period /= self.acceleration
            time.sleep(period)
            sender = self.transaction['payment']['sender']
            receiver = self.transaction['payment']['receiver']
            self.request.sender = sender
            self.request.receiver = receiver
            self.request.amount = amount
            self.request.id = str(self.id)

            self.time_list.append(time.time())
            self.pmt_list.append(str(self.id))

            try:
                self.stub.SendPayment(self.request)
            except Exception as er:
                print(er, 'is skipped for transaction', sender, '->', receiver)
            print(self.request)


def sent_transactions(stub, transseq, thread_limit, acceleration):
    time_init = time.time()
    i = 0

    time_list = list()
    pmt_list = list()

    flag = True

    while i < len(transseq):
        if len(time_list) > 1000 and flag is True:
            flag = False

            time_delta = list()
            for i in range(len(time_list)):
                time_delta.append(time_list[i] - time_list[0])

            for i in range(len(time_list)):
                pmt_list[i] += ' :: ' + str(time_delta[i])

            with open('pmt_list.json', 'w') as f:
                json.dump(pmt_list, f, sort_keys=True, indent=4 * ' ')

        if threading.active_count() < thread_limit:
            TransactionThread(stub, time.time() - time_init,
                              transseq[i], acceleration, i, time_list,
                              pmt_list).start()
            i += 1


def sent_transactions_new(stub, transseq, acceleration):
    time_init = time.time()

    time_list = list()
    pmt_list = list()

    flag = True
    trans_id = 0

    for transaction in transseq:
        amount = round(transaction['payment']['amount'])
        if amount > 0:

            sender = transaction['payment']['sender']
            receiver = transaction['payment']['receiver']
            trans_id += 1

            request = proto.SendPaymentRequest()
            request.sender = sender
            request.receiver = receiver
            request.amount = amount
            request.id = str(trans_id)

            time_shift = time.time() - time_init
            period = 1.E-9 * transaction['time'] - time_shift
            period /= acceleration
            time.sleep(period)

            time_list.append(time.time())
            pmt_list.append(str(trans_id))

            try:
                stub.SendPayment(request)
            except Exception as er:
                print(er, 'is skipped for transaction', sender, '->', receiver)
            print(request)

            if len(time_list) > 500 and flag is True:
                flag = False

                time_delta = list()
                for i in range(len(time_list)):
                    time_delta.append(time_list[i] - time_list[0])

                for i in range(len(time_list)):
                    pmt_list[i] += ' :: ' + str(time_delta[i])

                with open('pmt_list.json', 'w') as f:
                    json.dump(pmt_list, f, sort_keys=True, indent=4 * ' ')


def actrpc_gen(file_name_inlet):
    with open(file_name_inlet) as f:
        inlet = json.load(f)

    router_setts = RouterSetts()
    router_setts.set_from_file('../../optimizer/routermgt_inlet.json')
    blch_period = int(
        router_setts.blch_period * 1E+3 / router_setts.acceleration)
    blch_fee = router_setts.blch_fee

    with open(inlet['users_id_file_name']) as f:
        users_id = json.load(f)['users_id']

    del users_id['0']

    with open(inlet['user_balances_file_name']) as f:
        user_balances = json.load(f)['balances']

    with open(inlet['transseq_file_name']) as f:
        transseq = json.load(f)['transseq']

    time.sleep(1)

    stub = create_stub()

    set_blch_period(blch_period, stub)

    set_blch_fee(blch_fee, stub)

    channels_id = open_channels(users_id, user_balances, stub)

    print(3 + 2 * router_setts.blch_period / router_setts.acceleration * 1.1)

    time.sleep(3 + 2 * router_setts.blch_period / router_setts.acceleration * 1.1)

    # thread_limit = inlet['thread_limit']
    # sent_transactions(stub, transseq, thread_limit, router_setts.acceleration)

    sent_transactions_new(stub, transseq, router_setts.acceleration)

    with open(inlet['channels_id_file_name'], 'w') as f:
        json.dump({'channels_id': channels_id}, f,
                  sort_keys=True, indent=4 * ' ')


if __name__ == '__main__':
    actrpc_gen(file_name_inlet='actrpc_inlet.json')
