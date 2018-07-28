from watchdog.observers import Observer
import time
import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))

from watcher.watchoptimize import WatchOptimize
from watcher.protologutills import split_path_name
from core.routersetts import RouterSetts


def optimize(setts_file_name):
    setts = RouterSetts()
    setts.get_from_file(setts_file_name)

    obser = Observer()
    path = split_path_name(setts.router_log)['path']
    obser.schedule(WatchOptimize(setts), path)

    obser.start()

    try:
        while True:
            time.sleep(1)
    except KeyboardInterrupt:
        obser.stop()

    obser.join()


optimize(setts_file_name=sys.argv[1])
