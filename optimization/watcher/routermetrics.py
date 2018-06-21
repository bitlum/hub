import sys
import os
import time
import Gnuplot
import json

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))


class RouterMetrics:
    def __init__(self, smart_log, router_mgt):
        self.smart_log = smart_log
        self.init_time = time.time()

        self.time = list([float(0)])

        self.profit = list([int(0)])
        self.income = list([float(0)])
        self.balance_sum = list([int(0)])
        self.balance_sum_predict = list([int(0)])
        self.gain_sum_predict = list([float(0)])
        self.ROI = list([float(0)])
        self.ROI_predict = list([float(0)])

        self.profit_av = list([float(0)])
        self.income_av = list([float(0)])
        self.balance_sum_av = list([float(0)])
        self.balance_sum_predict_av = list([float(0)])
        self.gain_sum_predict_av = list([float(0)])
        self.ROI_av = list([float(0)])
        self.ROI_predict_av = list([float(0)])

        self.stat_period = router_mgt.setts.stat_period

        self.make_drawing = router_mgt.setts.make_drawing
        self.output_statistics = router_mgt.setts.output_statistics

        self.router_mgt = router_mgt

        Gnuplot.GnuplotOpts.default_term = 'qt noraise'
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
        if self.make_drawing:
            self.draw()

    def set_data(self):
        self.time.append(time.time() - self.init_time)

        self.profit.append(self.smart_log.profit)

        time_delta = self.time[-1] - self.time[-2]

        self.income.append((self.profit[-1] - self.profit[-2]) / time_delta)

        self.balance_sum.append(int(0))
        for _, balance in self.smart_log.router_balances.items():
            self.balance_sum[-1] += balance

        self.balance_sum_predict.append(int(0))
        for _, balance in self.router_mgt.balances_eff.items():
            self.balance_sum_predict[-1] += balance

        self.gain_sum_predict.append(int(0))
        for _, gain in self.router_mgt.gain_eff.items():
            self.gain_sum_predict[-1] += gain

        if self.balance_sum[-1] > 0:
            self.ROI.append(self.income[-1] / self.balance_sum[-1])
        else:
            self.ROI.append(float(0))

        if self.balance_sum_predict[-1] > 0:
            ROI = self.gain_sum_predict[-1] / self.balance_sum_predict[-1]
            self.ROI_predict.append(ROI)
        else:
            self.ROI_predict.append(float(0))

    def calc_data_av(self):
        self.profit_av.append(float(0))
        self.income_av.append(float(0))
        self.balance_sum_av.append(float(0))
        self.balance_sum_predict_av.append(float(0))
        self.gain_sum_predict_av.append(float(0))
        self.ROI_av.append(float(0))
        self.ROI_predict_av.append(float(0))

        count = int(0)
        for i in range(len(self.time) - 1, - 1, -1):
            if (self.time[-1] - self.time[i]) > self.stat_period:
                break
            else:
                count += 1
                self.profit_av[-1] += self.profit[i]
                self.income_av[-1] += self.income[i]
                self.balance_sum_av[-1] += self.balance_sum[i]
                self.balance_sum_predict_av[-1] += self.balance_sum_predict[i]
                self.gain_sum_predict_av[-1] += self.gain_sum_predict[i]
                self.ROI_av[-1] += self.ROI[i]
                self.ROI_predict_av[-1] += self.ROI_predict[i]

        self.profit_av[-1] /= count
        self.income_av[-1] /= count
        self.balance_sum_av[-1] /= count
        self.balance_sum_predict_av[-1] /= count
        self.gain_sum_predict_av[-1] /= count
        self.ROI_av[-1] /= count
        self.ROI_predict_av[-1] /= count

    def draw(self):

        ROI_av_curve = Gnuplot.Data(
            self.time, self.ROI_av,
            title="ROI",
            with_="lines lw 3 lt 1 lc 2")

        ROI_predict_av_curve = Gnuplot.Data(
            self.time, self.ROI_predict_av,
            title="predicted ROI",
            with_="lines lw 3 lt 0 lc 2")

        balance_sum_av_curve = Gnuplot.Data(
            self.time, self.balance_sum_av,
            axes='x1y2',
            title="locked funds",
            with_="lines lw 3 lt 1 lc 1")
        balance_sum_predict_av_curve = Gnuplot.Data(
            self.time,
            self.balance_sum_predict_av,
            axes='x1y2',
            title="predicted locked funds",
            with_="lines lw 3 lt 0 lc 1")

        self.gnuplot.plot(balance_sum_av_curve,
                          balance_sum_predict_av_curve,
                          ROI_av_curve,
                          ROI_predict_av_curve)

    def out_stat(self):
        if self.output_statistics:
            with open('outlet/statistics.json', 'w') as f:
                json.dump({'time': self.time,
                           'profit_av': self.profit_av,
                           'income_av': self.income_av,
                           'balance_sum_av': self.balance_sum_av,
                           'ROI_av': self.ROI_av},
                          f, sort_keys=True, indent=4 * ' ')
