#!/bin/sh

BINDIR=bin

if [ ! -d $BINDIR ]; then
   mkdir $BINDIR
fi
go build -o $BINDIR/scanIntranet cmd/scanIntranet/main.go
