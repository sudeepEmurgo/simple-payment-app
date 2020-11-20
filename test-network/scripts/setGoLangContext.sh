
export CC_RUNTIME_LANGUAGE=golang
export CC_SRC_PATH="../spa-chaincode-pdc/chaincode-go"
export VERSION=1

echo Vendoring Go dependencies ...
pushd ../spa-chaincode-pdc/chaincode-go
export GO111MODULE=on go mod vendor
popd
echo Finished vendoring Go dependencies