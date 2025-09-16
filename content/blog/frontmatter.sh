#run: sh -c "cd tf-intro; BUILD=build <src.md % "
frontmatter="$( <&0 jq --slurp --raw-input '.
  | .[0:index("\n}\n") + 2]
  | fromjson
' )" || exit "$?"

case "${BUILD:-local}"
in local)
  printf %s\\n "${frontmatter}" | jq --raw-output '[
    "# \(.title)",
    "Author: \(.author)",
    ""
  ] | join("\n")'
;; build)
  wd="$( pwd )" || exit "$?"
  printf %s\\n "${frontmatter}" | jq --raw-output --arg wd "${wd##*/}" '[
    "---",
    ".title  = \(.title | tojson),",
    ".date   = @date(\(.date | tojson)),",
    ".author = \( (.author // "Yueleshia") | tojson),",
    ".layout = \(.layout | tojson),",
    ".tags   = \(.tags | tojson),",
    ".draft  = \(.draft),",
    ".custom = {",
    "  .id = \($wd | tojson),",
    "  .series = \((.series // null) | tojson),",
    "  .updated = @date(\(.updated | tojson)),",
    "}",
    "---",
    ""
  ] | join("\n")'
;; test)
  printf %s\\n "${frontmatter}" >&2
esac
