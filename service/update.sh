#!/usr/bin/env bash # {{{1
#
# See also:
# - http://cr.yp.to/daemontools.html
# - http://thedjbway.b0llix.net/daemontools/blabbyd.html
# - https://golang.org/doc/code.html
# - https://www.balena.io/docs/learn/getting-started/raspberrypi3/go/

# Locals {{{1
GOPATH=$(go env GOPATH) # TODO Fix GOPATH in ../Makefile
CMD=$GOPATH/bin/$1
SVC_D=/var/svc.d/$1

install_service () { # {{{2
  if [ -L /service/$1 ]; then
    cd /service/$1; rm /service/$1; svc -dx . log
    rm -rf $SVC_D
    echo "- service $1 removed"
    cd - > /dev/null
  fi
  mkdir -p $SVC_D/env $SVC_D/log $SVC_D/util
  cp $CMD $SVC_D/$1; chmod +x $SVC_D/$1
  cp service/$1.sh $SVC_D/$1.sh; chmod +x $SVC_D/$1.sh
  cp service/$1_run.sh $SVC_D/run; chmod +x $SVC_D/run
  cp service/$1_log_run.sh $SVC_D/log/run; chmod +x $SVC_D/log/run
  cp util/*.* $SVC_D/util
  ln -s $SVC_D /service/$1
  echo "- service $1 installed"
}

# Install service $1 (with log service and env support) locally {{{1
# (if service $1 is already installed, first remove it completely)
install_service $1

# Access file $2 (sysout, the current log) for reading {{{1
while [ ! -r $2 ]; do
  echo "- waiting 2s to access $2"; sleep 2
done
