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

        self.smart_log = smart_log
        self.log_reader = LogReader(file_name, self.smart_log)
        self.router_mgt = router_mgt
        self.mgt_freq = int(mgt_freq)
        self.mgt_ticker = int(0)
        self.channels_change = list()

    def process(self, event):
        if (event.event_type == 'modified') and (
                event.src_path == self.log_reader.file_name):
            self.log_reader.process_log()

            self.channels_change.clear()
            for user, wane in self.router_mgt.wanes_dict.items():
                balance = self.smart_log.router_balances[user]
                bound = self.router_mgt.bounds_dict[user]
                if wane:
                    if balance < bound:
                        self.channels_change.append(user)
                else:
                    if balance > bound:
                        self.channels_change.append(user)

            print('channels_change', self.channels_change)

            if self.mgt_ticker % self.mgt_freq == 0:
                self.router_mgt.calc_parameters()
            self.mgt_ticker += 1

    def on_modified(self, event):
        self.process(event)

    def on_created(self, event):
        self.process(event)

    def on_moved(self, event):
        self.process(event)

    def on_deleted(self, event):
        self.process(event)
