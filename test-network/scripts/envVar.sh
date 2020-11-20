#
# Copyright IBM Corp All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#

# This is a collection of bash functions used by different scripts

source scriptUtils.sh
export ORDERER=orderer.example.com
export ORDERERMSP="NPCIMSP"

export ORDERER_CA=${PWD}/organizations/ordererOrganizations/example.com/orderers/${ORDERER}/msp/tlscacerts/tlsca.example.com-cert.pem

# Set OrdererOrg.Admin globals
setOrdererGlobals() {
  export CORE_PEER_LOCALMSPID=${ORDERERMSP}
  export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/ordererOrganizations/example.com/orderers/${ORDERER}/msp/tlscacerts/tlsca.example.com-cert.pem
  export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/ordererOrganizations/example.com/users/Admin@example.com/msp
}

# Set environment variables for the peer org
setGlobals() {
  local USING_ORG=""
  if [ -z "$OVERRIDE_ORG" ]; then
    USING_ORG=$1
  else
    USING_ORG="${OVERRIDE_ORG}"
  fi
  infoln "Using organization ${USING_ORG}"
  if [ $USING_ORG -eq 1 ]; then
    export ORGMSP="HDFCMSP"
    export ORGDOMAIN=hdfc.example.com
    export PEERPORT=7051
  elif [ $USING_ORG -eq 2 ]; then
    export ORGMSP="SBIMSP"
    export ORGDOMAIN=sbi.example.com
    export PEERPORT=9051
  elif [ $USING_ORG -eq 3 ]; then
    export ORGMSP="AXISMSP"
    export ORGDOMAIN=axis.example.com
    export PEERPORT=11051
  else
    errorln "ORG Unknown"
  fi

  export CORE_PEER_TLS_ENABLED=true
  export CORE_PEER_LOCALMSPID=${ORGMSP}
  export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/${ORGDOMAIN}/peers/peer0.${ORGDOMAIN}/tls/ca.crt
  export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/${ORGDOMAIN}/users/Admin@${ORGDOMAIN}/msp
  export CORE_PEER_ADDRESS=localhost:${PEERPORT}

  if [ "$VERBOSE" == "true" ]; then
    env | grep CORE
  fi
}

# parsePeerConnectionParameters $@
# Helper function that sets the peer connection parameters for a chaincode
# operation
parsePeerConnectionParameters() {

  PEER_CONN_PARMS=""
  PEERS=""
  while [ "$#" -gt 0 ]; do
    setGlobals $1
    # PEER="peer0.org$1"
    ## Set peer addresses
    PEERS="$PEERS $PEER"
    PEER_CONN_PARMS="$PEER_CONN_PARMS --peerAddresses $CORE_PEER_ADDRESS"
    ## Set path to TLS certificate
    TLSINFO=$(eval echo "--tlsRootCertFiles \$CORE_PEER_TLS_ROOTCERT_FILE")
    PEER_CONN_PARMS="$PEER_CONN_PARMS $TLSINFO"
    # shift by one to get to the next organization
    shift
  done
  # remove leading space for output
  PEERS="$(echo -e "$PEERS" | sed -e 's/^[[:space:]]*//')"
}

verifyResult() {
  if [ $1 -ne 0 ]; then
    fatalln "$2"
  fi
}
