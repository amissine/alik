#!/usr/bin/env bash # {{{1
#
# See also:
# - https://golang.org/doc/code.html

# Locals {{{1
GOPATH=$(go env GOPATH)
FEED=$GOPATH/bin/feed
SVC_D=/var/svc.d/feed

install_service_feed () { # {{{2
  [ -L /service/feed ] && return

  mkdir -p $SVC_D/env $SVC_D/log
  cp $FEED $SVC_D/feed; chmod +x $SVC_D/feed
  cp service/feed.sh $SVC_D/feed.sh; chmod +x $SVC_D/feed.sh
  cp service/feed_run.sh $SVC_D/run; chmod +x $SVC_D/run
  cp service/feed_log_run.sh $SVC_D/log/run; chmod +x $SVC_D/log/run
  ln -s $SVC_D /service/feed
  echo '- service feed installed'
}

# Install command feed {{{1
cd feed
go install
echo "- command feed installed to $GOPATH/bin"
cd - >/dev/null

# Install service feed (with log service and env support) locally {{{1
install_service_feed

# Access file $1 (sysout, the current log) for reading {{{1
while [ ! -r $1 ]; do
  echo "- waiting 2s to access $1"; sleep 2
done
echo '- updated'
