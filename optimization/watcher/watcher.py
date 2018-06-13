from watchdog.events import PatternMatchingEventHandler as PattMatchEvHand
import sys
import os
import time

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))

from watcher.protologutills import split_path_name
from samples.smartlog import SmartLog
from watcher.logreader import LogReader
from core.hubrpc import HubRPC
from core.routermgt import RouterMgt
from watcher.routermetrics import RouterMetrics


class Watcher(PattMatchEvHand):

    def __init__(self, router_setts):
        super().__init__(
            patterns='*' + split_path_name(router_setts.log_file_name)['name'],
            ignore_directories=True, case_sensitive=False)

        self.smart_log = SmartLog()
        self.log_reader = LogReader(router_setts.log_file_name, self.smart_log)
        self.router_mgt = RouterMgt(self.smart_log.transseq, router_setts)
        self.hubrpc = HubRPC(self.router_mgt.balances, router_setts)
        self.router_metrics = RouterMetrics(self.smart_log, router_setts)

        self.init_locked_funds = router_setts.init_locked_funds
        self.init_time_period = router_setts.init_time_period

        self.mgt_period = router_setts.mgt_period
        self.time_mgt_start = time.time()

    def process(self, event):
        if (event.event_type == 'modified') and (
                event.src_path == self.log_reader.file_name):

            self.log_reader.process_log()

            if time.time() - self.time_mgt_start >= self.mgt_period:
                self.router_metrics.process()
                self.router_mgt.calc_parameters()
                self.calc_update_set()
                self.set_init_update()
                self.hubrpc.update()
                self.time_mgt_start = time.time()

    def on_modified(self, event):
        self.process(event)

    def on_created(self, event):
        self.process(event)

    def on_moved(self, event):
        self.process(event)

    def on_deleted(self, event):
        self.process(event)

    def calc_update_set(self):
        self.hubrpc.update_set.clear()
        for user, wane in self.router_mgt.wanes.items():
            balance_cur = 0
            if user in self.smart_log.router_balances:
                balance_cur = self.smart_log.router_balances[user]
            bound = self.router_mgt.bounds[user]
            balance_opt = self.router_mgt.balances[user]
            if wane:
                if balance_cur < bound:
                    self.hubrpc.update_set.add(user)
            else:
                if balance_cur > bound or balance_cur < balance_opt:
                    self.hubrpc.update_set.add(user)

        print('update_set', self.hubrpc.update_set)
        for user in self.smart_log.blockage_set:
            self.hubrpc.update_set.discard(user)
        print('update_set', self.hubrpc.update_set)
        for user in self.smart_log.closure_set:
            self.hubrpc.update_set.discard(user)
        print('update_set', self.hubrpc.update_set)

        discard_newbie_set = set()
        for user in self.smart_log.newbie_set:
            period = time.time() - self.smart_log.open_time[user]
            if period > self.init_time_period:
                discard_newbie_set.add(user)
        for user in discard_newbie_set:
            self.smart_log.newbie_set.discard(user)

        for user in self.smart_log.newbie_set:
            self.hubrpc.update_set.discard(user)
        print('update_set', self.hubrpc.update_set)
        print()

    def set_init_update(self):
        for user in self.smart_log.just_opened_set:
            self.hubrpc.update_set.add(user)
            self.router_mgt.balances[user] = self.init_locked_funds
        self.smart_log.just_opened_set.clear()
