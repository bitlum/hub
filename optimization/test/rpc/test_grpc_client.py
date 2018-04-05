from __future__ import print_function
import time

import grpc

import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../../'))

import protobuffer.test.test_grpc_pb2 as proto
import protobuffer.test.test_grpc_pb2_grpc as proto_grpc


def run():
    channel = grpc.insecure_channel('localhost:50051')
    stub = proto_grpc.TimekeeperStub(channel)
    response = stub.SayTime(proto.TimeRequest(name='Aleks'))
    print("Timekeeper client received: " + response.time)


if __name__ == '__main__':

    while True:
        time.sleep(1)
        run()

