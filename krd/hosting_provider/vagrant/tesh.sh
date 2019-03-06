#! /bin/bash

krd_folder=$(pwd)
echo $pwd
krd_infra_folder=$krd_folder/../../krd_deployment_infra
krd_playbooks=$krd_infra_folder/playbooks
export krd_inventory_folder=$krd_folder/inventory
version=$(grep "kubespray_version" ${krd_playbooks}/krd-vars.yml | awk -F ': ' '{print $2}')
echo " Version is : $version"
echo "----DONE----"
