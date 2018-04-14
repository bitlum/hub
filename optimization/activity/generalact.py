import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))

from activity.flowvect import flowvect_gen
from activity.flowmatr import flowmatr_gen
from activity.periodmatr import periodmatr_gen
from activity.sizematr import sizematr_gen
from activity.actmatr import actmatr_gen
from activity.balances import balances_gen
from activity.transstream import transstream_gen


# Call all the functions of generating user activity in one place.


flowvect_gen(file_name_inlet='inlet/flowvect_inlet.json')

flowmatr_gen(file_name_inlet='inlet/flowmatr_inlet.json')

periodmatr_gen(file_name_inlet='inlet/periodmatr_inlet.json')

sizematr_gen(file_name_inlet='inlet/sizematr_inlet.json')

actmatr_gen(file_name_inlet='inlet/actmatr_inlet.json')

balances_gen(file_name_inlet='inlet/balances_inlet.json')

transstream_gen(file_name_inlet='inlet/transstream_inlet.json')
