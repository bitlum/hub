import json
import copy
import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))


# Calculation of router-locked funds and frequencies of channels re-creation
# by means of mean periods matrix, flow matrix, flow outlet vector and
# flow inlet vector.


def router_balance_calc(file_name_inlet):
    with open(file_name_inlet) as f:
        inlet = json.load(f)

    with open(inlet['periodmatr_mean_calc_file_name']) as f:
        periodmatr_mean = json.load(f)['periodmatr_mean_calc']

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

    # FIRST STEP
    # The simplest yield extremum.

    router_balances = [penalty / commission for _ in range(users_number)]
    if income:
        for balance in router_balances:
            balance *= 2

    # SECOND STEP
    # Limiting channel idle time due to re-creation.

    for i in range(len(router_balances)):
        balance_cur_lim = flowvect_out_calc[i] * time_p / alpha_p
        if router_balances[i] < balance_cur_lim:
            router_balances[i] = balance_cur_lim

    # THIRD STEP:
    # Matching the characteristic times of re-creation of channels
    # and transactions.

    periods_max = [0 for _ in range(users_number)]
    for i in range(users_number):
        for period_mean in periodmatr_mean[i]:
            if period_mean is not None:
                if periods_max[i] < period_mean:
                    periods_max[i] = period_mean

    for i in range(len(router_balances)):
        balance_cur_lim = alpha_T * flowvect_out_calc[i] * periods_max[i]
        if router_balances[i] < balance_cur_lim:
            router_balances[i] = balance_cur_lim

    # FOURTH STEP
    # Calculation of channel re-creation frequencies for outlet transactions.

    freqs_out = [flowvect_out_calc[i] / router_balances[i] for i in
                 range(users_number)]

    # FIFTH STEP
    # Calculation of channel re-creation frequencies for inlet transactions.

    freqs_in = copy.deepcopy(freqs_out)

    for i in range(len(flowmatr_calc)):
        for j in range(len(flowmatr_calc[i])):
            if flowmatr_calc[i][j] is not None:
                if freqs_in[j] > freqs_out[i]:
                    freqs_in[j] = freqs_out[i]

    # SIXTH AND SEVENTH STEP
    # Calculation of the final channels re-creation frequencies to ensure
    # all transactions.
    freqs = copy.deepcopy(freqs_out)
    for i in range(users_number):
        flow_delta_cur = flowvect_in_calc[i] - flowvect_out_calc[i]
        balance_delta_cur = flow_delta_cur / freqs_in[i]
        router_balances[i] = balance_delta_cur if flow_delta_cur > 0 else 0.
        freqs[i] = freqs_in[i] if flow_delta_cur > 0 else freqs_out[i]

    # write router-locked funds into a file

    with open(inlet['freqs_file_name'], 'w') as f:
        json.dump({'freqs': freqs}, f, sort_keys=True,
                  indent=4 * ' ')

    # write frequencies of channels re-creation into a file

    with open(inlet['router_balances_file_name'], 'w') as f:
        json.dump({'router_balances': router_balances}, f, sort_keys=True,
                  indent=4 * ' ')


if __name__ == '__main__':
    router_balance_calc(file_name_inlet='inlet/router_balances_inlet.json')
