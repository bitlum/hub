import sys
import os
import time

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))


class RouterState:
    def __init__(self):
        self.router_balances = dict()
        self.router_free_balance_ini = int()
        self.router_free_balance = int()
        self.router_balance_sum = int()
        self.sender = str()
        self.receiver = str()
        self.user = str()
        self.amount = int()
        self.earned = int()
        self.router_balance = int()
        self.profit = int()
        self.io_funds = int()

        self.id = {'payment': 'payment',
                   'state': 'state',
                   'change': 'channel_change',
                   'updated': 'updated',
                   'opened': 'opened',
                   'closed': 'closed',
                   'amount': 'amount',
                   'earned': 'earned',
                   'receiver': 'receiver',
                   'sender': 'sender',
                   'user': 'user_id',
                   'channels': 'channels',
                   'router_balance': 'router_balance',
                   'free_balance': 'free_balance',
                   'duration': 'average_change_update_duration',
                   'status': 'status',
                   'success': 'success'
                   }

    def set_amount(self, user, value):
        if user in self.router_balances:
            self.router_balances[user] += value
        else:
            self.router_balances[user] = value

    def get_payment_data(self, message):
        self.receiver = message[self.id['receiver']]
        self.sender = message[self.id['sender']]
        self.amount = message[self.id['amount']]
        self.earned = message[self.id['earned']]

    def get_change_data(self, message):
        self.user = message[self.id['user']]
        self.router_balance = message[self.id['router_balance']]

    def calc_router_balance_sum(self):
        self.router_balance_sum = 0
        for _, balance in self.router_balances.items():
            self.router_balance_sum += balance

    def calc_router_free_balance(self):
        difference = self.router_balance_sum - self.profit - self.io_funds
        self.router_free_balance = self.router_free_balance_ini - difference

    def set_payment(self, message):
        if message[self.id['status']] == self.id['success']:
            self.get_payment_data(message)

            if self.sender != '0':
                self.set_amount(self.sender, self.amount + self.earned)
            else:
                self.io_funds -= self.amount

            if self.receiver != '0':
                self.set_amount(self.receiver, -self.amount)
            else:
                self.io_funds += self.amount
            self.profit += self.earned

            self.calc_router_balance_sum()
            self.calc_router_free_balance()

    def set_state(self, message):
        self.router_balances.clear()

        self.router_free_balance = message[self.id['free_balance']]
        difference = self.router_balance_sum - self.profit - self.io_funds
        self.router_free_balance_ini = self.router_free_balance + difference

        for channel in message[self.id['channels']]:
            self.router_balances[channel[self.id['user']]] = channel[
                self.id['router_balance']]

        self.calc_router_balance_sum()

    def set_change(self, message):
        self.get_change_data(message)
        self.router_balances[self.user] = self.router_balance

        self.calc_router_balance_sum()
        self.calc_router_free_balance()
