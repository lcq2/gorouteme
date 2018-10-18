# gorouteme
Linux gateway management interface written in Go

## Introduction
GoRouteMe (a pun on goroutine, a very bad one...) is a simple Linux router/gateway management interface. I wrote it mostly for my personal use, after being unable to find a good gateway management interface
for my Linux system.

The main issue was that I did not want a Linux distro, I wanted to use a plain Debian 9 or Alpine Linux, and just download a working interface, I could not find anything that I liked, so I took the radical approach: wrote it myself.

Meanwhile, some friends of mine started to talk me about Go, so I wanted to give it a try and use it for this project. So to kill two birds with one stone, I decided to write the backend interface in Go.

My main goals for this project are:

* Keep the UI as ugly as possible, but functional - I cannot stand anymore gateway/router interfaces with nice window borders, all kind of eye candy, and then they're lacking the most basic features.
So my UI will be kept intentionally ugly and non appealing to the eye, as a silent protest for the current situation. This includes using the worst possible color combinations.

* Easy to use - I plan on using this for my daily use at home, I don't want to spend hours to write an iptables rule.

* As few dependencies as possible - I don't like to reinvent the wheel but my goal was to learn Go, not to learn package management...so I'm only using Go standard library (or whatever is called in Go), with the exception of "golang.org/x/crypto/pbkdf2",
that I needed for login key derivation.

* As few external files as possible - my goal is to run this inside an embedded device (sort-of...). I don't want to copy tens of files around. So the main html templates and static files can be packed into an archive, that will be extracted at runtime to /tmp.

* Try to be modern and to use a modern Linux system as the target

of course some goals might change, but it's just a draft.

## Status
Still work in progress..

## Hardware
My target hardware is an embedded x64 system, in particular my main hardware at the time is the APU2C4 platform (https://www.pcengines.ch/apu2c4.htm) mostly for the high quality Intel NICs.
If you plan on building your own Linux gateway, you should defintely try the APU2 platform, I've been using Pc Engines hardware for network appliances since forever and it never failed me once (no I'm not affiliated at all with it, I just believe that amazing work should be acknowledged).

Everything should work on ARM64 as well, if you're looking for a more low power solution, but finding a good ARM64 board with at least 2 REAL ethernet interfaces (not ethernet over usb crap) is a real challenge. If you find a suitable board let me know...

## Software
For now everything is hardcoded for a Debian 9 system, but I'm planning for some flexibility here. My main development system is Fedora Linux 27, so I continuosly test it also on Fedora.
However I feel that Debian is more suitable for an embedded system configuration.

## Setup and usage
I don't think this is ready yet for general usage, at least until I'm sure that I'm not going to wipe your system by some stupid path mistake...
So I prefer to not write anything here for now...I mean, there's not even a Makefile yet.
