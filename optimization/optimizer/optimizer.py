from watchdog.observers import Observer
import time
import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))

from samples.smartlog import SmartLog
from watcher.watcher import Watcher
from watcher.protologutills import split_path_name
from core.routersetts import RouterSetts

router_setts = RouterSetts()
router_setts.set_setts_from_file('../core/inlet/routermgt_inlet.json')

smart_log = SmartLog()

log_file = '/Users/bigelk/data/tmp/manager/test.log'

watcher = Watcher(log_file, smart_log, router_setts, 7)

obser = Observer()
obser.schedule(watcher, split_path_name(log_file)['path'])

obser.start()

try:
    while True:
        time.sleep(1)
except KeyboardInterrupt:
    obser.stop()

obser.join()
