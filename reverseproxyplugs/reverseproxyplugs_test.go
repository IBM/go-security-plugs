package reverseproxyplugs

import "testing"

func TestLoadPlugs(t *testing.T) {
	var ext []string
	var numTests int
	if numTests = LoadPlugs(nil, nil); numTests != 0 {
		t.Errorf("LoadPlugs expected 0 returned %d\n", numTests)
	}

	ext = []string{}
	if numTests = LoadPlugs(nil, ext); numTests != 0 {
		t.Errorf("LoadPlugs expected 0 returned %d\n", numTests)
	}
	ext = []string{"../examplegate/examplegate.so"}
	if numTests = LoadPlugs(nil, ext); numTests != 1 {
		t.Errorf("LoadPlugs expected 1 returned %d\n", numTests)
	}
	ext = []string{"../examplegate/examplegate.so", "../examplegate/examplegate.so"}
	if numTests = LoadPlugs(nil, ext); numTests != 2 {
		t.Errorf("LoadPlugs expected 2 returned %d\n", numTests)
	}
}

//handleRequest
//HandleRequestPlugs
//HandleResponsePlugs
//HandleErrorPlugs
//ShutdownPlugs
