#!/bin/bash

# options:
# remove stopped containers and untagged images
#   $ ./cleanup
# remove all stopped|running containers and untagged images
#   $ ./cleanup --reset
# remove containers|images|tags matching {repository|image|repository\image|tag|image:tag}
# pattern and untagged images
#   $ ./cleanup --purge {image}
# everything
#   $ ./cleanup --nuclear

function _clean_docker {
if [ "$1" == "--reset" ]; then
  # Remove all containers regardless of state
docker rm -vf $(docker ps -a -q) 2>/dev/null || echo "No more containers to remove."
elif [ "$1" == "--purge" ]; then
 # Attempt to remove running containers that are using the images we're trying to purge first.
(docker rm -vf $(docker ps -a | grep "$2/\|/$2 \| $2 \|:$2\|$2-\|$2:\|$2_" | awk '{print $1}') 2>/dev/null || echo "No containers using the \"$2\" image, continuing purge.") &&\
# Remove all images matching arg given after "--purge"
docker rmi $(docker images | grep "$2/\|/$2 \| $2 \|$2 \|$2-\|$2_" | awk '{print $3}') 2>/dev/null || echo "No images matching \"$2\" to purge."
else
# This alternate only removes "stopped" containers
docker rm -vf $(docker ps -a | grep "Exited" | awk '{print $1}') 2>/dev/null || echo "No stopped containers to remove."
fi

if [ "$1" == "--nuclear" ]; then
#docker rm -vf $(docker ps -a -q) 2>/dev/null || echo "No more containers to remove."
#docker rmi $(docker images -q) 2>/dev/null || echo "No more images to remove."
echo "Preparing to uninstall docker ....."
dpkg -l | grep Docker | awk '{print $2}' > /tmp/docker-list.txt
while read list; do
    sudo apt-get remove $list -y
done </tmp/docker-list.txt
rm /tmp/docker-list.txt
else
# Always remove untagged images
docker rmi $(docker images | grep "<none>" | awk '{print $3}') 2>/dev/null || echo "No untagged images to delete."
fi

}

function _clean_ansible {
if [ "$1" == "--nuclear" ]; then
version=$(grep "ansible_version" ${kud_playbooks}/kud-vars.yml | awk -F ': ' '{print $2}')
sudo pip uninstall ansible==$version
fi
}

#Defining config path
INSTALLER_DIR="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")"
kud_folder=${INSTALLER_DIR}
kud_infra_folder=$kud_folder/../../deployment_infra
kud_playbooks=$kud_infra_folder/playbooks

_clean_docker $1
_clean_ansible $1
