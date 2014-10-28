package streamon

import (
	"bufio"
	"errors"
	"io"
	//"log"
	"os/exec"
	"regexp"
)

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
			if this.FilterRe == nil {
				ch <- []string{text}
			} else if this.FilterRe.MatchString(text) {
				match := this.FilterRe.FindStringSubmatch(text)
				//log.Printf("match=%v\n", match)
				ch <- match
			} //else {
			//	log.Printf("else... text=%v but did not match %v\n", text, this.FilterRe)
			//}
		}
		//log.Println("NOTICE: reader loop exiting")
		close(ch)
	}()

	go func() {
		cmd := exec.Command(this.AttachCommand[0], this.AttachCommand[1:]...)
		cmd.Stdout = writer

		err := cmd.Run()
		if err != nil {
			//log.Printf("CommandListener.Attach: error running ssh listener: %v\n", err)
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
