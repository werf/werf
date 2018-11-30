---
title: Working with secrets
sidebar: reference
permalink: reference/deploy/secrets.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
---

Dapp secrets engine is recommended for storing: database passwords, files with encryption certificates, etc.

The idea is that sensitive data must be stored in a repository served by an application and remain independent from any specific server.

## Encryption key

A key is required for encryption and decryption of data. There are two locations from which dapp can read the key:
* from the `DAPP_SECRET_KEY` environment variable
* from a special `/.dapp_secret_key` file in the project root

You can promptly generate a key using the `dapp kube secret key generate` command:
```bash
$ dapp kube secret key generate
DAPP_SECRET_KEY=c85e100d4ff006b693b0555f09244fdf
```

For convenience, the command output already contains an environment variable and can be used in the `export` command.

> Encryption key must be **hex dump** of either 16, 24, or 32 bytes long to select AES-128, AES-192, or AES-256. `dapp kube secret key generate` command returns AES-128 encryption key.

If you want to generate key yourself don't forget about hex dump:
```bash
$ head -c16 </dev/random | xxd -p
5c350e9e22f97501c862396019436988
```

### Working with the `DAPP_SECRET_KEY` environment variable

If an environment variable is available in the environment where dapp is launched, dapp can use it.

In a local environment, you can declare it from the console.

For Gitlab CI, use [CI/CD Variables](https://docs.gitlab.com/ee/ci/variables/#variables) – they are only visible to repository masters, and regular developers will not see them.

### Working with the `/.dapp_secret_key` file

Using the `.dapp_secret_key` file is much safer and more convenient, because:
* users or release engineers are not required to add an encryption key for each launch;
* the secret value described in the file cannot be included into the cli `~/.bash_history` log.

**Attention! Do not save the file into the git repository. If you do it, the entire sense of encryption is lost, and anyone who has source files at hand can retrieve all the passwords. `/.dapp_secret_key` must be kept in `.gitignore`!**

## Encryption of values

The `.helm/secret-values.yaml` file is designed for storing secret values.

It is decoded in the course of deployment and used in helm as [additional values](https://github.com/kubernetes/helm/blob/master/docs/chart_template_guide/values_files.md). If no encryption key is available at the moment of dapp launch, the values are decoded into empty strings.

This is what a file containing encrypted values may look like:
```yaml
mysql:
  host: 100070c0e52ba2ff965ebd85f5fea9549392294e52aca006cf75
  user: 2ad80161428063803509eba8e9909ddcd0db0ddaada!b9ee47
  password: 80161428063803509eba8e9909ddcd0db0ddaab9ee47
  db: 406d3a4d2282ad80161428063803509eba8e9909ddcd0db0ddaab9ee47
```

### Encryption of a single value

For data encryption, a `dapp kube secret generate` command is used.
```bash
$ dapp kube secret generate
Enter secret: 
1000541517bccae1acce015629f4ec89996e0b4
```

The command also supports a redirected output, which is the performance outcome of other commands.
```bash
$ rake magic | dapp kube secret generate
1000541517bccae1acce015629f4ec89996e0b4
```

### Encryption of a yaml file

If you already have a prepared file with variables, e.g.:
```yaml
mysql:
  host: 192.168.1.1
  user: mydbuser
  password: password
  db: dbforapp
```

You can apply encryption to it using `dapp kube secret VALUES_PATH generate --values`, and it will output the file with encrypted keys:
```bash
$ dapp kube secret generate .helm/secret-values.yaml --values
mysql:
  host: 100070c0e52ba2ff965ebd85f5fea9549392294e52aca006cf75
  user: 2ad80161428063803509eba8e9909ddcd0db0ddaada!b9ee47
  password: 80161428063803509eba8e9909ddcd0db0ddaab9ee47
  db: 406d3a4d2282ad80161428063803509eba8e9909ddcd0db0ddaab9ee47
```

## Encryption of the entire file

Besides secret values, templates also use files that may not be stored unencrypted in the repository. For these files, the `.helm/secret` directory is allocated where encrypted files must be stored. Using the `dapp_secret_file` method (that generates dapp `_dapp_helpers.tpl` in the deployment process), you can get decrypted file content in a template (the method returns an empty string if no encryption key is available).

When applying file encryption, you must specify the file path:
```bash
$ dapp kube secret generate ~/certs/tls.key
100023b2d1c0ec145681183ec721dc06db34f7ebce9f328739f0350d7f3aea988b6d0b69e9f71ed5e2ad9d79449b7a7d830ee5148a30a50bd43b7e2ecaef1c657199a483f60322cf7727ddf3928b2f51b0fbb0b1cd931489c20061a5071cf4362cb7e91c79fdbfc6d950352535eac28affd47d8ea8af64559fa39d89e815ea2b95cb07e81ddba792bf0e834cbbdc2ef843394a23f0cd44a95a38dd1583c2ae8352af140fc3fcfa6da3485bbf9bd286e2864ad45e31bc8ce4239aa05aaa82beba58c0583d3e93141ae28d87f4ffdb3d089f18b86e42e88a0b065c604f92a1478e0bbaeee46136579895b803a4be80977135979c4022b83fb1787e7b1540ddc07cd287ba5a7442f8a3ce0f5177487751c25767c28fd6eacb7f021036d978301895d6f528f06d555c926ba617669348c7873ba98372ae75ee0fdb730cabe507c576371970a27476e557b8b250f83137535f1d466eb53756986160f75ef78075dd7f63f83d72c1daf04aa026000802d4bbc2832f6d63eb231b8e16af5f44fc2cd79220715cba783a495a9d25e778ec1c2aa8013ccc164b5fc51f3a061c1eeed1228f65867c25f962639c90d2398e48ad93744cab5f8fff1f9988ccdbc5778ff39c31bdd47950759f33bf126509d3105521571252823f523fcd4a478d9bce3ddf923f8f8cbe7bff5edc0e99fe908e8b737a6de2391729e6ada3d8069819a0857ceba1eb5a16ecc81d6bcd16e497c4e60af5d218d2d2e0064c07850e5aa2a8d83e0f0a2
```

To use the data in helm templates, you must save it to an appropriate file in the `.helm/secret` directory.

Calling the command with the `-o OUTPUT_FILE_PATH` option saves encrypted data to the file `OUTPUT_FILE_PATH`:
```bash
$ dapp kube secret generate ~/certs/tls.key -o .helm/secret/backend-saml/tls.key
```

Using a secret in a template may appear as follows:

{% raw %}
```yaml
...
data:
  tls.key: {{ tuple "/backend-saml/tls.key" . | include "dapp_secret_file" | b64enc }}
```
{% endraw %}

## Editing encrypted data

You can edit existing secrets using the `dapp kube secret edit` command. This command means you can work with data interactively.

## Inverse conversion of data

You can decrypt previously encrypted values using the `dapp kube secret extract` command.

```bash
$ dapp kube secret extract
Enter secret: 1000541517bccae1acce015629f4ec89996e0b4
42
```

Like with encryption, redirected output and secrets from files are also supported.

```bash
$ echo "1000541517bccae1acce015629f4ec89996e0b4" | dapp kube secret extract
42
```

```bash
$ dapp kube secret extract .helm/secret/sense_of_life.txt
The Ultimate Question of Life, the Universe, and Everything.
```

If you need to decrypt the secret-values file, you must also specify the `--values` option.

```bash
$ dapp kube secret extract .helm/secret-values.yaml --values
sense:
  of:
    life: 42
    lifes: [42, 42, 42]
```

## Regeneration of existing secrets

When launching the command, the secrets (`.helm/secret/**/*`) and secret values (`.helm/secret-values.yaml`) will be re-generated. In the course of generation, the current key is used along with the key (`--old-secret-key KEY`) that was used to encrypt the data.

```bash
$ dapp kube secret regenerate --old-secret-key c85e100d4ff006b693b0555f09244fdf
```

If the secret values are stored in multiple files, add paths as arguments.

```bash
$ dapp kube secret regenerate --old-secret-key c85e100d4ff006b693b0555f09244fdf .helm/secret-values2.yaml .helm/secret-staging.yaml
```
