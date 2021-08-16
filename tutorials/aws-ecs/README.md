# migrator on AWS

This is a walkthrough how to setup migrator on AWS in a secure way.

This is not a beginner level tutorial nor Jeff Barr style of blog post ;) It is assumed that you are already familiar with the following AWS services: IAM, ECS, ECR, Secrets Manager, RDS, and S3.

The goal of this tutorial is to deploy migrator on AWS ECS, load migrations from AWS S3 and apply them to AWS RDS DB while storing env variables securely in AWS Secrets Manager.

## The setup

There is a sample `migrator.yaml` provided in this folder. Please edit your S3 bucket name and follow the steps below.

Everywhere you see {region} or {aws_account_id} replace it with your values. If you see {version} change that to migrator version you use.

S3, Secrets Manager, ECR and the whole ECS cluster together with RDS database are all located in the same region.

It is assumed that v4.1.2+ is used (this tutorial was added in v4.1.2). v4.1.2 added support for `pathPrefix` which is used by Application Load Balancers/Application Gateways for application HTTP request routing. In the sample migrator config provided it is set to `/migrator`.

Finally, in steps below I stick to the AWS managed policies and AWS wizards to not complicate the whole setup and focus on the idea and the architecture of the solution.

## IAM

We need two roles which will be assumed by our task execution and later migrator itself.

Navigate to IAM page and launch create new role wizard. From "Select type of trusted entity" select AWS service and from the services list select "Elastic Container Service". Then from available options select "Elastic Container Service Task".

Create first role: `migrator-ecs` and assign 2 AWS managed policies: `SecretsManagerReadWrite` (there is no `ReadOnly` policy and we want our task to successful read secrets from AWS Secrets Manager), and a required policy `AmazonECSTaskExecutionRolePolicy` (otherwise task will not be able to pull images, publish metrics and logs).

Now, create second role: `migrator-task` and assign the following AWS managed policy: `AmazonS3ReadOnlyAccess` (our source migrations are stored in AWS S3 and that's the only permission migrator needs).

Above roles are fine in dev environment. Remember that when using migrator in production always apply least privileges rule.

## ECR

First we need to bake in `migrator.yaml` into official migrator image and store it securely in AWS ECR.

The user used for pushing the image should have access to AWS ECR. In dev environment you can use AWS managed `AmazonEC2ContainerRegistryFullAccess` policy.

If you don't have ECR already go ahead and create a new one (I recommend enabling both `Tag immutability` and `Scan on push`).

AWS ECR cli can generates docker login command for you. It generates output which is almost copy & paste ready. The only bit needed is to remove the `-e none` from generated command:

```sh
aws ecr get-login
```

Create docker image and push it to AWS ECR:

```sh
docker build --tag {aws_account_id}.dkr.{region}.amazonaws.com/migrator:{version} .
docker push {aws_account_id}.dkr.{region}.amazonaws.com/migrator:{version}
```

## ECS

Navigate to AWS ECS start page. Select "Create Cluster" and then from "Select cluster template" select "Networking only - Powered by AWS Fargate". Make sure you tick the VPC option (we will use this VPC for AWS RDS too).

That's all for now, let's switch to AWS RDS.

## AWS RDS

AWS RDS will be our database.

Navigate to AWS RDS start page and provision a new AWS RDS database.

When provisioning don't forget to run it in the same VPC as the ECS cluster. Ask AWS to generate the admin password. Copy the credentials. Unfold the "Additional connectivity configuration". Make sure DB will not have a public IP assigned, create a new Security Group for our DB. Call it `database`. For now, leave the inbound rules empty (ECS services does not exist yet).

## AWS Secrets Manager

If you take a look at the provided migrator config file it requires the following 4 env variables to be present:

- DATABASE_USERNAME
- DATABASE_PASSWORD
- DATABASE_NAME
- DATABASE_HOST

All above can be stored in AWS Secrets Manager DB credentials type of secret. It is a JSON which contains all of the above information. However, AWS ECS cannot at present parse & inject JSON secrets: [[ECS] [secrets]: support parameter name from AWS Secrets Manager #385](https://github.com/aws/containers-roadmap/issues/385).

Instead we will create 4 plain text secrets (important to click on the plain text tab not the default key/value JSON one). Navigate to AWS Secrets Manager starting page. Click "Store a new secret" and then "Other type of secrets". Then create 4 plaintext secrets call them: `migrator/test/DATABASE_USERNAME`, `migrator/test/DATABASE_PASSWORD`, etc.

## ECS Task Definition

Now let's create the task.

Navigate to ECS page. On the left there are Clusters and Task Definitions sections. Click "Task Definitions" and then "Create new Task Definition".

On the next page select AWS Fargate as launch type.

"Task Role" is the role which tasks assumes when running. migrator will assume it to access S3 source migrations. It needs `migrator-task` role.

"Task execution IAM role" is the role used by AWS to pull container images and publish container logs to Amazon CloudWatch. AWS provisions default role `ecsTaskExecutionRole` but it doesn't have permissions to fetch secrets so it is important to change that to `migrator-ecs` role.

"Task size" migrator has a small memory footprint and quite low CPU expectations. Any combination with 1 GB memory will do just fine.

Under the "Container Definitions" click "Add container". It's a simple step, provide values only for the required fields, image is our ECR image: `{aws_account_id}.dkr.ecr.{region}.amazonaws.com/migrator:{version}`. Port mapping is `8080`. Then scroll down to "Environment/Environment variables" section and add 4 env variables using the ARNs of the 4 secrets we created above. Remember to use "valueFrom" when key is the ARN reference. Ignore the rest of the fields (these are more advanced container settings which we don't need).

Click "Add" to add container definition and then "Create" to create task definition.

## ECS Service

Now navigate to ECS cluster. Under the services select "Create".

Select launch type: AWS Fargate. As a task definition use the one which we just created. Populate required field and hit "Next step".

In the "VPC and security groups" section select our cluster VPC and all its subnets. There is a security group auto created. Click edit next to it. First change its name to more meaningful name like `migrator-service`. This SG allows traffic on port `80` from anywhere. Remove this rule. We need a more strict inbound rule to only allow HTTP traffic from Application Load Balancer. Load balancer is created in next step... We will get back to `migrator-service` SG later.
In order for our service to be able to connect to Internet gateway make sure "Auto-assign public IP" is set to enabled. ECS Fargate template cluster creates VPC with public subnets and we need to assigning public IP to our service to have internet connection. The fact we are assigning public IP to our service is also a reason why we need a more strict inbound rule in our SG.

In "Load balancing" section select Application Load Balancer. Select the existing or create a new one.

If you don't have an existing ALB you need to create one. There is a link which opens new tab with EC2 Load Balancer Wizard. Click it. In the wizard select scheme: internet-facing, for listeners select HTTPS 443. In Availability zones section select our ECS VPC and select all its AZs.
In next step (as we selected HTTPS) you need to select a valid certificate or create a new one.
In next step, create new Security Group. Call it `frontend` and add HTTPS allow rules for your IP addresses.
Next step is "Configure routing". This step is required and creates a default routing target for path prefix `/`. This routing is required and we have to create it, leave all fields as they are. Call this target group default and hit "Next". Next step is to register targets for the default group we just created. As we don't really want default target and are only interested in having ALB for migrator we will skip this step. Last step is the review one, if everything is fine, click "Create".

Close the tab and go back to the create service wizard. Next to load balancers drop down there is a refresh icon. Hit it and select the newly created ALB. It will automatically populate container to load balance. Click "Add to load balancer" to review and confirm all the options. For "Production listener port" select "443:HTTPS" and for "Target group protocol" select "HTTP". AWS will pre-populate new migrator target group and provide default options. For example the `/migrator` prefix in path pattern. ALB can usually serve a number of applications and AWS adds this application/service prefix automatically for every target group. Set evaluation order for migrator (if new ALB set it to 1). If you want to have a dedicated ALB for migrator you can use `/` as path prefix but then you need to remove `pathPrefix` from `migrator.yaml` (and push new image to ECR and update Task Definition).
Important: set healthcheck to `/migrator/` (AWS pre-populated it as `/migrator`). The thing is that for `/migrator` migrator responds with 301 redirect to `/migrator/` and HTTP 301 doesn't count for ALB as a healthy HTTP response code (makes sense).

Service discovery step - accept defaults. Auto Scaling step is optional and migrator doesn't need auto scaling. Skip it.

Last step is the review. If all looks good, click "Create Service".

## Updating security groups

We created 3 SGs:

- `database` - edit it and add inbound allow traffic to DB port from `migrator-service` SG
- `migrator-service` - edit it and add inbound allow traffic to `8080` from `frontend` SG
- `frontend` - review that inbound rules have allow traffic to port `443` from you IP addresses

## Accessing migrator

Once this step is done, the migrator should be up and running. From AWS console ALB view you can copy the DNS name and then try the following URLs:

```sh
curl -v -k https://migrator-alb.{region}.elb.amazonaws.com/migrator/
curl -v -k https://migrator-alb.{region}.elb.amazonaws.com/migrator/v1/config
```

Check if migrator can load migrations from S3 and connect to DB:

```sh
curl -v -k https://migrator-alb.{region}.elb.amazonaws.com/migrator/migrations/source
```

When you're ready apply migrations:

```sh
curl -v -k -X POST -H "Content-Type: application/json" -d '{"mode": "apply", "response": "list"}' https://migrator-alb.{region}.elb.amazonaws.com/migrator/v1/migrations
```

Enjoy migrator!
