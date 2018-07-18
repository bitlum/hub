from __future__ import print_function
import time
import grpc
import json

import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))

import protobuffer.statrpc_pb2 as proto
import protobuffer.statrpc_pb2_grpc as proto_grpc


def run(port):
    channel = grpc.insecure_channel('localhost:' + port)
    stub = proto_grpc.GetStatisticsStub(channel)
    response = stub.GetStatParams(proto.GetStatParamsRequest())

    print('profit', response.profit)
    print('income', response.income)
    print('locked_balance', response.locked_balance)
    print('ROI_day', response.ROI_day)
    print('free_balance', response.free_balance)
    print('profit_av', response.profit_av)
    print('income_av', response.income_av)
    print('locked_balance_av', response.locked_balance_av)
    print('ROI_day_av', response.ROI_day_av)
    print('free_balance_av', response.free_balance_av)
    print()


if __name__ == '__main__':

    with open(sys.argv[1]) as f:
        inlet = json.load(f)
    port = inlet['port']

    while True:
        time.sleep(1)
        run(port)
