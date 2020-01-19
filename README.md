# YASMon

**Y**et **A**nother **S**tatus **Mon**itor

Simple, modular, customizable/hackable dwm status monitor written in go.

![Demo](https://raw.githubusercontent.com/maggisk/yasmon/master/demo.png)

Each component runs in its own goroutine so you can configure different update intervals for
different components.

## Goal
Show instant feedback when changing volume/backlight/keyboard layout etc. yet let other components
use long update intervals (e.g. weather component making http calls)

## How to
clone it and `go build && ./yasmon`

Components for keyboard, backlight, volume etc. that are normally controlled through dwm keyboard
actions are only updated when yasmon receives a SIGUSR1 signal. To get instant feedback after making
such changes send a `killall -s USR1 yasmon`

One simple way do to that is to make a bash file defining different action available through dwm
and send USR1 at the end of that file. e.g.
```
#!/usr/bin/env bash

case "$1" in
    volume)
        amixer set Master "$2"
        ;;
    backlight)
        xbacklight -inc "$2"
        ;;
esac

killall -s USR1 yasmon
```

## Customize
In spirit of dwm and simplicity, edit config.go and recompile.
Default config expects dwm to be configured to use a [nerd font](https://www.nerdfonts.com/font-downloads).

## Extend
Each component is just a go function accepting a `chan string` that writes to it any time to update the status bar.
See bottom of main.go for examples
