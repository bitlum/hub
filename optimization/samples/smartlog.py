import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))


class SmartLog:

    def __init__(self):
        self.channel_changes_all = []
        self.channel_changes_last = []
        self.payments_all = []
        self.payments_last = []
        self.states = []

        self.payment_name = 'payment'
        self.state_name = 'state'
        self.channel_change_name = 'channel_change'

    def append(self, message):
        if self.payment_name in message:
            self.payments_all.append(message)
            self.payments_last.append(message)

        if self.state_name in message:
            self.states.append(message)
            self.payments_last.clear()
            self.channel_changes_last.clear()

        if self.channel_change_name in message:
            if message['channel_change']['type'] == 'udpated':
                self.channel_changes_all.append(message)
                self.channel_changes_last.append(message)

    def __str__(self):
        out_str = ''
        out_str += 'Number of channel_changes_all is ' + str(len(
            self.channel_changes_all)) + '\n'
        out_str += 'Number of channel_changes_last is ' + str(len(
            self.channel_changes_last)) + '\n'
        out_str += 'Number of payments_all is ' + str(len(
            self.payments_all)) + '\n'
        out_str += 'Number of payments_last is ' + str(len(
            self.payments_last)) + '\n'
        out_str += 'Number of states is ' + str(len(self.states)) + '\n'
        return out_str
