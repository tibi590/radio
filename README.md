# Remote Controlled Radio Station (RCRS)

Made for our amateur radio club to control the equipments remotely and see if everything works as it supposed to.

## Configuration
Every configuration can be found in libs/myconst/myconst.go

## Running
Just run `go run .` in the folder where you found `README.md`.<br>
The first time it generates the `pins.txt` with the help of `myconst.MAX_NUMBER_OF_PINS` then quits to let you configure the file.<br>
After that you can restart it and it'll start on port `8080` (if you haven't changed it)

If you want port support you'll need root privileges, so run it with sudo.

You can connect a camera anytime you want (before or even after starting the server), but don't disconnect it because the server will crash (can't do anything about it sadly).

There is a nginx config which is needed if you want to use the page outside of localhost. Listens on port `80`, proxies to `8080`<br>
It uses these paths:
- `/`: Main page (and only usable page)
- `/radio_ws`: websocket connection

## Techinal informations

### Abbreviations
#### Websocket communication:
- `RE`: read error
- `WE`: write error
- `h`: user names who hold button
- `u`: user list
#### Pin file:
- `T`: toggle button
- `P`: push button

### Syntaxts
#### Pin file
```
<name of button>;<status: 0|1|->;<mode: T|P>
...
```

#### Websocket
First character detones the command 
##### User list update
```
u[name of 1. user],...
```
##### Holding list update
```
h[<name of 1. user>;<button number>],...
```
##### Button status update
```
<status of 1. button><status of 2. button>...
```