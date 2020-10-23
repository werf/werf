```shell
# add ~/bin into PATH
export PATH=$PATH:$HOME/bin
echo 'export PATH=$PATH:$HOME/bin' >> ~/.bashrc

# install multiwerf into ~/bin directory
mkdir -p ~/bin
cd ~/bin
curl -L https://raw.githubusercontent.com/werf/multiwerf/master/get.sh | bash
```

##### Using werf in the current shell

This will create `werf` shell function which calls to the werf binary which multiwerf has been prepared for your session:

```shell
source $(multiwerf use {{ include.version }} {{ include.channel }} --as-file)
werf version
...
```

##### Optional: activate werf on terminal startup

```shell
echo '. $(multiwerf use {{ include.version }} {{ include.channel }} --as-file)' >> ~/.bashrc
```
