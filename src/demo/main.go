package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	aroz "imuslab.com/arozos/demo/aroz"
)

var (
	handler *aroz.ArozHandler
)

/*
	Demo for showing the implementation of ArOZ Online Subservice Structure

	Proxy url is get from filepath.Dir(StartDir) of the serviceInfo.
	In this example, the proxy path is demo/*
*/

//Kill signal handler. Do something before the system the core terminate.
func SetupCloseHandler() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("\r- Shutting down demo module.")
		//Do other things like close database or opened files

		os.Exit(0)
	}()
}

func main() {
	//If you have other flags, please add them here

	//Start the aoModule pipeline (which will parse the flags as well). Pass in the module launch information
	handler = aroz.HandleFlagParse(aroz.ServiceInfo{
		Name:     "Demo Subservice",
		Desc:     "A simple subservice code for showing how subservice works in ArOZ Online",
		Group:    "Development",
		IconPath: "demo/icon.png",
		Version:  "0.0.1",
		//You can define any path before the actualy html file. This directory (in this case demo/ ) will be the reverse proxy endpoint for this module
		StartDir:     "demo/home.html",
		SupportFW:    true,
		LaunchFWDir:  "demo/home.html",
		SupportEmb:   true,
		LaunchEmb:    "demo/embedded.html",
		InitFWSize:   []int{720, 480},
		InitEmbSize:  []int{720, 480},
		SupportedExt: []string{".txt", ".md"},
	})

	//Register the standard web services urls
	fs := http.FileServer(http.Dir("./web"))
	http.HandleFunc("/api_test", apiTestDemo)
	http.Handle("/", fs)

	//To receive kill signal from the System core, you can setup a close handler to catch the kill signal
	//This is not nessary if you have no opened files / database running
	SetupCloseHandler()

	//Any log println will be shown in the core system via STDOUT redirection. But not STDIN.
	log.Println("Demo module started. Listening on " + handler.Port)
	err := http.ListenAndServe(handler.Port, nil)
	if err != nil {
		log.Fatal(err)
	}

}

//API Test Demo. This showcase how can you access arozos resources with RESTFUL API CALL
func apiTestDemo(w http.ResponseWriter, r *http.Request) {
	//Get username and token from request
	username, token := handler.GetUserInfoFromRequest(w, r)
	log.Println("Received request from: ", username, " with token: ", token)

	//Create an AGI Call that get the user desktop files
	script := `
		if (requirelib("filelib")){
			var filelist = filelib.glob("user:/Desktop/*")
			sendJSONResp(JSON.stringify(filelist));
		}else{
			sendJSONResp(JSON.stringify({
				error: "Filelib require failed"
			}));
		}
	`

	//Execute the AGI request on server side
	resp, err := handler.RequestGatewayInterface(token, script)
	if err != nil {
		//Something went wrong when performing POST request
		log.Println(err)
	} else {
		//Try to read the resp body
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			w.Write([]byte(err.Error()))
			return
		}
		resp.Body.Close()

		//Relay the information to the request using json header
		//Or you can process the information within the go program
		w.Header().Set("Content-Type", "application/json")
		w.Write(bodyBytes)

	}
}
