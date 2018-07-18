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
    response = stub.GetProfit(proto.GetProfitRequest())
    profit = response.profit
    response = stub.GetTime(proto.GetTimeRequest())
    time = response.time
    print("profit ", profit, "time ", time)


if __name__ == '__main__':

    with open(sys.argv[1]) as f:
        inlet = json.load(f)
    port = inlet['port']

    while True:
        time.sleep(1)
        run(port)
