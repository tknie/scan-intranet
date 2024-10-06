#!/bin/sh

BINDIR=bin

mkdir $BINDIR
go build -o $BINDIR/scanIntranet cmd/scanIntranet/main.go
