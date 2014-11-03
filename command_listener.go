package streamon

import (
	"bufio"
	"errors"
	"io"
	"log"
	"os/exec"
	"regexp"
)

var debugEnabled bool

var ErrEmptyAttachCommand = errors.New("attachCommand must not be empty")

type CommandListener struct {
	AttachCommand []string
	FilterRe      *regexp.Regexp
	Error         error
}

// Optionally, filterRe may be nil.  When this is the case, each line of command output will be sent
// directly to the channel with no regular expression filtering applied.
func NewCommandListener(attachCommand []string, filterRe *regexp.Regexp) (*CommandListener, error) {
	if len(attachCommand) == 0 {
		return nil, ErrEmptyAttachCommand
	}
	cl := CommandListener{
		AttachCommand: attachCommand,
		FilterRe:      filterRe,
	}
	return &cl, nil
}

func (this *CommandListener) Attach(ch chan []string) *CommandListener {
	if len(this.AttachCommand) == 0 {
		this.Error = ErrEmptyAttachCommand
		return this
	}
	//if this.Running {
	//	this.Error = fmt.Errorf("CommandListener is already running")
	//	return this
	//}
	//this.Running = true

	reader, writer := io.Pipe()

	go func() {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			text := scanner.Text()
			debug("text=%v\n", text)
			if this.FilterRe == nil {
				debug("text=%v and auto forwarding because filter is empty\n", text)
				ch <- []string{text}
			} else if this.FilterRe.MatchString(text) {
				match := this.FilterRe.FindStringSubmatch(text)
				debug("match=%v text=%v filter=%v\n", match, text, this.FilterRe.String())
				ch <- match
			} //else {
			//	debug("text=%v but did not match filter=%v\n", text, this.FilterRe.String())
			//}
		}
		debug("reader loop exiting for AttachCommand=%v FilterRe=%v\n", this.AttachCommand, this.FilterRe)
		close(ch)
	}()

	go func() {
		cmd := exec.Command(this.AttachCommand[0], this.AttachCommand[1:]...)
		cmd.Stdout = writer
		//cmd.Stderr = writer
		debug("cmd.Args=%v\n", cmd.Args)

		err := cmd.Run()
		if err != nil {
			debug("running command resulted in error=%v\n", err)
			this.Error = err
		}
		err = reader.Close()
		if this.Error == nil && err != nil {
			this.Error = nil
		}
		err = writer.Close()
		if this.Error == nil && err != nil {
			this.Error = nil
		}
	}()

	return this
}

func debug(message string, args ...interface{}) {
	if debugEnabled {
		log.Printf("DEBUG: streamon: "+message, args...)
	}
}
