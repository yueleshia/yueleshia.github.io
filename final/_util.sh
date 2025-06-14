#!/bin/sh

# run: IS_LOCAL=true % link tf-intro.adoc
#run: IS_LOCAL=false % header tf-intro.adoc "date = 2000-01-01"

cmd="${1}"
id="${2}"

case "${cmd}"
in link)  nickel export "_series.ncl" --field "adoc.\"${id}\".link" --format raw
;; header)
  export extra="${3}"

  nickel export _series.ncl \
    --field "adoc.\"${id}\".header" \
    --format raw \
    -- \
    --override is_local="${IS_LOCAL:-false}" \
    --override extra=@env:extra \
  # end
;; *)   printf %s\\n "Unsupported command '${cmd}'" exit 1
esac
