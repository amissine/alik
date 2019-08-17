#!/usr/bin/env bash
(sleep 2; sudo svc -d $2; sleep 1; sudo svstat $2; echo '- hit Ctrl-C to exit') &
sudo -E tail -F $1 | tai64nlocal
