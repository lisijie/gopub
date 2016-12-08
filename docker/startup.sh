#!/bin/sh

set -e

until telnet mysql_server 3306 > /dev/null 2>&1 ; do
	echo "waiting for mysql startup..."
	sleep 3
done

exec /gopub/gopub