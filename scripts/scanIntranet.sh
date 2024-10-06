#!/bin/sh


#fping -g 192.168.178.0/24 -r 1 >/dev/null 2>&1

POSTGRES_URL=postgres://<user>:<password>@<host>:<port>/<database>
export POSTGRES_URL

$HOME/bin/scanIntranet

