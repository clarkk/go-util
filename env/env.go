package env

import (
	"fmt"
	"log"
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
	
	Environment struct {
		Env_data
		lang	lang.Lang
	}
	
	Properties map[string]any
	
	ctx_key		string
)

func New(d Env_data) *Environment {
	return &Environment{
		Env_data:	d,
		lang:		lang.New(d.Lang(), nil),
	}
}

func New_request(r *http.Request, d Env_data) *Environment {
	e := &Environment{
		Env_data:	d,
		lang:		lang.New(d.Lang(), req.Accept_lang(r)),
	}
	
	ctx := context.WithValue(r.Context(), ctx_env, e)
	r2 := r.WithContext(ctx)
	*r = *r2
	
	return e
}

func Request(r *http.Request) *Environment {
	if e, ok := r.Context().Value(ctx_env).(*Environment); ok {
		return e
	}
	return nil
}

func Fatal_log(err error) error {
	log.Printf("env data: %v", err)
	return err
}

func Key_error(key string) error {
	return fmt.Errorf("Invalid env key: %s", key)
}

func Type_error(key string, value any) error {
	return fmt.Errorf("Invalid env key type: %s (%T)", key, value)
}

func (e *Environment) Lang() string {
	return e.lang.Get()
}

func (e *Environment) Lang_string(key string, replace map[string]any) string {
	return e.lang.String(key, replace)
}

func (e *Environment) Lang_error(key string, replace map[string]any) error {
	return e.lang.Error(key, replace)
}

func (e *Environment) Lang_printer() *lang.Printer {
	return e.lang.Printer()
}