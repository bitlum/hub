from watchdog.observers import Observer
import time
import sys
import os
import json

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))

from samples.smartlog import SmartLog
from watcher.watcher import Watcher
from watcher.protologutills import split_path_name
from core.routersetts import RouterSetts
from core.routermgt import RouterMgt

router_setts = RouterSetts()
router_setts.set_setts_from_file('../core/inlet/routermgt_inlet.json')

smart_log = SmartLog()

router_mgt = RouterMgt(smart_log.transseq, router_setts)

log_file = '/Users/bigelk/data/tmp/manager/test.log'

obser = Observer()
watcher = Watcher(log_file, smart_log, router_mgt)
obser.schedule(watcher, split_path_name(log_file)['path'])
obser.start()

try:
    while True:
        time.sleep(1)
except KeyboardInterrupt:
    obser.stop()

obser.join()
