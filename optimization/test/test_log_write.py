import time
import random

import protobuf.test.test_pb2 as proto
from datetime import datetime


def log_generate():
    message = proto.Event()
    message.time = datetime.now().strftime('%Y-%m-%d %H:%M:%S')
    values = ('error', 'overhead', 'lack', 'check', 'stop', 'test', 'continue')
    for _ in range(random.randint(1, 7)):
        message.value.append(values[random.randint(0, len(values) - 1)])
    return message


log_file = 'message.log'
with open(log_file, 'wb') as f:
    pass

messages_number = 100

for i in range(messages_number):
    time.sleep(random.uniform(1, 3))
    message_ = log_generate()
    print(message_)
    with open(log_file, 'ab') as f:
        f.write(len(message_.SerializeToString()).to_bytes(2, byteorder='big',
                                                           signed=False))
        f.write(message_.SerializeToString())
