import os
from watchdog.events import PatternMatchingEventHandler

import sys

sys.path.append('../')

import protobuf.test.test_pb2 as proto


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
        event = proto.Event()
        self.file.seek(self.pos_cur)
        event.ParseFromString(self.file.read(self.size_message_cur))
        return event

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
                print(self.smart_log)


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
