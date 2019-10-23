package pkg

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewManager(t *testing.T) {
	Convey("When creating a new manager", t, func() {
		Convey("should fail with a invalid pool size", func() {
			manager, err := NewManager("test", 0)
			manager.DisableLogging()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "[vizier] source: manager. message: pool size must be greater than 0. details: test. invalid size 0")
		})
		Convey("should pass with a valid pool size", func() {
			manager, err := NewManager("test", 1)
			manager.DisableLogging()
			So(err, ShouldBeNil)
			So(manager, ShouldNotBeNil)
			So(manager.GetSize(), ShouldEqual, 1)
			So(manager.name, ShouldEqual, "test")
		})
	})
}

func TestGetSize(t *testing.T) {
	Convey("When getting the pool size", t, func() {
		Convey("should return the correct value", func() {
			manager, err := NewManager("test", 25)
			manager.DisableLogging()
			So(err, ShouldBeNil)
			So(manager, ShouldNotBeNil)
			So(manager.GetSize(), ShouldEqual, 25)
		})
	})
}

func TestNode(t *testing.T) {
	Convey("When creating a node", t, func() {
		Convey("should fail on duplicates", func() {
			manager, _ := NewManager("test", 25)
			manager.DisableLogging()
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
			manager.DisableLogging()
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
	Convey("When creating an output edge", t, func() {
		Convey("should fail if the node does not exist", func() {
			manager, _ := NewManager("test", 25)
			manager.DisableLogging()
			So(func() {
				manager.Output("test", "output")
			}, ShouldPanic)
		})
		Convey("should fail if the edge already exists", func() {
			manager, _ := NewManager("test", 25)
			manager.DisableLogging()
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
			manager.DisableLogging()
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
		})
	})
}

func TestEdge(t *testing.T) {
	Convey("When creating an edge", t, func() {
		Convey("should fail if the destination node does not exist", func() {
			manager, _ := NewManager("test", 25)
			manager.DisableLogging()
			So(func() {
				manager.Edge("source", "destination", "edge")
			}, ShouldPanic)
		})
		Convey("should fail if the source node does not exist", func() {
			manager, _ := NewManager("test", 25)
			manager.DisableLogging()
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
			manager.DisableLogging()
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
			manager.DisableLogging()
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
