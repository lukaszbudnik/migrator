# migrator on AWS EKS

The goal of this tutorial is to deploy migrator on AWS EKS with Fargate and IAM Roles for Service Accounts (IRSA), load migrations from AWS S3 and apply them to AWS RDS DB.

In all below commands I use these env variables:

```
AWS_REGION=us-east-2
CLUSTER_NAME=awesome-product
AWS_ACCOUNT_ID=XXX
```

## S3 - upload test migrations

Create S3 bucket in same region you will be deploying AWS EKS.

You can use test migrations to play around with migrator:

```
cd test/migrations
aws s3 cp --recursive migrations s3://your-bucket-migrator/migrations
```

## ECR - build and publish migrator image

Update `baseLocation` property in `migrator.yaml` to your AWS S3 bucket. Now that `migrator.yaml` is ready, build the migrator image and push it to ECR.

You can find detailed instructions in previous tutorial: [tutorials/aws-ecs](../aws-ecs).

```
aws ecr get-login --region $AWS_REGION
docker build --tag $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/migrator:v2020.1.0 .
docker push $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/migrator:v2020.1.0
```

Don't forget to edit `migrator-deployment.yaml` line 22 and update the image name to the image built above.

## EKS - create and setup the cluster

Note: There is no AWS managed policy which allows creation of EKS clusters. You need to explicitly add `eks:*` permissions to the user executing the create cluster command.

Then create the cluster with Fargate profile:

```
eksctl create cluster \
  --name $CLUSTER_NAME \
  --version 1.16 \
  --region $AWS_REGION \
  --external-dns-access \
  --alb-ingress-access \
  --full-ecr-access \
  --fargate
```

It will take around 15 minutes to complete. Make sure kubectl points to your cluster:

```
kubectl config current-context
```

Create an IAM OIDC provider and associate it with your cluster (prerequisite for IAM integration):

```
eksctl utils associate-iam-oidc-provider \
  --region $AWS_REGION \
  --cluster $CLUSTER_NAME \
  --approve
```

## Application Load Balancer - create ingress controller

### Service Account

First, we need to create Kubernetes service account for the Application Load Balancer ingress controller. This Kubernetes service account will be assigned AWS IAM Policy (via automatically created AWS IAM Role) allowing Kubernetes to manage AWS ALB on your behalf.

Code is available in `kubernetes-sigs/aws-alb-ingress-controller` repo:

```
kubectl apply -f https://raw.githubusercontent.com/kubernetes-sigs/aws-alb-ingress-controller/v1.1.7/docs/examples/rbac-role.yaml
```

Next, we need to create IAM policy allowing the ingress controller to provision the ALB for us. Again, we will use code available in `kubernetes-sigs/aws-alb-ingress-controller` repo:

```
aws iam create-policy \
  --policy-name ALBIngressControllerIAMPolicy \
  --policy-document https://raw.githubusercontent.com/kubernetes-sigs/aws-alb-ingress-controller/v1.1.7/docs/examples/iam-policy.json
```

Finally, we need to attach the IAM policy to our service account:

```
eksctl create iamserviceaccount \
  --region $AWS_REGION \
  --name alb-ingress-controller \
  --namespace kube-system \
  --cluster $CLUSTER_NAME \
  --attach-policy-arn arn:aws:iam::$AWS_ACCOUNT_ID:policy/ALBIngressControllerIAMPolicy \
  --override-existing-serviceaccounts \
  --approve
```

### ALB ingress controller

I will use helm chart `incubator/aws-alb-ingress-controller` to provision ALB ingress controller.

When creating the ALB ingress controller we need to explicitly pass the following parameters:

* cluster name
* region (EC2 metadata are disabled)
* VPC ID (EC2 metadata are disabled)
* service account with permissions to manage AWS ALB - the one we created above

```
# get VPC ID
vpcid=$(eksctl get cluster -n $CLUSTER_NAME -r $AWS_REGION | tail -1 | awk '{print $5}')
# add helm repo
helm repo add incubator http://storage.googleapis.com/kubernetes-charts-incubator
# install the chart
helm install alb-ingress-controller incubator/aws-alb-ingress-controller \
  --set clusterName=$CLUSTER_NAME \
  --set awsRegion=$AWS_REGION \
  --set awsVpcID=$vpcid \
  --set rbac.serviceAccount.create=false \
  --set rbac.serviceAccount.name=alb-ingress-controller \
  --namespace kube-system
```

## migrator setup

### Service Account

migrator needs to connect to S3 to read source migrations. We will create IAM service account for it called `migrator-serviceaccount`.

As a policy I will use the AWS-managed `AmazonS3ReadOnlyAccess`:

```
eksctl create iamserviceaccount \
  --cluster $CLUSTER_NAME \
  --region $AWS_REGION \
  --name migrator-serviceaccount \
  --namespace default \
  --attach-policy-arn arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess \
  --approve
```

### RDS

Create a new RDS DB. The example uses PostgreSQL. Launch the create wizard. Let AWS generate password for you. In "Connectivity" section make sure new DB will be provisioned in same VPC in which you created the EKS cluster. In same section expand "Additional connectivity configuration" and instead of using the default VPC security group create a new one called `database`.

Hit "Create database".

You cannot add inbound rules to DB security group in the wizard so we have to do it after DB is created. migrator pod need access to DB. Update `database` security group to allow inbound traffic on the DB port from the following SG: `eks-cluster-sg-$CLUSTER_NAME-{suffix}` (the suffix is random, replace it with your SG, simply start typing the name of the SG and AWS web console will autocomplete it for you).

For handling secrets securely you may want to checkout the latest version of [godaddy/kubernetes-external-secrets](https://godaddy.github.io/kubernetes-external-secrets/) and see if it works correctly with Fargate (as of the time of writing this tutorial it didn't work with IRSA). If not proceed with standard Kubernetes secrets.

Copy credentials and connection information and update the secret in `kustomization.yaml`.

Then create `database-credentials` secret:

```
kubectl apply -k .
```

The generated secret name has a suffix appended by hashing the contents. This ensures that a new secret is generated each time the contents is modified. Open `migrator-deployment.yaml` and update references to the secret name on lines: 30, 35, 40, 45.

While you are updating secrets don't forget to set a valid region name on line 26.

### ALB configuration

Last bit required is to update the `migrator-ingress.yaml` if we want migrator to be secure:

* alb.ingress.kubernetes.io/certificate-arn (line 9) - as we want HTTPS listener we need to provide ARN to ACM certificate
* alb.ingress.kubernetes.io/inbound-cidrs (line 11) - restrict access to your IP addresses (or leave default allow all mask)

And that should be us.

## Deploy migrator

Review the config files and if all good deploy migrator:

```
kubectl apply -f migrator-deployment.yaml
kubectl apply -f migrator-service.yaml
kubectl apply -f migrator-ingress.yaml
```

Wait a few moments for alb-ingress-controller to provision the ALB.

## Access migrator

The migrator is up and running. From Kubernetes ingress get the ALB DNS name:

```
address=$(kubectl get ingress | tail -1 | awk '{print $3}')
```

and then try the following URLs:

```
curl -v -k https://$address/migrator/
curl -v -k https://$address/migrator/v1/config
```

Check if migrator can load migrations from S3 and connect to DB:

```
curl -v -k https://$address/migrator/v1/migrations/source
```

When you're ready apply migrations:

```
curl -v -k -X POST -H "Content-Type: application/json" -d '{"mode": "apply", "response": "list"}' https://$address/migrator/v1/migrations
```

Enjoy migrator!

## Cleanup

```
kubectl delete -k .
kubectl delete -f migrator-ingress.yaml
kubectl delete -f migrator-service.yaml
kubectl delete -f migrator-deployment.yaml
eksctl delete iamserviceaccount migrator-serviceaccount --cluster $CLUSTER_NAME
eksctl delete iamserviceaccount alb-ingress-controller --namespace kube-system --cluster $CLUSTER_NAME
kubectl delete -f https://raw.githubusercontent.com/kubernetes-sigs/aws-alb-ingress-controller/v1.1.7/docs/examples/rbac-role.yaml
helm uninstall alb-ingress-controller --namespace kube-system
aws iam delete-policy --policy-arn arn:aws:iam::$AWS_ACCOUNT_ID:policy/ALBIngressControllerIAMPolicy
```

and delete the whole cluster:

```
eksctl delete cluster --region $AWS_REGION --name $CLUSTER_NAME
```
