#!/usr/local/bin/bash
#echo 'mia-hub' > env/REMOTE_FEED

. util/common.sh

touch ./syserr; chgrp admin ./syserr; chmod 640 ./syserr;

log started
exec envdir ./env ./gobble.sh
log "exiting with $?..."
