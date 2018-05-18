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


def optimize(file_name_inlet):
    with open(file_name_inlet) as f:
        inlet = json.load(f)

    router_setts = RouterSetts()
    router_setts.set_setts_from_file(file_name_inlet)

    smart_log = SmartLog()

    log_file = inlet['log_file_name']

    optimizer_idle = inlet['optimizer_idle']

    watcher = Watcher(log_file, smart_log, router_setts, optimizer_idle)

    obser = Observer()
    obser.schedule(watcher, split_path_name(log_file)['path'])

    obser.start()

    try:
        while True:
            time.sleep(1)
    except KeyboardInterrupt:
        obser.stop()

    obser.join()


optimize(file_name_inlet='routermgt_inlet.json')
