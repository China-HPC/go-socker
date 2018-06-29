package socker

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestQueryChildPIDs(t *testing.T) {
	Convey("Test QueryChildPIDs", t, func() {
		pid := fmt.Sprintf("%d", os.Getpid())
		pids, err := QueryChildPIDs(pid)
		So(err, ShouldBeNil)
		So(pids, ShouldBeNil)
		go func() {
			exec.Command("bash", "-c", "sleep 1").Run()
		}()
		pids, err = QueryChildPIDs(pid)
		So(err, ShouldBeNil)
		So(len(pids), ShouldEqual, 1)
	})
}
