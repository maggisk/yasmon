package main

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

/*** CORE ***/

type Status struct {
	name, value string
}

var templateRegex = regexp.MustCompile(`\{\w+\}`)

func formatTemplate(template string, values map[string]string) string {
	r := templateRegex.ReplaceAllFunc([]byte(template), func(key []byte) []byte {
		return []byte(values[string(key)])
	})
	return string(r)
}

func main() {
	master := make(chan Status)
	values := make(map[string]string)

	template := Configure(func(name string, component func(chan string)) {
		name = "{" + name + "}"
		values[name] = ""
		ch := make(chan string)
		go component(ch)
		go func() {
			for {
				value := <-ch
				master <- Status{name: name, value: value}
			}
		}()
	})

	for {
		status := <-master
		// fmt.Printf("%s=%s (was %s)\n", status.name, status.value, values[status.name])
		if status.value != values[status.name] {
			values[status.name] = status.value
			checkErr(exec.Command("xsetroot", "-name", formatTemplate(template, values)).Run())
		}
	}
}

/*** UTILITY FUNCTIONS FOR COMPONENTS ***/

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func bash(cmd string) string {
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Command `%s` failed with error `%s`\n", cmd, err)
		fmt.Fprintln(os.Stderr, "Maybe there's a dependency missing?")
		return "?"
	}
	return strings.TrimSpace(string(out))
}

func tickTime(interval time.Duration) <-chan time.Time {
	intNano := interval.Nanoseconds()
	ch := make(chan time.Time)
	go func() {
		for {
			ch <- time.Now()
			time.Sleep(time.Duration(intNano - (time.Now().UnixNano() % intNano)))
		}
	}()
	return ch
}

func tickSignal() <-chan os.Signal {
	ch := make(chan os.Signal, 1)
	ch <- syscall.SIGUSR1
	signal.Notify(ch, syscall.SIGUSR1)
	return ch
}

/*** COMPONENTS ***/

func DateComp(interval time.Duration, format string) func(chan string) {
	return func(ch chan string) {
		for now := range tickTime(interval) {
			ch <- now.Format(format)
		}
	}
}

func KeyboardLayoutComp(ch chan string) {
	for _ = range tickSignal() {
		ch <- bash(`setxkbmap -query | grep layout | sed -n -e 's/^layout:\s*//p'`)
	}
}

func BatteryPowerComp(interval time.Duration, charging string, discharging string) func(chan string) {
	return func(ch chan string) {
		for _ = range tickTime(interval) {
			if bash(`cat /sys/class/power_supply/BAT*/status`) == "Discharging" {
				ch <- discharging
			} else {
				ch <- charging
			}
		}
	}
}

func BatteryChargeComp(interval time.Duration) func(chan string) {
	return func(ch chan string) {
		for _ = range tickTime(interval) {
			charge_full, err := strconv.Atoi(bash(`cat /sys/class/power_supply/BAT*/charge_full`))
			checkErr(err)
			charge_now, err := strconv.Atoi(bash(`cat /sys/class/power_supply/BAT*/charge_now`))
			checkErr(err)
			charge := math.Min(100.0, float64(charge_now)/float64(charge_full)*100.0)
			ch <- fmt.Sprintf("%.0f", charge)
		}
	}
}

func VolumeComp(ch chan string) {
	for _ = range tickSignal() {
		ch <- bash(`amixer sget Master | grep 'Left:' | awk -F'[][%]' '{ print $2 }'`)
	}
}

func NetworkComp(interval time.Duration) func(chan string) {
	return func(ch chan string) {
		for _ = range tickTime(interval) {
			// this is a surprisingly complex topic - this command might not work for everyone
			ip := bash(`ip route get 1 | sed -n 's/^.*src \([0-9.]*\) .*$/\1/p'`)
			if ip == "" {
				ip = "no network"
			}
			ch <- ip
		}
	}
}

func BacklightComp(ch chan string) {
	for _ = range tickSignal() {
		ch <- strings.Split(bash(`xbacklight -get`), ".")[0]
	}
}

func BashComp(interval time.Duration, cmd string) func(chan string) {
	return func(ch chan string) {
		for _ = range tickTime(interval) {
			ch <- bash(cmd)
		}
	}
}
