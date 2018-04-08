import json
import copy
import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))


def balance_calc(file_name_inlet):
    with open(file_name_inlet) as f:
        inlet = json.load(f)

    with open(inlet['periodmatr_mean_file_name']) as f:
        periodmatr_mean = json.load(f)['periodmatr_mean']

    with open(inlet['flowmatr_calc_file_name']) as f:
        flowmatr_calc = json.load(f)['flowmatr_calc']

    with open(inlet['flowvect_out_calc_file_name']) as f:
        flowvect_out_calc = json.load(f)['flowvect_out_calc']

    with open(inlet['flowvect_in_calc_file_name']) as f:
        flowvect_in_calc = json.load(f)['flowvect_in_calc']

    users_number = len(flowvect_out_calc)
    income = inlet['income']
    penalty = inlet['penalty']
    commission = inlet['commission']
    time_p = inlet['time_p']
    alpha_p = inlet['alpha_p']
    alpha_T = inlet['alpha_T']

    # first step:

    balances = [penalty / commission for _ in range(users_number)]
    if income:
        for balance in balances:
            balance *= 2

    # second step:

    for i in range(len(balances)):
        balance_cur_lim = alpha_p * flowvect_out_calc[i] * time_p
        if balances[i] < balance_cur_lim:
            balances[i] = balance_cur_lim

    # third step:

    periods_max = [0 for _ in range(users_number)]
    for i in range(users_number):
        for period_mean in periodmatr_mean[i]:
            if period_mean is not None:
                if periods_max[i] < period_mean:
                    periods_max[i] = period_mean

    for i in range(len(balances)):
        balance_cur_lim = alpha_T * flowvect_out_calc[i] * periods_max[i]
        if balances[i] < balance_cur_lim:
            balances[i] = balance_cur_lim

    # fourth step:

    freqs_out = [flowvect_out_calc[i] / balances[i] for i in
                 range(users_number)]

    # fifth step:

    freqs_in = copy.deepcopy(freqs_out)

    for i in range(len(flowmatr_calc)):
        for j in range(len(flowmatr_calc[i])):
            if flowmatr_calc[i][j] is not None:
                if freqs_in[j] > freqs_out[i]:
                    freqs_in[j] = freqs_out[i]

    # sixth and seventh step:

    freqs = copy.deepcopy(freqs_out)
    for i in range(users_number):
        flow_delta_cur = flowvect_in_calc[i] - flowvect_out_calc[i]
        balance_delta_cur = flow_delta_cur / freqs_in[i]
        balances[i] = balance_delta_cur if flow_delta_cur > 0 else 0.
        freqs[i] = freqs_in[i] if flow_delta_cur > 0 else freqs_out[i]

    with open(inlet['freqs_file_name'], 'w') as f:
        json.dump({'freqs': freqs}, f, sort_keys=True,
                  indent=4 * ' ')

    with open(inlet['balances_file_name'], 'w') as f:
        json.dump({'balances': balances}, f, sort_keys=True,
                  indent=4 * ' ')


if __name__ == '__main__':
    balance_calc(file_name_inlet='inlet/balances_inlet.json')
