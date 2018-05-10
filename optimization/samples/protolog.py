import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))


class ProtoLog:

    def __init__(self):
        self.messages = []
        self.total_length = 0
        self.number_messages = 0

    def append(self, message):
        self.messages.append(message)
        self.total_length += len(message.SerializeToString())
        self.number_messages = len(self.messages)

    def __str__(self):
        out_str = ''
        for i in range(0, self.number_messages):
            out_str += self.messages[i].__str__()
            out_str += '\n'
        out_str += 'Number of massages are ' + str(self.number_messages) + '\n'
        out_str += 'Total length of massages is ' + str(
            self.total_length) + ' bytes'
        return out_str
