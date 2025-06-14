#!/bin/sh

# run: ../pages/make.sh %
#run: % server

# Print 'a' to prevent `$( )` from stripping whitespace
wd="$( dirname "${0}"; printf a )"; wd="${wd%?a}"
cd "${wd}" || exit "$?"

compile() {
  mkdir -p hugo/content/blog
  printf %s\\n final/*.adoc | IS_LOCAL="${1}" parallel --will-cite --color-failed '
    path={}
    relpath="${path#final/}"

    printf %s\\n "=== ${relpath} ===" >&2
    cd "final" || exit "$?"
    tetra parse "${relpath}" >"../hugo/content/blog/${relpath}" || exit "$?"
  '
}

make() {
  case "${1}"
  in local)
    compile true

  ;; server)
    compile false

  #;; pages)
  #  git reset --hard main

  #  make "server"

  #  git add .
  #  git commit --message "Publish $( date +"%Y-%m-%d" )"
  #  git push --force origin pages

  ;; *)
    printf %s\\n "${0}: Unknown command '${1}'" >&2; exit 1

  esac
}

if [ -z "$*" ]; then
  make "server"
else
  for x in "$@"; do
    make "${x}"
  done
fi
