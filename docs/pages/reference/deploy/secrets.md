---
title: Working with secrets
sidebar: reference
permalink: reference/deploy/secrets.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
---

Werf secrets engine is recommended for storing database passwords, files with encryption certificates, etc.

The idea is that sensitive data must be stored in a repository served by an application and remain independent from any specific server.

## Encryption key

A key is required for encryption and decryption of data. There are two locations from which werf can read the key:
* from the `WERF_SECRET_KEY` environment variable
* from a special `.werf_secret_key` file in the project root

> Encryption key must be **hex dump** of either 16, 24, or 32 bytes long to select AES-128, AES-192, or AES-256. `werf helm secret keygen` command returns AES-128 encryption key.

You can promptly generate a key using the `werf helm secret generate-secret-key` command.

### werf helm secret generate-secret-key command

{% include /cli/werf_helm_secret_generate_secret_key.md header="####" %}

### Working with the WERF_SECRET_KEY environment variable

If an environment variable is available in the environment where werf is launched, werf can use it.

In a local environment, you can declare it from the console.

For Gitlab CI, use [CI/CD Variables](https://docs.gitlab.com/ee/ci/variables/#variables) – they are only visible to repository masters, and regular developers will not see them.

### Working with the .werf_secret_key file

Using the `.werf_secret_key` file is much safer and more convenient, because:
* users or release engineers are not required to add an encryption key for each launch;
* the secret value described in the file cannot be included into the cli `~/.bash_history` log.

> **Attention! Do not save the file into the git repository. If you do it, the entire sense of encryption is lost, and anyone who has source files at hand can retrieve all the passwords. `.werf_secret_key` must be kept in `.gitignore`!**

## Secret values encryption

The `.helm/secret-values.yaml` file is designed for storing secret values.

It is decoded in the course of deployment and used in helm as [additional values](https://github.com/kubernetes/helm/blob/master/docs/chart_template_guide/values_files.md).

This is what a file containing encrypted values may look like:
```yaml
mysql:
  host: 10005968c24e593b9821eadd5ea1801eb6c9535bd2ba0f9bcfbcd647fddede9da0bf6e13de83eb80ebe3cad4
  user: 100016edd63bb1523366dc5fd971a23edae3e59885153ecb5ed89c3d31150349a4ff786760c886e5c0293990
  password: 10000ef541683fab215132687a63074796b3892d68000a33a4a3ddc673c3f4de81990ca654fca0130f17
  db: 1000db50be293432129acb741de54209a33bf479ae2e0f53462b5053c30da7584e31a589f5206cfa4a8e249d20
```

### werf helm secret values edit command

{% include /cli/werf_helm_secret_values_edit.md header="####" %}

### werf helm secret values encrypt command

{% include /cli/werf_helm_secret_values_encrypt.md header="####" %}

### werf helm secret values decrypt command

{% include /cli/werf_helm_secret_values_decrypt.md header="####" %}

## Secret file encryption

Besides secret values, templates also use files that may not be stored unencrypted in the repository. For these files, the `.helm/secret` directory is allocated where encrypted files must be stored. Using the `werf_secret_file` method (that generates werf `_werf_helpers.tpl` in the deployment process), you can get decrypted file content in a template.

To use secret data in helm templates, you must save it to an appropriate file in the `.helm/secret` directory.

Using a secret in a template may appear as follows:

{% raw %}
```yaml
...
data:
  tls.key: {{ tuple "/backend-saml/tls.key" . | include "werf_secret_file" | b64enc }}
```
{% endraw %}

### werf helm secret file edit command

{% include /cli/werf_helm_secret_values_edit.md header="####" %}

### werf helm secret file encrypt command

{% include /cli/werf_helm_secret_values_encrypt.md header="####" %}

### werf helm secret file decrypt command

{% include /cli/werf_helm_secret_values_decrypt.md header="####" %}

## Secret key rotation

### werf helm secret rotate secret key command

{% include /cli/werf_helm_secret_rotate_secret_key.md header="####" %}
