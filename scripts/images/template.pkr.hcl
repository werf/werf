source "qemu" "werf-ubuntu-amd64-20-04" {
  disk_image           = true
  floppy_files         = ["cloud-data/user-data", "cloud-data/meta-data"]
  floppy_label         = "cidata"
  skip_resize_disk     = true
  format               = "qcow2"
  iso_checksum         = "sha256:c33e4d8f8fd52d639aa114b237274da428be7fd95025e20787b11cf77507d111"
  iso_target_extension = "qcow2"
  iso_url              = "https://cloud-images.ubuntu.com/releases/focal/release-20220322/ubuntu-20.04-server-cloudimg-amd64.img"
  memory               = "2048"
  output_directory     = "image"
  qemuargs = [
    ["-smp", "2"],
  ]
  headless               = true
  use_default_display    = true
  ssh_password           = "werf"
  ssh_port               = 22
  ssh_username           = "werf"
  ssh_timeout            = "10m"
  ssh_handshake_attempts = 150
  vm_name                = "werf-ubuntu-amd64-20.04.qcow2"
}

build {
  sources = ["source.qemu.werf-ubuntu-amd64-20-04"]

  provisioner "shell" {
    inline  = ["/usr/bin/cloud-init status --wait"]
    timeout = "10m"
  }

  provisioner "shell" {
    env = {
      "DEBIAN_FRONTEND": "noninteractive"
    }
    inline  = [
      "echo 'user.max_user_namespaces = 15000' | sudo tee -a /etc/sysctl.conf",
      "echo 'kernel.unprivileged_userns_clone = 1' | sudo tee -a /etc/sysctl.conf",
      "sudo apt-get update -yq",
      "sudo apt-get upgrade -yq",
      "sudo apt-get install -yq linux-image-virtual-hwe-20.04 crun uidmap",
      "mkdir -p ~/.local/share/containers",
      "sudo apt-get -yq autoremove --purge snapd",
      "sudo apt-get -yq purge vim git linux-headers-* rsync strace",
      "sudo apt-get autoremove -yq",
      "sudo rm -rf /var/cache/snapd/ /snap /var/cache/apt /var/lib/apt/lists ~/snap",
      "sync",
      "sudo dd if=/dev/zero of=/zeroing || true; sync && sudo rm -rf /zeroing",
      "sync",
    ]
  }
}
