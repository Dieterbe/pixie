#!/bin/bash
# generate sets (albums) in a very cheap (symlinks) and
# deterministic/idempotent way (so that you can throw them away if you want to),
# based on tmsu tags.

# dependencies: bash, tmsu, exiv2, python2

source "$(dirname "$0")"/lib.sh

exports_dir=$(grep exports_dir ~/.pixie/config.ini | cut -d' ' -f3 | tr -d '"')
[ -n "$exports_dir" ] || die_error "could not get exports_dir from ~/.pixie/config.ini"
exports_dir=$(sed "s#^~/#$HOME/#" <<< "$exports_dir")

cfg_dir="$HOME/.pixie/exports"

function process_set() {
    local set=$1
    echo "Processing '$set'"
    source $set || die_error "Can't source '$set'"
    dir="$exports_dir/$(basename $set .sh)"
    echo "about to mkdir $dir.new"
    mkdir -p "$dir.new" || die_error "Can't make '$dir.new'"
    cd "$dir.new" || die_error "Can't cd '$dir.new'"
    for matching_file in $(tmsu files $match); do
        link_file "$matching_file"
    done
    cd - >/dev/null || die_error "Can't cd back to start dir"
    rm -rf "$dir" || die_error "Can't rm '$dir'"
    mv "$dir.new" "$dir" || die_error "Can't mv '$dir.new' '$dir'"
}

if [ -n "$1" ]; then
    process_set $1
else
    for set in "$cfg_dir"/*; do
        process_set $set
    done
fi
