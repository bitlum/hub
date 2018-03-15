import time
from watchdog.observers import Observer

from smartlog import *
from test.test_watch_log_read import *

log = SmartLog()

log_file = 'message.log'

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
