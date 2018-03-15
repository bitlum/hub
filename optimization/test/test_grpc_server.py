from concurrent import futures
import time
from datetime import datetime

import grpc

import protobuf.test.test_grpc_pb2 as proto
import protobuf.test.test_grpc_pb2_grpc as proto_grpc

_ONE_DAY_IN_SECONDS = 60 * 60 * 24


class Timekeeper(proto_grpc.TimekeeperServicer):

    def SayTime(self, request, context):
        print(request.name + ' is interested in')
        return proto.TimeReply(time=request.name + ', time is ' + str(
            datetime.now().strftime('%Y-%m-%d %H:%M:%S')))


def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    proto_grpc.add_TimekeeperServicer_to_server(Timekeeper(), server)
    server.add_insecure_port('[::]:50051')
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
