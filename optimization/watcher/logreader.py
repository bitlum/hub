import datetime
import sys
import os
import json

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))

import protobuffer.log_pb2 as proto
from protobuffer.protobuf3_to_dict_patch import protobuf_to_dict
from watcher.protologutills import print_massege
from state.protolog import ProtoLog


class LogReader:

    def __init__(self, smart_log, setts):
        self.router_log = setts.router_log
        self.serialized_log = setts.serialized_log
        self.proto_log = ProtoLog()
        self.smart_log = smart_log
        self.pos = int(0)
        self.size_message = int(0)
        self.size_file = int(0)
        self.file = None
        self.message_names = ['payment', 'state', 'channel_change']
        self.show_log = setts.show_log
        self.show_pretty_log = setts.show_pretty_log
        self.output_log = setts.output_log

    def process_log(self):
        with open(self.router_log, "rb") as self.file:
            self.size_file = os.path.getsize(self.router_log)
            while True:
                if self.pos >= self.size_file:
                    break
                self.file.seek(self.pos)
                self.size_message = int.from_bytes(self.file.read(2),
                                                   byteorder='big',
                                                   signed=False)
                self.pos += 2
                self.proto_log.append(self.read_message())
                self.pos += self.size_message

                self.convert()

                if self.show_pretty_log:
                    print(self.smart_log)

    def read_message(self):
        log = proto.Log()
        self.file.seek(self.pos)
        log.ParseFromString(self.file.read(self.size_message))
        return log

    def convert(self):
        dict_massege = protobuf_to_dict(
            self.proto_log.messages[-1],
            use_enum_labels=True,
            including_default_value_fields=True)

        dict_massege['date'] = datetime.datetime.fromtimestamp(
            self.proto_log.messages[-1].time * 1e-9).__str__()

        dict_massege['message_type'] = 'unknown'
        for name in self.message_names:
            if self.proto_log.messages[-1].HasField(name):
                dict_massege['message_type'] = name

        self.smart_log.append(dict_massege)

        if self.show_log:
            print_massege(dict_massege)

    def out_log(self):
        if self.output_log:
            with open(self.serialized_log, 'w') as f:
                json.dump(self.smart_log.messages, f, sort_keys=True,
                          indent=4 * ' ')
