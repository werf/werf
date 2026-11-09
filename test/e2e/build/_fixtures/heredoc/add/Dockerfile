FROM registry.werf.io/base/ubuntu:22.04

ADD <<FILE1 /file1.txt
Content of file 1
FILE1

ADD --chmod=755 <<"EOF" /install.sh
#!/bin/sh
echo "$GREETING"
EOF
