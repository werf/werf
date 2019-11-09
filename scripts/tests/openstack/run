#!/bin/bash -e

cd "$(dirname "${BASH_SOURCE[0]}")"

if [[ $# -eq 0 ]]; then
  if [ ! -z "SSH_PRIVATE_KEY_PATH" ] ; then
    eval "$(ssh-agent)" > /dev/null
    trap 'kill '"$SSH_AGENT_PID" EXIT
    ssh-add ${SSH_PRIVATE_KEY_PATH} || exit $?
  else
    echo "Set env variable SSH_PRIVATE_KEY_PATH"
    exit 1
  fi
  if [ ! -d .terraform ] ; then
    terraform init tf
  fi
  terraform apply tf
  export ANSIBLE_SSH_ARGS
  ANSIBLE_SSH_ARGS="${ANSIBLE_SSH_ARGS:-"-C -o ControlMaster=auto -o ControlPersist=600s"} -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null"
  export ANSIBLE_PIPELINING=True
  ansible-playbook -i ~/.ansible-terraform-inventory ansible/main.yaml

elif [[ $# -eq 1 && "x$1" == "xdestroy" ]] ; then
  terraform destroy
  wait
else
  exit 1
fi
