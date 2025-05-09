package main

import (
	// "log"
	// "syscall"
	// "unsafe"
	"context"
	"fmt"
	"log"
	"msr-helper/pkg/rdmsr"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli/v3"
)

type flagValues struct {
	monitorDuration time.Duration
	monitorMode     bool
	logMode         bool
	hex             bool
	raw             bool
	bitfield        string
}

func main() {

	var flagValues flagValues

	var addr uint64

	var cmd = cli.Command{
		Name:  "rdmsr",
		Usage: "provide a hex msr address",
		Arguments: []cli.Argument{
			&cli.UintArg{
				Name:        "addr",
				Destination: &addr,
				Config:      cli.IntegerConfig{Base: 16},
				Max:         1,
				Min:         1,
			},
		},
		MutuallyExclusiveFlags: []cli.MutuallyExclusiveFlags{
			{Flags: [][]cli.Flag{
				{&cli.BoolFlag{Name: "raw", Aliases: []string{"R"}, Destination: &flagValues.raw}},
			},
			},
			{Flags: [][]cli.Flag{
				{

					&cli.StringFlag{Name: "bitfield", Aliases: []string{"B"}, Destination: &flagValues.bitfield, Validator: func(arg string) error {
						if match, _ := regexp.MatchString(`\d+:\d+`, arg); !match {
							return fmt.Errorf("invalid format. 63:0 expected")
						}
						return nil
					}},
				},
			},
			},
		},
		Flags: []cli.Flag{
			&cli.DurationFlag{Name: "mon", Value: time.Second * 2, Aliases: []string{"M"}, Destination: &flagValues.monitorDuration},
			&cli.BoolFlag{Name: "log", Aliases: []string{"l"}, Usage: "to be used only with monitor mode", Destination: &flagValues.logMode},
			&cli.BoolFlag{Name: "hex", Aliases: []string{"H"}, Destination: &flagValues.hex},
		},

		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.IsSet("mon") && flagValues.monitorDuration != 0 {
				flagValues.monitorMode = true
				vals, err := rdmsr.Monitor(uint32(addr), flagValues.monitorDuration)

				for {
					select {
					case val := <-vals:
						{
							outputMSRValue(val, flagValues)
						}
					case er := <-err:
						{
							if er != nil {
								log.Println(er)
								cli.HandleExitCoder(er)
								return er
							}
						}
					}
				}
			} else {
				msr, err := rdmsr.ReadMsr(uint32(addr))
				if err != nil {
					log.Println(err)
					cli.HandleExitCoder(err)
					return err
				}
				outputMSRValue(msr, flagValues)
			}

			return nil
		},
	}
	cmd.Run(context.Background(), os.Args)
}

func outputMSRValue(msr rdmsr.Msr, flagValues flagValues) {
	var outputString string
	var outputValue uint64

	if flagValues.raw {
		high, low := msr.ToBinary()
		outputString = high + " " + low
	} else {
		if flagValues.bitfield != "" {
			bits := strings.Split(flagValues.bitfield, ":")
			beg, _ := strconv.Atoi(bits[0])
			end, _ := strconv.Atoi(bits[1])
			outputValue = msr.ToBitfield(beg, end)
		} else {
			outputValue = msr.Value()
		}

		if flagValues.hex {
			outputString = fmt.Sprintf("%x", outputValue)
		} else {
			outputString = fmt.Sprintf("%d", outputValue)
		}

	}

	if flagValues.monitorMode && flagValues.logMode {
		fmt.Println(time.Now().Local().Format(time.DateTime), outputString)

	} else {
		fmt.Printf("\033[2K\r%s", outputString) //https://stackoverflow.com/a/52367312
	}

}
