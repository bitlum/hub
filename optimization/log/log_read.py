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

file_inlet = '../core/inlet/routermgt_inlet.json'

with open(file_inlet) as f:
    inlet = json.load(f)

router_setts = RouterSetts()
router_setts.set_income(inlet['income'])
router_setts.set_penalty(inlet['penalty'])
router_setts.set_commission(inlet['commission'])
router_setts.set_time_p(inlet['time_p'])
router_setts.set_alpha_p(inlet['alpha_p'])
router_setts.set_alpha_T(inlet['alpha_T'])

smart_log = SmartLog()

router_mgt = RouterMgt(smart_log.transseq, router_setts)

log_file = '/Users/bigelk/data/tmp/manager/test.log'

log_file_path = split_path_name(log_file)['path']

obser = Observer()
watcher = Watcher(log_file, smart_log, router_mgt)
obser.schedule(watcher, log_file_path)
obser.start()

try:
    while True:
        time.sleep(1)
except KeyboardInterrupt:
    obser.stop()

obser.join()
