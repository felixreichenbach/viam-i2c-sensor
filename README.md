# Viam Sensehat Module



-> Ran into blocker as custom models cannot use the I2C bus through the board



## Module Creation Commands

```go mod init github.com/felixreichenbach/viam-i2c-sensor```

```go build -o bin/sensehat```

Cross-Compile For Raspberry Pi (to be verified!):
```env GOOS=linux GOARCH=arm64 go build -o bin/sensehat_rapi```



## Install go to compile on the Pi

```wget https://dl.google.com/go/go1.20.7.linux-arm64.tar.gz -O go.tar.gz```


```sudo tar -C /usr/local/ -xzf go.tar.gz```

Add 

PATH=$PATH:/usr/local/go/bin
GOPATH=$HOME/go

to .profile


```source .profile```

## Viam Module Configuration

Module:
```
{
    "name": "sensehat",
    "executable_path": "<-- YOUR PATH -->/bin/sensehat"
}
```

Component:
```
{
    "name": "lsp25h",
    "model": "sensehat:sensor:lps25h",

    "attributes": {},
    "depends_on": [],
    "type": "sensor"
 }
```
