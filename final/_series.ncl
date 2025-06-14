let remove = fun suffix id =>
  let len = std.string.length id in
  id |> std.string.substring (len - (std.string.length id)) len
in

{
  is_local    | Bool   | default = false,
  path_prefix | String | default = "",
  extra       | String | default = "",

  author = "Yueleshia",

  files = {
    "tf-intro.adoc"       = "Terraform/OpenTofu for Devs: An Introduction",
    "tf-secrets.adoc"     = "TF for Devs: Managing Secrets Pre-Ephemeral",
    "tf-config.adoc"      = "TF for Devs: Managing Configuration at Scale",
    "tf-enterprise.adoc"  = "TF for Devs: Scaling to Enterprise",
    "tf-advanced.adoc"    = "TF for Devs: Advanced Tips and Tricks",

    "table-driven-lexing.adoc" = "Table Driven Parsing",
  },

  by_series = {
    tf_development = {
      name =  "OpenTofu for Devs",
      list = [
        "tf-intro.adoc",
        "tf-secrets.adoc",
        "tf-config.adoc",
        "tf-enterprise.adoc",
        "tf-advanced.adoc",
      ],
    },
  },

  by_file =
    files
    |> std.record.map (fun path v => {
      title       = v,
      "%{"path"}" = path,
      series_list =
        by_series
        |> std.record.fields
        |> std.array.filter (fun id => by_series."%{id}".list |> std.array.any ((==) path))
      ,
    })
  ,

  adoc = by_file |> std.record.map_values (fun file => {
    link   =
      if is_local
      then "link:%{file.path |> remove ".adoc"}.html"
      else m%"link:{{< ref "%{file.path |> remove ".adoc"}" >}}"%
    ,
    series = file.series_list
      |> std.array.map (fun series_id => by_series."%{series_id}")
      |> std.array.map (fun series => m%"
        This post is part of a series __%{series.name}__:

        %{
          series.list
          |> std.array.map (fun path =>
            let title = files."%{path}" in

            if path == file.path
            then ". %{title}"
            else ". %{adoc."%{path}".link}[%{title}]"
          )
          |> std.string.join "\n"
        }
      "%)
    ,
    frontmatter =
      if is_local
      then "= %{file.title}"
      else m%"
        +++
        title  = '%{file.title}'
        author = '%{author}'
        %{extra}
        +++
      "%
    ,
    header = m%"
      %{frontmatter}

      :toc:

      :source-highlighter: rouge
      :rouge-style: magritte

      %{
        if std.array.length series == 0
        then ""
        else "====\n%{series |> std.string.join "\n"}\n====\n"
      }
    "%,
  }),
} | {
  files | {
    _ | String
  },
  by_series | {
    _ | {
      name | String,
      list | std.contract.all_of [
        Array String,
        Array (std.contract.from_validator (fun path =>
          if files |> std.record.has_field path
          then 'Ok
          else 'Error { message = "%{path} is not present in 'files'" }
        ))
      ],
    }
  },
  ..
}
