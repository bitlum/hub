import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))

from core.actmatr_from_stream import actmatr_calc
from core.actmatr_smart import actmatr_smart_gen
from core.flows import flows_calc
from core.balances import balance_calc

actmatr_calc(file_name_inlet='inlet/actmatr_from_stream_inlet.json')
actmatr_smart_gen(file_name_inlet='inlet/actmatr_smart_inlet.json')
flows_calc(file_name_inlet='inlet/flows_inlet.json')
balance_calc(file_name_inlet='inlet/balances_inlet.json')
