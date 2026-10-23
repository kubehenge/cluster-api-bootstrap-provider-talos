#!/usr/bin/env bash

# Fails when a node of the integration tests' cluster holds an image that was not pulled by one of the
# digests the given files pin. containerd records an image pulled by tag under that tag, and one
# pulled by digest only under the digest, so a tag among the images means something was resolved by
# tag, and a digest not pinned means something came from elsewhere.
#
#   hack/check-images.sh <talosctl> <talosconfig> <nodes, comma-separated> <file with pins>...

set -euo pipefail

if [ $# -lt 4 ]; then
	echo "usage: $0 <talosctl> <talosconfig> <nodes, comma-separated> <file with pins>..." >&2
	exit 2
fi

talosctl=$1
talosconfig=$2
nodes=$3
shift 3

# repository@digest, whatever tag the pin names, as containerd names an image pulled by digest.
pinned=$(grep -ohE '[a-z0-9.-]+(:[0-9]+)?/[a-z0-9._/-]+:[A-Za-z0-9._-]+@sha256:[0-9a-f]{64}' "$@" |
	sed -E 's/:[A-Za-z0-9._-]+@/@/' | sort -u)

status=0
pulled=

for namespace in system cri; do
	listing=$("$talosctl" --talosconfig="$talosconfig" --nodes="$nodes" image list --namespace="$namespace")
	echo "Images in the $namespace namespace:"
	echo "$listing"
	while read -r node image _; do
		case $image in
			IMAGE | '') continue ;;
			# containerd also records every image under its configuration digest alone.
			sha256:*) continue ;;
			*@sha256:*)
				reference=$(printf '%s' "$image" | sed -E 's/:[A-Za-z0-9._-]+@/@/')
				if ! grep -qxF "$reference" <<< "$pinned"; then
					echo "$node holds $image in $namespace, a digest nothing pins" >&2
					status=1
				fi
				pulled+="$reference"$'\n'
				;;
			*)
				echo "$node holds $image in $namespace, pulled by tag" >&2
				status=1
				;;
		esac
	done <<< "$listing"
done

while read -r reference; do
	if [ -n "$reference" ] && ! grep -qxF "$reference" <<< "$pulled"; then
		echo "pinned, not pulled by any node: $reference"
	fi
done <<< "$pinned"

if [ $status -ne 0 ]; then
	exit 1
fi

echo "Every image on $nodes was pulled by a pinned digest."
