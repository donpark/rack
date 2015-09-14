package cli_test

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/convox/cli/Godeps/_workspace/src/github.com/codegangsta/cli"
)

func ExampleApp() {
	// set args for examples sake
	os.Args = []string{"greet", "--name", "Jeremy"}

	app := cli.NewApp()
	app.Name = "greet"
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "name", Value: "bob", Usage: "a name to say"},
	}
	app.Action = func(c *cli.Context) {
		fmt.Printf("Hello %v\n", c.String("name"))
	}
	app.Author = "Harrison"
	app.Email = "harrison@lolwut.com"
	app.Authors = []cli.Author{cli.Author{Name: "Oliver Allen", Email: "oliver@toyshop.com"}}
	app.Run(os.Args)
	// Output:
	// Hello Jeremy
}

func ExampleAppSubcommand() {
	// set args for examples sake
	os.Args = []string{"say", "hi", "english", "--name", "Jeremy"}
	app := cli.NewApp()
	app.Name = "say"
	app.Commands = []cli.Command{
		{
			Name:        "hello",
			Aliases:     []string{"hi"},
			Usage:       "use it to see a description",
			Description: "This is how we describe hello the function",
			Subcommands: []cli.Command{
				{
					Name:        "english",
					Aliases:     []string{"en"},
					Usage:       "sends a greeting in english",
					Description: "greets someone in english",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "name",
							Value: "Bob",
							Usage: "Name of the person to greet",
						},
					},
					Action: func(c *cli.Context) {
						fmt.Println("Hello,", c.String("name"))
					},
				},
			},
		},
	}

	app.Run(os.Args)
	// Output:
	// Hello, Jeremy
}

func ExampleAppHelp() {
	// set args for examples sake
	os.Args = []string{"greet", "h", "describeit"}

	app := cli.NewApp()
	app.Name = "greet"
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "name", Value: "bob", Usage: "a name to say"},
	}
	app.Commands = []cli.Command{
		{
			Name:        "describeit",
			Aliases:     []string{"d"},
			Usage:       "use it to see a description",
			Description: "This is how we describe describeit the function",
			Action: func(c *cli.Context) {
				fmt.Printf("i like to describe things")
			},
		},
	}
	app.Run(os.Args)
	// Output:
	// NAME:
	//    describeit - use it to see a description
	//
	// USAGE:
	//    command describeit [arguments...]
	//
	// DESCRIPTION:
	//    This is how we describe describeit the function
}

func ExampleAppBashComplete() {
	// set args for examples sake
	os.Args = []string{"greet", "--generate-bash-completion"}

	app := cli.NewApp()
	app.Name = "greet"
	app.EnableBashCompletion = true
	app.Commands = []cli.Command{
		{
			Name:        "describeit",
			Aliases:     []string{"d"},
			Usage:       "use it to see a description",
			Description: "This is how we describe describeit the function",
			Action: func(c *cli.Context) {
				fmt.Printf("i like to describe things")
			},
		}, {
			Name:        "next",
			Usage:       "next example",
			Description: "more stuff to see when generating bash completion",
			Action: func(c *cli.Context) {
				fmt.Printf("the next example")
			},
		},
	}

	app.Run(os.Args)
	// Output:
	// describeit
	// d
	// next
	// help
	// h
}

func TestApp_Run(t *testing.T) {
	s := ""

	app := cli.NewApp()
	app.Action = func(c *cli.Context) {
		s = s + c.Args().First()
	}

	err := app.Run([]string{"command", "foo"})
	expect(t, err, nil)
	err = app.Run([]string{"command", "bar"})
	expect(t, err, nil)
	expect(t, s, "foobar")
}

var commandAppTests = []struct {
	name     string
	expected bool
}{
	{"foobar", true},
	{"batbaz", true},
	{"b", true},
	{"f", true},
	{"bat", false},
	{"nothing", false},
}

func TestApp_Command(t *testing.T) {
	app := cli.NewApp()
	fooCommand := cli.Command{Name: "foobar", Aliases: []string{"f"}}
	batCommand := cli.Command{Name: "batbaz", Aliases: []string{"b"}}
	app.Commands = []cli.Command{
		fooCommand,
		batCommand,
	}

	for _, test := range commandAppTests {
		expect(t, app.Command(test.name) != nil, test.expected)
	}
}

func TestApp_CommandWithArgBeforeFlags(t *testing.T) {
	var parsedOption, firstArg string

	app := cli.NewApp()
	command := cli.Command{
		Name: "cmd",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "option", Value: "", Usage: "some option"},
		},
		Action: func(c *cli.Context) {
			parsedOption = c.String("option")
			firstArg = c.Args().First()
		},
	}
	app.Commands = []cli.Command{command}

	app.Run([]string{"", "cmd", "my-arg", "--option", "my-option"})

	expect(t, parsedOption, "my-option")
	expect(t, firstArg, "my-arg")
}

func TestApp_RunAsSubcommandParseFlags(t *testing.T) {
	var context *cli.Context

	a := cli.NewApp()
	a.Commands = []cli.Command{
		{
			Name: "foo",
			Action: func(c *cli.Context) {
				context = c
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "lang",
					Value: "english",
					Usage: "language for the greeting",
				},
			},
			Before: func(_ *cli.Context) error { return nil },
		},
	}
	a.Run([]string{"", "foo", "--lang", "spanish", "abcd"})

	expect(t, context.Args().Get(0), "abcd")
	expect(t, context.String("lang"), "spanish")
}

func TestApp_CommandWithFlagBeforeTerminator(t *testing.T) {
	var parsedOption string
	var args []string

	app := cli.NewApp()
	command := cli.Command{
		Name: "cmd",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "option", Value: "", Usage: "some option"},
		},
		Action: func(c *cli.Context) {
			parsedOption = c.String("option")
			args = c.Args()
		},
	}
	app.Commands = []cli.Command{command}

	app.Run([]string{"", "cmd", "my-arg", "--option", "my-option", "--", "--notARealFlag"})

	expect(t, parsedOption, "my-option")
	expect(t, args[0], "my-arg")
	expect(t, args[1], "--")
	expect(t, args[2], "--notARealFlag")
}

func TestApp_CommandWithNoFlagBeforeTerminator(t *testing.T) {
	var args []string

	app := cli.NewApp()
	command := cli.Command{
		Name: "cmd",
		Action: func(c *cli.Context) {
			args = c.Args()
		},
	}
	app.Commands = []cli.Command{command}

	app.Run([]string{"", "cmd", "my-arg", "--", "notAFlagAtAll"})

	expect(t, args[0], "my-arg")
	expect(t, args[1], "--")
	expect(t, args[2], "notAFlagAtAll")
}

func TestApp_Float64Flag(t *testing.T) {
	var meters float64

	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.Float64Flag{Name: "height", Value: 1.5, Usage: "Set the height, in meters"},
	}
	app.Action = func(c *cli.Context) {
		meters = c.Float64("height")
	}

	app.Run([]string{"", "--height", "1.93"})
	expect(t, meters, 1.93)
}

func TestApp_ParseSliceFlags(t *testing.T) {
	var parsedOption, firstArg string
	var parsedIntSlice []int
	var parsedStringSlice []string

	app := cli.NewApp()
	command := cli.Command{
		Name: "cmd",
		Flags: []cli.Flag{
			cli.IntSliceFlag{Name: "p", Value: &cli.IntSlice{}, Usage: "set one or more ip addr"},
			cli.StringSliceFlag{Name: "ip", Value: &cli.StringSlice{}, Usage: "set one or more ports to open"},
		},
		Action: func(c *cli.Context) {
			parsedIntSlice = c.IntSlice("p")
			parsedStringSlice = c.StringSlice("ip")
			parsedOption = c.String("option")
			firstArg = c.Args().First()
		},
	}
	app.Commands = []cli.Command{command}

	app.Run([]string{"", "cmd", "my-arg", "-p", "22", "-p", "80", "-ip", "8.8.8.8", "-ip", "8.8.4.4"})

	IntsEquals := func(a, b []int) bool {
		if len(a) != len(b) {
			return false
		}
		for i, v := range a {
			if v != b[i] {
				return false
			}
		}
		return true
	}

	StrsEquals := func(a, b []string) bool {
		if len(a) != len(b) {
			return false
		}
		for i, v := range a {
			if v != b[i] {
				return false
			}
		}
		return true
	}
	var expectedIntSlice = []int{22, 80}
	var expectedStringSlice = []string{"8.8.8.8", "8.8.4.4"}

	if !IntsEquals(parsedIntSlice, expectedIntSlice) {
		t.Errorf("%v does not match %v", parsedIntSlice, expectedIntSlice)
	}

	if !StrsEquals(parsedStringSlice, expectedStringSlice) {
		t.Errorf("%v does not match %v", parsedStringSlice, expectedStringSlice)
	}
}

func TestApp_ParseSliceFlagsWithMissingValue(t *testing.T) {
	var parsedIntSlice []int
	var parsedStringSlice []string

	app := cli.NewApp()
	command := cli.Command{
		Name: "cmd",
		Flags: []cli.Flag{
			cli.IntSliceFlag{Name: "a", Usage: "set numbers"},
			cli.StringSliceFlag{Name: "str", Usage: "set strings"},
		},
		Action: func(c *cli.Context) {
			parsedIntSlice = c.IntSlice("a")
			parsedStringSlice = c.StringSlice("str")
		},
	}
	app.Commands = []cli.Command{command}

	app.Run([]string{"", "cmd", "my-arg", "-a", "2", "-str", "A"})

	var expectedIntSlice = []int{2}
	var expectedStringSlice = []string{"A"}

	if parsedIntSlice[0] != expectedIntSlice[0] {
		t.Errorf("%v does not match %v", parsedIntSlice[0], expectedIntSlice[0])
	}

	if parsedStringSlice[0] != expectedStringSlice[0] {
		t.Errorf("%v does not match %v", parsedIntSlice[0], expectedIntSlice[0])
	}
}

func TestApp_DefaultStdout(t *testing.T) {
	app := cli.NewApp()

	if app.Writer != os.Stdout {
		t.Error("Default output writer not set.")
	}
}

type mockWriter struct {
	written []byte
}

func (fw *mockWriter) Write(p []byte) (n int, err error) {
	if fw.written == nil {
		fw.written = p
	} else {
		fw.written = append(fw.written, p...)
	}

	return len(p), nil
}

func (fw *mockWriter) GetWritten() (b []byte) {
	return fw.written
}

func TestApp_SetStdout(t *testing.T) {
	w := &mockWriter{}

	app := cli.NewApp()
	app.Name = "test"
	app.Writer = w

	err := app.Run([]string{"help"})

	if err != nil {
		t.Fatalf("Run error: %s", err)
	}

	if len(w.written) == 0 {
		t.Error("App did not write output to desired writer.")
	}
}

func TestApp_BeforeFunc(t *testing.T) {
	beforeRun, subcommandRun := false, false
	beforeError := fmt.Errorf("fail")
	var err error

	app := cli.NewApp()

	app.Before = func(c *cli.Context) error {
		beforeRun = true
		s := c.String("opt")
		if s == "fail" {
			return beforeError
		}

		return nil
	}

	app.Commands = []cli.Command{
		cli.Command{
			Name: "sub",
			Action: func(c *cli.Context) {
				subcommandRun = true
			},
		},
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "opt"},
	}

	// run with the Before() func succeeding
	err = app.Run([]string{"command", "--opt", "succeed", "sub"})

	if err != nil {
		t.Fatalf("Run error: %s", err)
	}

	if beforeRun == false {
		t.Errorf("Before() not executed when expected")
	}

	if subcommandRun == false {
		t.Errorf("Subcommand not executed when expected")
	}

	// reset
	beforeRun, subcommandRun = false, false

	// run with the Before() func failing
	err = app.Run([]string{"command", "--opt", "fail", "sub"})

	// should be the same error produced by the Before func
	if err != beforeError {
		t.Errorf("Run error expected, but not received")
	}

	if beforeRun == false {
		t.Errorf("Before() not executed when expected")
	}

	if subcommandRun == true {
		t.Errorf("Subcommand executed when NOT expected")
	}

}

func TestApp_AfterFunc(t *testing.T) {
	afterRun, subcommandRun := false, false
	afterError := fmt.Errorf("fail")
	var err error

	app := cli.NewApp()

	app.After = func(c *cli.Context) error {
		afterRun = true
		s := c.String("opt")
		if s == "fail" {
			return afterError
		}

		return nil
	}

	app.Commands = []cli.Command{
		cli.Command{
			Name: "sub",
			Action: func(c *cli.Context) {
				subcommandRun = true
			},
		},
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "opt"},
	}

	// run with the After() func succeeding
	err = app.Run([]string{"command", "--opt", "succeed", "sub"})

	if err != nil {
		t.Fatalf("Run error: %s", err)
	}

	if afterRun == false {
		t.Errorf("After() not executed when expected")
	}

	if subcommandRun == false {
		t.Errorf("Subcommand not executed when expected")
	}

	// reset
	afterRun, subcommandRun = false, false

	// run with the Before() func failing
	err = app.Run([]string{"command", "--opt", "fail", "sub"})

	// should be the same error produced by the Before func
	if err != afterError {
		t.Errorf("Run error expected, but not received")
	}

	if afterRun == false {
		t.Errorf("After() not executed when expected")
	}

	if subcommandRun == false {
		t.Errorf("Subcommand not executed when expected")
	}
}

func TestAppNoHelpFlag(t *testing.T) {
	oldFlag := cli.HelpFlag
	defer func() {
		cli.HelpFlag = oldFlag
	}()

	cli.HelpFlag = cli.BoolFlag{}

	app := cli.NewApp()
	err := app.Run([]string{"test", "-h"})

	if err != flag.ErrHelp {
		t.Errorf("expected error about missing help flag, but got: %s (%T)", err, err)
	}
}

func TestAppHelpPrinter(t *testing.T) {
	oldPrinter := cli.HelpPrinter
	defer func() {
		cli.HelpPrinter = oldPrinter
	}()

	var wasCalled = false
	cli.HelpPrinter = func(w io.Writer, template string, data interface{}) {
		wasCalled = true
	}

	app := cli.NewApp()
	app.Run([]string{"-h"})

	if wasCalled == false {
		t.Errorf("Help printer expected to be called, but was not")
	}
}

func TestAppVersionPrinter(t *testing.T) {
	oldPrinter := cli.VersionPrinter
	defer func() {
		cli.VersionPrinter = oldPrinter
	}()

	var wasCalled = false
	cli.VersionPrinter = func(c *cli.Context) {
		wasCalled = true
	}

	app := cli.NewApp()
	ctx := cli.NewContext(app, nil, nil)
	cli.ShowVersion(ctx)

	if wasCalled == false {
		t.Errorf("Version printer expected to be called, but was not")
	}
}

func TestAppCommandNotFound(t *testing.T) {
	beforeRun, subcommandRun := false, false
	app := cli.NewApp()

	app.CommandNotFound = func(c *cli.Context, command string) {
		beforeRun = true
	}

	app.Commands = []cli.Command{
		cli.Command{
			Name: "bar",
			Action: func(c *cli.Context) {
				subcommandRun = true
			},
		},
	}

	app.Run([]string{"command", "foo"})

	expect(t, beforeRun, true)
	expect(t, subcommandRun, false)
}

func TestGlobalFlag(t *testing.T) {
	var globalFlag string
	var globalFlagSet bool
	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "global, g", Usage: "global"},
	}
	app.Action = func(c *cli.Context) {
		globalFlag = c.GlobalString("global")
		globalFlagSet = c.GlobalIsSet("global")
	}
	app.Run([]string{"command", "-g", "foo"})
	expect(t, globalFlag, "foo")
	expect(t, globalFlagSet, true)

}

func TestGlobalFlagsInSubcommands(t *testing.T) {
	subcommandRun := false
	parentFlag := false
	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.BoolFlag{Name: "debug, d", Usage: "Enable debugging"},
	}

	app.Commands = []cli.Command{
		cli.Command{
			Name: "foo",
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "parent, p", Usage: "Parent flag"},
			},
			Subcommands: []cli.Command{
				{
					Name: "bar",
					Action: func(c *cli.Context) {
						if c.GlobalBool("debug") {
							subcommandRun = true
						}
						if c.GlobalBool("parent") {
							parentFlag = true
						}
					},
				},
			},
		},
	}

	app.Run([]string{"command", "-d", "foo", "-p", "bar"})

	expect(t, subcommandRun, true)
	expect(t, parentFlag, true)
}

func TestApp_Run_CommandWithSubcommandHasHelpTopic(t *testing.T) {
	var subcommandHelpTopics = [][]string{
		{"command", "foo", "--help"},
		{"command", "foo", "-h"},
		{"command", "foo", "help"},
	}

	for _, flagSet := range subcommandHelpTopics {
		t.Logf("==> checking with flags %v", flagSet)

		app := cli.NewApp()
		buf := new(bytes.Buffer)
		app.Writer = buf

		subCmdBar := cli.Command{
			Name:  "bar",
			Usage: "does bar things",
		}
		subCmdBaz := cli.Command{
			Name:  "baz",
			Usage: "does baz things",
		}
		cmd := cli.Command{
			Name:        "foo",
			Description: "descriptive wall of text about how it does foo things",
			Subcommands: []cli.Command{subCmdBar, subCmdBaz},
		}

		app.Commands = []cli.Command{cmd}
		err := app.Run(flagSet)

		if err != nil {
			t.Error(err)
		}

		output := buf.String()
		t.Logf("output: %q\n", buf.Bytes())

		if strings.Contains(output, "No help topic for") {
			t.Errorf("expect a help topic, got none: \n%q", output)
		}

		for _, shouldContain := range []string{
			cmd.Name, cmd.Description,
			subCmdBar.Name, subCmdBar.Usage,
			subCmdBaz.Name, subCmdBaz.Usage,
		} {
			if !strings.Contains(output, shouldContain) {
				t.Errorf("want help to contain %q, did not: \n%q", shouldContain, output)
			}
		}
	}
}

func TestApp_Run_SubcommandFullPath(t *testing.T) {
	app := cli.NewApp()
	buf := new(bytes.Buffer)
	app.Writer = buf

	subCmd := cli.Command{
		Name:  "bar",
		Usage: "does bar things",
	}
	cmd := cli.Command{
		Name:        "foo",
		Description: "foo commands",
		Subcommands: []cli.Command{subCmd},
	}
	app.Commands = []cli.Command{cmd}

	err := app.Run([]string{"command", "foo", "bar", "--help"})
	if err != nil {
		t.Error(err)
	}

	output := buf.String()
	if !strings.Contains(output, "foo bar - does bar things") {
		t.Errorf("expected full path to subcommand: %s", output)
	}
	if !strings.Contains(output, "command foo bar [arguments...]") {
		t.Errorf("expected full path to subcommand: %s", output)
	}
}

func TestApp_Run_Help(t *testing.T) {
	var helpArguments = [][]string{{"boom", "--help"}, {"boom", "-h"}, {"boom", "help"}}

	for _, args := range helpArguments {
		buf := new(bytes.Buffer)

		t.Logf("==> checking with arguments %v", args)

		app := cli.NewApp()
		app.Name = "boom"
		app.Usage = "make an explosive entrance"
		app.Writer = buf
		app.Action = func(c *cli.Context) {
			buf.WriteString("boom I say!")
		}

		err := app.Run(args)
		if err != nil {
			t.Error(err)
		}

		output := buf.String()
		t.Logf("output: %q\n", buf.Bytes())

		if !strings.Contains(output, "boom - make an explosive entrance") {
			t.Errorf("want help to contain %q, did not: \n%q", "boom - make an explosive entrance", output)
		}
	}
}

func TestApp_Run_Version(t *testing.T) {
	var versionArguments = [][]string{{"boom", "--version"}, {"boom", "-v"}}

	for _, args := range versionArguments {
		buf := new(bytes.Buffer)

		t.Logf("==> checking with arguments %v", args)

		app := cli.NewApp()
		app.Name = "boom"
		app.Usage = "make an explosive entrance"
		app.Version = "0.1.0"
		app.Writer = buf
		app.Action = func(c *cli.Context) {
			buf.WriteString("boom I say!")
		}

		err := app.Run(args)
		if err != nil {
			t.Error(err)
		}

		output := buf.String()
		t.Logf("output: %q\n", buf.Bytes())

		if !strings.Contains(output, "0.1.0") {
			t.Errorf("want version to contain %q, did not: \n%q", "0.1.0", output)
		}
	}
}

func TestApp_Run_DoesNotOverwriteErrorFromBefore(t *testing.T) {
	app := cli.NewApp()
	app.Action = func(c *cli.Context) {}
	app.Before = func(c *cli.Context) error { return fmt.Errorf("before error") }
	app.After = func(c *cli.Context) error { return fmt.Errorf("after error") }

	err := app.Run([]string{"foo"})
	if err == nil {
		t.Fatalf("expected to recieve error from Run, got none")
	}

	if !strings.Contains(err.Error(), "before error") {
		t.Errorf("expected text of error from Before method, but got none in \"%v\"", err)
	}
	if !strings.Contains(err.Error(), "after error") {
		t.Errorf("expected text of error from After method, but got none in \"%v\"", err)
	}
}

func TestApp_Run_SubcommandDoesNotOverwriteErrorFromBefore(t *testing.T) {
	app := cli.NewApp()
	app.Commands = []cli.Command{
		cli.Command{
			Name:   "bar",
			Before: func(c *cli.Context) error { return fmt.Errorf("before error") },
			After:  func(c *cli.Context) error { return fmt.Errorf("after error") },
		},
	}

	err := app.Run([]string{"foo", "bar"})
	if err == nil {
		t.Fatalf("expected to recieve error from Run, got none")
	}

	if !strings.Contains(err.Error(), "before error") {
		t.Errorf("expected text of error from Before method, but got none in \"%v\"", err)
	}
	if !strings.Contains(err.Error(), "after error") {
		t.Errorf("expected text of error from After method, but got none in \"%v\"", err)
	}
}