package main

import (
	"bufio"
	"fmt"
	"log"
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
		name = "{" + name + "}"
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
		fmt.Fprintf(os.Stderr, "Command `%s` failed with error `%s`", cmd, err)
		return ""
	}
	return strings.TrimSpace(string(out))
}

func assertCommandExists(cmd, reason string) {
	_, err := exec.LookPath(cmd)
	if err != nil {
		log.Fatalf("Could not find %s in PATH. It is needed for %s", cmd, reason)
	}
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

func newMonitor(s string, args ...string) *bufio.Scanner {
	cmd := exec.Command(s, args...)
	stdout, err := cmd.StdoutPipe()
	checkErr(err)
	monitor := bufio.NewScanner(stdout)
	checkErr(cmd.Start())
	return monitor
}

func KbLayoutPollComp(interval time.Duration) func(chan string) {
	assertCommandExists("setxkbmap", "detecting keyboard layout")

	return func(ch chan string) {
		for _ = range tick(interval) {
			ch <- getKeyboardLayout()
		}
	}
}

func KbLayoutComp(ch chan string) {
	assertCommandExists("setxkbmap", "detecting keyboard layout")
	assertCommandExists("xev", "watching for keyboard layout changes")

	// I'm not sure how reliable/portable this is and it requires xev
	// but works for me and gives instant keyboard layout indication
	ch <- getKeyboardLayout()
	monitor := newMonitor("xev", "-root", "-event", "property")
	for monitor.Scan() {
		if strings.Contains(monitor.Text(), "_XKB_RULES_NAMES") {
			ch <- getKeyboardLayout()
		}
	}
	checkErr(monitor.Err())
}

func getKeyboardLayout() string {
	return bash(`setxkbmap -query | grep layout | sed -n -e 's/^layout:\s*//p'`)
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
	assertCommandExists("alsactl", "watching for changes in audio volume")
	assertCommandExists("amixer", "detecting audio volume")

	cmd := `amixer sget Master | grep 'Left:' | awk -F'[][]' '{ print $2 }'`
	ch <- bash(cmd)
	monitor := newMonitor("stdbuf", "-oL", "alsactl", "monitor")
	for monitor.Scan() {
		ch <- bash(cmd)
	}
}

func NetworkComp(interval time.Duration) func(chan string) {
	assertCommandExists("ip", "detecting network status")

	return func(ch chan string) {
		for _ = range tick(interval) {
			ch <- bash(`((ip -br a | grep -v "^lo" | grep -o '[0-9]*\.[0-9\.]*') || echo "no network") | head -1`)
		}
	}
}
