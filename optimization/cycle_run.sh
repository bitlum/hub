#!/usr/bin/env bash
printf '' > router_log.log
screen -dmS manager manager --updateslog=router_log.log
cd activity/rpc/
screen -dmS activity python3 actrpc.py
cd ../../optimizer
python3 optimizer.py
cd ..
screen -S manager -p 0 -X stuff $'\003'
screen -S activity -p 0 -X stuff $'\003'

