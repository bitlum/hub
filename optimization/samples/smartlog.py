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
        router_funds = 0
        for _, balance in self.router_balances.items():
            router_funds += balance
        out_str = ''
        out_str += 'Profit is ' + str(self.profit) + '\n'
        out_str += 'Router funds is ' + str(router_funds) + '\n'
        out_str += 'Number of channel_changes is ' + str(len(
            self.channel_changes)) + '\n'
        out_str += 'Number of payments is ' + str(len(
            self.transseq)) + '\n'
        out_str += 'Number of states is ' + str(len(self.states)) + '\n'

        lines_number = len(self.transseq)
        if lines_number >= 20:
            lines_number = 20
        for _ in range(20 - lines_number):
            print()

        if len(self.transseq) > 0:
            for i in range(lines_number):
                ind = len(self.transseq) - lines_number + i
                sender = self.transseq[ind]["payment"]["sender"]
                receiver = self.transseq[ind]["payment"]["receiver"]
                amount = str(self.transseq[ind]["payment"]["amount"])
                earned = str(self.transseq[ind]["payment"]["earned"])
                status = self.transseq[ind]["payment"]["status"]
                out_str += '* ' + sender + ' -> ' + receiver + \
                           ' :: ' + amount + ' : ' + earned + \
                           ' : ' + status + '\n'

            for key in list(SortedDict(self.router_balances).keys()):
                out_str += key + ' ' + str(
                    self.router_balances[key]) + '\n'

        return out_str
