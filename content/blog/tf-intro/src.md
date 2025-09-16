{
  "title":   "Terraform: Introduction",
  "date":    "2025-06-11",
  "updated": "2025-09-21",
  "layout":  "post.shtml",
  "tags":    ["archive", "text"],
  "draft":   false,
  "_": ""
}
!{|@sh ../frontmatter.sh }

OpenTofu is maintained by the Linux Foundation and open-source ([MPL 2.0](https://github.com/opentofu/opentofu/blob/main/LICENSE)) fork of Terraform in response to HashiCorp changing Terraform's software license from the MPL to the BSL[^1].

[^1]: [Updating the license from MPL to Business Source License](https://github.com/hashicorp/terraform/commit/b145fbcaadf0fa7d0e7040eac641d9aef2a26433)

# Introduction: The Advent of the Cloud

Imagine you are trying make a music streaming service.
You will need a website (HTML/CSS/JS), a backend, and a database to store music.
You buy a computer, install your application, install MySQL, configure MySQL, configure your internet router, and install some security like a firewall.
To make changes, you just connect a monitor, keyboard, and mouse.
Maybe you'll need hire a sysadmin to know how to deal with hardware failures like a faulty hard drive disk.

As serve customers, you find your double the CPU but half the RAM, but repurchasing hardware sounds like a waste.
Enter hardware as a service (HaaS).
You rent a VPS that meets your CPU/RAM needs, letting others use your "unused" resources.
You no longer need to manage hardware, but now you are picking up sysadmin skills as SSH and manage your software install through the terminal.
Maybe you'll start writing some bash scripts to automate installing and updating.

But now, your application load is high enough that you want to setup a read replica of your database on a second VPS, and have a second application talk to it.
But it's a real drag to now manage two SSH keys and swap back and forth to do all the setup.
Enter the cloud where hardware provisioning is now exposed through a web interface or through an HTTP-accessible API.
You want to double your database size?
Just click that button there.

Oh wait, there's an REST API to manage hardware?
Now you are a DevOps engineer using OpenTofu/Terraform to provision infrastructure.

# What is OpenTofu/Terraform

OpenTofu is one solution to the problem of: I have a recipe for my infrastructure, but sometimes I (or someone on my team) tweak it to experiment, but I hit undo or save those changes.
At its core, OpenTofu is a CLI tool that reads a recipe written in HCL (HashiCorp Configuration Language) that specify CRUD (Create, Read, Update, Delete) actions, and then orchestrates said actions as REST APIs calls.

If you were to reimplement OpenTofu you would do the following:

1. Parse your HCL recipe files
1. Fetch the previous state to see what to delete if a resource was removed from the recipe
1. Calculate a directed acyclic graph of what CRUD tasks need to be run. Read first, then Create/Update/Delete as directed by the recipe
1. Call the CRUD APIs as necessary, orchestrating them in parallel if possible
1. Save your new state

Here is a sample recipe:

```
#hcl
resource "digitalocean_droplet" "vps" {
  image   = "ubuntu-20-04-x64"
  name    = "sample-vps-1"
  region  = "nyc2"
  size    = "s-1vcpu-1gb"
}
```

Given, OpenTofu supports several services, you might ask: who is doing the work to map HCL concepts to the REST APIs that for these services?
So yes, there is a package registry to which anyone can publish that HashiCorp (for Terraform) or the Linux Foundation (for OpenTofu) are maintaining.
In HCL parlance, these are known as providers.
Providers are full-on compiled executables that are run on your local machine.


# HCL Basics

In HCL, the following are the core keywords:

* `variable`: Inputs (essentially JSON format)
* `output`: Outputs (essentially JSON format)
* `resource`: A full CRUD resource
* `data`: Meant for only reading a value without changing it.
* `module`: A directory of *.tf files.
This a single compilation block.
File seperation is just for the programmer.
Effectively, all *.tf files are combined into single file.
These can be imported from other modules and function like a resource.

HCL has a standard library of functions, list comprehensions, and object comprehensions but no loops.
So the base language is not Turing complete, but there are several escape hatches to execute code.
There are additional concepts like locals (variables) and provider functions.

Here is a sample code snippet that checks out repos based on what is in `local.repos`.

```
#hcl
variable "unused" { type = string }

locals {
  base_dir = ".."
  repos = {
    "core"     = "yueleshia/tetra"
    "markdown" = "yueleshia/tetra-markdown"
    "typst"    = "yueleshia/tetra-typst"
    "site"     = "yueleshia/site"
  }
}

#data "github_repository" "repo" {}  # This is a comment

resource "terraform_data" "repo" {
  for_each = local.repos

  triggers_replace = each.value

  provisioner "local-exec" {
    command = <<-EOT
      dir="${base_dir}/${each.value}"
      [ -d "$dir" ] && { printf %s\\n "The path '${each.value}' already exists" >&2; exit 1; }
      mkdir "$dir" || exit "$?"
      git -C "$dir" init _bare || exit "$?"
      git -C "$dir/_bare" remote add origin "git@github.com:${each.value}.git" || exit "$?"
      git -C "$dir/_bare" config --local core.sshCommand "ssh -i ~/.ssh/github" || exit "$?"

      git -C "$dir/_bare" switch --create _bare
      git -C "$dir/_bare" fetch  --depth 1
      git -C "$dir/_bare" worktree add "../main" "origin/main" || exit "$?"
    EOT
  }

  provisioner "local-exec" {
    when = destroy
    command = <<-EOT
      [ -d "${self.triggers_replace}" ] && rm -r "${self.triggers_replace}"
    EOT
  }
}
```

