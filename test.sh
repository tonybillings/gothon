#!/bin/bash
export GOTHON_KEEP_TEMP_DIR=true
cd test || exit
go test -v
cd ..
