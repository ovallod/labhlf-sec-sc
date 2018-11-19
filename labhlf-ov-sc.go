package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/lib/cid"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// Labhlfsc Describes the structure of the labhlf
type Labhlfsc struct {
}

// Package Describes the package that will be shipped to the destination
type Package struct {
	PackageID   string `json:"packageId"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Destination string `json:"destination"`
}

var chaincodeVersion = "2018-11-nn"
var logger = shim.NewLogger("SmartContract Labhlfsc with security")

//Init function of the smartcontract
func (t *Labhlfsc) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success([]byte(chaincodeVersion))
}

//Invoke function of the smartcontract : gathers the transactions of the smartcontract
func (t *Labhlfsc) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	// Verification des arguments
	var args = stub.GetArgs()
	var err error

	logger.Info("Function: ", string(args[0]))

	// traitement de la function
	switch string(args[0]) {
	case "OrderShippment":
		// Create a Package asset
		logger.Info(" OrderShippment function, Value:  ", string(args[1]))
		// Check security
		orgsroles := map[string][]string{"org1": {"supplier"}}

		err = checkOrgAndRole(stub, orgsroles)
		if err != nil {
			fmt.Println(err)
			logger.Error(" OrderShippment function, ERROR user not authorized ", err)
			jsonResp := "User not authorized"
			return shim.Error(jsonResp)
		}

		// Check input format
		var pack Package
		err = json.Unmarshal(args[1], &pack)
		if err != nil {
			fmt.Println(err)
			jsonResp := "Failed to Unmarshal package input data"
			return shim.Error(jsonResp)
		}
		// verify mandatory fields
		missingFields := false
		missingFieldsList := "Missing or empty attributes: "
		if pack.Destination == "" {
			missingFieldsList += "Destination"
			missingFields = true
		}
		// ..... Complete the fields checking
		if missingFields {
			fmt.Println(missingFieldsList)
			return shim.Error(missingFieldsList)
		}
		pack.Status = "READY"

		var packageBytes []byte
		packageBytes, err = json.Marshal(pack)
		if err != nil {
			fmt.Println(err)
			return shim.Error("Failed to marshal Package object")
		}

		err = stub.PutState(pack.PackageID, packageBytes)
		if err != nil {
			fmt.Println(err)
			return shim.Error("Failed to write packageBytes data")
		}
		logger.Info(" OrderShippment function, txid: ", stub.GetTxID())
		return shim.Success([]byte(`{"txid":"` + stub.GetTxID() + `","err":null}`))
	case "Ship":
		// Change the status of the package
		// parameters : packageId, status
		logger.Info(" Shipp function, Package id:  ", string(args[1]), " - Status ", string(args[2]))
		// Check security
		orgsroles := map[string][]string{"org2": {"carrier"}}
		err = checkOrgAndRole(stub, orgsroles)
		if err != nil {
			fmt.Println(err)
			jsonResp := "User not authorized"
			return shim.Error(jsonResp)
		}

		// Retrieve the parameters
		var packageID, status string
		var packageBytes []byte
		var pack Package
		packageID = string(args[1])
		status = string(args[2])

		// Test the value of the status : it should be SHIPMENT, SHIPPED, DELIVERED
		// ... TBC .... Implement the test

		// Get the Package in the ledger based on the PackageId
		packagebytes, err := stub.GetState(packageID)
		if err != nil {
			return shim.Error("Failed to get package " + packageID)
		}

		err = json.Unmarshal(packagebytes, &pack)
		if err != nil {
			fmt.Println(err)
			jsonResp := "Failed to Unmarshal package data " + packageID
			return shim.Error(jsonResp)
		}
		// Test the value of the current status :
		// if the new status is SHIPMENT, the current one should be READY
		// if the new status is SHIPPED, the current one should be SHIPMENT
		// if the new status is DELIVERED, the current one should be SHIPPED
		// ... TBC .... Implement the test

		//Update the status
		pack.Status = status

		packageBytes, err = json.Marshal(pack)
		if err != nil {
			fmt.Println(err)
			return shim.Error("Failed to marshal Package object")
		}
		// Update the package in the ledger
		err = stub.PutState(pack.PackageID, packageBytes)
		if err != nil {
			fmt.Println(err)
			return shim.Error("Failed to write packageBytes data")
		}
		logger.Info(" shipp function, txid: ", stub.GetTxID())
		return shim.Success([]byte(`{"txid":"` + stub.GetTxID() + `","err":null}`))
	case "Acknowledgement":
		// Acknowledge the reception of the package
		// parameters : packageId
		logger.Info(" Acknowledgement function, Package id:  ", string(args[1]), " - Status ", string(args[2]))
		// Check security
		orgsroles := map[string][]string{"org1": {"consumer"}}

		err = checkOrgAndRole(stub, orgsroles)
		if err != nil {
			fmt.Println(err)
			jsonResp := "User not authorized"
			return shim.Error(jsonResp)
		}

		// Retrieve the parameters
		var packageID, status string
		var packageBytes []byte
		var pack Package
		packageID = string(args[1])
		status = string(args[2])

		// Get the Package in the ledger based on the PackageId
		packagebytes, err := stub.GetState(packageID)
		if err != nil {
			return shim.Error("Failed to get package " + packageID)
		}

		err = json.Unmarshal(packagebytes, &pack)
		if err != nil {
			fmt.Println(err)
			jsonResp := "Failed to Unmarshal package data " + packageID
			return shim.Error(jsonResp)
		}
		//Update the status
		pack.Status = status

		packageBytes, err = json.Marshal(pack)
		if err != nil {
			fmt.Println(err)
			return shim.Error("Failed to marshal Package object")
		}
		// Update the package in the ledger
		err = stub.PutState(pack.PackageID, packageBytes)
		if err != nil {
			fmt.Println(err)
			return shim.Error("Failed to write packageBytes data")
		}
		logger.Info(" shipp function, txid: ", stub.GetTxID())
		return shim.Success([]byte(`{"txid":"` + stub.GetTxID() + `","err":null}`))
	case "GetPackageStatus":
		// Get the status of the package
		// parameters : packageId
		logger.Info(" GetPackageStatus function, Package id:  ", string(args[1]))
		// Check security
		orgsroles := map[string][]string{
			"org1": {"supplier", "consumer"},
			"org2": {"carrier"}}

		err = checkOrgAndRole(stub, orgsroles)
		if err != nil {
			fmt.Println(err)
			jsonResp := "User not authorized"
			return shim.Error(jsonResp)
		}

		// Retrieve the parameters
		var packageID string
		var pack Package
		packageID = string(args[1])

		// Get the Package in the ledger based on the PackageId
		packagebytes, err := stub.GetState(packageID)
		if err != nil {
			return shim.Error("Failed to get package " + packageID)
		}

		err = json.Unmarshal(packagebytes, &pack)
		if err != nil {
			fmt.Println(err)
			jsonResp := "Failed to Unmarshal package data " + packageID
			return shim.Error(jsonResp)
		}

		return shim.Success([]byte(`{"PackageID":"` + pack.PackageID + `","Status":"` + pack.Status + `"}`))

	default:
		//The transaction name is not known
		return shim.Error("unkwnon function")

	}
}

func main() {
	err := shim.Start(new(Labhlfsc))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}

// =========================================================================================
// checkOrgAndRole Check the user is in the Org and has the Role provided in parameter .
// the orgsroles has the following shape
//      orgsroles := map[string]string{
//    	"org1": {"supplier","consumer"},
//  	"org2": {"carrier"}
//      }
// =========================================================================================
func checkOrgAndRole(stub shim.ChaincodeStubInterface, orgsroles map[string][]string) error {

	clientIdentity, err := cid.New(stub)
	if err != nil {
		logger.Error(" Error trying to retrieve the clientIdentity ")
		return err
	}
	mspid, err := clientIdentity.GetMSPID()

	if err != nil {
		logger.Error(" Error trying to retrieve the mspid ")
		return err
	}
	id, err := clientIdentity.GetID()

	logger.Info("user is", id, " and user.org is ", mspid)
	//Search if one of authorized organization is the one of the user
	roles, ok := orgsroles[mspid]
	if !ok {
		logger.Error("User not in the expected organization ", mspid)
		err = errors.New("Error: user not in the expected organization")
		return err
	}

	//Search if one of authorized roles is the one of the user
	for i, role := range roles {
		logger.Info(" roles: n° ", string(i), " Value: ", role)
		//if the role is the one of the user, return without error
		err = clientIdentity.AssertAttributeValue("role", role)
		if err == nil {
			return nil
		}
	}
	r, ok, _ := clientIdentity.GetAttributeValue("role")
	logger.Error("Error AssertAttributeValue: ", r)
	err = errors.New("Error: role of the user is not authorized")
	return err
}
