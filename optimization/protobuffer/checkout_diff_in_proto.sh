#!/usr/bin/env bash
echo 'log:'
diff log.proto ../../manager/logger/log.proto

echo 'hubrpc:'
diff hubrpc.proto ../../manager/hubrpc/hubrpc.proto

echo 'rpc:'
diff rpc.proto ../../manager/router/emulation/rpc.proto


