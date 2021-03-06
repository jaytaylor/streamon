package streamon

import (
	"regexp"
	"testing"
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

// Test for map[string]string equivalence.
func eq(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if w, ok := b[k]; !ok || v != w {
			return false
		}
	}
	return true
}

func testAttach(t *testing.T, attachCommand []string, filterRe *regexp.Regexp, expectedState map[string]string) {
	commandListener, err := NewCommandListener(attachCommand, filterRe)
	if err != nil {
		t.Fatal("unexpected error creating command with `NewCommandListener`: %v", err)
	}
	ch := make(chan []string)
	if commandListener.Attach(ch).Error != nil {
		t.Fatalf("unexpected error attaching: %v", commandListener.Error)
	}
	state := map[string]string{}
	for ch != nil {
		select {
		case match, ok := <-ch:
			if !ok {
				ch = nil
				break
			}
			t.Logf("match=%v\n", match)
			if filterRe == nil && len(match) != 1 {
				t.Fatalf("expected match to contain 1 element in match list but found %v (%v)", len(match), match)
			}
			if filterRe != nil && len(match) != 3 {
				t.Fatalf("expected match to find 3 elements in match list but found %v (%v)", len(match), match)
			}
			if filterRe != nil {
				state[match[1]] = match[2]
			} else {
				state[match[0]] = match[0]
			}
		}
	}
	if !eq(state, expectedState) {
		t.Fatalf("unexpected state=%v, expected=%v", state, expectedState)
	}
}

func Test_CommandListenerNilFilterRe(t *testing.T) {
	attachCommand := []string{
		"bash",
		"-c",
		`echo "'test-app_v1_web_10023' changed state to [STARTING]"
echo "'test-app_v1_web_10023' changed state to [RUNNING]"
echo "'test-app_v1_web_10023' changed state to [STOPPING]"
echo "'test-app_v1_web_10024' changed state to [STARTING]"
echo "'test-app_v1_web_10023' changed state to [STOPPED]"
echo "'test-app_v1_web_10024' changed state to [RUNNING]"
echo "'test-app_v1_web_10023' changed state to [STARTING]"
echo "'test-app_v1_web_10024' changed state to [STOPPED]"
echo "'test-app_v1_web_10023' changed state to [RUNNING]"`,
	}
	expectedState := map[string]string{
		"'test-app_v1_web_10023' changed state to [RUNNING]":  "'test-app_v1_web_10023' changed state to [RUNNING]",
		"'test-app_v1_web_10023' changed state to [STARTING]": "'test-app_v1_web_10023' changed state to [STARTING]",
		"'test-app_v1_web_10023' changed state to [STOPPED]":  "'test-app_v1_web_10023' changed state to [STOPPED]",
		"'test-app_v1_web_10023' changed state to [STOPPING]": "'test-app_v1_web_10023' changed state to [STOPPING]",
		"'test-app_v1_web_10024' changed state to [RUNNING]":  "'test-app_v1_web_10024' changed state to [RUNNING]",
		"'test-app_v1_web_10024' changed state to [STARTING]": "'test-app_v1_web_10024' changed state to [STARTING]",
		"'test-app_v1_web_10024' changed state to [STOPPED]":  "'test-app_v1_web_10024' changed state to [STOPPED]",
	}
	testAttach(t, attachCommand, nil, expectedState)
}

func Test_CommandListenerTypical(t *testing.T) {
	attachCommand := []string{
		"bash",
		"-c",
		`echo "'test-app_v1_web_10023' changed state to [STARTING]"
echo "'test-app_v1_web_10023' changed state to [RUNNING]"
echo "'test-app_v1_web_10023' changed state to [STOPPING]"
echo "'test-app_v1_web_10024' changed state to [STARTING]"
echo "'test-app_v1_web_10023' changed state to [STOPPED]"
echo "'test-app_v1_web_10024' changed state to [RUNNING]"
echo "'test-app_v1_web_10023' changed state to [STARTING]"
echo "'test-app_v1_web_10024' changed state to [STOPPED]"
echo "'test-app_v1_web_10023' changed state to [RUNNING]"`,
	}
	filterRe := regexp.MustCompile(`'([^']+)' changed state to \[([^\]]+)\]`)
	expectedState := map[string]string{"test-app_v1_web_10023": "RUNNING", "test-app_v1_web_10024": "STOPPED"}
	testAttach(t, attachCommand, filterRe, expectedState)
}

func Test_CommandListenerSleepy(t *testing.T) {
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
	expectedState := map[string]string{"test-app_v1_web_10023": "RUNNING"}
	testAttach(t, attachCommand, filterRe, expectedState)
}
