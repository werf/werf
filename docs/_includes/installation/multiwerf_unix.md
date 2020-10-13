##### Installing multiwerf

```shell
# add ~/bin into PATH
export PATH=$PATH:$HOME/bin
echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc

# install multiwerf into ~/bin directory
mkdir -p ~/bin
cd ~/bin
curl -L https://raw.githubusercontent.com/werf/multiwerf/master/get.sh | bash
```

##### Adding werf alias to the current shell session

```shell
. $(multiwerf use {{ include.version }} {{ include.channel }} --as-file)
```

##### CI usage tip

To ensure that multiwerf exists and is executable, use the `type` command:

```shell
type multiwerf && . $(multiwerf use {{ include.version }} {{ include.channel }} --as-file)
```

The command prints a message to stderr if multiwerf is not found. Thus, diagnostics in a CI environment becomes simpler.

##### Optional: run command on terminal startup

```shell
echo '. $(multiwerf use {{ include.version }} {{ include.channel }} --as-file)' >> ~/.bashrc
```
