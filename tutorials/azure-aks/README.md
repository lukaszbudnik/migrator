# migrator on Azure AKS

The goal of this tutorial is to publish migrator image to Azure ACR private container repository, deploy migrator to Azure AKS, load migrations from Azure Blob Container and apply them to Azure Database for PostgreSQL.

In order to simplify the customisation of the whole deployment I will be using the following env variables. Update them before you start this tutorial.

```
# resource group name (resource group must exist)
RG_NAME=lukaszbudnik
# Azure ACR repo name to create
ACR_NAME=migrator
# Azure ACK cluster name
ACK_CLUSTER_NAME=awesome-product
```

## Recommendations for production deployment

For keeping secrets in production workloads I recommend using [Azure Key Vault Provider for Secrets Store CSI Driver](https://github.com/Azure/secrets-store-csi-driver-provider-azure). It allows you to get secret contents stored in an Azure Key Vault instance and use the Secrets Store CSI driver interface to mount them into Kubernetes pods. It can also sync secrets from Azure Key Vault to Kubernetes secrets. Further, Azure Key Vault Provider can use managed identity so there is no need to juggle any credentials.

As an ingress controller I recommend using Azure Application Gateway. For more information refer to: https://docs.microsoft.com/en-us/azure/developer/terraform/create-k8s-cluster-with-aks-applicationgateway-ingress and https://azure.github.io/application-gateway-kubernetes-ingress/.

Another nice addition is to use [Service Catalog](https://svc-cat.io) and [Open Service Broker for Azure](https://osba.sh) to provision Azure Database and Azure Storage on your behalf. To use it check my gists: [Service Catalog with Open Service Broker for Azure to provision Azure Database](https://gist.github.com/lukaszbudnik/b2734c250e71b0c7f18dd93fb882cc42) and [Service Catalog with Open Service Broker for Azure to provision Azure Storage Account](https://gist.github.com/lukaszbudnik/c03549cfa9728d9e4957e6bc54ef3c6e).

In this tutorial I decided to keep things simple. Azure Key Vault Provider for Secrets Store CSI Driver and Azure Application Gateway are a little bit more complex to setup and are outside of the scope of this tutorial. Also, Azure Database and Azure Storage are created manually.

## Azure Blob Container - upload test migrations

Create new Azure Storage Account and a new container. Then:

* update `baseLocation` property in `migrator.yaml`.
* update `storage-account` credentials in `kustomization.yaml`

You can use test migrations to play around with migrator. They are located in `test/migrations` directory.

## ACR - build and publish migrator image

Create container repository:

```
az acr create --resource-group $RG_NAME --name $ACR_NAME --sku Basic
```

Build image and push it to ACR:

```
az acr build --registry $ACR_NAME --image migrator:v4.2-azure .
```

From the output you can see:

```yaml
- image:
    registry: migrator.azurecr.io
    repository: migrator
    tag: v4.2-azure
    digest: sha256:e72108eec96204f01d4eaa87d83ea302bbb651194a13636597c32f7434de5933
  runtime-dependency:
    registry: registry.hub.docker.com
    repository: lukasz/migrator
    tag: dev-v4.2
    digest: sha256:b414ea94960048904a137ec1655b7e6471e57a483ccd095fca744c7a449a118e
```

which means that the ACR image is: `migrator.azurecr.io/migrator:v4.2-azure`.

Edit `migrator-deployment.yaml` and update the image name to the one built above (line 21).

## Create and setup the AKS

Create AKS cluster and attach our ACR repository to it (ACR attach is required otherwise AKS won't be able to pull our migrator image):

```
az aks create --name $ACK_CLUSTER_NAME \
  --resource-group $RG_NAME \
  --load-balancer-sku basic \
  --vm-set-type AvailabilitySet \
  --node-count 1 \
  --enable-addons monitoring \
  --attach-acr $ACR_NAME \
  --no-ssh-key
```

Wait a moment for Azure to create the cluster. Fetch credentials so that kubectl can successfully connect to AKS.

```
az aks get-credentials --resource-group $RG_NAME --name $ACK_CLUSTER_NAME
```

Make sure kubectl points to our cluster:

```
kubectl config current-context
```

## NGINX ingress controller

Let's create a minimalistic (one replica) `nginx-ingress` controller, disable port 80 as we want to listen only on port 443. NGINX ingress will deploy self-signed cert (read documentation about how to set it up with your own existing certs or let's encrypt). Also, I'm allowing two specific IP address ranges for my app. And yes, comma needs to be escaped. For testing you may replace it with 0.0.0.0/0 or remove at all:

```
helm repo add stable https://kubernetes-charts.storage.googleapis.com/

helm install nginx-ingress stable/nginx-ingress \
    --set controller.replicaCount=1 \
    --set controller.service.enableHttp=false \
    --set controller.service.loadBalancerSourceRanges={1.2.3.4/32\,5.6.7.8/32} \
    --set controller.nodeSelector."beta\.kubernetes\.io/os"=linux \
    --set defaultBackend.nodeSelector."beta\.kubernetes\.io/os"=linux
```

## Azure Database for PostgreSQL

The example uses PostgreSQL. Go ahead and create a new database using Azure Database for PostgreSQL.

By default newly provisioned DB blocks all traffic. Once the DB is up and running, open it in Azure portal and navigate to "Settings" -> "Connection security". Toggle on the "Allow access to Azure services" option and click "Save".

Open `kustomization.yaml` and update `database-credentials` secret.

## Kubernetes Secrets

Now that we have storage account and database credentials in `kustomization` it's time to create secrets:

```
kubectl apply -k .
```

The generated secret names have a suffix appended by hashing the contents. This ensures that a new Secret is generated each time the contents is modified. Open `migrator-deployment.yaml` and update references to:

* storage-account secret on lines: 27, 32
* database-credentials secret on lines: 37, 42, 47, 52

## Deploy migrator

Review the config files and if all good deploy migrator:

```
kubectl apply -f migrator-deployment.yaml
kubectl apply -f migrator-service.yaml
kubectl apply -f migrator-ingress.yaml
```

Wait a few moments and check the external IP of the NGINX ingress controller:

```
kubectl get service nginx-ingress-controller
```

## Accessing migrator

The migrator is up and running. You can now access it by external IP address:

```
curl -v -k https://65.52.0.0/migrator/
curl -v -k https://65.52.0.0/migrator/v1/config
```

Check if migrator can load migrations from Azure Blob Storage and connect to Azure Database for PostgreSQL:

```
curl -v -k https://65.52.0.0/migrator/v1/migrations/source
```

When you're ready apply migrations:

```
curl -v -k -X POST -H "Content-Type: application/json" -d '{"mode": "apply", "response": "list"}' https://65.52.0.0/migrator/v1/migrations
```

Enjoy migrator!

## Cleanup

```
kubectl delete -k .
kubectl delete -f migrator-ingress.yaml
kubectl delete -f migrator-service.yaml
kubectl delete -f migrator-deployment.yaml
helm del nginx-ingress
```
