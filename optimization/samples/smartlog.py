import sys
import os

from sortedcontainers import SortedDict

from samples.routerstate import RouterState

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))


class SmartLog(RouterState):

    def __init__(self):
        super().__init__()
        self.channel_changes = list()
        self.transseq = list()
        self.states = list()

    def append(self, message):
        if self.id['payment'] in message:
            self.transseq.append(message)
            self.set_payment(message[self.id['payment']])

        if self.id['state'] in message:
            self.states.append(message)
            self.set_state(message[self.id['state']])

        if self.id['change'] in message:
            if message[self.id['change']]['type'] == self.id['updated']:
                self.channel_changes.append(message)
                self.set_change(message[self.id['change']])

    def __str__(self):
        out_str = ''
        out_str += 'Number of channel_changes is ' + str(len(
            self.channel_changes)) + '\n'
        out_str += 'Number of payments is ' + str(len(
            self.transseq)) + '\n'
        out_str += 'Number of states is ' + str(len(self.states)) + '\n'
        if len(self.transseq) > 0:
            out_str += 'Last transaction: ' + self.sender + ' -> ' + \
                       self.receiver + ' : ' + str(self.amount) + '\n'

        for key in list(SortedDict(self.router_balances).keys()):
            out_str += key + ' ' + str(self.router_balances[key]) + '\n'

        return out_str
