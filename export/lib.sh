#!/bin/bash

die_error () {
    echo "ERROR: $@"
    exit 2
}

filename () {
    # some things (ex. facebook) look at filenames, not exif timestamps.
    # so all pictures must be ordered by time irrespective of original
    # filename or src dir, we can easily do this by divising
    # appropriate filenames based on exif time.
    local file=$1
    local file_base="$(basename "$file")"
    # the tail is needed, because in some rare cases it returns more than 1 date, i know of one example:
    # fotos/originals/dieter-s3/extSdCard_DCIM/Camera/20121228_142756-1.jpg
    # and in that case it seems the 2nd date is most appropriate.
    local fname=$(exiv2 -g Exif.Image.DateTime -P v "$file" | tail -n 1 | tr ' ' '_')_"$file_base"
    echo "$fname"
}

# http://stackoverflow.com/questions/2564634/bash-convert-absolute-path-into-relative-path-given-a-current-directory
relpath(){ python2 -c "import os.path; print os.path.relpath('$1','${2:-$PWD}')" ; }

link_file () {
    local file=$1
    fname="$(filename "$file")"
    [ ! -e "$fname" ] || die_error "'$fname' already exists in '$(pwd)'? wanted to link to '$file'"
    # unfortunately, there is no way to make a `ln -r -s` link that doesn't dereference the target.
    # the target itself can be a symlink to somebody else, we wish to symlink to the target, using a relative path.
    echo file $file
    relpath "$file"
    echo fname $fname
    ln -s "$(relpath "$file")" "$fname"
}

