package main

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

/*** CORE ***/

var templateRegex = regexp.MustCompile(`\{\w+\}`)

type Status struct {
	name, value string
}

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

func tickSignal(which syscall.Signal) <-chan os.Signal {
	ch := make(chan os.Signal, 1)
	ch <- which
	signal.Notify(ch, which)
	return ch
}

func readFile(path string) string {
	s, err := ioutil.ReadFile(path)
	checkErr(err)
	return strings.TrimSpace(string(s))
}

func atoi(s string) int {
	i, err := strconv.Atoi(s)
	checkErr(err)
	return i
}

func hasNetworkConnection() bool {
	return exec.Command("ping", "-c", "1", "8.8.8.8").Run() == nil
}

func batteryPath() string {
	// Change this if it doesn't work for you!
	return bash(`echo /sys/class/power_supply/BAT*`)
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
	for _ = range tickSignal(syscall.SIGUSR1) {
		ch <- bash(`setxkbmap -query | grep layout | sed -n -e 's/^layout:\s*//p'`)
	}
}

func BatteryPowerComp(interval time.Duration, charging string, discharging string) func(chan string) {
	return func(ch chan string) {
		root := batteryPath()
		for _ = range tickTime(interval) {
			if readFile(path.Join(root, "status")) == "Discharging" {
				ch <- discharging
			} else {
				ch <- charging
			}
		}
	}
}

func BatteryChargeComp(interval time.Duration) func(chan string) {
	return func(ch chan string) {
		root := batteryPath()
		for _ = range tickTime(interval) {
			chargeFull := atoi(readFile(path.Join(root, "charge_full")))
			chargeNow := atoi(readFile(path.Join(root, "charge_now")))
			chargePct := math.Min(100.0, float64(chargeNow)/float64(chargeFull)*100.0)
			ch <- fmt.Sprintf("%.0f", chargePct)
		}
	}
}

func VolumeComp(format func(bool, int) string) func(chan string) {
	return func(ch chan string) {
		for _ = range tickSignal(syscall.SIGUSR1) {
			pair := strings.Fields(bash(`amixer get Master | grep 'Left:' | awk -F'[][%]' '{ print $2, $5 }'`))
			if len(pair) >= 2 {
				volume, err := strconv.Atoi(pair[0])
				checkErr(err)
				ch <- format(pair[1] == "off", volume)
			}
		}
	}
}

func SimpleVolumeFormatter(muted, zero, low, high string) func(bool, int) string {
	return func(isMuted bool, volume int) string {
		icon := ""
		if isMuted {
			icon = muted
		} else if volume == 0 {
			icon = zero
		} else if volume < 50 {
			icon = low
		} else {
			icon = high
		}
		return fmt.Sprintf("%s %d", icon, volume)
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

func HasNetworkComp(interval time.Duration, yes, no string) func(chan string) {
	return func(ch chan string) {
		for _ = range tickTime(interval) {
			if hasNetworkConnection() {
				ch <- yes
			} else {
				ch <- no
			}
		}
	}
}

func BacklightComp(ch chan string) {
	for _ = range tickSignal(syscall.SIGUSR1) {
		line := bash(`xbacklight -get`)
		v, err := strconv.ParseFloat(line, 32)
		if err != nil {
			ch <- line
		} else {
			ch <- fmt.Sprintf("%.0f", v)
		}
	}
}

func BashComp(interval time.Duration, cmd string) func(chan string) {
	return func(ch chan string) {
		for _ = range tickTime(interval) {
			ch <- bash(cmd)
		}
	}
}

// wrapper around component that delays component initialization until we
// have network connection
func NeedsNetwork(component func(chan string)) func(chan string) {
	return func(ch chan string) {
		for !hasNetworkConnection() {
			time.Sleep(time.Second / 10)
		}
		component(ch)
	}
}
