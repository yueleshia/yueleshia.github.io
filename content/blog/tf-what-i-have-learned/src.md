{
  "title":   "Terraform: What I have learned",
  "date":    "2025-12-14",
  "updated": "2025-12-14",
  "layout":  "post.shtml",
  "tags":    ["opentofu", "terraform"],
  "draft":   false,
  "_": ""
}
!{|@sh ../frontmatter.sh }

!{#run: BUILD=build tetra run file % >index.smd }

# Introduction

What I have learned after fours years of being the primary contributor and maintainer.
The bulk of my experience is with Azure and AWS.
And I have experience that you probably can only get at large enterprise: projects with hundreds of modules (folders) and is multi-account (requires swapping credentials).
Here are my thoughts.

I will be assuming a working knowledge of Terraform.
Unfortunately, I do not use OpenTofu in my day job, so I cannot speak to it at the depth of this post.

# Terminology

I will refer to dev/QA/prod/etc. as (release) _phases_.
The assumption is that, by in large, you deploy the exact same infrastructure to all phases.
Typically, you purchase different accounts attached to a single tenant/organization to segregate the different phases.
In AWS, these accounts are called _landing zones_.
In Azure, these accounts are called _subscriptions_.

# Compile-time known keys

Always use compile-time know keys.
If not, a plan/apply will fail with the message: `"for_each" map includes keys derived from resource attributes that cannot be determined`.
Sometimes, it is fine on the first apply, but not on the second after one of the keys changes.
Here is an example that you can run:

```
data "external" "run_time" {
  program = ["echo", jsonencode({ a = "1" })]
}
resource "terraform_data" "example" {
  for_each = data.external.run_time.result
}
resource "terraform_data" "example2" {
  for_each = { for k, v in terraform_data.example : v.id => v }
}

```

The compile-time property is preserved transitively as in the following:

```
data "external" "run_time" {
  program = ["echo", jsonencode({ x = "a" })]
}
locals {
  source     = { a = data.external.run_time.result.x }
  transitive = { for k, v in local.source : k => 1 }
}
resource "terraform_data" "example" {
  for_each = local.transitive
}
```

# Iteration Speed

It is hard to state just how important iteration speed is.
I've seen the person hits compile/run and opens social media immediately after.
Some changes take five, if not ten, times longer when I retain details in active memory.
Let's not get started on flow state.
If you have ever done creative work, UI design, game design, debugging, exploration/research/POC work, etc., you know the difference between seeing a change in 1 second and seeing it 30 seconds.

Terraform is unfortunately rather slow at the things it should be fast at.
In terms of productivity, it makes you much faster at maintaining projects through the AWS UI.
But in terms of what it does, it is quite slow.[^napkin-math]
Fundamentally, Terraform does following:

[^napkin-math]: See https://github.com/sirupsen/napkin-math

1. It parses a DSL + runs executables to know data structure shapes (should be milliseconds)
2. It resolves constraint for dependencies, creation order, etc. (maybe should sub-second for complex trees)
3. It load balances jobs into threads and prepares API network calls (website speeds * 100 since your project is probably non-trivial)

Most 3 is inherent slowness given an online-Cloud-based model of infrastructure.
But for 1 and 2, Terraform regularly takes me 10s of seconds when I'd expect 1 second at most even for complex projects.
So, this is likely a Terraform code issue, so out of our scope.
However, for 3, we reduce of the time spent in `terraform apply` to fix loop by moving it to compile time.

## Empowering Terraform Validate

Although, not present in any Terraform education material I have ever read, there is a meaningful difference between compile time and run time.
If all your variables are

!{# @TODO: remind myself how I do admonitions }

The guiding principal is: avoid tfvars as much as possible, ideally, only enough to identify an phase.
The keys (not necessarily the values) used in for_each or count meta arguments should always be compile-time known.

To demonstrate this, let us go through an example.
Your solution deploys two microservices as containers: auth_api (service), transaction_api (service), and invoicing-generator (job); on three accounts: dev, QA, prod.

```
.
├── tfvars
│   ├── dev.tfvars
│   ├── prod.tfvars
│   └── qa.tfvars
├── accounts.yaml
├── workloads.yaml
└── main.tf
```

```yaml
# workloads.yaml
auth_api:
  type:          "ecs-service"
  load-balancer: "api"
  cluster:       "transaction"
  cpu:           { "d": 256, "q": 256, "p": 1028 }
  memory:        { "d": 512, "q": 512, "p": 2048 }
transaction_api:
  type:          "ecs-service"
  load-balancer: "api"
  cluster:       "transaction"
  cpu:           { "d": 256, "q": 256, "p": 1028 }
  memory:        { "d": 512, "q": 512, "p": 2048 }
invoicing_generator:
  type:          "ecs-task"
  cluster:       "invoicing"
  cpu:           { "d": 256, "q": 256, "p": 1028 }
  memory:        { "d": 512, "q": 512, "p": 1024}
```

```yaml
# accounts.yaml
d:
  name:       "Dev"
  phase_abv:  "d"
  account_id: 1234567890
q:
  name:       "QA"
  phase_abv:  "q"
  account_id: 2234567890
p:
  name:       "Prod"
  phase_abv:  "p"
  account_id: 3234567890
```

```
# dev.tfvars
phase_abv = "d"

# qa.tfvars
phase_abv = "q"

# prod.tfvars
phase_abv = "p"
```

```
# main.tf
variable "phase_abv" { type = string }

locals {
  accounts  = yamldecode("accounts.yaml")[var.phase_abv]
  workloads = yamldecode("workloads.yaml")
}

resource "aws_iam_role" "workload" {
  for_each = local.workloads

  name               = "aws${var.phase_abv}-role-${each.key}"
  assume_role_policy = data.aws_iam_policy_document.workload_trust
}
# Using this gives you better diffs than doing a jsonencode()
data "aws_iam_policy_document" "workload_trust" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}

resource "aws_ecs_cluster" "cluster" {
  for_each = toset(distinct([ for _, v in local.workloads : v.cluster ]))

  name = "aws${var.phase_abv}-ecscluster-${each.key}"
}

resource "aws_ecs_service" "workload" {
  for_each = { for k, v in local.workloads : k => v if contains(["ecs-service"], v.type) }

  name            = "aws${var.phase_abv}-ecsservice-${each.key}"
  cluster         = aws_ecs_cluster.cluster[each.value.cluster]
  task_definition = aws_ecs_task_definition.workload[each.key].family # No revision to use latest
  desired_count   = 1
  iam_role        = aws_iam_role.workload.name

  lifecycle {
    ignore_changes = [desired_count]
  }
}

resource "aws_ecs_task_definition" "workload" {
  for_each = { for k, v in local.workloads : k => v if contains(["ecs-service", "ecs-task"], v.type) }

  family = "aws${var.phase_abv}-ecstaskdef-${each.key}"
  container_definitions = jsonencode([
    {
      name         = each.key
      image        = "service-first"
      cpu          = each.value.cpu[variable.phase_abv]
      memory       = each.value.memory[variable.phase_abv]
      essential    = true
      portMappings = [{ containerPort = 443, hostPort = 443 }]
    },
  ])

  volume {
    name      = "service-storage"
    host_path = "/ecs/service-storage"
  }
}
```

The points I wish to highlight are:

* Keys are always all compile-time known. In addition to `plan`/`apply` needing it when resources change, it also enables `terraform validate` to start constraint resolving within a resource.

* Many details of workloads.yaml are independent of phase, fall into pattern, but still need to be configured within said pattern, e.g. different clusters for different ECS services/tasks.

* It is easy to establish and rely on a naming convention for when you have to refer to resources outside the IaC context (e.g. in CI/CD pipelines). Bonus points if you use the workloads.yaml file as the single source of truth for your CI/CD pipelines as well.

* Never include generated unique IDs (e.g. AWS security group IDs) nor phase-specific details as part of your for_each keys. e.g. Avoid `aws_security_group_rule.sg["sg-a1b2c3d4f5"]` and `aws_ecs_cluster.workload["dev_transaction_api"]`. These make refactoring (creating migration scripts) much more difficult later on.

* It makes `terraform validate` catch errors when you workload-specific resource. Like for permissions and networking, e.g.:

!{# @TODO: add other necessary permission }

```
data "aws_iam_policy_document" "transaction_api" {
  statement {
    effect    = "Allow"
    actions   = ["ecs:RunTask"]
    resources = [aws_ecs_task_definition.workload["invoicing_generator"].arn]
  }
}
```

It really helps to catch if I had made an typoed, in "invoicing_generator", or marked it as a non-ECS resource type, etc.
Now imagine you are scaling to this to hundred of microservices.
If you follow this principle of compile-time known keys, `terraform validate` catches _a lot_ more bugs, and in just seconds.

# Least-Access Privileges

When working in teams, the temptation is always get it to work first, then make it make it high quality later (i.e. never).
For given only the permissions a workload needs and nothing more, the temptation is make give it permission all, and narrow never.
Never refactoring tends to be feature of tech leadership, so we'll mark it off as out of scope.
But the piece of advice I do want to pass on is, in infrastructure work, I have found that struggling with least-access privileges is short-term loss for is faster implementation long-term.

Once you have implemented least-access privileges, you understand exactly what is necessary for every operation you do, and this understanding lessens the cognitive load.
Once you've done it once, everything tends to fall into place.
Understanding begets confidence and ability to conceptualize what you are doing.

Additionally, the your Terraform code becomes a lot more self-documenting.
Thus you can remove permissions when you deprecate features without a fear of breaking something randomly.

# Versioning

## Providers (.hcl.lock Files)

Both Terraform and provider changes can cause your tfstate to be backwards incompatible for `terraform apply` (typically because the schema for resources changes).
This typically is inconsequential and just means you have to do a `terraform init -upgrade`, but there is an off chance something breaks.
This why you probably want to commit your lock files, but in the worse case, you can always revert to an older version, `terraform state rm; terraform import`, or edit the state file directly.

## Terraform Codebase

The promise of Terraform is that your IaC is reproducible.
In theory, this means that you can git tag your Terraform codebase releases, and if you ever have to rollback software, you simply apply a previous tag.
Unfortunately, infra is not so easy.

There are two issues that arise:

* Concerns around solution uptime.
If you ever recreate resources or destroy important resources between versions, this often means that a something gets restarted, and causes downtime.
If there has been significant time between releases, any operational knowledge of specific tasks you had to do for an infra change often gets lost or is never documented.

* The deployment process often includes more than just the Terraform codebase.
Most shops with a DevOps process will have both a Terraform IaC and a CI/CD pipeline process.
Terraform often dictates the permissions for your CI/CD pipelines, and you often makes improvements to those pipelines.
Even if your pipelines are also versioned, the transition state between infra and CI/CD with their respective counterparts will probably mean something breaks.

You can treat infra deployment much the same as you do database (schema changes) deployments.
In database deployments/rollbacks, you lose data if a table or column goes away that affects ever row.
Thus you often have to treat infra and database rollbacks as a new release, depending on the specifics of what is being lost of course.

# Codebase File Layout

The main guiding principal, is single source of truth.
For a given folder (i.e. the applying module), the same *.tf files should serve all phases (dev, prod, etc.).
The entire idea of different accounts or release phases for reproducibility.
If you have a account-specific resource the classic trick is `count = var.phase_abv != 'p' ? 0 : 1` or a similar thing with for_each.

How to split partition your project into subfolders has two aspects to it:
* deployment constraints, i.e. are these set of resource deploy at the same cadence as each other, does one lead the other etc.
* `terraform apply` times. How parallelized vs simple you wish your `terraform apply` pipeline to be

An example for deployment concerns is invoicing might depend updates to transaction_api and transaction_api might update frequently, so might put them in separate folders to maintain their separate lifecycles.
A central `workloads.yaml` shared across both is quite nice.
Good old [Conway's Law](https://en.wikipedia.org/wiki/Conway%27s_law) might be a guiding light.

For apply times, note that GitHub runners have a timeout of 1 hour.
You should design your automation workflow around making it deterministic and removing as many manual steps as possible.

## Infra-based vs Functional-based Hierarchy

I am not a fan of project layouts that separate based off infrastructure resource types (stacks) like [this](https://spacelift.io/blog/iac-architecture-patterns-terragrunt).
(Terragrunt is equivalent to Terraform for the purposes of this section.)

```
.
├── modules
│  ├── app
│  │   ├── main.tf
│  │   ├── outputs.tf
│  │   └── variables.tf
│  ├── mysql
│  │   ├── main.tf
│  │   ├── outputs.tf
│  │   └── variables.tf
│  └── vpc
│      ├── main.tf
│      ├── outputs.tf
│      └── variables.tf
└── env
    ├── global.tfvars
    ├── dev
    │   ├── app
    │   │   └── terragrunt.tfvars
    │   ├── mysql
    │   │   └── terragrunt.tfvars
    │   └── vpc
    │       └── terragrunt.tfvars
    ├── prod
    │   ├── app
    │   │   └── terragrunt.tfvars
    │   ├── mysql
    │   │   └── terragrunt.tfvars
    │   └── vpc
    │       └── terragrunt.tfvars
    └── qa
        ├── app
        │   └── terragrunt.tfvars
        ├── mysql
        │   └── terragrunt.tfvars
        └── vpc
            └── terragrunt.tfvars
```

* The specific infra you use, is more of a developer concern (based on the how to implement the technical requirements) rather than an infra concern.
When business needs inevitably dictate that you have to change a container into a VM or anything else, this makes transitioning between the two more challenging.
Files farther apart are conceptually more distance by the principal of [locality of behaviour](https://htmx.org/essays/locality-of-behaviour/). 
And doing a migration would require `terraform import; terraform state rm` in different tfstates vs a `terraform state mv` within single tfstate.
Import syntax is always more painful because it is different depending on the specific resource.

* This creates a lot of folder sprawl.
* When two workloads have to reference each other (as in roles, event-based message passing systems like SNS/SQS, security groups, etc.) it you are forced to communicate via Terraform's output system and/or the [terraform_remote_state](https://developer.hashicorp.com/terraform/language/state/remote-state-data) data source.
This creates a lot of unnecessary boilerplate code because you have to assign names to outputs when you could have just done a direct reference if they were in the same folder, e.g. `aws_iam_role.workload.name`.

I would much prefer the following:

```
.
├── modules
│   ├── globals.tfvars
│   ├── networking
│   │   ├── outputs.tf
│   │   ├── variables.tf
│   │   └── vpc.tf
│   └── use_case_1
│      ├── app.tf
│      ├── mysql.tf
│      ├── outputs.tf
│      └── variables.tf
└── env
    ├── networking
    │   ├── dev.tf
    │   ├── qa.tf
    │   └── prod.tf
    └── use_case_1
        ├── dev.tf
        ├── qa.tf
        └── prod.tf

```

## Takeways

It is difficult to give prescriptions for all situations as different projects call for different approaches.
Though I think a functional-based (i.e. operational/business-based) approach is more sensible, the actual guiding principals are:

* Single source of truth: single IaC codebase for all phases.
* Split modules based on automation needs
* Split modules based on deployment needs

# Refactoring

Refactoring is an essential part of codebase maintenance.

There are four ways of refactoring:

* Destroy and recreate, this is fine for operationally-low-impact resources.
* `terraform state mv` is the ideal way method.
* `terraform state import` for is necessary for things you were not previously managing with terraform
* Editing the state file directly when all else fails.[^state-file-refactor]
The tfstate is a JSON file.
If you do make changes, you will have to add one to the `serial` field before .

[^state-file-refactor]: I have had to do this for [mongodb](https://registry.terraform.io/providers/mongodb/mongodbatlas/latest/docs/resources/database_user.html) when the restrictions around the naming made it impossible to import.

When you have to do several imports, it is much faster to do them on a local file.
My favourite workflow is a variant of following:

```sh
#migrate.sh
export AWS_PROFILE=develop

main() {
  if true; then
    echo 'terraform { backend "s3" }' >BACKEND.tf
    terraform init -reconfigure
    terraform state pull >terraform.tfstate
    echo 'terraform { backend "local" }' >BACKEND.tf
    terraform init -reconfigure
  fi
  echo 'terraform { backend "local" }' >BACKEND.tf

  terraform state mv ... 
  #IS_LOCAL=true terragrunt state mv ... 

  echo 'terraform { backend "s3" }' >BACKEND.tf
  # This init will ask you if you wish to migrate
  # Saying yes here is equivalent to doing `terraform state push terraform.tfstate`
  terraform init -reconfigure
}

main "$@"
```

I like to use Terragrunt to generate BACKEND.tf, if I provide the environment variable `IS_LOCAL=true`.
Ideally, you want to do this for all phases.

# Destory Order

Sometimes you have to specify the [depends_on](https://developer.hashicorp.com/terraform/language/meta-arguments/depends_on) meta argument to specify destroy order.

# A Better Terraform

There are five pain points I have with Terraform.

* The main thing that holds back Terraform is that the terraform tfstate cannot be configured, which makes configuring auth for a multi-phase solution more onerous.
[Terragrunt](https://github.com/gruntwork-io/terragrunt) resolves this.

* The constraint solver of Terraform is not as powerful as it could be. @TODO: find a good example.

* Providers are not sandboxed. They are fullblown executables.

* The Providers are under-docuemnted so writing one in any language other than Go is difficult.
But it seems things [are changing](https://github.com/aneoconsulting/tf-provider) with the fork of OpenTofu.

* Terraform validate is kind of slow for what it does.

* There are no user-space functions in Terraform, so data wrangling tends require a lot of repeating yourself.
I typically resort to using configuration languages, my favourite of which is [nickel](https://github.com/tweag/nickel?tab=readme-ov-file#comparison).

* I would like a `lifecycle { never_provision }`for resources that I wish Terraform never provisions.
For example, in enterprise, I do not have permission to create repositories, so I would like to represent that without resorting to local-provisioner workarounds.

