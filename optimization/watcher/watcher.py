from watchdog.events import PatternMatchingEventHandler as PattMatchEvHand
import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))

from watcher.protologutills import split_path_name
from watcher.logreader import LogReader


class Watcher(PattMatchEvHand):

    def __init__(self, file_name, smart_log, router_mgt, mgt_freq=2):
        super().__init__(patterns='*' + split_path_name(file_name)['name'],
                         ignore_directories=True, case_sensitive=False)

        self.log_reader = LogReader(file_name, smart_log)
        self.router_mgt = router_mgt
        self.mgt_freq = int(mgt_freq)
        self.mgt_ticker = int(0)

    def process(self, event):
        if (event.event_type == 'modified') and (
                event.src_path == self.log_reader.file_name):
            self.log_reader.process_log()
            if self.mgt_ticker % self.mgt_freq == 0:
                self.router_mgt.calc_parameters()
                print('balances ', self.router_mgt.balances)
                print('freqs ', self.router_mgt.freqs)
                print('wanes ', self.router_mgt.wanes)
            self.mgt_ticker += 1

    def on_modified(self, event):
        self.process(event)

    def on_created(self, event):
        self.process(event)

    def on_moved(self, event):
        self.process(event)

    def on_deleted(self, event):
        self.process(event)
