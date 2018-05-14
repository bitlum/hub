import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))


class RouterState:
    def __init__(self):
        self.router_balances = dict()
        self.sender = str()
        self.receiver = str()
        self.user = str()
        self.amount = int()
        self.router_balance = int()
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
        self.router_balances[key] = self.router_balances[
                                 key] + val if key in self.router_balances else val

    def get_payment_data(self, message):
        self.receiver = message[self.id['receiver']]
        self.sender = message[self.id['sender']]
        self.amount = message[self.id['amount']]

    def get_change_data(self, message):
        self.user = message[self.id['user']]
        self.router_balance = message[self.id['router_balance']]

    def set_payment(self, message):
        if message[self.id['status']] == self.id['success']:
            self.get_payment_data(message)
            self.set_amount(self.sender, self.amount)
            self.set_amount(self.receiver, -self.amount)

    def set_state(self, message):
        self.router_balances.clear()
        self.duration = message[self.id['duration']]
        for channel in message[self.id['channels']]:
            self.router_balances[channel[self.id['user']]] = channel[
                self.id['router_balance']]

    def set_change(self, message):
        self.get_change_data(message)
        self.router_balances[self.user] = self.router_balance
