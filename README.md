# Stampy Admission Controller for AWS/EKS

This webhook will make sure to validate docker image signatures before creating deployments/pods.

The webhook needs read only access to S3 buckets so make sure your EKS worker node role has AmazonS3ReadOnlyAccess attached as shown below.

```
resource "aws_iam_role_policy_attachment" "jit-node-AmazonS3ReadOnlyAccess" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess"
  role       = "${aws_iam_role.jit-node.name}"
}
```

# Install Helm on your cluster

```
kubectl create serviceaccount --namespace kube-system tiller
kubectl create clusterrolebinding tiller-cluster-rule --clusterrole=cluster-admin --serviceaccount=kube-system:tiller
kubectl patch deploy --namespace kube-system tiller-deploy -p '{"spec":{"template":{"spec":{"serviceAccount":"tiller"}}}}'      
helm init --service-account tiller --upgrade
```

# Install Stampy Admission Controller Webhook via Helm 

```
# helm install ./stampy-webhook-admission-controller-0.2.0.tgz --set controller.image=121924372514.dkr.ecr.us-east-2.amazonaws.com/stampy-webhook-admission-controller --set controller.imageTag=v0.2.0 --set controller.region=us-east-2 --set controller.bucket=docker-signatures
```

