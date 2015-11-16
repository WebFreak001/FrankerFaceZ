package main // import "bitbucket.org/stendec/frankerfacez/socketserver/cmd/ffzsocketserver"

import (
	"bitbucket.org/stendec/frankerfacez/socketserver/server"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var configFilename = flag.String("config", "config.json", "Configuration file, including the keypairs for the NaCl crypto library, for communicating with the backend.")
var flagGenerateKeys = flag.Bool("genkeys", false, "Generate NaCl keys instead of serving requests.\nArguments: [int serverId] [base64 backendPublic]\nThe backend public key can either be specified in base64 on the command line, or put in the json file later.")

func main() {
	flag.Parse()

	if *flagGenerateKeys {
		generateKeys(*configFilename)
		return
	}

	confFile, err := os.Open(*configFilename)
	if os.IsNotExist(err) {
		fmt.Println("Error: No config file. Run with -genkeys and edit config.json")
		os.Exit(3)
	}
	if err != nil {
		log.Fatal(err)
	}
	conf := &server.ConfigFile{}
	confBytes, err := ioutil.ReadAll(confFile)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(confBytes, &conf)
	if err != nil {
		log.Fatal(err)
	}

	httpServer := &http.Server{
		Addr: conf.ListenAddr,
	}

	logFile, err := os.OpenFile("output.log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal("Could not create logfile: ", err)
	}

	server.SetupServerAndHandle(conf, nil)

	go commandLineConsole()

	log.SetOutput(logFile)

	if conf.UseSSL {
		err = httpServer.ListenAndServeTLS(conf.SSLCertificateFile, conf.SSLKeyFile)
	} else {
		err = httpServer.ListenAndServe()
	}

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func generateKeys(outputFile string) {
	if flag.NArg() < 1 {
		fmt.Println("Specify a numeric server ID after -genkeys")
		os.Exit(2)
	}
	if flag.NArg() >= 2 {
		server.GenerateKeys(outputFile, flag.Arg(0), flag.Arg(1))
	} else {
		server.GenerateKeys(outputFile, flag.Arg(0), "")
	}
	fmt.Println("Keys generated. Now edit config.json")
}
