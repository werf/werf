---
title: Secrets
permalink: advanced/helm/configuration/secrets.html
---

werf secrets engine is recommended for storing database passwords, files with encryption certificates, etc.

The idea is that sensitive data must be stored in a repository served by an application and remain independent of any specific server.

werf supports passing secrets as:
 - separate [secret values]({{ "/advanced/helm/configuration/values.html#user-defined-secret-values" | true_relative_url }}) yaml file (`.helm/secret-values.yaml` by default, or any file passed by the `--secret-values` option);
 - secret files — raw encoded files, which can be used in the templates.

## Encryption key

A key is required for encryption and decryption of data. There are two locations from which werf can read the key:
* from the `WERF_SECRET_KEY` environment variable
* from a special `.werf_secret_key` file in the project root
* from `~/.werf/global_secret_key` (globally)

> Encryption key must be **hex dump** of either 16, 24, or 32 bytes long to select AES-128, AES-192, or AES-256. [werf helm secret generate-secret-key command]({{ "reference/cli/werf_helm_secret_generate_secret_key.html" | true_relative_url }}) returns AES-128 encryption key

You can promptly generate a key using the [werf helm secret generate-secret-key command]({{ "reference/cli/werf_helm_secret_generate_secret_key.html" | true_relative_url }}).

### Working with the WERF_SECRET_KEY environment variable

If an environment variable is available in the environment where werf is launched, werf can use it.

In a local environment, you can declare it from the console.

For GitLab CI, use [CI/CD Variables](https://docs.gitlab.com/ee/ci/variables/#variables) – they are only visible to repository masters, and regular developers will not see them.

### Working with the .werf_secret_key file

Using the `.werf_secret_key` file is much safer and more convenient, because:
* users or release engineers are not required to add an encryption key for each launch;
* the secret value described in the file cannot be included into the cli `~/.bash_history` log.

> **ATTENTION! Do not save the file into the git repository. If you do it, the entire sense of encryption is lost, and anyone who has source files at hand can retrieve all the passwords. `.werf_secret_key` must be kept in `.gitignore`!**

## Secret key rotation

To regenerate secret files and values with new secret key use [werf helm secret rotate-secret-key command]({{ "reference/cli/werf_helm_secret_rotate_secret_key.html" | true_relative_url }}).

## Secret values

The secret values file is designed for storing secret values. **By default** werf uses `.helm/secret-values.yaml` file, but user can specify arbitrary number of such files.

Secret values file may look like:
```yaml
mysql:
  host: 10005968c24e593b9821eadd5ea1801eb6c9535bd2ba0f9bcfbcd647fddede9da0bf6e13de83eb80ebe3cad4
  user: 100016edd63bb1523366dc5fd971a23edae3e59885153ecb5ed89c3d31150349a4ff786760c886e5c0293990
  password: 10000ef541683fab215132687a63074796b3892d68000a33a4a3ddc673c3f4de81990ca654fca0130f17
  db: 1000db50be293432129acb741de54209a33bf479ae2e0f53462b5053c30da7584e31a589f5206cfa4a8e249d20
```

To manage secret values files use the following commands:
- [`werf helm secret values edit` command]({{ "reference/cli/werf_helm_secret_values_edit.html" | true_relative_url }})
- [`werf helm secret values encrypt` command]({{ "reference/cli/werf_helm_secret_values_encrypt.html" | true_relative_url }})
- [`werf helm secret values decrypt` command]({{ "reference/cli/werf_helm_secret_values_decrypt.html" | true_relative_url }})

### Using in a chart template

The secret values files are decoded in the course of deployment and used in helm as [additional values](https://helm.sh/docs/chart_template_guide/values_files/). Thus, given the following secret values yaml:

```yaml
# .helm/secret-values.yaml
mysql:
  user: 10003c7f513b1ba1a0eb3d2cfb8294c93fddda8701850aa8adc1d9032229ddb4fd3b
  password: 1000cd6674285b65f55b739ee2e5130cfc6d01d87772c9e62c1c917d9b10194f14ef
```

— usage of these values is the same as regular values:

{% raw %}
```yaml
...
env:
- name: MYSQL_USER
  value: {{ .Values.mysql.user }}
- name: MYSQL_PASSWORD
  value: {{ .Values.mysql.password }}
```
{% endraw %}

## Secret files

Secret files are excellent for storing sensitive data such as certificates and private keys in the project repository. For these files, the `.helm/secret` directory is allocated where encrypted files must be stored.

To use secret data in helm templates, you must save it to an appropriate file in the `.helm/secret` directory.

To manage secret files use the following commands:
 - [`werf helm secret file edit` command]({{ "reference/cli/werf_helm_secret_file_edit.html" | true_relative_url }})
 - [`werf helm secret file encrypt` command]({{ "reference/cli/werf_helm_secret_file_encrypt.html" | true_relative_url }})
 - [`werf helm secret file decrypt` command]({{ "reference/cli/werf_helm_secret_file_decrypt.html" | true_relative_url }})

### Using in a chart template

<!-- Move to reference -->

The `werf_secret_file` runtime function allows using decrypted file content in a template. The required function argument is a secret file path relative to `.helm/secret` directory.

Using the decrypted secret `.helm/secret/backend-saml/tls.key` in a template may appear as follows:

{% raw %}
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: myproject-backend-saml
type: kubernetes.io/tls
data:
  tls.crt: {{ werf_secret_file "backend-saml/stage/tls.crt" | b64enc }}
  tls.key: {{ werf_secret_file "backend-saml/stage/tls.key" | b64enc }}
```
{% endraw %}

Note that `backend-saml/stage/` is an arbitrary file structure. User can place all files into the single directory `.helm/secret` or create subdirectories at his own discretion.
