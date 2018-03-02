#!/usr/bin/env bash

protoc --go_out=. *.proto
protoc --python_out=. *.proto