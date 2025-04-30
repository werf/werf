#/bin/bash

export LD_LIBRARY_PATH=./lib 
export PATH=$PWD:$PATH
export WERF_ELF_PGP_PRIVATE_KEY_BASE64=$(cat key)
werf build --sign-elf-files --dev --repo localhost:5000/test
