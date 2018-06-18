import sys
import os

from sortedcontainers import SortedDict

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))


def print_massege(massege, nesting=-4):
    if type(massege) == dict:
        print('')
        nesting += 4

        keys = list(SortedDict(massege).keys())
        for key in keys:
            if key == 'time':
                keys.remove(key)
                keys.insert(0, key)
        for key in keys:
            if key == 'date':
                keys.remove(key)
                keys.insert(0, key)
        for key in keys:
            if key == 'message_type':
                keys.remove(key)
                keys.insert(0, key)

        for key in keys:
            print(nesting * ' ', end='')
            print(key, end=': ')

            if massege[key] == 0:
                if key == 'type':
                    massege[key] = 'openning'
                elif key == 'status':
                    massege[key] = 'null'

            print_massege(massege[key], nesting)
    elif type(massege) == list:
        print('')
        nesting += 4
        for user_state in massege:
            print(nesting * ' ', end='')
            print(user_state['user_id'], end=': ')
            print_massege(user_state, nesting)
    else:
        print(massege)


def split_path_name(file_name):
    split = 0
    for i in range(len(file_name)):
        if file_name[-1 - i] == '/':
            split = - i
            break
    if split == 0:
        return {'path': './', 'name': file_name}
    else:
        return {'path': file_name[:split], 'name': file_name[split:]}
