package dcmd

import (
	"testing"
)

func TestMiddlewareOrder(t *testing.T) {
	container := &Container{}

	i := 0

	container.AddMidlewares(func(inner RunFunc) RunFunc {
		return func(d *Data) (interface{}, error) {
			// This middleware should be ran first
			if i != 0 {
				t.Error("Unordered:", i, "expected 0")
			}
			i = 1
			t.Log("mw 1")
			return inner(d)
		}
	}, func(inner RunFunc) RunFunc {
		return func(d *Data) (interface{}, error) {
			// This middleware should be ran first
			if i != 1 {
				t.Error("Unordered:", i, "expected 1")
			}
			i = 2
			t.Log("mw 2")
			return inner(d)
		}
	})

	cmd := &TestCommand{}
	container.AddCommand(cmd, NewTrigger("test").SetMiddlewares(func(inner RunFunc) RunFunc {
		return func(d *Data) (interface{}, error) {
			// This middleware should be ran first
			if i != 2 {
				t.Error("Unordered:", i, "expected 2")
			}
			i = 3
			t.Log("mw 3")
			return inner(d)
		}
	}))

	sub := container.Sub("sub")
	sub.AddMidlewares(func(inner RunFunc) RunFunc {
		return func(d *Data) (interface{}, error) {
			// This middleware should be ran first
			if i != 2 {
				t.Error("Unordered:", i, "expected 2")
			}
			i = 3
			t.Log("mw 4")
			return inner(d)
		}
	})

	sub.AddCommand(cmd, NewTrigger("test").SetMiddlewares(func(inner RunFunc) RunFunc {
		return func(d *Data) (interface{}, error) {
			// This middleware should be ran first
			if i != 3 {
				t.Error("Unordered:", i, "expected 3")
			}
			i = 4
			t.Log("mw 5")
			return inner(d)
		}
	}))

	data1 := &Data{
		MsgStrippedPrefix: "test",
		Source:            PrefixSource,
	}
	data2 := &Data{
		MsgStrippedPrefix: "sub test",
		Source:            PrefixSource,
	}

	container.Run(data1)
	i = 0
	resp, _ := container.Run(data2)
	if i != 4 {
		t.Error("i: ", i, "Expected 4")
	}
	if resp != TestResponse {
		t.Error("Response: ", resp, ", Expected: ", TestResponse)
	}
}
