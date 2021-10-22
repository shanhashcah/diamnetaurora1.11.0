#!/usr/bin/env bash

set -e

source /etc/profile

echo "using config:"
cat diamnet-core.cfg

# initialize new db
diamnet-core new-db

if [ "$1" = "standalone" ]; then
  # start a network from scratch
  diamnet-core force-scp

  # initialze history archive for stand alone network
  diamnet-core new-hist vs

  # serve history archives to aurora on port 1570
  pushd /history/vs/
  python3 -m http.server 1570 &
  popd
fi

exec /init -- diamnet-core run