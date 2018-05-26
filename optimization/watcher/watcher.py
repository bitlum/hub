from watchdog.events import PatternMatchingEventHandler as PattMatchEvHand
import sys
import os
import time

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))

from watcher.protologutills import split_path_name
from watcher.logreader import LogReader
from core.hubrpc import HubRPC
from core.routermgt import RouterMgt
from watcher.routermetrics import RouterMetrics


class Watcher(PattMatchEvHand):

    def __init__(self, file_name, smart_log, router_setts, mgt_freq,
                 draw_period_av):
        super().__init__(patterns='*' + split_path_name(file_name)['name'],
                         ignore_directories=True, case_sensitive=False)

        self.smart_log = smart_log
        self.log_reader = LogReader(file_name, smart_log)
        self.router_mgt = RouterMgt(self.smart_log.transseq, router_setts)
        self.mgt_freq = int(mgt_freq)
        self.mgt_ticker = int(0)
        self.update_set = set()

        self.hubrpc = HubRPC(self.router_mgt.balances, self.update_set)
        self.hubrpc.set_payment_fee_base(router_setts.payment_fee_base)
        self.hubrpc.set_payment_fee_proportional(
            router_setts.payment_fee_proportional)

        self.router_metrics = RouterMetrics(self.smart_log, time.time(),
                                            draw_period_av)

    def process(self, event):
        if (event.event_type == 'modified') and (
                event.src_path == self.log_reader.file_name):
            self.log_reader.process_log()

            if self.mgt_ticker % self.mgt_freq == 0:

                self.router_metrics.process()

                self.router_mgt.calc_parameters()

                self.update_set.clear()
                for user, wane in self.router_mgt.wanes.items():
                    balance_cur = 0
                    if user in self.smart_log.router_balances:
                        balance_cur = self.smart_log.router_balances[user]
                    bound = self.router_mgt.bounds[user]
                    balance_opt = self.router_mgt.balances[user]
                    if wane:
                        if balance_cur < bound:
                            self.update_set.add(user)
                    else:
                        if balance_cur > bound or balance_cur < balance_opt:
                            self.update_set.add(user)

                for user in self.smart_log.blockage_set:
                    self.update_set.discard(user)

                self.hubrpc.update()

            self.mgt_ticker += 1

    def on_modified(self, event):
        self.process(event)

    def on_created(self, event):
        self.process(event)

    def on_moved(self, event):
        self.process(event)

    def on_deleted(self, event):
        self.process(event)
