from __future__ import print_function
import time
import grpc

import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))

import protobuffer.statrpc_pb2 as proto
import protobuffer.statrpc_pb2_grpc as proto_grpc
from statrpc.statsetts import StatSetts


def run(port):
    channel = grpc.insecure_channel('localhost:' + port)
    stub = proto_grpc.GetStatisticsStub(channel)
    response = stub.GetStatParams(proto.GetStatParamsRequest())

    print('profit', response.profit)
    print('gain_sum', response.gain_sum)
    print('locked_balance', response.locked_balance)
    print('ROI_day', response.ROI_day)
    print('free_balance', response.free_balance)
    print('profit_av', response.profit_av)
    print('gain_sum_av', response.gain_sum_av)
    print('locked_balance_av', response.locked_balance_av)
    print('ROI_day_av', response.ROI_day_av)
    print('free_balance_av', response.free_balance_av)
    print()


if __name__ == '__main__':

    setts = StatSetts()
    setts.get_from_file(sys.argv[1])
    print(setts)

    while True:
        time.sleep(1)
        run(setts.port)
