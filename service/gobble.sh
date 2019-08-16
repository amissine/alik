#!/usr/bin/env bash
(sleep 2; sudo svc -d /service/feed; echo '- service feed is down') &
sudo -E tail -F $1 | tai64nlocal
