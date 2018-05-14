import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../../'))

from core.testbed.actmatr_transseq import actmatr_calc
from core.testbed.actmatr_smart import actmatr_smart_gen
from core.testbed.flows import flows_calc
from core.testbed.router_balances import router_balance_calc

# Call all the functions of core optimization.

actmatr_calc(file_name_inlet='inlet/actmatr_from_stream_inlet.json')
actmatr_smart_gen(file_name_inlet='inlet/actmatr_smart_inlet.json')
flows_calc(file_name_inlet='inlet/flows_inlet.json')
router_balance_calc(file_name_inlet='inlet/router_balances_inlet.json')
