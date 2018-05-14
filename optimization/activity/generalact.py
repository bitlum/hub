import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))

from activity.flowvect import flowvect_gen
from activity.flowmatr import flowmatr_gen
from activity.periodmatr import periodmatr_gen
from activity.amountmatr import amountmatr_gen
from activity.actmatr import actmatr_gen
from activity.user_balances import user_balances_gen
from activity.transseq import transseq_gen


# Call all the functions of generating user activity in one place.


flowvect_gen(file_name_inlet='inlet/flowvect_inlet.json')

flowmatr_gen(file_name_inlet='inlet/flowmatr_inlet.json')

periodmatr_gen(file_name_inlet='inlet/periodmatr_inlet.json')

amountmatr_gen(file_name_inlet='inlet/amountmatr_inlet.json')

actmatr_gen(file_name_inlet='inlet/actmatr_inlet.json')

user_balances_gen(file_name_inlet='inlet/user_balances_inlet.json')

transseq_gen(file_name_inlet='inlet/transseq_inlet.json')
