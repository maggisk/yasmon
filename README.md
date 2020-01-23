# YASMon

**Y**et **A**nother **S**tatus **Mon**itor

Simple, modular, customizable/hackable dwm status monitor written in go.

![Demo](https://raw.githubusercontent.com/maggisk/yasmon/master/demo.png)

## Goal
Show instant feedback when changing volume/backlight/keyboard layout etc. yet let other components
use long update intervals (e.g. weather component making http calls)

## How to
* Install dependencies (see below)
* clone it and `go build && ./yasmon`

Components for keyboard, backlight, volume etc. that are normally controlled through dwm keyboard
actions are only updated when yasmon receives a SIGUSR1 signal. To get instant feedback after making
such changes send a `pkill -xo --signal SIGUSR1 yasmon`

One simple way to do that is to make a bash file defining different commands available through dwm
and send USR1 at the end of that file. e.g.
```
#!/usr/bin/env bash

case "$1" in
    volume)
        amixer set Master "$2"
        ;;
    backlight)
        xbacklight -time 0 -steps 1 -inc "$2"
        ;;
esac

pkill -xo --signal SIGUSR1 yasmon
```

Alternatively, if you just want it to update periodically, you can
```
while true; do
    sleep 1
    pkill -xo --signal SIGUSR1 yasmon
done &
```

##  Builtin components
See `config.go`

## Customize
In spirit of dwm and simplicity, edit config.go and recompile.
Default config expects dwm to be configured to use a [nerd font](https://www.nerdfonts.com/font-downloads).

## Extend
Each component is just a go function accepting a `chan string`. It runs in its own goroutine and
can send a string through the channel at any time to update the status bar.
See bottom of main.go for examples

## Dependencies
* go compiler | [arch](https://www.archlinux.org/packages/community/x86_64/go/)
* xsetroot to send status line to dwm | [arch](https://www.archlinux.org/packages/extra/x86_64/xorg-xsetroot/)
* xbacklight to use BacklightComp | [arch](https://www.archlinux.org/packages/extra/x86_64/xorg-xbacklight/)
* setxkbmap to use KeyboardLayoutComp | [arch](https://www.archlinux.org/packages/extra/x86_64/xorg-setxkbmap/)
* amixer to use VolumeComp | [arch](https://www.archlinux.org/packages/extra/x86_64/alsa-utils/)
