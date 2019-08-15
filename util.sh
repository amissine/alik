# util.sh {{{1

declare -i CAAS_N=500
let LIMIT=CAAS_N+CAAS_N

# }}}1
log () { # {{{1
  echo `date +%s` "$1" 1>&2
}

pipe_in () { # {{{1
  log "pipe_in: started sending sysin to $1, shall kill $2 on EOF"
  tail -n 999999 >> $1
  local rc=$?
  kill $2
  log "pipe_in: got EOF (rc $rc), killed $2, now exiting"
}

file_can_grow () { # {{{1
  let capacity=LIMIT-$1; let rc=0
  log "file_can_grow: count $1 capacity $capacity"
  [ $capacity -gt 0 ] || let rc=1
  return $rc
}

pour () { # {{{1
  rm -f $1 2>/dev/null; touch $1
  tail -n 999999 -F $1 &
  local pid2=$!
  log "pour: process $pid2 is sending $1 to sysout"
  while true; do
    pipe_in $1 $pid2 &
    local pid=$!
    log "pour: process $pid (pipe_in) is sending sysin to $1"
    while file_can_grow `wc -l $1`; do
      kill -0 $pid2 2>/dev/null; [ $? -eq 0 ] || return 0
      sleep 5
    done
    log "pour: rotating $1"
    local timestamp=`date +%s`
    kill $pid; sleep 1; mv $1 archive/$1.$timestamp; touch $1; wait $pid
  done
}

into () { # {{{1
  cat -u >> $1
  sleep 0.2; kill $2
}

pour () { # {{{1
  rm -f $1 2>/dev/null; touch $1
  tail -n 999999 -F $1 & local pout=$!
  log "pour: pout $pout"
  into $1 $pout; return

  while true; do
    into $1 $pout & local pin=$!
    if kill -0 $pout 2>/dev/null; then break; fi
    sleep 1
  done
  kill $pin
}
