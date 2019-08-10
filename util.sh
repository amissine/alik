# util.sh {{{1

declare -i CAAS_N=500
let LIMIT=CAAS_N+CAAS_N

# }}}1
log () { # {{{1
  echo "$1" 1>&2
}

pipe_in () { # {{{1
  log "pipe_in: started, shall kill $2"
  cat -u >> $1
  sleep 1; kill $2
  log "pipe_in: killed $2, now exiting"
}

pipe_out () { # {{{1
  tail -n 999999 -f $1 &
}

file_can_grow () { # {{{1
  let capacity=LIMIT-$1; let rc=0
  log "file_can_grow: count $1 capacity $capacity"
  [ $capacity -gt 0 ] || let rc=1
  return $rc
}

pour () { # {{{1
  rm -f $1 2>/dev/null; touch $1
  pipe_out $1
  local pid2kill=$!
  pipe_in $1 $pid2kill &
  while true; do
    while file_can_grow `wc -l $1`; do
      kill -0 $pid2kill 2>/dev/null; [ $? -eq 0 ] || return 0
      sleep 5
    done
    log "pour: archiving and rotating $1"
    let LIMIT=LIMIT+LIMIT+LIMIT+LIMIT
  done
}
