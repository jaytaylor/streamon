package commandmonitor

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"regexp"
)

type (
	// CommandOutputProcessor struct {
	// 	ch chan string
	// }

	CommandListener struct {
		AttachCommand []string
		FilterRe      *regexp.Regexp
		Running       bool
		Channel       chan string
		//GetAttachCommand() *exec.Cmd
		//getChannel() chan string
		//chan string
	}

	/*NodeDynoStateChangeListener struct {
		host string
		//ch   chan string
	}

	DynoState struct {
		Dyno  Dyno
		State string
	}*/
)

// var (
// 	// 'test-app_v1_web_10023' changed state to [STARTING]
// 	dynoStateParserRe = regexp.MustCompile(`'([^']+) changed state to \[([^\]]+)\]'`)
// )

/*func NewDynoState(host, message string) (*DynoState, error) {
	parsed := dynoStateParserRe.FindStringSubmatch(strings.TrimSpace(message))
	state := parsed[]
	dyno, err := ContainerToDyno(host, container)
	if err != nil {
		return nil, err
	}
	dynoState := DynoState{
		Dyno: dyno,
		State: state,
	}
	return dynoState, nil
}*/

func NewCommandListener(attachCommand []string, filterRe *regexp.Regexp) (*CommandListener, error) {
	if len(attachCommand) == 0 {
		return nil, fmt.Errorf("NewCommandListener: attachCommand must not be empty")
	}
	cl := CommandListener{
		AttachCommand: attachCommand,
		FilterRe:      filterRe,
		Running:       false,
		//Channel:       make(chan string),
	}
	return &cl, nil
}

func (this *CommandListener) Attach(ch chan []string) error {
	if len(this.AttachCommand) == 0 {
		return fmt.Errorf("CommandListener.Attach: AttachCommand must not be empty")
	}
	if this.Running {
		return fmt.Errorf("CommandListener is already running")
	}
	this.Running = true

	r, w := io.Pipe()

	go func(reader io.Reader) {
		scanner := bufio.NewScanner(r)
		for this.Running && scanner.Scan() {
			text := scanner.Text()
			if this.FilterRe != nil && this.FilterRe.MatchString(text) {
				fmt.Printf("HEY... so v=%v\n", text)
				match := this.FilterRe.FindStringSubmatch(text)
				ch <- match
			} else {
				fmt.Printf("well... text=%v but did not match %v\n", text, this.FilterRe)
			}
		}
		fmt.Println("OYOYOYOY: reader loop exiting")
	}(r)

	cmd := exec.Command(this.AttachCommand[0], this.AttachCommand[1:]...)
	cmd.Stdout = w

	err := cmd.Run()
	if err != nil {
		//fmt.Printf("CommandListener.Attach: error running ssh listener: %v\n", err)
		return err
	}
	this.Running = false
	return nil
}
