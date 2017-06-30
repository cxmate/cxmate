The github.com/ericsage/cxmate/proto package is generated from the cxmate.proto file.

To generate it, you need the protoc v3 and the protoc-gen-go tools.

Download the protoc build (version 3.x) for your platform
and put it in your PATH.

https://github.com/google/protobuf/releases

Use 'go get' to install the latest protoc-gen-go (v4 at time of writing):

$ go get -u github.com/golang/protobuf/protoc-gen-go

Then, run the following from this directory to re-generate the package:

$ protoc -I=. cxmate.proto --go_out=plugins=grpc:.

If you encounter unexpected diffs, such as changes to the
"ProtoPackageIsVersion3" constants, then you may be using the wrong versions of
protoc or protoc-gen-go.
To debug, run 'which protoc' and 'which protoc-gen-go' and check that the
commands in your path are the ones you just installed.
