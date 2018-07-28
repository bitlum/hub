from watchdog.observers import Observer
import time

import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))

from watcher.protologutills import split_path_name
from statrpc.watchstat import WatchStat
from statrpc.statsetts import StatSetts


def start_stat(setts_file_name):
    setts = StatSetts()
    setts.get_from_file(setts_file_name)

    path = split_path_name(setts.router_log)['path']

    obser = Observer()
    obser.schedule(WatchStat(setts), path)
    obser.start()
    try:
        while True:
            time.sleep(1)
    except KeyboardInterrupt:
        obser.stop()


start_stat(setts_file_name=sys.argv[1])
