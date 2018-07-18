from watchdog.observers import Observer
import time

import sys
import os

current_path = os.path.dirname(os.path.abspath(__file__))
sys.path.append(os.path.join(current_path, '../'))

from watcher.protologutills import split_path_name
from statrpc.statwatcher import WatcherStat
from statrpc.statsetts import StatSetts


def start_stat(log_file_name, inlet_file_name):
    path = split_path_name(log_file_name)['path']
    setts = StatSetts()
    setts.set_from_file(inlet_file_name)

    obser = Observer()
    obser.schedule(WatcherStat(setts, log_file_name), path)
    obser.start()
    try:
        while True:
            time.sleep(1)
    except KeyboardInterrupt:
        obser.stop()


start_stat(log_file_name=sys.argv[1], inlet_file_name='statrpc_inlet.json')
