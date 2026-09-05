package router

import (
	"testing"
)

func TestRadixRouter(t *testing.T) {
	r := New()

	r.GET("/api/users/{id:int}", func(ctx *Context) (interface{}, error) {
		return "user:" + ctx.Param("id"), nil
	})

	v1 := r.Group("/v1")
	v1.GET("/profile", func(ctx *Context) (interface{}, error) {
		return "profile-ok", nil
	})

	// Match int
	res, err := r.Dispatch(NewContext("GET", "/api/users/99"))
	if err != nil || res != "user:99" {
		t.Errorf("failed matching int param: %v, err=%v", res, err)
	}

	// Reject non-int
	_, err = r.Dispatch(NewContext("GET", "/api/users/not-an-int"))
	if err == nil {
		t.Errorf("expected 404 for non-int param")
	}

	// Match group
	res, err = r.Dispatch(NewContext("GET", "/v1/profile"))
	if err != nil || res != "profile-ok" {
		t.Errorf("failed matching group route: %v, err=%v", res, err)
	}
}
