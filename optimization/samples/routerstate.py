import sys
import os

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
