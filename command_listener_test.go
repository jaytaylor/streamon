package commandmonitor

import (
	//"os/exec"
	"regexp"
	"testing"
	"time"
)

type (
	TestCommandListener struct {
	}
)

var (
	// 'test-app_v1_web_10023' changed state to [STARTING]
	dynoStateParserRe = regexp.MustCompile(`'([^']+) changed state to \[([^\]]+)\]'`)
)

func Test_NewCommandListener(t *testing.T) {
	cl, err := NewCommandListener([]string{}, nil)
	if cl != nil {
		t.Fatalf("CommandListener should have been nil when 0 attach args were provided (cl=%v, err=%v)", cl, err)
	}
	if err == nil {
		t.Fatalf("NewCommandListener should have returned an error when 0 attach args were provided (cl=%v, err=%v)", cl, err)
	}
	cl, err = NewCommandListener([]string{"echo"}, nil)
	if cl == nil {
		t.Fatalf("CommandListener should not have been nil when 1 or more attach args were provided (cl=%v, err=%v)", cl, err)
	}
	if err != nil {
		t.Fatalf("NewCommandListener should not have returned an error when 1 or more attach args were provided (cl=%v, err=%v)", cl, err)
	}
}

func Test_CommandListener(t *testing.T) {
	attachCommand := []string{
		"bash",
		"-c",
		`echo "'test-app_v1_web_10023' changed state to [STARTING]" ; sleep 1 ;
echo "'test-app_v1_web_10023' changed state to [RUNNING]" ; sleep 1 ;
echo "'test-app_v1_web_10023' changed state to [STOPPING]" ; sleep 1 ;
echo "'test-app_v1_web_10023' changed state to [STOPPED]" ; sleep 1 ;
echo "'test-app_v1_web_10023' changed state to [STARTING]" ; sleep 1 ;
echo "'test-app_v1_web_10023' changed state to [RUNNING]" ; sleep 1 ; `,
	}
	filterRe := regexp.MustCompile(`'([^']+)' changed state to \[([^\]]+)\]`)
	commandListener, err := NewCommandListener(attachCommand, filterRe)
	if err != nil {
		t.Fatal("error creating command with `NewCommandListener`: %v", err)
	}
	//t.Log("cl=%v", commandListener)
	ch := make(chan []string)
	go commandListener.Attach(ch)
	//time.Sleep(500 * time.Millisecond)
	state := map[string]string{}
	for {
		time.Sleep(500 * time.Millisecond)
		if !commandListener.Running {
			t.Log("not runnni")
			break
		}
		select {
		case match := <-ch:
			t.Logf("match=%v\n", match)
			if len(match) != 3 {
				t.Fatalf("expected match to find 3 elements in match list but found %v (%v)", len(match), match)
			}
			state[match[1]] = match[2]
		}
	}
	expectedState := map[string]string{"test-app_v1_web_10023": "RUNNING"}
	if state["test-app_v1_web_10023"] != expectedState["test-app_v1_web_10023"] {
		t.Fatalf("unexpected state=%v, expected=%v", state, expectedState)
	}
	//time.Sleep(10 * time.Second)
	// commandListener
}

//func (this *TestDynoStateChangeListener) Attach() {

//}

func NewState(input string) (string, error) {
	return input, nil
}

// func Test_StatusMonitor(t *testing.T) {
// 	//listener := NodeDynoStateChangeListener{"testlab-sb.threatstream.com"}
// 	listener := TestCommandListener{}
// 	ch := make(chan string)
// 	go AttachCommandListener(listener, ch)
// 	time.Sleep(10000000000)
// }

func Test_NewDynoState(t *testing.T) {
	type StateTest struct {
		Input          string
		ExpectedResult string
		ExpectedError  error
	}

	testCases := []StateTest{
		StateTest{
			Input:          `'test-app_v1_web_10023' changed state to [STARTING]`,
			ExpectedResult: "okay",
			ExpectedError:  nil,
		},
		StateTest{
			Input:          `'test-app_v1_web_10023' changed state to [STARTING]`,
			ExpectedResult: "hrm",
			ExpectedError:  nil,
		},
	}

	for _, testCase := range testCases {
		_, err := NewState(testCase.Input)
		if err != nil {
			t.Errorf(`got unexpcted error with input "%v"`, testCase.Input)
		}
	}
}
