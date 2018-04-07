#!/usr/bin/env bash
echo 'diff log:'
diff log.proto ../../manager/logger/log.proto

echo 'diff hubrpc:'
diff hubrpc.proto ../../manager/hubrpc/hubrpc.proto

echo 'diff rpc:'
diff rpc.proto ../../manager/router/emulation/rpc.proto


