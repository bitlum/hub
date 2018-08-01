from concurrent import futures
import time
import grpc

import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))

import protobuffer.statrpc_pb2 as proto
import protobuffer.statrpc_pb2_grpc as proto_grpc

_ONE_DAY_IN_SECONDS = 60 * 60 * 24


class StatParamsTmp:
    def __init__(self):
        self.profit = [float(42)]


class GetStatistics(proto_grpc.GetStatisticsServicer):
    def __init__(self, stat_params):
        self.stat_params = stat_params

    def GetStatParams(self, request, context):
        response = proto.GetStatParamsResponse()
        response.profit = self.stat_params.profit[-1]
        response.gain_sum = self.stat_params.gain_sum[-1]
        response.locked_balance = self.stat_params.locked_balance[-1]
        response.ROI_day = self.stat_params.ROI_day[-1]
        response.free_balance = self.stat_params.free_balance[-1]
        response.profit_av = self.stat_params.profit_av
        response.gain_sum_av = self.stat_params.gain_sum_av
        response.locked_balance_av = self.stat_params.locked_balance_av
        response.ROI_day_av = self.stat_params.ROI_day_av
        response.free_balance_av = self.stat_params.free_balance_av
        return response


def stat_serve(stat_params, setts):
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    proto_grpc.add_GetStatisticsServicer_to_server(
        GetStatistics(stat_params), server)
    server.add_insecure_port('[::]:' + setts.port)
    server.start()
    print('serve() is started')
    try:
        while True:
            time.sleep(_ONE_DAY_IN_SECONDS)
    except KeyboardInterrupt:
        server.stop(0)
