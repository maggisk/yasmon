package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)


type Status struct {
	name, value string
}

type RegFunc func(string, func(chan string))

var templateRegex = regexp.MustCompile(`\{\w+\}`)

func formatTemplate(template string, values map[string]string) string {
	r := templateRegex.ReplaceAllFunc([]byte(template), func(key []byte) []byte {
		return []byte(values[string(key)])
	})
	return string(r)
}

func consumer(name string, ch chan string, master chan Status) {
	for {
		value := <-ch
		master <- Status{name: name, value: value}
	}
}

func main() {
	master := make(chan Status)
	values := make(map[string]string)

	template := Configure(func(name string, producer func(chan string)) {
		name = "{"+name+"}"
		values[name] = ""
		ch := make(chan string)
		go producer(ch)
		go consumer(name, ch, master)
	})

	for {
		status := <-master
		if status.value != values[status.name] {
			values[status.name] = status.value
			exec.Command("xsetroot", "-name", formatTemplate(template, values)).Run()
		}
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func bash(cmd string) string {
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Command `%s` failed with `%s`. Maybe there is a dependency missing?\n", cmd, err)
		return ""
	}
	return strings.TrimSpace(string(out))
}

func tick(interval time.Duration) <-chan time.Time {
	// TODO make this more precise - e.g. trigger tick at start of every second
	ch := make(chan time.Time)
	go func() {
		ch <- time.Now()
		for t := range time.Tick(interval) {
			ch <- t
		}
	}()
	return ch
}

func newMonitor(s string, args ...string) *bufio.Reader {
	cmd := exec.Command(s, args...)
	stdout, err := cmd.StdoutPipe()
	checkErr(err)
	reader := bufio.NewReader(stdout)
	cmd.Start()
	return reader
}

func KbLayoutComp(interval time.Duration) func(chan string) {
	return func(ch chan string) {
		for _ = range tick(interval) {
			ch <- bash(`setxkbmap -query | grep layout | sed -n -e 's/^layout:\s*//p'`)
		}
	}
}

func DateComp(format string) func(chan string) {
	return func(ch chan string) {
		for now := range tick(time.Second) {
			ch <- now.Format(format)
		}
	}
}

func BatteryComp(interval time.Duration, formatter func(bool, float64) string) func(chan string) {
	return func(ch chan string) {
		for _ = range tick(interval) {
			charge_full, _ := strconv.Atoi(bash(`cat /sys/class/power_supply/BAT*/charge_full`))
			charge_now, _ := strconv.Atoi(bash(`cat /sys/class/power_supply/BAT*/charge_now`))
			percent := math.Min(1.0, float64(charge_now)/float64(charge_full)) * 100.0
			status := bash(`cat /sys/class/power_supply/BAT*/status`)
			ch <- formatter(status != "Discharging", percent)
		}
	}
}

func BatteryFormat(charging bool, percent float64) string {
	// We could pick more precise icons based on battery percentage but...
	icon := ""
	if charging {
		icon = " "
	} else {
		icon = " "
	}

	return fmt.Sprintf("%s%.0f%%", icon, percent)
}

func VolumeComp(ch chan string) {
	monitor := newMonitor("stdbuf", "-oL", "alsactl", "monitor")
	for {
		ch <- bash(`amixer sget Master | grep 'Left:' | awk -F'[][]' '{ print $2 }'`)
		monitor.ReadLine()
	}
}

func NetworkComp(ch chan string) {
	for _ = range tick(time.Second) {
		ch <- bash(`(ip -br a | grep -v "^lo" | grep -o '[0-9]*\.[0-9\.]*') || echo "no network"`)
	}
}
