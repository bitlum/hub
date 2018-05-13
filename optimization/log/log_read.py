from watchdog.observers import Observer
import time
import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))

from samples.protolog import ProtoLog
from samples.smartlog import SmartLog
from watcher.watchlogread import WatchLogRead
from watcher.protologutills import split_path_name

proto_log = ProtoLog()
smart_log = SmartLog()

log_file = '/Users/bigelk/data/tmp/manager/test.log'

obser = Observer()
watch_log_read = WatchLogRead(
    split_path_name(log_file)['path'] + split_path_name(log_file)['name'],
    proto_log, smart_log)
obser.schedule(watch_log_read, split_path_name(log_file)['path'])
obser.start()

try:
    while True:
        time.sleep(1)
except KeyboardInterrupt:
    obser.stop()

obser.join()
