from watchdog.events import PatternMatchingEventHandler
import datetime
from sortedcontainers import SortedDict

import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))

import protobuffer.log_pb2 as proto

from protobuffer.protobuf3_to_dict_patch import protobuf_to_dict


def print_massege(massege, nesting=-4):
    if type(massege) == dict:
        print('')
        nesting += 4

        keys = list(SortedDict(massege).keys())
        for key in keys:
            if key == 'time':
                keys.remove(key)
                keys.insert(0, key)
        for key in keys:
            if key == 'date':
                keys.remove(key)
                keys.insert(0, key)

        for key in keys:
            print(nesting * ' ', end='')
            print(key, end=': ')

            if massege[key] == 0:
                if key == 'type':
                    massege[key] = 'openning'
                elif key == 'status':
                    massege[key] = 'null'

            print_massege(massege[key], nesting)
    elif type(massege) == list:
        print('')
        nesting += 4
        for ind in range(len(massege)):
            print(nesting * ' ', end='')
            print(ind, end=': ')
            print_massege(massege[ind], nesting)
    else:
        print(massege)


class WatchLogRead(PatternMatchingEventHandler):

    def __init__(self, file_name, smart_log):
        super().__init__(patterns='*' + split_path_name(file_name)['name'],
                         ignore_directories=True, case_sensitive=False)
        self.file_name = file_name
        self.smart_log = smart_log
        self.pos_cur = 0
        self.size_message_cur = 0
        self.size_file = 0
        self.file = None

    def process(self, event):
        # print(event.event_type, datetime.datetime.now())
        # print(event.event_type,
        #       datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S'))
        # print(event.event_type, time.time())
        if (event.event_type == 'modified') and (
                event.src_path == self.file_name):
            self.read_new_messages()

    def on_modified(self, event):
        self.process(event)

    def on_created(self, event):
        self.process(event)

    def on_moved(self, event):
        self.process(event)

    def on_deleted(self, event):
        self.process(event)

    def read_message(self):
        log = proto.Log()
        self.file.seek(self.pos_cur)
        log.ParseFromString(self.file.read(self.size_message_cur))
        return log

    def read_new_messages(self):
        with open(self.file_name, "rb") as self.file:
            self.size_file = os.path.getsize(self.file_name)
            while True:
                if self.pos_cur >= self.size_file:
                    break
                self.file.seek(self.pos_cur)
                self.size_message_cur = int.from_bytes(self.file.read(2),
                                                       byteorder='big',
                                                       signed=False)
                self.pos_cur += 2
                self.smart_log.append(self.read_message())
                self.pos_cur += self.size_message_cur

                # TODO handle defaults:

                dict_massege = protobuf_to_dict(
                    self.smart_log.messages[-1],
                    use_enum_labels=True,
                    including_default_value_fields=True)
                dict_massege['date'] = datetime.datetime.fromtimestamp(
                    self.smart_log.messages[-1].time * 1e-9).__str__()
                print_massege(dict_massege)


def split_path_name(file_name):
    split = 0
    for i in range(len(file_name)):
        if file_name[-1 - i] == '/':
            split = - i
            break
    if split == 0:
        return {'path': './', 'name': file_name}
    else:
        return {'path': file_name[:split], 'name': file_name[split:]}
