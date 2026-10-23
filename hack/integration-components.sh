#!/usr/bin/env bash

# Builds the local clusterctl repository the integration tests install Cluster API and cert-manager
# from, out of the manifests and images components.txt pins, and writes the clusterctl configuration
# that points at it.
#
#   hack/integration-components.sh <components.txt> <output directory> <Cluster API version>

set -euo pipefail

if [ $# -ne 3 ]; then
	echo "usage: $0 <components.txt> <output directory> <Cluster API version>" >&2
	exit 2
fi

pins=$1
out=$2
capi_version=$3
repository=$out/repository

rm -rf "$repository"
mkdir -p "$repository"

manifests=()
images=()

while read -r kind first second; do
	case $kind in
		'' | '#'*) continue ;;
		image) images+=("$first") ;;
		*)
			url=$second
			version=${url%/*}
			version=${version##*/}
			file=${url##*/}
			case $kind in
				cluster-api | infrastructure-docker)
					# The tests build against the Cluster API go.mod requires; a different release here
					# would test the bootstrap provider against controllers it was not built for.
					if [ "$version" != "$capi_version" ]; then
						echo "$pins names Cluster API $version, go.mod requires $capi_version" >&2
						exit 1
					fi
					;;
			esac
			dir=$repository/$kind/$version
			mkdir -p "$dir"
			curl -fsSL --proto '=https' --retry 5 -o "$dir/$file.download" "$url"
			echo "$first  $dir/$file.download" | sha256sum -c --quiet -
			mv "$dir/$file.download" "$dir/$file"
			if [ "$file" != metadata.yaml ]; then
				manifests+=("$kind/$version/$file")
			fi
			;;
	esac
done < "$pins"

for ref in "${images[@]}"; do
	case $ref in
		*:*@sha256:*) ;;
		*)
			echo "$pins: $ref is no image named by tag and digest" >&2
			exit 1
			;;
	esac
	named=${ref%@*}
	# The reference as a manifest names it, with every character a regular expression gives meaning to
	# escaped; it must end where the reference ends, so a longer tag or a pinned one is left alone.
	pattern=$(printf '%s' "$named" | sed 's/[][\.*^$/]/\\&/g')
	found=
	for manifest in "${manifests[@]}"; do
		sed -E -i "s/(^|[^[:alnum:]._/-])${pattern}([\"'[:space:]]|$)/\1${ref//\//\\/}\2/g" "$repository/$manifest"
		if grep -qF "$ref" "$repository/$manifest"; then
			found=1
		fi
	done
	if [ -z "$found" ]; then
		echo "$pins pins $ref, which no manifest names" >&2
		exit 1
	fi
done

status=0
for manifest in "${manifests[@]}"; do
	# Every container image and every image passed as an argument must now carry a digest.
	unpinned=$(grep -nE '(image:|-image=)' "$repository/$manifest" | grep -vE '@sha256:[0-9a-f]{64}' || true)
	if [ -n "$unpinned" ]; then
		echo "$manifest names images components.txt does not pin:" >&2
		echo "$unpinned" >&2
		status=1
	fi
done
[ $status -eq 0 ] || exit 1

component() {
	for manifest in "${manifests[@]}"; do
		case $manifest in
			"$1"/*) echo "$repository/$manifest" ;;
		esac
	done
}

cert_manager=$(component cert-manager)
cert_manager_version=${cert_manager%/*}
cert_manager_version=${cert_manager_version##*/}

# clusterctl reads a provider from a local repository when its URL is a path; the providers named
# here replace its built-in GitHub entries of the same name.
cat > "$out/clusterctl.yaml" << EOF
providers:
  - name: cluster-api
    type: CoreProvider
    url: $(component cluster-api)
  - name: docker
    type: InfrastructureProvider
    url: $(component infrastructure-docker)
cert-manager:
  url: $cert_manager
  version: $cert_manager_version
EOF

echo "clusterctl repository with pinned images in $repository"
