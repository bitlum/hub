from watchdog.observers import Observer
import time
import sys
import os
import json

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))

from watcher.watcher import Watcher
from watcher.protologutills import split_path_name
from core.routersetts import RouterSetts


def optimize(file_name_inlet):
    router_setts = RouterSetts()
    router_setts.set_setts_from_file(file_name_inlet)

    obser = Observer()
    path = split_path_name(router_setts.log_file_name)['path']
    obser.schedule(Watcher(router_setts), path)

    obser.start()

    try:
        while True:
            time.sleep(1)
    except KeyboardInterrupt:
        obser.stop()

    obser.join()


optimize(file_name_inlet='routermgt_inlet.json')
