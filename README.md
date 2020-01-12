# YASMon

**Y**et **A**nother **S**tatus **Mon**itor

Super simple, modular, customizable/hackable dwm status monitor written in go.

![Demo](https://raw.githubusercontent.com/maggisk/yasmon/master/demo.png)

## How to
clone it and `go build && ./yasmon`

## Customize
In spirit of dwm and simplicity, edit config.go and recompile.
Default config expects dwm to be configured to use a [nerd font](https://www.nerdfonts.com/font-downloads).

## Extend
Each module/component is just a go function accepting a `chan string` that writes to it any time to update the status bar.
See bottom of main.go for examples

## TODO
* document dependencies for each module
* more features (cpu, memory, disk space etc.)
