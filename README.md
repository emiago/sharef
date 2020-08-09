# Sharef 

Sharef is Command Line Tool for easy sharing files.  
Focus is to have tool that can sync/stream securely file between two machines/containers easily. 
It can help to avoid:
- uploading files to webservice, just to download them on different machine 
- downloading file just to upload on different machine
- making important files public

It uses WebRTC technology underneath.
Webrtc makes encrypted and p2p communication which makes sharing files and streaming highly secure. There is no third-party dependency.
It is totally written in GO language using pion/webrtc library.


### Features list:
- Sending files 
- Sending directories
- Sending multiple files in one command
- Sending and streaming changes of file/directory from sender
- Making connection once send files on the fly

**NOTE:** It is experimental right now. Code should be improved.


# Install

For now only linux users :)

```
sudo wget -P /usr/bin https://github.com/emiraganov/sharef/releases/download/v0.2/sharef
sudo chmod +x /usr/bin/sharef
```

# Usage

Before file streaming can begin, SDP offer and answer must be exchanged. This encoded string
will establish **p2p** connection and it is unique for each session. 

## > Send files/directories
![SENDDEMO](docs/SharefSendDemo.gif)

**Sender:**

```
sharef push file1 file2 dir1 dir2 ...
```

**Receiver:**

You just call pull in directory where you want to get files from sender.
```
sharef pull
```



## >> Keep Us Synced
#### #DEMO

You can make sender to listen for file/dir changes after inital sending. Sender will listen for any changes under file/directory and automaticaly resend changed file. This will keep sender and receiver in sync.
This is useful if you are working on some directory and you want
those changes to be sent to receiver automatically.

*Probably this will be improved but for now do not use this on large files*

**Sender:**

```
sharef push -f file/dir
```




## >|> Make connection once and send on the fly

After exchanching SDP, all sending will be done by deamon running in background.

![SENDDEMO](docs/SharefDeamonDemo.gif)

**Sender:**

Calling this will deamonize sender
```
sharef push -d
```

After making connection with receiver, you send same as:
```
sharef push file
sharef push dir
sharef push file1 file2
...
```

Sharef will detect deamon is running and it will just tell deamon to do the job. 
Deamon is listening on 9876 port by default, using HTTP2 protocol.
BE AWARE receiver will put everything in same directory where it is run.


# Feedback 

Any feedback is welcome!

# Contribute

Building sharef

```
make //build sharef
make test //run unit tests
make integrationtest //run integration tests
```

# References

- https://github.com/pion/webrtc
- https://github.com/Antonito/gfile