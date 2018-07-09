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
        self.balance_sum_max = list([int(0)])
        self.balance_sum_predict = list([int(0)])
        self.gain_sum_predict = list([float(0)])
        self.ROI = list([float(0)])
        self.ROI_accum = list([float(0)])
        self.ROI_predict = list([float(0)])
        self.amount_mean_forw = list([float(0)])
        self.amount_mean_io = list([float(0)])

        self.profit_av = list([float(0)])
        self.income_av = list([float(0)])
        self.balance_sum_av = list([float(0)])
        self.balance_sum_predict_av = list([float(0)])
        self.gain_sum_predict_av = list([float(0)])
        self.ROI_av = list([float(0)])
        self.ROI_predict_av = list([float(0)])
        self.amount_mean_forw_av = list([float(0)])
        self.amount_mean_io_av = list([float(0)])

        self.average_period = router_mgt.setts.average_period
        self.average_period /= router_mgt.setts.acceleration

        self.plot_period = router_mgt.setts.plot_period
        self.plot_period /= router_mgt.setts.acceleration

        self.make_drawing = router_mgt.setts.make_drawing
        self.output_stat = router_mgt.setts.output_stat

        self.router_mgt = router_mgt

        Gnuplot.GnuplotOpts.default_term = 'qt noraise'
        self.gnuplot = Gnuplot.Gnuplot()
        self.gnuplot.title("Average Metrics vs Time")
        self.gnuplot("set ytics nomirror tc rgb '#008000'")
        self.gnuplot("set y2tics nomirror tc rgb '#800080'")
        self.gnuplot("set xtics nomirror")
        self.gnuplot("set xlabel 'time, s'")
        self.gnuplot("set ylabel 'fraction/s'")
        self.gnuplot("set y2label 'satoshi'")
        self.gnuplot('set xtics ' + str(self.plot_period / 5))
        self.gnuplot('set xrange[0:' + str(self.plot_period) + ']')

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

        if self.balance_sum_max[-1] > self.balance_sum[-1]:
            self.balance_sum_max.append(self.balance_sum_max[-1])
        else:
            self.balance_sum_max.append(self.balance_sum[-1])

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

        if self.balance_sum_max[-1] > 0:
            self.ROI_accum.append(
                self.profit[-1] / self.time[-1] / self.balance_sum_max[-1])
        else:
            self.ROI_accum.append(0)

        self.amount_mean_forw.append(self.router_mgt.amount_mean_forw)
        self.amount_mean_io.append(self.router_mgt.amount_mean_io)

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
        self.amount_mean_forw_av.append(float(0))
        self.amount_mean_io_av.append(float(0))

        count = int(0)
        for i in range(len(self.time) - 1, - 1, -1):
            if (self.time[-1] - self.time[i]) > self.average_period:
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
                self.amount_mean_forw_av[-1] += self.amount_mean_forw[i]
                self.amount_mean_io_av[-1] += self.amount_mean_io[i]

        self.profit_av[-1] /= count
        self.income_av[-1] /= count
        self.balance_sum_av[-1] /= count
        self.balance_sum_predict_av[-1] /= count
        self.gain_sum_predict_av[-1] /= count
        self.ROI_av[-1] /= count
        self.ROI_predict_av[-1] /= count
        self.amount_mean_forw_av[-1] /= count
        self.amount_mean_io_av[-1] /= count

    def draw(self):

        ROI_av_curve = Gnuplot.Data(
            self.time, self.ROI_av,
            title="ROI",
            with_="lines lw 3 lt 1 lc rgb '#008000'")

        ROI_predict_av_curve = Gnuplot.Data(
            self.time, self.ROI_predict_av,
            title="predicted ROI",
            with_="lines lw 3 lt 0 lc rgb '#008000'")

        ROI_accum_curve = Gnuplot.Data(
            self.time, self.ROI_accum,
            title="accumulated ROI",
            with_="lines lw 1 lt 1 lc rgb '#008000'")

        balance_sum_av_curve = Gnuplot.Data(
            self.time, self.balance_sum_av,
            axes='x1y2',
            title="locked funds",
            with_="lines lw 3 lt 1 lc rgb '#800080'")
        balance_sum_predict_av_curve = Gnuplot.Data(
            self.time,
            self.balance_sum_predict_av,
            axes='x1y2',
            title="predicted locked funds",
            with_="lines lw 3 lt 0 lc rgb '#800080'")
        balance_sum_max_curve = Gnuplot.Data(
            self.time,
            self.balance_sum_max,
            axes='x1y2',
            title="max locked funds",
            with_="lines lw 1 lt 1 lc rgb '#800080'")

        time_min = self.time[-1] - self.plot_period

        if time_min > 0:
            self.gnuplot(
                'set xrange[' + str(time_min) + ':' + str(self.time[-1]) + ']')

        self.gnuplot.plot(balance_sum_av_curve,
                          balance_sum_predict_av_curve,
                          balance_sum_max_curve,
                          ROI_av_curve,
                          ROI_predict_av_curve,
                          ROI_accum_curve)

    def out_stat(self):
        if self.output_stat:
            ROI_day = self.ROI_accum[-1] * 60 * 60 * 24
            ROI_day /= self.router_mgt.setts.acceleration
            pmnt_fee_prop = 1.E-6 * self.router_mgt.setts.pmnt_fee_prop
            with open('outlet/statistics.json', 'w') as f:
                json.dump({'time': self.time,
                           'acceleration': self.router_mgt.setts.acceleration,
                           'profit_av': self.profit_av,
                           'balance_sum_av': self.balance_sum_av,
                           'balance_sum_max_final': self.balance_sum_max[-1],
                           'ROI_day': ROI_day,
                           'amount_mean_forw_av': self.amount_mean_forw_av,
                           'amount_mean_io_av': self.amount_mean_io_av,
                           'pmnt_fee_prop': pmnt_fee_prop},
                          f, sort_keys=True, indent=4 * ' ')
