package env

import (
	"fmt"
	"context"
	"net/http"
	"github.com/clarkk/go-util/lang"
	"github.com/clarkk/go-util/serv/req"
	"github.com/clarkk/go-util/sess"
)

const ctx_env ctx_key = ""

type (
	Env_data interface {
		Session() *sess.Session
		Lang() string
		Data() Properties
		Update(map[string]any) error
	}
	
	Environment[T Env_data] struct {
		Env_data T
		lang	lang.Lang
	}
	
	Properties map[string]any
	
	ctx_key		string
)

func Key_error(key string) error {
	return fmt.Errorf("Invalid env key: %s", key)
}

func Type_error(key string, value any) error {
	return fmt.Errorf("Invalid env key type: %s (%T)", key, value)
}

func New[T Env_data](d T) *Environment[T] {
	return &Environment[T]{
		Env_data:	d,
		lang:		lang.New(d.Lang(), nil),
	}
}

func New_request[T Env_data](r *http.Request, d T) *Environment[T] {
	e := &Environment[T]{
		Env_data:	d,
		lang:		lang.New(d.Lang(), req.Accept_lang(r)),
	}
	
	ctx := context.WithValue(r.Context(), ctx_env, e)
	r2 := r.WithContext(ctx)
	*r = *r2
	
	return e
}

func Request[T Env_data](r *http.Request) *Environment[T] {
	e, ok := r.Context().Value(ctx_env).(*Environment[T])
	if !ok {
		return nil
	}
	return e
}

func (e *Environment[T]) Lang_string(key string, replace map[string]any) string {
	return e.lang.String(key, replace)
}

func (e *Environment[T]) Lang_error(key string, replace map[string]any) error {
	return e.lang.Error(key, replace)
}