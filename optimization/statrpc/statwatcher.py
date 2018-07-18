from watchdog.events import PatternMatchingEventHandler as PattMatchEvHand
from watchdog.observers import Observer
import time
from threading import Thread
import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))

from watcher.protologutills import split_path_name
from samples.smartlog import SmartLog
from watcher.logreader import LogReader
from statrpc.statserver import stat_serve
from statrpc.statsetts import StatSetts


class StatServerThread(Thread):
    def __init__(self, smart_log, setts):
        Thread.__init__(self)
        self.smart_log = smart_log
        self.setts = setts

    def run(self):
        stat_serve(self.smart_log, self.setts)


class WatcherStat(PattMatchEvHand):

    def __init__(self, setts, log_file_name):
        super().__init__(
            patterns='*' + split_path_name(log_file_name)['name'],
            ignore_directories=True, case_sensitive=False)
        self.smart_log = SmartLog()
        self.log_reader = LogReader(log_file_name, self.smart_log, setts)
        self.stat_server_thread = StatServerThread(self.smart_log, setts)
        self.stat_server_thread.start()

    def process(self, event):
        if (event.event_type == 'modified') and (
                event.src_path == self.log_reader.file_name):
            self.log_reader.process_log()

    def on_modified(self, event):
        self.process(event)

    def on_created(self, event):
        self.process(event)

    def on_moved(self, event):
        self.process(event)

    def on_deleted(self, event):
        self.process(event)


def watch(log_file_name, inlet_file_name):
    path = split_path_name(log_file_name)['path']
    setts = StatSetts()
    setts.set_from_file(inlet_file_name)

    obser = Observer()
    obser.schedule(WatcherStat(setts, log_file_name), path)
    obser.start()
    try:
        while True:
            time.sleep(1)
    except KeyboardInterrupt:
        obser.stop()


if __name__ == '__main__':
    watch(log_file_name=sys.argv[1], inlet_file_name=sys.argv[2])
    print('watch() is stoped')
