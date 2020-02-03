# migrator on AWS EKS

The goal of this tutorial is to deploy migrator on AWS EKS with Fargate IAM integration, load migrations from AWS S3 and apply them to AWS RDS DB.

In all below commands I use a cluster name of `awesome-product` and `ap-northeast-1` region. Everywhere you see {aws_account_id} replace it with your AWS account ID.

## Fargate with IAM

In this tutorial I'm using AWS Fargate with IAM. This is a new feature and as of the time of writing this tutorial there were only 4 AWS regions supported.

I tried to used the following helm charts to simplify the whole deployment but I couldn't get it working with Fargate and IAM Roles for Service Accounts (IRSA). I will try to revisit them in the future, so before you start the below tutorial be sure to check out their latest versions as they may ease your life (and if you do this don't forget to send me a pull with amended tutorial):

* https://github.com/helm/charts/tree/master/incubator/aws-alb-ingress-controller - issue to watch: [[incubator/aws-alb-ingress-controller] Fargate and IRSA: Not authorized to perform sts:AssumeRoleWithWebIdentity](https://github.com/helm/charts/issues/20504)
* https://github.com/godaddy/kubernetes-external-secrets - issue to watch: [Fargate Support](https://github.com/godaddy/kubernetes-external-secrets/issues/254)

## S3 - upload test migrations

Create S3 bucket in same region you will be deploying AWS EKS. Update `baseDir` property in `migrator.yaml`.

You can also use test migrations to play around with migrator:

```
cd test/migrations
aws s3 cp --recursive migrations s3://your-bucket-migrator/migrations
```

## ECR - build and publish migrator image

You can find detailed instructions in [contrib/aws-ecs-ecr-secretsmanager-rds-s3](../aws-ecs-ecr-secretsmanager-rds-s3).

```
aws ecr get-login --region ap-northeast-1
docker build --tag {aws_account_id}.dkr.ecr.ap-northeast-1.amazonaws.com/migrator:v4.1.2 .
docker push {aws_account_id}.dkr.ecr.ap-northeast-1.amazonaws.com/migrator:v4.1.2
```

Don't forget to edit `migrator-deployment.yaml` and update the image name to the one built above (line 22).

## Create and setup the cluster

There is no AWS managed policy which would allow creation of EKS clusters. You need to explicitly add `eks:*` permissions to the user executing the create cluster command. Mind that we are using fargate profile (last param).

```
eksctl create cluster \
--name awesome-product \
--version 1.14 \
--region ap-northeast-1 \
--external-dns-access \
--alb-ingress-access \
--full-ecr-access \
--fargate
```

It will take around 15 minutes to complete. Make sure kubectl points to `awesome-product` cluster:

```
kubectl config current-context
```

Create an IAM OIDC provider and associate it with your cluster (prerequisite for IAM integration):

```
eksctl utils associate-iam-oidc-provider \
    --region ap-northeast-1 \
    --cluster awesome-product \
    --approve
```

## ALB

As mentioned at the top, you may review latest version of [incubator/aws-alb-ingress-controller](https://github.com/helm/charts/tree/master/incubator/aws-alb-ingress-controller) and see if Fargate and IAM are now working.

If not, proceed with the below instructions.

### IAM

We need to create Kubernetes service account for the ALB ingress controller. Code is available at kubernetes-sigs/aws-alb-ingress-controller:

```
kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/aws-alb-ingress-controller/v1.1.5/docs/examples/rbac-role.yaml
```

Next, we need to create IAM policy allowing the ingress controller to provision the Application Load Balancer for us. Again, we will use code available at kubernetes-sigs/aws-alb-ingress-controller:

```
aws iam create-policy \
    --policy-name ALBIngressControllerIAMPolicy \
    --policy-document https://raw.githubusercontent.com/kubernetes-sigs/aws-alb-ingress-controller/v1.1.5/docs/examples/iam-policy.json
```

Next, we need to create the service account. Please change the policy ARN to the one created above (if you used same policy name then just update {aws_account_id}).

```
eksctl create iamserviceaccount \
    --region ap-northeast-1 \
    --name alb-ingress-controller \
    --namespace kube-system \
    --cluster awesome-product \
    --attach-policy-arn arn:aws:iam::{aws_account_id}:policy/ALBIngressControllerIAMPolicy \
    --override-existing-serviceaccounts \
    --approve
```

### ALB ingress controller

Open and edit `alb-ingress-controller.yaml` and update 3 things:

* cluster name - line 39
* AWS region - line 43
* EKS VPC - line 48

Finally, create the ingress controller itself:

```
kubectl apply -f alb-ingress-controller.yaml
```

### S3

migrator needs to connect to S3 to read source migrations. We will create IAM service account for it called `migrator-serviceaccount`.

As a policy I will use the AWS-managed `AmazonS3ReadOnlyAccess`:

```
eksctl create iamserviceaccount \
  --region ap-northeast-1 \
  --name migrator-serviceaccount \
  --namespace default \
  --cluster awesome-product \
  --attach-policy-arn arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess \
  --approve
```

## RDS

Create a new RDS DB. The example uses PostgreSQL. Launch the create wizard. Let AWS generate password for you. In "Connectivity" section make sure new DB will be provisioned in same VPC in which you created the EKS cluster. In same section expand "Additional connectivity configuration" and instead of using the default VPC security group create a new one called `database`.

Hit "Create database".

You cannot add inbound rules to DB security group in the wizard so we have to do it after DB is created. migrator pod need access to DB. Update `database` security group to allow inbound traffic on the DB port from the following SG: `eks-cluster-sg-awesome-product-{suffix}` (the suffix is random, replace it with your SG, simply start typing the name of the SG and AWS web console will load and autocomplete it for you).

In this place I would like to remind you to checkout the latest version of [godaddy/kubernetes-external-secrets](https://godaddy.github.io/kubernetes-external-secrets/) and see if this helm chart works correctly with Fargate and IAM.

If not proceed with standard Kubernetes secrets.

Copy credentials and connection information and update the secret in `kustomization.yaml`.

Then create `database-credentials` secret:

```
kubectl apply -k .
```

The generated secret name has a suffix appended by hashing the contents. This ensures that a new Secret is generated each time the contents is modified. Open `migrator-deployment.yaml` and update references to the secret name on lines: 30, 35, 40, 45.

## Deploy migrator

Last bit required is to update the `migrator-ingress.yaml` to make it a little bit more secure:

* alb.ingress.kubernetes.io/certificate-arn (line 9) - as we want HTTPS listener we need to provide ARN to ACM certificate
* alb.ingress.kubernetes.io/inbound-cidrs (line 11) - restrict access to your IP addresses (or leave default allow all mask)

And that should be us.

Review the config files and if all good deploy migrator:

```
kubectl apply -f migrator-deployment.yaml
kubectl apply -f migrator-service.yaml
kubectl apply -f migrator-ingress.yaml
```

Wait a few moments for alb-ingress-controller to provision the ALB.

## Accessing migrator

The migrator is up and running. From AWS console copy the ALB DNS name and then try the following URLs:

```
curl -v -k https://3e283ba1-default-migratori-3cc8-xxx.ap-northeast-1.elb.amazonaws.com/migrator/
curl -v -k https://3e283ba1-default-migratori-3cc8-xxx.ap-northeast-1.elb.amazonaws.com/migrator/v1/config
```

Check if migrator can load migrations from S3 and connect to DB:

```
curl -v -k https://3e283ba1-default-migratori-3cc8-xxx.ap-northeast-1.elb.amazonaws.com/migrator/v1/migrations/source
```

When you're ready apply migrations:

```
curl -v -k -X POST -H "Content-Type: application/json" -d '{"mode": "apply", "response": "list"}' https://3e283ba1-default-migratori-xxx.ap-northeast-1.elb.amazonaws.com/migrator/v1/migrations
```

Enjoy migrator!

## Cleanup

```
kubectl delete -k .
kubectl delete -f migrator-ingress.yaml
kubectl delete -f migrator-service.yaml
kubectl delete -f migrator-deployment.yaml
kubectl delete -f alb-ingress-controller.yaml
kubectl delete -f https://raw.githubusercontent.com/kubernetes-sigs/aws-alb-ingress-controller/v1.1.5/docs/examples/rbac-role.yaml
```
