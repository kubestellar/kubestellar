#!/usr/bin/env bash

# Usage: $0

# This script will write the file counts/$githash/matrix.csv, where
# $githash is $(git rev-parse --short HEAD).
# The matrix has a row per directory (with certain ones excluded).
# The matrix has the columns documented in count-directory.sh.

# This script also writes counts/$githash/sum-over-directories.csv.
# This has one line, with the same columns as the matrix except that
# the directory is replaced by two columns: one holding $githash
# and one holding the unix timestamp of the commit date of HEAD.
# Thus, the concatenation of all those sum files makes one CSV table.

bindir=$(dirname $0)
githash=$(git rev-parse --short HEAD)
cmt=$(git show --no-patch --no-notes --format=%ct $githash)
mkdir -p "counts/$githash"

# find . \( -name '.[a-zA-Z]*' -prune \) -or \( -name venv -prune \) -or \( -name bin \) -or -type d -exec ${bindir}/count-directory.sh \{\} \; > "counts/$githash/matrix.csv"

find . \( -path './.git/*' -prune \) -or -type d -exec ${bindir}/count-directory.sh \{\} \; > "counts/$githash/matrix.csv"

grep '^\.,' "counts/$githash/matrix.csv" | sed "s/.,/${githash},${cmt},/" > "counts/$githash/sum-over-directories.csv"
