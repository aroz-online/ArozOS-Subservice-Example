# ArozOS-Subservice-Example
An example of an Aroz OS Subservice named "demo"

## Build Instruction
```
# On Linux with Bash
cd src/demo/
./build.sh

# On Windows with CMD
cd src\demo\
build.bat
```

AFter th build complete, you should see the following binaries
- demo.exe
- demo_darwin_amd64
- demo_linux_amd64
- demo_linux_arm
- demo_linux_arm64

The binary name are in the format of {subservice name}_{GOOS}_{GOOARCH}

## How Subservice Launching works
The subservice launching is done by the arozos core binary with exec function call. The sequence is as follow.
1. The arozos check if the binary exists for the current running platform. Exit if binary not found.
2. Execute the binary with -info flag. The subservice binary will return a JSON string of module info and exit.
3. The arozos will start the binary again with two new flags: -port and -rpt. The port define the subservice listinging HTTP port and the rpt defines the RESTFUL API endpoint of the arozos. More information in the RESTFUL section.
4. The subservice should do a blocking loop and run in the background (e.g. start HTTP server ListenAndServe() function)
5. The arozos will then start a reverse proxy to the given port and restart the subservice if proxying failed.

## Module Register JSON (aka -info)
### Create Subservice with aroz go module
To create a subservice with aroz go module (included in src/demo/aroz), involve the aroz module as follow.
```
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
```

The example above register a new module to arozos as "Demo Subservice" and set the reverse proxy path of the module to /demo (extracted from the StartDir paramter)

The meaning of each field is documented in the doc.txt under src/demo/aroz. For short: 
```
type ServiceInfo struct {
	Name         string   //Name of this module. e.g. "Audio"
	Desc         string   //Description for this module
	Group        string   //Group of the module, e.g. "system" / "media" etc
	IconPath     string   //Module icon image path e.g. "Audio/img/function_icon.png"
	Version      string   //Version of the module. Format: [0-9]*.[0-9][0-9].[0-9]
	StartDir     string   //Default starting dir, e.g. "Audio/index.html"
	SupportFW    bool     //Support floatWindow. If yes, floatWindow dir will be loaded
	LaunchFWDir  string   //This link will be launched instead of 'StartDir' if fw mode
	SupportEmb   bool     //Support embedded mode
	LaunchEmb    string   //This link will be launched instead of StartDir / Fw if a file is opened with this module
	InitFWSize   []int    //Floatwindow init size. [0] => Width, [1] => Height
	InitEmbSize  []int    //Embedded mode init size. [0] => Width, [1] => Height
	SupportedExt []string //Supported File Extensions. e.g. ".mp3", ".flac", ".wav"
}
```

After defining the module info, the handler will return the required information on how to start the module. Example usage of the handler is as follow.
```
err := http.ListenAndServe(handler.Port, nil)
	if err != nil {
		log.Fatal(err)
	}
```

### Create a respond JSON manually
See aroz.go HandleFlagParse function for more information.

### Web Request Handling
The subservice host its own web server and handle all its function requests.
You can use http.FileServer and http.HandleFunc as what you usually use in developing other Go applications.

The arozos will rewrite all proxy request to the relative root of your web directory. For example, you registered your StartDir as ```demo/start.html```, then the reverse proxy path to your module will be ```demo/```. 

#### Examples
Assuming your arozos is hosted on localhost:80 and your subservice -port is set to :12810 by the arozos core system.
All request that go thorugh ```http://localhost:80/demo/``` will be forwarded to ```http://localhost:12810/```.

Here are some rewrite examples
| Request URL by client|Forward URL|http.HandleFunc URL|
| ------------- |-------------| -----|
|http://localhost:80/demo/|http://localhost:12810/|/|
|http://localhost:80/demo/test.html|http://localhost:12810/test.html|/test.html|
|http://localhost:80/demo/api/|http://localhost:12810/api/|/api|



## arozos Core System Interaction

To interact with the arozos core system, there are two methods.

1. agi gateway via AJAX request
2. agi gateway via RESTFUL server endpoint

**agi stands for ArOZ Gateway Interface. See arozos documentation for more information*

### Agi gateway via AJAX request

This method is intended for front end code to call for arozos core resources.

The first method is similar to what other WebApps in arozos does: Create an ajax request to the agi gateway endpoint.

To use the agi gateway, create an js file in .js or .agi file extension and place it under a location that is servable by your subservice web server.

In your web interface, do the following.

1. Include ao_module.js from the arozos script folder. ao_module.js also requires jquery. Hence, you can include these two lines in your ```<head>``` section of your UI file.

   ```javascript
   <script type="text/javascript" src="../../script/jquery.min.js"></script>
   <script type="text/javascript" src="../../script/ao_module.js"></script>
   ```

2. Created a js file that is executable by the agi runtime and servable via your own web server relative path. For example,  ```/test.agi```

3. Call to the agi gateway function wrapper and provide the js file relative to your reverse proxy path. For example, 

   ```javascript
    ao_module_agirun("demo/test.agi", {message: "Hello"},function(data){
   	//Do something with the results here
   	console.log(data);
   }
   ```

   In the example above, the test.agi is called with an inserted paramter "message" that can be accessed inside the agi file. See agi section in arozos documentation for more information.



### Agi Gateway via RESTFUL server endpoint

This method is designed for subservice server (aka the golang written server) t call for arozos core resources. 



Example is included in the main.go file. In short, you can create an request to the given endpoint and process the respond using aroz module's RequestGatewayInterface function. Here is an example for that:

```go
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
    panic(err)
} else {
    //Try to read the resp body
    bodyBytes, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        panic(err)
    return
}
resp.Body.Close()

//Parse the json
desktopFileList := []string{}
err = json.Unmarshal(bodyBytes, &desktopFileList)
if err != nil{
	panic(err)
}

//Print the list of this user desktop files
log.Println(desktopFileList)

```

The idea is that there are 3 steps involved in running AGI script in the agi server endpoint.

1. Get the request username and token from the request header (done with ```handler.GetUserInfoFromRequest(w, r)```)
2. Create a request with a section of agi code included. This will be executed in the arozos core AGI runtime virtual machine.
3. The returned value (generated by calling ```sendJSONResp```)  will be the respond body of the HTTP request.

and you can parse the output for further processing.



