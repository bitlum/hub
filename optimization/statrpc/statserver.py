from concurrent import futures
import time
import grpc
import json

import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))

import protobuffer.statrpc_pb2 as proto
import protobuffer.statrpc_pb2_grpc as proto_grpc
from statrpc.statsetts import StatSetts

_ONE_DAY_IN_SECONDS = 60 * 60 * 24


class SmartLog:
    def __init__(self):
        self.profit = float(42)


class GetStatistics(proto_grpc.GetStatisticsServicer):
    def __init__(self, smart_log):
        self.smart_log = smart_log
        self.init_time = time.time()

    def GetProfit(self, request, context):
        print('somebody is interested in profit')
        return proto.GetProfitResponse(profit=self.smart_log.profit)

    def GetTime(self, request, context):
        print('somebody is interested in time')
        return proto.GetTimeResponse(time=(time.time() - self.init_time))


def stat_serve(smart_log, setts):
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    proto_grpc.add_GetStatisticsServicer_to_server(
        GetStatistics(smart_log), server)
    server.add_insecure_port('[::]:' + setts.port)
    server.start()
    print('serve() is started')
    try:
        while True:
            time.sleep(_ONE_DAY_IN_SECONDS)
    except KeyboardInterrupt:
        server.stop(0)


if __name__ == '__main__':
    setts = StatSetts()
    setts.set_from_file(sys.argv[1])
    stat_serve(SmartLog(), setts)
    print('serve() is stoped')
