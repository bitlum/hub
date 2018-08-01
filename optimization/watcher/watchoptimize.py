from watchdog.events import PatternMatchingEventHandler as PattMatchEvHand
import sys
import os
import time
from threading import Thread

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))

from watcher.protologutills import split_path_name
from state.smartlog import SmartLog
from watcher.logreader import LogReader
from core.hubrpc import HubRPC
from core.routermgt import RouterMgt
from state.routermetrics import RouterMetrics


class WatchOptimize(PattMatchEvHand):

    def __init__(self, setts):
        super().__init__(
            patterns='*' + split_path_name(setts.router_log)['name'],
            ignore_directories=True, case_sensitive=False)

        self.init_time = time.time()

        self.smart_log = SmartLog()
        self.log_reader = LogReader(self.smart_log, setts)
        self.router_mgt = RouterMgt(self.smart_log.transseq, setts)
        self.hubrpc = HubRPC(self.router_mgt.balances, setts)
        self.router_metrics = RouterMetrics(self.smart_log, self.router_mgt,
                                            setts)

        self.init_period = setts.init_period
        self.init_period /= setts.acceleration

        self.init_mult = setts.init_mult

        self.mgt_period = setts.mgt_period
        self.mgt_period /= setts.acceleration

        self.time_mgt_start = time.time()

        self.output_period = setts.output_period
        self.output_period /= setts.acceleration

        output_sleep = setts.output_period / setts.acceleration
        self.sleep_thread = SleepThread(self.router_metrics, self.log_reader,
                                        output_sleep)
        self.sleep_thread.start()

    def process(self, event):
        if (event.event_type == 'modified') and (
                event.src_path == self.log_reader.router_log):

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
                if balance_cur < bound or balance_cur > balance_opt:
                    self.hubrpc.update_set.add(user)
            else:
                if balance_cur > bound or balance_cur < balance_opt:
                    self.hubrpc.update_set.add(user)

        for user in self.smart_log.blockage_set:
            self.hubrpc.update_set.discard(user)

        for user in self.smart_log.closure_set:
            self.hubrpc.update_set.discard(user)

        discard_newbie_set = set()
        for user in self.smart_log.newbie_set:
            period = time.time() - self.smart_log.open_time[user]
            if period > self.init_period:
                discard_newbie_set.add(user)
        for user in discard_newbie_set:
            self.smart_log.newbie_set.discard(user)

        for user in self.smart_log.newbie_set:
            self.hubrpc.update_set.discard(user)

    def set_init_update(self):
        for user in self.smart_log.just_opened_set:
            self.hubrpc.update_set.add(user)
            user_balance_ini = self.smart_log.users_balance_ini[user]
            self.router_mgt.balances[user] = user_balance_ini * self.init_mult
        self.smart_log.just_opened_set.clear()

        # TODO tmp
        if self.router_mgt.setts.init_mult == 0:
            self.hubrpc.update_set.clear()


class SleepThread(Thread):
    def __init__(self, router_metrics, log_reader, sleep):
        Thread.__init__(self)
        self.router_metrics = router_metrics
        self.log_reader = log_reader
        self.sleep = sleep

    def run(self):
        time.sleep(self.sleep)
        self.router_metrics.out_stat()
        self.log_reader.out_log()
        print('\nMade json output\n')
