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



