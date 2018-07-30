import sys
import os
import time
import statistics

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))


class StatParams:
    def __init__(self, smart_log, setts):
        self.smart_log = smart_log
        self.setts = setts
        self.average_period = self.setts.average_period
        self.average_period /= self.setts.acceleration
        self.init_time = time.time()

        self.time = list([float(0)])

        self.profit = list([int(0)])
        self.income = list([float(0)])
        self.locked_balance = list([int(0)])
        self.ROI_day = list([float(0)])
        self.free_balance = list([int(0)])

        self.profit_av = float(0)
        self.income_av = float(0)
        self.locked_balance_av = float(0)
        self.ROI_day_av = float(0)
        self.free_balance_av = float(0)

    def process(self):
        self.set_data()
        self.cut_samples()
        self.calc_data_av()

    def set_data(self):
        self.time.append(time.time() - self.init_time)
        self.profit.append(self.smart_log.profit)

        time_delta = self.time[-1] - self.time[-2]
        profit_delta = self.profit[-1] - self.profit[-2]
        self.income.append(profit_delta / time_delta)

        self.locked_balance.append(int(0))
        for _, balance in self.smart_log.router_balances.items():
            self.locked_balance[-1] += balance
        self.locked_balance[-1] -= self.smart_log.io_funds

        if self.locked_balance[-1] > 0:
            self.ROI_day.append(self.income[-1] / self.locked_balance[-1])
            self.ROI_day[-1] *= 60 * 60 * 24 * 100
        else:
            self.ROI_day.append(float(0))

        self.free_balance.append(self.smart_log.router_free_balance)

    def cut_samples(self):
        while self.time[-1] - self.time[0] > self.average_period:
            del self.time[0]
            del self.profit[0]
            del self.income[0]
            del self.locked_balance[0]
            del self.ROI_day[0]
            del self.free_balance[0]

    def calc_data_av(self):
        for i in range(len(self.time)):
            self.profit_av = statistics.mean(self.profit)
            self.income_av = statistics.mean(self.income)
            self.locked_balance_av = statistics.mean(self.locked_balance)
            self.ROI_day_av = statistics.mean(self.ROI_day)
            self.free_balance_av = statistics.mean(self.free_balance)
