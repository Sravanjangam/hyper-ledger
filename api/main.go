package main

import (
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"golang.org/x/net/context"
)


type Asset struct {
	DEALERID    string  `json:"DEALERID"`
	MSISDN      string  `json:"MSISDN"`
	MPIN        string  `json:"MPIN"`
	BALANCE     float64 `json:"BALANCE"`
	STATUS      string  `json:"STATUS"`
	TRANSAMOUNT float64 `json:"TRANSAMOUNT"`
	TRANSTYPE   string  `json:"TRANSTYPE"`
	REMARKS     string  `json:"REMARKS"`
}

var contract *client.Contract

func main() {
	
	mspID := os.Getenv("MSP_ID") 
	certPath := os.Getenv("CERT_PATH") 
	keyDir := filepath.Dir(os.Getenv("KEY_PATH")) 
	keyPath := os.Getenv("KEY_PATH") 
	tlsCertPath := os.Getenv("TLS_CERT_PATH") 
	peerEndpoint := os.Getenv("PEER_ENDPOINT")
	gatewayPeer := os.Getenv("GATEWAY_PEER") 
	channelName := "mychannel"
	chaincodeName := "asset"

	
	certPem, err := os.ReadFile(certPath)
	if err != nil {
		panic(fmt.Sprintf("failed to read cert: %v", err))
	}
	cert, err := identity.NewX509Identity(mspID, certPem)
	if err != nil {
		panic(err)
	}

	
	keyPem, err := os.ReadFile(keyPath)
	if err != nil {
		panic(fmt.Sprintf("failed to read key: %v", err))
	}
	sign, err := identity.NewPrivateKeySign(keyPem)
	if err != nil {
		panic(err)
	}

	
	tlsCertPem, err := os.ReadFile(tlsCertPath)
	if err != nil {
		panic(fmt.Sprintf("failed to read TLS cert: %v", err))
	}
	rootCAs := x509.NewCertPool()
	rootCAs.AppendCertsFromPEM(tlsCertPem)
	tlsCreds := credentials.NewClientTLSFromCert(rootCAs, gatewayPeer)

	
	conn, err := grpc.Dial(peerEndpoint, grpc.WithTransportCredentials(tlsCreds))
	if err != nil {
		panic(fmt.Sprintf("failed to create gRPC connection: %v", err))
	}
	defer conn.Close()

	
	gateway, err := client.Connect(
		cert,
		client.WithSign(sign),
		client.WithClientConnection(conn),
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		panic(err)
	}
	defer gateway.Close()

	network := gateway.GetNetwork(channelName)
	contract = network.GetContract(chaincodeName)

	
	http.HandleFunc("/create", createHandler)
	http.HandleFunc("/update", updateHandler)
	http.HandleFunc("/read/", readHandler)
	http.HandleFunc("/history/", historyHandler)
	fmt.Println("REST API listening on :8080")
	http.ListenAndServe(":8080", nil)
}


func createHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var asset Asset
	if err := json.Unmarshal(body, &asset); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, err = contract.SubmitTransaction(
		"CreateAsset",
		asset.DEALERID,
		asset.MSISDN,
		asset.MPIN,
		fmt.Sprintf("%f", asset.BALANCE),
		asset.STATUS,
		fmt.Sprintf("%f", asset.TRANSAMOUNT),
		asset.TRANSTYPE,
		asset.REMARKS,
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to submit transaction: %v", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "Asset created")
}


func updateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var asset Asset
	if err := json.Unmarshal(body, &asset); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, err = contract.SubmitTransaction(
		"UpdateAsset",
		asset.DEALERID,
		asset.MSISDN,
		asset.MPIN,
		fmt.Sprintf("%f", asset.BALANCE),
		asset.STATUS,
		fmt.Sprintf("%f", asset.TRANSAMOUNT),
		asset.TRANSTYPE,
		asset.REMARKS,
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to submit transaction: %v", err), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, "Asset updated")
}


func readHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	dealerID := r.URL.Path[len("/read/"):]
	if dealerID == "" {
		http.Error(w, "dealerID required", http.StatusBadRequest)
		return
	}
	result, err := contract.EvaluateTransaction("ReadAsset", dealerID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to evaluate transaction: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
}


func historyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	dealerID := r.URL.Path[len("/history/"):]
	if dealerID == "" {
		http.Error(w, "dealerID required", http.StatusBadRequest)
		return
	}
	result, err := contract.EvaluateTransaction("GetHistoryForAsset", dealerID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to evaluate transaction: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
}