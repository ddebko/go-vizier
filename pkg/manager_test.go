package pkg

import (
	"runtime"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewManager(t *testing.T) {
	DisableLogging()
	Convey("When creating a new manager", t, func() {
		Convey("should fail with a invalid pool size", func() {
			_, err := NewManager("test", 0)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "[vizier] source: manager. message: pool size must be greater than 0. details: test. invalid size 0")
		})
		Convey("should pass with a valid pool size", func() {
			manager, err := NewManager("test", 1)
			So(err, ShouldBeNil)
			So(manager, ShouldNotBeNil)
			So(manager.GetSize(), ShouldEqual, 1)
			So(manager.name, ShouldEqual, "test")
		})
	})
}

func TestGetSize(t *testing.T) {
	DisableLogging()
	Convey("When getting the pool size", t, func() {
		Convey("should return the correct value", func() {
			manager, err := NewManager("test", 25)
			So(err, ShouldBeNil)
			So(manager, ShouldNotBeNil)
			So(manager.GetSize(), ShouldEqual, 25)
		})
	})
}

func TestNode(t *testing.T) {
	DisableLogging()
	Convey("When creating a node", t, func() {
		Convey("should fail on duplicates", func() {
			manager, _ := NewManager("test", 25)
			So(func() {
				manager.Node("test", func(payload interface{}) map[string]interface{} {
					return map[string]interface{}{}
				})
			}, ShouldNotPanic)
			So(func() {
				manager.Node("test", func(payload interface{}) map[string]interface{} {
					return map[string]interface{}{}
				})
			}, ShouldPanic)
		})
		Convey("should succeed on unqiue name", func() {
			manager, _ := NewManager("test", 25)
			So(func() {
				manager.Node("test", func(payload interface{}) map[string]interface{} {
					return map[string]interface{}{}
				})
			}, ShouldNotPanic)
			val, exists := manager.states["test"]
			state, ok := val.(State)
			So(exists, ShouldBeTrue)
			So(ok, ShouldBeTrue)
			So(state.Name, ShouldEqual, "test")
		})
	})
}

func TestOutput(t *testing.T) {
	DisableLogging()
	Convey("When creating an output edge", t, func() {
		Convey("should fail if the node does not exist", func() {
			manager, _ := NewManager("test", 25)
			So(func() {
				manager.Output("test", "output")
			}, ShouldPanic)
		})
		Convey("should fail if the edge already exists", func() {
			manager, _ := NewManager("test", 25)
			So(func() {
				manager.Node("test", func(payload interface{}) map[string]interface{} {
					return map[string]interface{}{}
				})
			}, ShouldNotPanic)
			So(func() {
				manager.Output("test", "output")
			}, ShouldNotPanic)
			So(func() {
				manager.Output("test", "output")
			}, ShouldPanic)
		})
		Convey("should return a channel Packet on success", func() {
			manager, _ := NewManager("test", 25)
			var output chan Packet
			So(func() {
				manager.Node("test", func(payload interface{}) map[string]interface{} {
					return map[string]interface{}{}
				})
			}, ShouldNotPanic)
			So(func() {
				output = manager.Output("test", "output")
			}, ShouldNotPanic)

			So(output, ShouldNotBeNil)

			val, exists := manager.states["test"]
			state, ok := val.(State)
			So(exists, ShouldBeTrue)
			So(ok, ShouldBeTrue)
			So(state.Name, ShouldEqual, "test")

			edge, ok := state.edges["output"]
			So(ok, ShouldBeTrue)
			So(len(state.edges), ShouldEqual, 1)
			So(edge.isOutput, ShouldBeTrue)

			_, ok = state.buffers["output"]
			So(ok, ShouldBeTrue)
			So(len(state.buffers), ShouldEqual, 2)
		})
	})
}

func TestEdge(t *testing.T) {
	DisableLogging()
	Convey("When creating an edge", t, func() {
		Convey("should fail if the destination node does not exist", func() {
			manager, _ := NewManager("test", 25)
			So(func() {
				manager.Edge("source", "destination", "edge")
			}, ShouldPanic)
		})
		Convey("should fail if the source node does not exist", func() {
			manager, _ := NewManager("test", 25)
			So(func() {
				manager.Node("destination", func(payload interface{}) map[string]interface{} {
					return map[string]interface{}{}
				})
			}, ShouldNotPanic)
			So(func() {
				manager.Edge("source", "destination", "edge")
			}, ShouldPanic)
		})
		Convey("should fail the source edge contains duplicate", func() {
			manager, _ := NewManager("test", 25)
			So(func() {
				manager.Node("destination", func(payload interface{}) map[string]interface{} {
					return map[string]interface{}{}
				})
			}, ShouldNotPanic)
			So(func() {
				manager.Node("source", func(payload interface{}) map[string]interface{} {
					return map[string]interface{}{}
				})
			}, ShouldNotPanic)
			So(func() {
				manager.Edge("source", "destination", "edge")
			}, ShouldNotPanic)
			So(func() {
				manager.Edge("source", "destination", "edge")
			}, ShouldPanic)
		})
		Convey("should succeed", func() {
			manager, _ := NewManager("test", 25)
			So(func() {
				manager.Node("destination", func(payload interface{}) map[string]interface{} {
					return map[string]interface{}{}
				})
			}, ShouldNotPanic)
			So(func() {
				manager.Node("source", func(payload interface{}) map[string]interface{} {
					return map[string]interface{}{}
				})
			}, ShouldNotPanic)
			So(func() {
				manager.Edge("source", "destination", "edge")
			}, ShouldNotPanic)

			val, exists := manager.states["source"]
			state, ok := val.(State)
			So(exists, ShouldBeTrue)
			So(ok, ShouldBeTrue)
			So(state.Name, ShouldEqual, "source")

			edge, ok := state.edges["source_to_destination_edge"]
			So(ok, ShouldBeTrue)
			So(len(state.edges), ShouldEqual, 1)
			So(edge.isOutput, ShouldBeFalse)

			_, ok = state.buffers["source_to_destination_edge"]
			So(ok, ShouldBeTrue)

			_, ok = state.buffers[state.Name]
			So(ok, ShouldBeTrue)

			So(len(state.buffers), ShouldEqual, 2)

			val, exists = manager.states["destination"]
			state, ok = val.(State)
			So(exists, ShouldBeTrue)
			So(ok, ShouldBeTrue)
			So(state.Name, ShouldEqual, "destination")
			So(len(state.edges), ShouldBeZeroValue)

			_, ok = state.buffers[state.Name]
			So(ok, ShouldBeTrue)

			So(len(state.buffers), ShouldEqual, 1)
		})
	})
}

func TestStart(t *testing.T) {
	DisableLogging()
	Convey("When starting the pool", t, func() {
		Convey("should return an error when the graph is empty", func() {
			manager, err := NewManager("test", 25)
			So(err, ShouldBeNil)
			So(manager, ShouldNotBeNil)

			vizErr := manager.Start()
			So(vizErr, ShouldNotBeNil)
			So(vizErr.Err().Error(), ShouldEqual, "[vizier] source: manager. message: pool requires at least one state. details: test")
		})
		Convey("should return an error if the graph is already running", func() {
			manager, err := NewManager("test", 25)
			So(err, ShouldBeNil)
			So(manager, ShouldNotBeNil)

			manager.Node("test", func(payload interface{}) map[string]interface{} {
				return map[string]interface{}{}
			})

			vizErr := manager.Start()
			So(vizErr, ShouldBeNil)

			vizErr = manager.Start()
			So(vizErr, ShouldNotBeNil)
			So(vizErr.Err().Error(), ShouldEqual, "[vizier] source: manager. message: pool is already running. details: test")

			vizErr = manager.Stop()
			So(vizErr, ShouldBeNil)
		})
	})
}

func TestStop(t *testing.T) {
	DisableLogging()
	Convey("When stopping the pool", t, func() {
		Convey("should return an error if the pool is not running", func() {
			manager, err := NewManager("test", 25)
			So(err, ShouldBeNil)
			So(manager, ShouldNotBeNil)

			vizErr := manager.Stop()
			So(vizErr, ShouldNotBeNil)
			So(vizErr.Err().Error(), ShouldEqual, "[vizier] source: manager. message: pool is not running. details: test")
		})
		Convey("should successfully stop", func() {
			manager, err := NewManager("test", 25)
			So(err, ShouldBeNil)
			So(manager, ShouldNotBeNil)

			manager.Node("test", func(payload interface{}) map[string]interface{} {
				return map[string]interface{}{}
			})

			vizErr := manager.Start()
			So(vizErr, ShouldBeNil)
			So(runtime.NumGoroutine(), ShouldBeGreaterThan, 24)

			vizErr = manager.Stop()
			So(vizErr, ShouldBeNil)
			time.Sleep(100 * time.Millisecond)
			So(runtime.NumGoroutine(), ShouldBeLessThan, 25)
		})
	})
}

func TestSetLogLevel(t *testing.T) {
	Convey("When setting the log level", t, func() {
		Convey("the level is set", func() {
			SetLogLevel(log.ErrorLevel)
			So(log.GetLevel(), ShouldEqual, log.ErrorLevel)
		})
	})
}

func TestSetSize(t *testing.T) {
	Convey("When setting the pool size", t, func() {
		Convey("should return an error if the graph is empty", func() {
			manager, err := NewManager("test", 5)
			So(err, ShouldBeNil)
			So(manager, ShouldNotBeNil)

			vizErr := manager.SetSize(5)
			So(vizErr, ShouldNotBeNil)
			So(vizErr.Err().Error(), ShouldEqual, "[vizier] source: manager. message: pool requires at least one state. details: test")
		})
		Convey("should return an error if the size is invalid", func() {
			manager, err := NewManager("test", 5)
			So(err, ShouldBeNil)
			So(manager, ShouldNotBeNil)

			manager.Node("test", func(payload interface{}) map[string]interface{} {
				return map[string]interface{}{}
			})

			vizErr := manager.SetSize(0)
			So(vizErr, ShouldNotBeNil)
			So(vizErr.Err().Error(), ShouldEqual, "[vizier] source: manager. message: pool size must be greater than 0. details: test. invalid size 0")
		})
		Convey("should increase the pool size", func() {
			manager, err := NewManager("test", 5)
			So(err, ShouldBeNil)
			So(manager, ShouldNotBeNil)

			manager.Node("test", func(payload interface{}) map[string]interface{} {
				return map[string]interface{}{}
			})

			vizErr := manager.Start()
			So(vizErr, ShouldBeNil)
			So(manager.GetSize(), ShouldEqual, 5)

			vizErr = manager.SetSize(10)
			time.Sleep(100 * time.Millisecond)
			So(vizErr, ShouldBeNil)
			So(manager.GetSize(), ShouldEqual, 10)
			So(runtime.NumGoroutine(), ShouldBeGreaterThan, 9)

			vizErr = manager.Stop()
			So(vizErr, ShouldBeNil)
		})
		Convey("should decrease the pool size", func() {
			manager, err := NewManager("test", 25)
			So(err, ShouldBeNil)
			So(manager, ShouldNotBeNil)

			manager.Node("test", func(payload interface{}) map[string]interface{} {
				return map[string]interface{}{}
			})

			vizErr := manager.Start()
			So(vizErr, ShouldBeNil)
			So(manager.GetSize(), ShouldEqual, 25)

			vizErr = manager.SetSize(5)
			time.Sleep(100 * time.Millisecond)
			So(vizErr, ShouldBeNil)
			So(manager.GetSize(), ShouldEqual, 5)
			So(runtime.NumGoroutine(), ShouldBeLessThan, 25)

			vizErr = manager.Stop()
			So(vizErr, ShouldBeNil)
		})
	})
}

func TestRecover(t *testing.T) {
	Convey("When a worker panics", t, func() {
		Convey("a new worker will spawn", func() {
			manager, err := NewManager("test", 1)
			So(err, ShouldBeNil)
			So(manager, ShouldNotBeNil)

			manager.Node("test", func(payload interface{}) map[string]interface{} {
				panic("testing recover")
				return map[string]interface{}{}
			})

			vizErr := manager.Start()
			So(vizErr, ShouldBeNil)
			So(runtime.NumGoroutine(), ShouldEqual, 3)

			_, vizErr = manager.Invoke("test", "hello world")
			time.Sleep(100 * time.Millisecond)
			So(vizErr, ShouldBeNil)
			So(runtime.NumGoroutine(), ShouldEqual, 3)

			vizErr = manager.Stop()
			So(vizErr, ShouldBeNil)
		})
	})
}

func TestInvoke(t *testing.T) {
	Convey("When invoking a state", t, func() {
		Convey("the payload is processed", func() {
			manager, err := NewManager("test", 1)
			So(err, ShouldBeNil)
			So(manager, ShouldNotBeNil)

			manager.Node("test", func(payload interface{}) map[string]interface{} {
				return map[string]interface{}{}
			})

			vizErr := manager.Start()
			So(vizErr, ShouldBeNil)

			_, vizErr = manager.Invoke("dne", "hello world")
			So(vizErr, ShouldNotBeNil)
			So(vizErr.Err().Error(), ShouldEqual, "[vizier] source: manager. message: state does not exist. details: failed to invoke state dne.")

			_, vizErr = manager.Invoke("test", "hello world")
			So(vizErr, ShouldBeNil)

			vizErr = manager.Stop()
			So(vizErr, ShouldBeNil)
		})
	})
}

func TestBatchInvoke(t *testing.T) {
	Convey("When batch invoking a state", t, func() {
		Convey("the payload is processed", func() {

		})
	})
}

func TestGetResults(t *testing.T) {
	Convey("When getting the results to a invoke", t, func() {

	})
}
