package main

import (
	"bufio"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var (
	serverStdinPath = "spigot-server.in"
	serverStdoutPath = "spigot-server.out"
	log = logrus.New()
)

func getStdioFiles(cleanUpOld bool) (in *os.File, out *os.File, err error) {
	if cleanUpOld {
		_ = os.Remove(serverStdinPath)
		_ = os.Remove(serverStdoutPath)
	}

 	in, err = os.OpenFile(serverStdinPath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		return
	}

	out, err = os.OpenFile(serverStdoutPath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		return
	}

	return
}

func captureOutput(stdout io.Reader) (err error) {
	go func(){
		buff := bufio.NewReader(stdout)
		var b []byte

		for {
			if b, err = ioutil.ReadAll(buff); err != nil {
				return
			}

			if len(b) > 0 {
				if _, err = os.Stdout.Write(b); err != nil {
					return
				}
			}

			time.Sleep(time.Millisecond * 200)
		}
	}()

	return
}

var StartCommandAction cli.ActionFunc = func(context *cli.Context) error {
	spigotJarPath := context.String("spigot-path")

	in, out, err := getStdioFiles(true)

	if err != nil {
		return err
	}

	cmd := exec.Command("java", "-jar", spigotJarPath, "nogui")
	cmd.Stdin = in
	cmd.Stdout = io.MultiWriter(out, os.Stdout)
	cmd.Stderr = os.Stderr

	err = cmd.Run()

	if err != nil {
		return  err
	}
	return nil
}

var StopCommandAction cli.ActionFunc = func(context *cli.Context) error {
	in, _, err := getStdioFiles(false)

	if err != nil {
		return err
	}

	log.Info("Stopping the server...")
	// пишем команду остановки сервера в консоль сервера
	if _, err := in.Write([]byte("stop\n")); err != nil {
		return err
	}
	return nil
}

var ConsoleCommandAction cli.ActionFunc = func(context *cli.Context) error {
	in, out, err := getStdioFiles(false)

	if err != nil {
		return err
	}

	w := bufio.NewWriter(in)
	err = captureOutput(out)

	if err != nil {
		return  err
	}

	for {
		line, _ := bufio.NewReader(os.Stdin).ReadString('\n')

		if line == "exit\n" {
			break
		}

		if _, err := w.WriteString(line); err != nil {
			return err
		}

		if err := w.Flush(); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	app := &cli.App{
		Name: "spigot-cli",
		Usage: "commands for spigot mc server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "spigot-path",
				Aliases: []string{"sp"},
				Usage: "Path to spigot jar `FILE`",
				Value: "spigot.jar",
			},
		},
		Commands: []*cli.Command{
			{
				Name: "start",
				Usage: "Start spigot server",
				Action: StartCommandAction,
			},
			{
				Name: "stop",
				Usage: "Stop spigot server",
				Action: StopCommandAction,
			},
			{
				Name: "console",
				Usage: "Launch server console",
				Action: ConsoleCommandAction,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
