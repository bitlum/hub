import sys
import os

from sortedcontainers import SortedDict

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))


class RouterState:
    def __init__(self):
        self.balances = dict()
        self.sender = str()
        self.receiver = str()
        self.user = str()
        self.transaction = int()
        self.amount = int()
        self.duration = int()

        self.id = {'payment': 'payment',
                   'state': 'state',
                   'change': 'channel_change',
                   'updated': 'udpated',
                   'amount': 'amount',
                   'receiver': 'receiver',
                   'sender': 'sender',
                   'user': 'user_id',
                   'channels': 'channels',
                   'router_balance': 'router_balance',
                   'duration': 'average_change_update_duration',
                   'status': 'status',
                   'success': 'success'
                   }

    def set_amount(self, key, val):
        self.balances[key] = self.balances[
                                 key] + val if key in self.balances else val

    def get_transaction_data(self, message):
        self.receiver = message[self.id['receiver']]
        self.sender = message[self.id['sender']]
        self.transaction = message[self.id['amount']]

    def get_change_data(self, message):
        self.user = message[self.id['user']]
        self.amount = message[self.id['router_balance']]

    def set_payment(self, message):
        if message[self.id['status']] == self.id['success']:
            self.get_transaction_data(message)
            self.set_amount(self.sender, self.transaction)
            self.set_amount(self.receiver, -self.transaction)

    def set_state(self, message):
        self.balances.clear()
        self.duration = message[self.id['duration']]
        for channel in message[self.id['channels']]:
            self.balances[channel[self.id['user']]] = channel[
                self.id['router_balance']]

    def set_change(self, message):
        self.get_change_data(message)
        self.balances[self.user] = self.amount


class SmartLog(RouterState):

    def __init__(self):
        super().__init__()
        self.channel_changes = []
        self.payments = []
        self.states = []

    def append(self, message):
        if self.id['payment'] in message:
            self.payments.append(message)
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
            self.payments)) + '\n'
        out_str += 'Number of states is ' + str(len(self.states)) + '\n'
        if len(self.payments) > 0:
            out_str += 'Last transaction: ' + self.sender + ' -> ' + \
                       self.receiver + ' : ' + str(self.transaction) + '\n'

        for key in list(SortedDict(self.balances).keys()):
            out_str += key + ' ' + str(self.balances[key]) + '\n'

        return out_str
