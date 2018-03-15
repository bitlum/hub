import time
from watchdog.observers import Observer

from smartlog import *
from watchlogread import *

log = SmartLog()

log_file = '/Users/bigelk/data/tmp/manager/test.log'

obser = Observer()
watch_log_read = WatchLogRead(
    split_path_name(log_file)['path'] + split_path_name(log_file)['name'], log)
obser.schedule(watch_log_read, split_path_name(log_file)['path'])
obser.start()

try:
    while True:
        time.sleep(1)
except KeyboardInterrupt:
    obser.stop()

obser.join()
