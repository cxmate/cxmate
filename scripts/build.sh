
OS=( "linux" "darwin" "windows" )
ARCH=( "amd64" )

for os in ${OS[@]}; do
  for arch in ${ARCH[@]}; do
    export GOOS=$os
    export GOARCH=$arch
    go build -o cxmate-${os}-${arch} ..
  done
done
