#!/bin/sh

export PATH="${GOPATH}/bin:$PATH"

if [ "$DEBUG_AS_ROOT" = "true" ]; then
	DLV=$(which dlv)
	exec sudo "PATH=$PATH" -E "$DLV" --only-same-user=false "$@"
else
	exec dlv "$@"
fi