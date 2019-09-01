#!/usr/local/bin/bash

. ../util/common.sh

exec multilog $LOG_SIZE $LOG_NUM ./main
