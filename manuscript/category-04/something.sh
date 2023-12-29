# Source: https://gist.github.com/vfarcic/29deee52bcd5720625e969205de2f0e1

#######################################################################################################
# Mastering Continuous Integration and Continuous Deployment: A Complete CI/CD Tutorial for Beginners #
#######################################################################################################

# Additional Info:
# - Something: acme.com
# - Argo CD - Applying GitOps Principles To Manage A Production Environment In Kubernetes: https://youtu.be/vpWQeoaiRM4
# - How To Apply GitOps To Everything - Combining Argo CD And Crossplane: https://youtu.be/yrj4lmScKHQ
# - Crossplane - GitOps-based Infrastructure as Code 

#########
# Setup #
#########

# Install `nix` by following the instructions at https://nix.dev/install-nix.

#########################################
# Ephemeral Shell Environments with Nix #
#########################################

gh repo clone vfarcic/crossplane-tutorial dsds ds dasd asd \
    dsa dsad asdas das dasd asd asd

nix-shell --packages gh kubectl awscli2

PS1="$ "

gh repo clone vfarcic/crossplane-tutorial

which gh

which kubectl

cd crossplane-tutorial

chmod +x setup/01-managed-resources.sh

./setup/01-managed-resources.sh

cat setup/01-managed-resources-nix.sh

chmod +x setup/01-managed-resources-nix.sh

./setup/01-managed-resources-nix.sh

# Press `ctrl+c` to stop

./setup/01-managed-resources-nix.sh

# Press `ctrl+c` to stop

exit

cd crossplane-tutorial

cat shell.nix

nix-shell

PS1="$ "

gum

./setup/01-managed-resources.sh

# Press `ctrl+c` to stop

exit

nix-shell --run $SHELL

exit

nix-store --gc

###########
# Destroy #
###########

cd ..

rm -rf crossplane-tutorial

# Delete the fork from [GitHub](https://github.com)
