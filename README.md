# aws-name-asg-instances

Automatically apply Name tags to instances in an ASG based on their custom
tags.

Names that appear in the AWS console beside EC2 instances are set by
creating a special tag `Name`.  When members of an ASG start, they are not
given any names, so often what happens is that instance is allowed to
name itself inside of user data.  However, if user data fails for any
reason, the instance will never name itself and it can be difficult to
find the instance's purpose in the AWS console.

The issue with an instance naming itself is that the instance profile (and
underlying IAM role) provides the instance with the `ec2:CreateTags`
permission, which cannot have a scoped `Resource` declaration.  This
violates least privilege and provides the instance with the ability to
create (and overwrite) tags on any instance in the same AWS account.

This project creates a CloudWatch Event rule that watches for AutoScaling
events, specifically the successful launch of new EC2 instances, and names
them based on their tags.  Thus, only the Lambda function that backs the
CloudWatch Event rule has the abiility to name EC2 instances, and only in
a specific format.

## Costs

The resources created under this CloudFormation template will cost either
very little or nothing.  The only element that costs anything is the
Lambda function, and Amazon has a generous free tier that should cover
just about everyone's use case for this tool, making it free to run.

## Naming Format

The instances are named based on the following convention:

```
<project>-<environment>-<instance_id>
```

The tags `project` and `environment` must be available on the instance and
given a non-empty string value.  The `instance_id` is already known by the
auto-scaling group during launch, so you do not need to provide it.

The `instance_id` is stripped of its `i-` prefix, leaving only the unique
ID.

The resulting name is then limited to 255 characters, as that is the
limit of tag values.

An example of this is, using a project `donny` and environment `staging`
is:

```
donny-staging-029d0202d1a
```

## Project Requirements

* Amazon Web Services account
* Permissions to create AWS resources:

  Specifically: CloudFormation, CloudWatch Events, Lambda, IAM roles

## Launching the Stack

The stack must be launched in any region where auto-scaling groups are
used and you want to name its members.  However, CloudWatch Event rules
may not be available in every region, so the following Ansible playbook
ensures that the stack is launched only in the regions where all AWS
services are supported.

The project is launch in 4 phases:
1. Deploy the IAM stack (this happens only once -- globally)
1. Deploy the Event stack (this happens only once -- per region)
1. Prepare the Event stack for Go code (this happens only once -- per region)
1. Deploy the Go code (this happens at least once -- per region)

Run `make help` to see how to deploy the project in each phase.

**Important**: Each phase of the stack requires both an AWS region and an AWS profile to be specified.  The given profile can be different IAM roles depending on your organization's structure (for instance, if administrators are only allowed to launch IAM resources).

## Changes between v1.x and v2.x

This project changed drastically between v1.x and v2.x in the following ways:

* The orchestration by Ansible was removed and replaced with a Makefile.

  Ansible is heavyweight for the type of orchestration required by this
  project.  The project's CloudFormation infrastructure was retooled so
  that it is longer-lived and less changes are required.  For instance,
  the Lambda code is the only phase that changes often, so optimization
  was focused there instead of updating the CFN stack constantly.  As
  the infrastructure didn't change often, the Ansible orchestration no
  longer seemed necessary and could be replaced with a simple Makefile.

* The Lambda function code was converted from Python to Go.

  The Go code is faster to run, saving time and money, and enforces
  type checking, something which Python consistently did not meet. The
  Go code abstracts the JSON payloads by converting them to structures,
  which Python cannot handle and consistently requires dictionary key
  checks, obfuscating the code.  Lastly, tests could be easily added
  the handle the business logic (versus the type checking) required by
  the project.

* The stacks were split between global IAM and regional Event stacks.

  In v1.x, the project created a separate IAM role in each region where
  the event resources were created.  This created too many IAM resources
  can AWS limits could be hit, which impacts other teams.  It also looks
  messy and isn't needed if there is some prior dependency checking. The
  IAM stack is now deployed to a single region, since it is global, and
  then the ARN of the role is passed to each event stack, which is
  regionally deployed, to complete the process.  This also means
  different teams (administrators, developers) can deploy the different
  parts of the project, enforcing any organization security constraints.

## License

tl;dr MIT license.

Please read [LICENSE](LICENSE) to view the license for this project.

