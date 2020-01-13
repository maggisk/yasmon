# YASMon

**Y**et **A**nother **S**tatus **Mon**itor

Super simple, modular, customizable/hackable dwm status monitor written in go.

![Demo](https://raw.githubusercontent.com/maggisk/yasmon/master/demo.png)

Each component runs in a goroutine which makes instant feedback possible of things that can be monitored (e.g. volume via alsactl monitor)

## How to
clone it and `go build && ./yasmon`

## Customize
In spirit of dwm and simplicity, edit config.go and recompile.
Default config expects dwm to be configured to use a [nerd font](https://www.nerdfonts.com/font-downloads).

## Extend
Each component is just a go function accepting a `chan string` that writes to it any time to update the status bar.
See bottom of main.go for examples

## TODO
* document dependencies for each component
* see if we can give instant feedback for more components (ip monitor? keyboard layout?)
* more features (cpu, memory, disk space etc.)
