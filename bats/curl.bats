#!/usr/bin/env bats
# See also: {{{1
#
#   https://gist.github.com/tkuchiki/041a401041530c05f73a

@test "addition using bc" { # {{{1
  result="$(echo 2+2 | bc)"
  [ "$result" -eq 4 ]
}

@test "addition using dc" { # {{{1
  result="$(echo 2 2+p | dc)"
  [ "$result" -eq 4 ]
}

@test "use the run helper" { # {{{1
  run echo $PWD
  [ "$status" -eq 0 ]
  [ "$output" = $PWD ]
}

@test "use the run helper's \$lines array" { # {{{1
  run echo $PWD
  [ "$status" -eq 0 ]
  [ "${lines[0]}" = $PWD ]
}

@test "check if \$PWD/util/curl.sh exists and is executable" { # {{{1
  [ -x ./util/curl.sh ]
}

@test "get BTC order book from SDEX" { # {{{1
  . util/BTC.sh
  sdex_ob BTC >> ./bats.log
  [ "$LOG_NUM" = 'n44' ]
}

@test "get BTC trades from SDEX" { # {{{1
  . util/BTC.sh
  sdex_t BTC >> ./bats.log
  [ "$LOG_NUM" = 'n44' ]
}

