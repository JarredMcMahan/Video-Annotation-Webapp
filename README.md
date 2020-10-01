## Education Surgery

### To Run
The code block below could be executed from the atn root, and it would build the project:

```
cd code/backend
go build
cd ../frontend
npm run build
```

Then one would query for their video device name. This would usually be like /dev/video2 on 
Linux or "FHD Capture" on macOS. Move into the directory `code/backend` and run
```
./asv --video-device $MY_DEVICE
```

Where `$MY_DEVICE` is the name of the video device. This will start a server on the computer
running on port 3000. Visit that machine at port 3000 on the local network to start the app.

### Other Requirements
* The host machine currently needs to be linux/macOS
* The client can't be Firefox on macOS, for some reason
* FFMPEG needs to be installed and on the PATH
* USB 3.0+ is _very_ important, unfortunately. It makes a major difference in quality

### To Move to the Greater Internet
Essentially, this project should be ready to connect computers that aren't on the same
local network with just a little tweaking. Specifically, move the HTTP server to it's
own application. That is the majority of the work needed.

