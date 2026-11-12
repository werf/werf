FROM registry.werf.io/base/ubuntu:22.04
ARG GREETING=hello

COPY <<FILE1 /file1.txt
Content of file 1
FILE1

COPY --chmod=755 <<"EOF" /install.sh
#!/bin/sh
echo "$GREETING"
EOF

COPY --chown=1000:1000 <<EOF /owned.txt
owned
EOF

COPY <<EOF /expanded.txt
$GREETING world
EOF

COPY context.txt <<A <<B /dir/
a
A
b
B
