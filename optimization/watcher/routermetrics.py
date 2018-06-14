import sys
import os
import time
import Gnuplot
import json

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))


class RouterMetrics:
    def __init__(self, smart_log, router_setts):
        self.smart_log = smart_log
        self.init_time = time.time()

        self.time = list([float(0)])

        self.profit = list([int(0)])
        self.income = list([float(0)])
        self.balance_sum = list([int(0)])
        self.ROI = list([float(0)])

        self.profit_av = list([float(0)])
        self.income_av = list([float(0)])
        self.balance_sum_av = list([float(0)])
        self.ROI_av = list([float(0)])

        self.draw_period = router_setts.draw_period

        Gnuplot.GnuplotOpts.default_term = 'qt'
        self.gnuplot = Gnuplot.Gnuplot()
        self.gnuplot.title("metrics vs time")
        self.gnuplot("set y2tics")
        self.gnuplot("set xlabel 'time, s'")
        self.gnuplot("set ylabel 'ROI, fraction/s'")
        self.gnuplot("set y2label 'locked funds, satoshi'")
        self.gnuplot("set grid")

    def process(self):
        self.set_data()
        self.calc_data_av()
        self.draw()
        self.json_outlet()

    def set_data(self):
        self.time.append(time.time() - self.init_time)

        self.profit.append(self.smart_log.profit)

        time_delta = self.time[-1] - self.time[-2]

        self.income.append((self.profit[-1] - self.profit[-2]) / time_delta)

        self.balance_sum.append(int(0))
        for _, balance in self.smart_log.router_balances.items():
            self.balance_sum[-1] += balance

        if self.income[-1] > 0:
            self.ROI.append(self.income[-1] / self.balance_sum[-1])
        else:
            self.ROI.append(float(0))

    def calc_data_av(self):
        self.profit_av.append(float(0))
        self.income_av.append(float(0))
        self.balance_sum_av.append(float(0))
        self.ROI_av.append(float(0))

        count = int(0)
        for i in range(len(self.time) - 1, - 1, -1):
            if (self.time[-1] - self.time[i]) > self.draw_period:
                break
            else:
                count += 1
                self.profit_av[-1] += self.profit[i]
                self.income_av[-1] += self.income[i]
                self.balance_sum_av[-1] += self.balance_sum[i]
                self.ROI_av[-1] += self.ROI[i]

        self.profit_av[-1] /= count
        self.income_av[-1] /= count
        self.balance_sum_av[-1] /= count
        self.ROI_av[-1] /= count

    def draw(self):

        ROI_av_curve = Gnuplot.Data(self.time, self.ROI_av,
                                    title="ROI",
                                    with_="lines lw 2")

        balance_sum_av_curve = Gnuplot.Data(self.time, self.balance_sum_av,
                                            axes='x1y2',
                                            title="locked funds",
                                            with_="lines lw 2")

        self.gnuplot.plot(balance_sum_av_curve, ROI_av_curve)

    def json_outlet(self):

        with open('outlet/statistics.json', 'w') as f:
            json.dump({'time': self.time,
                       'profit_av': self.profit_av,
                       'income_av': self.income_av,
                       'balance_sum_av': self.balance_sum_av,
                       'ROI_av': self.ROI_av},
                      f, sort_keys=True, indent=4 * ' ')
