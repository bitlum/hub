from concurrent import futures
import time

import grpc

import sys

sys.path.append('../../')

import protobuffer.rpc_pb2 as proto
import protobuffer.rpc_pb2_grpc as proto_rpc

_ONE_DAY_IN_SECONDS = 60 * 60 * 24


class Emulator(proto_rpc.EmulatorServicer):

    def SendPayment(self, request, context):
        print('send:', request, sep='\n')
        return proto.SendPaymentResponse()

    def OpenChannel(self, request, context):
        print('open:', request, sep='\n')
        response = proto.OpenChannelResponse()
        response.chan_id = request.user_id
        return response

    def CloseChannel(self, request, context):
        print('close:', request, sep='\n')
        return proto.CloseChannelResponse()


def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    proto_rpc.add_EmulatorServicer_to_server(Emulator(), server)
    server.add_insecure_port('[::]:9393')
    server.start()
    print('serve() is started')
    try:
        while True:
            time.sleep(_ONE_DAY_IN_SECONDS)
    except KeyboardInterrupt:
        server.stop(0)


if __name__ == '__main__':
    serve()
    print('serve() is stoped')
