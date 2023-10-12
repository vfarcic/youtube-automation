# Source: https://gist.github.com/vfarcic/be4e956babf230d709c5d52d80dc5ee3

#########################################################
# Mastering CI/CD: Streamline Your Development Workflow #
#########################################################

# Additional Info:
# - Fancy projectxxx: acme.com
# - Argo CD - Applying GitOps Principles To Manage A Production Environment In Kubernetes: https://youtu.be/vpWQeoaiRM4
# - How To Apply GitOps To Everything - Combining Argo CD And Crossplane: https://youtu.be/yrj4lmScKHQ
# - Crossplane - GitOps-based Infrastructure as Code through Kubernetes API: https://youtu.be/n8KjVmuHm7A
# - How To Shift Left Infrastructure Management Using Crossplane Compositions: https://youtu.be/AtbS1u2j7po
# - SchemaHero - Database Schema Migrations Inside Kubernetes: https://youtu.be/SofQxb4CDQQ
# - Is Timoni With CUE a Helm Replacement?: https://youtu.be/bbE1BFCs548
# - GitHub CLI - How to manage repositories more efficiently: https://youtu.be/BII6ZY2Rnlc

#########
# Setup #
#########

# Create a Kubernetes cluster
# The demo was tested on GKE but it should work on any Kubernetes
#   cluster, except those running inside containers like KinD or
#   Civo.

git clone https://github.com/vfarcic/falco-demo

cd falco-demo

helm upgrade --install falco falco \
    --repo https://falcosecurity.github.io/charts \
    --values values.yaml --namespace falco --create-namespace \
    --wait

kubectl create namespace demo

kubectl --namespace demo run demo --image alpine \
    -- sh -c "sleep infinity"

# Install `jq` by following the instructions at
#   https://jqlang.github.io/jq/download

####################################
# Falco Threat-Detection in Action #
####################################

kubectl --namespace demo exec --stdin --tty demo \
    -- sh -c "ls /"

kubectl --namespace falco logs \
    --selector app.kubernetes.io/name=falco --container falco \
    | grep Notice | jq .

kubectl --namespace falco get pods

# Replace `[...]` with the name of one of the `falco-*` pods
export POD=[...]

kubectl --namespace falco exec -it $POD \
    -- sh -c "cat /etc/falco/falco_rules.yaml"

cat rule-example.yaml

###########
# Destroy #
###########

# Destroy or reset the cluster
