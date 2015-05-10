package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/EricLagerg/go-gnulib/stdlib"
	"github.com/EricLagerg/go-gnulib/utmp"

	flag "github.com/ogier/pflag"
)

const (
	Help1 = `Usage: uptime [OPTION]... [FILE]
Print the current time, the length of time the system has been up,
the number of users on the system, and the average number of jobs
in the run queue over the last 1, 5 and 15 minutes.  Processes in
an uninterruptible sleep state also contribute to the load average.
If FILE is not specified, use`
	Help2 = `as FILE is common.

      --help     display this help and exit
      --version  output version information and exit

Report wc bugs to ericscottlagergren@gmail.com
Go coreutils home page: <https://www.github.com/EricLagerg/go-coreutils/>
`

	Version = `
	uptime (Go coreutils) 1.0
Copyright (C) 2015 Eric Lagergren
License GPLv3+: GNU GPL version 3 or later <http://gnu.org/licenses/gpl.html>.
This is free software: you are free to change and redistribute it.
There is NO WARRANTY, to the extent permitted by law.
`

	delim = " "
)

var (
	version = flag.BoolP("version", "v", false, "")

	// fatal = log.New(os.Stderr, "", log.Lshortfile)
	fatal = log.New(os.Stderr, "", 0)
)

func printUptime(us []utmp.Utmp) {

	var (
		bootTime int32
		entries  int64
		now      utmp.TimeVal

		days, hours, mins int
		uptime            float64
	)

	file, err := os.Open("/proc/uptime")
	if err != nil {
		fatal.Fatalln(err)
	}
	defer file.Close()

	buf := make([]byte, 256)

	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		fatal.Fatalln(err)
	}

	// /proc/uptime's output is in the format of "%f %f\n"
	// The first space in the buffer will be the end of the first number
	line := string(buf[:bytes.IndexByte(buf[:n], ' ')])

	secs, err := strconv.ParseFloat(line, 64)
	if err != nil {
		fatal.Fatalln(err)
	}

	if 0 <= secs || secs < math.MaxFloat64 {
		uptime = secs
	} else {
		uptime = -1
	}

	for _, v := range us {

		if v.IsUserProcess() {
			entries++
		}

		if v.Type == utmp.BootTime {
			bootTime = v.Time.Sec
		}
	}

	now.GetTimeOfDay()
	if uptime == 0 {
		if bootTime == 0 {
			fatal.Fatalln("can't get boot time")
		}

		uptime = float64(now.Sec - bootTime)
	}

	days = int(uptime) / 86400
	hours = (int(uptime) - (days * 86400)) / 3600
	mins = (int(uptime) - (days * 86400) - (hours * 3600)) / 60

	fmt.Print(time.Now().Local().Format(" 15:04pm  "))

	if uptime == -1 {
		fmt.Print("up ???? days ??:??,  ")
	} else {
		if 0 < days {
			if days > 1 {
				fmt.Printf("up %d days %2d:%02d,  ", days, hours, mins)
			} else {
				fmt.Printf("up %d day %2d:%02d,  ", days, hours, mins)
			}
		} else {
			fmt.Printf("up  %2d:%02d,  ", hours, mins)
		}
	}

	if len(us) > 1 || len(us) == 0 {
		fmt.Printf("%d users", entries)
	} else {
		fmt.Printf("%d user", entries)
	}

	var avg [3]float64
	loads := stdlib.GetLoadAvg(&avg)

	if loads == -1 {
		fmt.Printf("%s", "\n")
	} else {
		if loads > 0 {
			fmt.Printf(",  load average: %.2f", avg[0])
		}

		if loads > 1 {
			fmt.Printf(", %.2f", avg[1])
		}

		if loads > 2 {
			fmt.Printf(", %.2f", avg[2])
		}

		if loads > 0 {
			fmt.Printf("%s", "\n")
		}
	}
}

func uptime(fname string, opts int) {
	entries := uint64(0)
	us := make([]utmp.Utmp, 0)
	err := utmp.ReadUtmp(fname, &entries, &us, opts)
	if err != nil {
		fatal.Fatalln(err)
	}

	printUptime(us)
}

func main() {
	flag.Usage = func() {
		// This is a little weird because I want to insert the correct
		// UTMP/WTMP file names into the Help output, but usually my
		// Help constants are raw string literals, so I had to
		// break it up into a couple chunks and move around some formatting.
		fmt.Fprintf(os.Stderr, "%s %s.  %s %s",
			Help1, utmp.UtmpFile, utmp.WtmpFile, Help2)
		os.Exit(1)
	}
	flag.Parse()

	if *version {
		fmt.Printf("%s\n", Version)
		os.Exit(0)
	}

	switch flag.NArg() {
	case 0:
		uptime(utmp.UtmpFile, utmp.CheckPIDs)
	case 1:
		uptime(flag.Arg(0), 0)
	default:
		fatal.Fatalf("extra operand %s\n", flag.Arg(1))
	}
}
