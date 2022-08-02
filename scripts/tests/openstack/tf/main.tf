data "openstack_compute_keypair_v2" "ssh" {
  name = "kubernetes"
}

data "openstack_compute_flavor_v2" "medium" {
  name = "m1.medium"
}

data "openstack_dns_zone_v2" "werf" {
  name = "ci.asidorovj.ru."
}

resource "random_string" "docker_registry_password" {
  length = 16
  special = true
  override_special = "_%@"
}

resource "openstack_compute_instance_v2" "kubernetes" {
  name = join("-", ["kubernetes", "1-${11+count.index}"])
  count = 6
  image_name = "ubuntu-18-04-cloud-amd64"
  flavor_name = data.openstack_compute_flavor_v2.medium.name
  key_pair = data.openstack_compute_keypair_v2.ssh.name

  network {
    name = "public"
    access_network = true
  }

  connection {
    host = self.access_ip_v4
    user = "ubuntu"
  }
  provisioner "remote-exec" {
    inline = [
      "echo Ready!"
    ]
  }
}

resource "openstack_dns_recordset_v2" "kubernetes" {
  name = join("-", ["kubernetes", "1-${11+count.index}.${data.openstack_dns_zone_v2.werf.name}"])
  count = 6
  zone_id = data.openstack_dns_zone_v2.werf.id
  description = "Record for kubernetes cluster"
  ttl = 30
  type = "A"
  records = ["${openstack_compute_instance_v2.kubernetes[count.index].access_ip_v4}"]
}

resource "openstack_compute_instance_v2" "registry" {
  name = "docker-registry"
  image_name = "ubuntu-18-04-cloud-amd64"
  flavor_name = data.openstack_compute_flavor_v2.medium.name
  key_pair = data.openstack_compute_keypair_v2.ssh.name

  network {
    name = "public"
    access_network = true
  }

  connection {
    host = self.access_ip_v4
    user = "ubuntu"
  }

  provisioner "remote-exec" {
    inline = [
      "echo Ready!"
    ]
  }
}

resource "openstack_dns_recordset_v2" "registry" {
  name = "registry.${data.openstack_dns_zone_v2.werf.name}"
  zone_id = data.openstack_dns_zone_v2.werf.id
  description = "Record for docker registry"
  ttl = 30
  type = "A"
  records = ["${openstack_compute_instance_v2.registry.access_ip_v4}"]
}

resource "ansible_host" "kubernetes" {
  count = 6
  inventory_hostname = openstack_compute_instance_v2.kubernetes[count.index].name
  groups = ["kubernetes"]
  vars = {
    ansible_user = "ubuntu"
    ansible_host = openstack_compute_instance_v2.kubernetes[count.index].access_ip_v4
    ansible_become = "yes"
    kubernetes_version = "1.${11+count.index}"
    kubernetes_domain = join("-", ["kubernetes", "1-${11+count.index}.${data.openstack_dns_zone_v2.werf.name}"])
  }
}

resource "ansible_host" "registry" {
  inventory_hostname = openstack_compute_instance_v2.registry.name
  groups = ["registry"]
  vars = {
    ansible_user = "ubuntu"
    ansible_host = openstack_compute_instance_v2.registry.access_ip_v4
    ansible_become = "yes"
    domain = "registry.${data.openstack_dns_zone_v2.werf.name}"
    docker_registry_user = "werfuser"
    docker_registry_pass = random_string.docker_registry_password.result
  }
}

output "docker_registry_password" {
  value = random_string.docker_registry_password.result
  description = "Password for docker registry user: werfuser"
  sensitive = false
}
