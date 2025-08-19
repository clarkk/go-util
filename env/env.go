package env

import (
	"fmt"
	"context"
	"net/http"
	"github.com/clarkk/go-util/lang"
	"github.com/clarkk/go-util/serv/req"
)

const ctx_env ctx_key = ""

type (
	Env_data interface {
		Session() *sess.Session
		Lang() string
		Data() Properties
		Update(map[string]any) error
	}
	
	Environment struct {
		Env_data
		Lang	lang.Lang
		ctx		context.Context
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

func New(d Env_data) *Environment {
	return &Environment{
		Env_data:	d,
		Lang:		lang.New(d.Lang(), nil),
	}
}

func New_request(r *http.Request, d Env_data) *Environment {
	e := &Environment{
		Env_data:	d,
		Lang:		lang.New(d.Lang(), req.Accept_lang(r)),
	}
	
	ctx := context.WithValue(r.Context(), ctx_env, e)
	r2 := r.WithContext(ctx)
	*r = *r2
	
	return e
}

func Request(r *http.Request) *Environment {
	e, ok := r.Context().Value(ctx_env).(*Environment)
	if !ok {
		return nil
	}
	return e
}