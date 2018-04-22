# Generated by the gRPC Python protocol compiler plugin. DO NOT EDIT!
import grpc

import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../../'))

import protobuffer.menagerie.test_grpc_pb2 as test__grpc__pb2


class TimekeeperStub(object):
  # missing associated documentation comment in .proto file
  pass

  def __init__(self, channel):
    """Constructor.

    Args:
      channel: A grpc.Channel.
    """
    self.SayTime = channel.unary_unary(
        '/whattime.Timekeeper/SayTime',
        request_serializer=test__grpc__pb2.TimeRequest.SerializeToString,
        response_deserializer=test__grpc__pb2.TimeReply.FromString,
        )


class TimekeeperServicer(object):
  # missing associated documentation comment in .proto file
  pass

  def SayTime(self, request, context):
    # missing associated documentation comment in .proto file
    pass
    context.set_code(grpc.StatusCode.UNIMPLEMENTED)
    context.set_details('Method not implemented!')
    raise NotImplementedError('Method not implemented!')


def add_TimekeeperServicer_to_server(servicer, server):
  rpc_method_handlers = {
      'SayTime': grpc.unary_unary_rpc_method_handler(
          servicer.SayTime,
          request_deserializer=test__grpc__pb2.TimeRequest.FromString,
          response_serializer=test__grpc__pb2.TimeReply.SerializeToString,
      ),
  }
  generic_handler = grpc.method_handlers_generic_handler(
      'whattime.Timekeeper', rpc_method_handlers)
  server.add_generic_rpc_handlers((generic_handler,))