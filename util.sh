# Utils {{{1

declare -i PID2KILL=-1

sip () { # {{{2
  for file2tail in $@; do
    tail -n 999999 -f $file2tail &
    PID2KILL=$!
  done
}

pour () { # {{{2
  local file2tail=$1
  rm -f $file2tail 2>/dev/null; touch $file2tail
  sip $file2tail
  tail -n 999999 -f >> $file2tail
  echo "pour exiting PID2KILL $PID2KILL" 1>&2
  sleep 1; kill $PID2KILL
}
